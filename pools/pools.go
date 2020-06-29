package pools

import (
	"context"

	"github.com/marosmars/resourceManager/ent"
	resource "github.com/marosmars/resourceManager/ent/resource"
	resourcePool "github.com/marosmars/resourceManager/ent/resourcepool"
	tagged "github.com/marosmars/resourceManager/ent/tag"
	"github.com/pkg/errors"
)

// ResourceTag is a unique identifier of a resource claim
type ResourceTag struct {
	ResourceTag string
}

type PoolLabel struct {
	PoolLabel string
}

// Pool is a resource provider
type Pool interface {
	LabeledPool
	ClaimResource(tag ResourceTag) (*ent.Resource, error)
	FreeResource(tag ResourceTag) error
	QueryResource(tag ResourceTag) (*ent.Resource, error)
	QueryResources() (ent.Resources, error)
	Destroy() error
}

type LabeledPool interface {
	AddLabel(label PoolLabel) error
}

// SingletonPool provides only a single resource that can be reclaimed under various tags
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

	pool, err := newSingletonPoolInner(ctx, client, resourceType, propertyValues, poolName)

	if err != nil {
		return nil, err
	}

	return &SingletonPool{pool, ctx, client}, nil
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

	tag, err := client.Tag.Create().
		SetTag(SINGLETON_BLUEPRINT_RESOURCE).
		Save(ctx)

	// Create blueprint resource
	_, err = client.Resource.Create().
		SetPool(pool).
		SetTag(tag).
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
	// Check if there are no more claims
	claims, err := pool.QueryResources()
	if err != nil {
		return err
	}

	if len(claims) > 1 {
		return errors.Errorf("Unable to destroy pool \"%s\", there are allocated claims",
			pool.ResourcePool.Name)
	}

	// Delete resource blueprint
	err = pool.freeResourceInner(ResourceTag{SINGLETON_BLUEPRINT_RESOURCE})
	if err != nil {
		return errors.Wrapf(err, "Cannot destroy pool \"%s\"", pool.ResourcePool.Name)
	}

	// Delete pool itself
	err = pool.client.ResourcePool.DeleteOne(pool.ResourcePool).Exec(pool.ctx)
	if err != nil {
		return errors.Wrapf(err, "Cannot destroy pool \"%s\"", pool.ResourcePool.Name)
	}

	return nil
}

func (pool SingletonPool) AddLabel(label PoolLabel) error {
	// TODO implement labeling
	return errors.Errorf("NOT IMPLEMENTED")
}

func (pool SingletonPool) ClaimResource(tag ResourceTag) (*ent.Resource, error) {
	res, err := pool.QueryResource(tag)

	// Resource exists for the tag, return the same one
	if res != nil {
		return res, nil
	}

	// Allocate new resource for this tag
	if ent.IsNotFound(err) {
		blueprintRes, err := pool.queryBlueprintResourceEager()
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to find singleton blueprint resource in pool \"%s\"",
				pool.ResourcePool.Name)
		}

		newResource, err := pool.copyResourceWithNewTag(blueprintRes, tag)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to claim a resource in pool \"%s\"", pool.ResourcePool.Name)
		}
		return newResource, err
	}

	return nil, err
}

func (pool SingletonPool) copyResourceWithNewTag(res *ent.Resource, tag ResourceTag) (*ent.Resource, error) {
	props, err := res.QueryProperties().WithType().All(pool.ctx)
	if err != nil {
		return nil, err
	}

	// copy properties from blueprint so that each claim has its own
	var copiedProps ent.Properties
	for _, pp := range props {
		builder := pool.client.Property.CreateFrom(pp)
		if copiedProp, err := builder.Save(pool.ctx); err != nil {
			return nil, err
		} else {
			copiedProps = append(copiedProps, copiedProp)
		}

	}

	createdTag, err := pool.client.Tag.Create().
		SetTag(tag.ResourceTag).
		Save(pool.ctx)
	if err != nil {
		return nil, err
	}

	// start with a copy of blueprint resource
	return pool.client.Resource.CreateFrom(res).
		// set new tag
		SetTag(createdTag).
		// set copied property instances
		AddProperties(copiedProps...).
		Save(pool.ctx)
}

func (pool SingletonPool) FreeResource(tag ResourceTag) error {
	if tag.ResourceTag == SINGLETON_BLUEPRINT_RESOURCE {
		// Do not actually free this, it serves as blueprint
		return nil
	}

	return pool.freeResourceInner(tag)
}

func (pool SingletonPool) freeResourceInner(tag ResourceTag) error {
	res, err := pool.FindResource(tag).
		WithProperties().
		WithTag().
		Only(pool.ctx)
	if err != nil {
		return err
	}

	// clean resources
	for _, pp := range res.Edges.Properties {
		if err = pool.client.Property.DeleteOne(pp).Exec(pool.ctx); err != nil {
			return errors.Wrapf(err, "Unable to free a resource in pool \"%s\"", pool.ResourcePool.Name)
		}
	}
	
	// clean tag if present
	if res.Edges.Tag != nil {
		if err = pool.client.Tag.DeleteOne(res.Edges.Tag).Exec(pool.ctx); err != nil {
			return errors.Wrapf(err, "Unable to free a resource in pool \"%s\"", pool.ResourcePool.Name)
		}
	}

	if err = pool.client.Resource.DeleteOne(res).Exec(pool.ctx); err != nil {
		return errors.Wrapf(err, "Unable to free a resource in pool \"%s\"", pool.ResourcePool.Name)
	}

	return nil
}

func (pool SingletonPool) FindResource(tag ResourceTag) *ent.ResourceQuery {
	return pool.client.Resource.Query().
		Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).
		Where(resource.HasTagWith(tagged.TagEQ(tag.ResourceTag)))
}

func (pool SingletonPool) QueryResource(tag ResourceTag) (*ent.Resource, error) {
	resource, err := pool.FindResource(tag).
		Only(pool.ctx)

	return resource, err
}

// load eagerly with some edges, ready to be copied
func (pool SingletonPool) queryBlueprintResourceEager() (*ent.Resource, error) {
	resource, err := pool.FindResource(ResourceTag{SINGLETON_BLUEPRINT_RESOURCE}).
		WithPool().
		Only(pool.ctx)

	return resource, err
}

func (pool SingletonPool) QueryResources() (ent.Resources, error) {
	resource, err := pool.client.Resource.Query().
		Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).
		All(pool.ctx)

	return resource, err
}
