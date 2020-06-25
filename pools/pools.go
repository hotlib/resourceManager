package pools

import (
	"context"

	"github.com/marosmars/resourceManager/ent"
	resource "github.com/marosmars/resourceManager/ent/resource"
	resourcePool "github.com/marosmars/resourceManager/ent/resourcepool"
	"github.com/pkg/errors"
)

// Scope is a unique identifier of a resource claim
type Scope struct {
	Scope string
}

// Pool is a resource provider
type Pool interface {
	ClaimResource(scope Scope) (*ent.Resource, error)
	FreeResource(scope Scope) error
	QueryResource(scope Scope) (*ent.Resource, error)
	QueryResources() (ent.Resources, error)
	Destroy() error
}

// SingletonPool provides only a single resource that can be reclaimed under various scopes
type SingletonPool struct {
	*ent.ResourcePool

	ctx    context.Context
	client *ent.Client
}

// SINGLETON_BLUEPRINT_RESOURCE identifies resource blueprint
const SINGLETON_BLUEPRINT_RESOURCE string = "SINGLETON_BLUEPRINT_RESOURCE"

// NewSingletonPool creates a brand new pool allocating DB entities in the process
func NewSingletonPool(
	ctx context.Context,
	client *ent.Client,
	resourceType *ent.ResourceType,
	propertyValues map[string]interface{},
	poolName string) (Pool, error) {

	pool, err := WithTx(ctx, client, func(tx *ent.Tx) (interface{}, error) {
		return newSingletonPoolInner(ctx, tx.Client(), resourceType, propertyValues, poolName)
	})

	if err != nil {
		return nil, err
	}

	return &SingletonPool{pool.(*ent.ResourcePool).Unwrap(), ctx, client}, nil
}

func newSingletonPoolInner(ctx context.Context,
	client *ent.Client,
	resourceType *ent.ResourceType,
	propertyValues map[string]interface{},
	poolName string) (*ent.ResourcePool, error) {
	pool, err := client.ResourcePool.Create().
		SetName(poolName).
		SetPoolType("singleton").
		SetResourceType(resourceType).
		Save(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create new pool \"%s\". Error creating pool", poolName)
	}

	// Parse & create the props
	var props ent.Properties
	if props, err = parseProps(ctx, client, resourceType, propertyValues); err != nil {
		return nil, errors.Wrapf(err, "Unable to create new pool \"%s\". Error parsing properties", poolName)
	}

	// Create blueprint resource
	_, err = client.Resource.Create().
		SetPool(pool).
		SetScope(SINGLETON_BLUEPRINT_RESOURCE).
		AddProperties(props...).
		Save(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create new pool \"%s\". Error creating singleton resource", poolName)
	}

	return pool, nil
}

// ExistingSingletonPool creates a new pool
func ExistingSingletonPool(
	ctx context.Context,
	client *ent.Client,
	poolName string) (*SingletonPool, error) {

	pool, err := client.ResourcePool.Query().
		Where(resourcePool.NameEQ(poolName)).
		Only(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot create pool from existing entity")
	}

	if pool.PoolType != resourcePool.PoolTypeSingleton {
		return nil, errors.Errorf("Wrong pool type \"%s\", expected \"%s\" for pool \"%s\"",
			pool.PoolType, resourcePool.PoolTypeSingleton, pool.Name)
	}

	_, err = pool.QueryClaims().
		Where(resource.ScopeEQ(SINGLETON_BLUEPRINT_RESOURCE)).
		Only(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot create pool from existing entity due to blueprint resource")
	}

	return &SingletonPool{pool, ctx, client}, nil
}

func parseProps(
	ctx context.Context,
	tx *ent.Client,
	resourceType *ent.ResourceType,
	propertyValues map[string]interface{}) (ent.Properties, error) {

	var props ent.Properties
	propTypes, err := resourceType.QueryPropertyTypes().All(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to determine property types for \"%s\"", resourceType)
	}

	for _, pt := range propTypes {
		pv := propertyValues[pt.Name]

		if pt.Mandatory {
			if pv == nil {
				return nil, errors.Errorf("Missing mandatory property \"%s\"", pt.Name)
			}
		} else {
			if pv == nil {
				continue
			}
		}

		ppBuilder := tx.Property.Create().SetType(pt)

		// TODO is there a better way of parsing individual types ? Reuse something from inv ?
		// TODO add additional types
		switch pt.Type {
		case "int":
			ppBuilder.SetIntVal(pv.(int))
		case "string":
			ppBuilder.SetStringVal(pv.(string))
		default:
			return nil, errors.Errorf("Unsupported property type \"%s\"", pt.Type)
		}

		pp, err := ppBuilder.Save(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to instantiate property of type \"%s\"", pt.Type)
		}
		props = append(props, pp)
	}

	return props, nil
}

// Destroy removes the pool from DB if there are no more claims
func (pool SingletonPool) Destroy() error {
	_, err := WithTx(pool.ctx, pool.client, func(tx *ent.Tx) (interface{}, error) {
		// Check if there are no more claims
		claims, err := pool.queryResourcesInner(tx.Client())
		if err != nil {
			return nil, err
		}

		if len(claims) > 1 {
			return nil, errors.Errorf("Unable to delete pool \"%s\", there are allocated claims",
				pool.ResourcePool.Name)
		}

		// Delete resource blueprint
		err = pool.freeResourceInner(tx.Client(), Scope{SINGLETON_BLUEPRINT_RESOURCE})
		if err != nil {
			return nil, err
		}

		// Delete pool itself
		err = tx.ResourcePool.DeleteOne(pool.ResourcePool).Exec(pool.ctx)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return errors.Wrapf(err, "Cannot destroy pool \"%s\"", pool.ResourcePool.Name)
	}

	return nil
}

func (pool SingletonPool) ClaimResource(scope Scope) (*ent.Resource, error) {
	claim, err := WithTx(pool.ctx, pool.client, func(tx *ent.Tx) (interface{}, error) {
		return pool.claimResourceInner(tx.Client(), scope)
	})

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to claim a resource in pool \"%s\"", pool.ResourcePool.Name)
	}

	return claim.(*ent.Resource).Unwrap(), nil
}

func (pool SingletonPool) claimResourceInner(client *ent.Client, scope Scope) (*ent.Resource, error) {
	res, err := pool.queryResourceInner(client, scope)

	// Resource exists for the scope, return the same one
	if res != nil {
		return res, nil
	}

	// Allocate new resource for this scope
	if ent.IsNotFound(err) {
		blueprintRes, err := pool.queryBlueprintResourceEager(client)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to find singleton blueprint resource in pool \"%s\"",
				pool.ResourcePool.Name)
		}

		newResource, err := pool.copyResourceWithNewScope(client, blueprintRes, scope)
		return newResource, err
	}

	return res, err
}

func (pool SingletonPool) copyResourceWithNewScope(client *ent.Client, res *ent.Resource, scope Scope) (*ent.Resource, error) {
	props, err := res.QueryProperties().WithType().All(pool.ctx)
	if err != nil {
		return nil, err
	}
	
	// copy properties from blueprint so that each claim has its own
	var copiedProps ent.Properties
	for _, pp := range props {
		builder := client.Property.CreateFrom(pp)
		if copiedProp, err := builder.Save(pool.ctx); err != nil {
			return nil, err
		} else {
			copiedProps = append(copiedProps, copiedProp)
		}

	}

	// start with a copy of blueprint resource
	return client.Resource.CreateFrom(res).
		// override scope
		SetScope(scope.Scope). 
		// set copied property instances
		AddProperties(copiedProps...).
		Save(pool.ctx)
}

func (pool SingletonPool) FreeResource(scope Scope) error {
	if scope.Scope == SINGLETON_BLUEPRINT_RESOURCE {
		// Do not actually free this, it serves as blueprint
		return nil
	}

	_, err := WithTx(pool.ctx, pool.client, func(tx *ent.Tx) (interface{}, error) {
		return nil, pool.freeResourceInner(tx.Client(), scope)
	})

	if err != nil {
		return errors.Wrapf(err, "Unable to free a resource in pool \"%s\"", pool.ResourcePool.Name)
	}

	return nil
}

func (pool SingletonPool) freeResourceInner(client *ent.Client, scope Scope) error {
	res, err := client.Resource.Query().
		Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).
		Where(resource.ScopeEQ(scope.Scope)).
		WithProperties().
		Only(pool.ctx)
	if err != nil {
		return err
	}

	for _, pp := range res.Edges.Properties {
		if err = client.Property.DeleteOne(pp).Exec(pool.ctx); err != nil {
			return err
		}
	}

	if err = client.Resource.DeleteOne(res).Exec(pool.ctx); err != nil {
		return err
	}

	return nil
}

func (pool SingletonPool) QueryResource(scope Scope) (*ent.Resource, error) {
	return pool.queryResourceInner(pool.client, scope)
}

func (pool SingletonPool) queryResourceInner(client *ent.Client, scope Scope) (*ent.Resource, error) {
	resource, err := client.Resource.Query().
		Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).
		Where(resource.ScopeEQ(scope.Scope)).
		Only(pool.ctx)

	return resource, err
}

// load eagerly with all edges
func (pool SingletonPool) queryBlueprintResourceEager(client *ent.Client) (*ent.Resource, error) {
	resource, err := client.Resource.Query().
		Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).
		Where(resource.ScopeEQ(SINGLETON_BLUEPRINT_RESOURCE)).
		WithPool().
		Only(pool.ctx)

	return resource, err
}

func (pool SingletonPool) QueryResources() (ent.Resources, error) {
	return pool.queryResourcesInner(pool.client)
}

func (pool SingletonPool) queryResourcesInner(client *ent.Client) ([]*ent.Resource, error) {
	resource, err := client.Resource.Query().
		Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).
		All(pool.ctx)

	return resource, err
}
