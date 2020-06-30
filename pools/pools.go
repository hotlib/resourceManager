package pools

import (
	"context"

	"github.com/marosmars/resourceManager/ent"
	"github.com/marosmars/resourceManager/ent/predicate"
	property "github.com/marosmars/resourceManager/ent/property"
	propertyType "github.com/marosmars/resourceManager/ent/propertytype"
	resource "github.com/marosmars/resourceManager/ent/resource"
	resourcePool "github.com/marosmars/resourceManager/ent/resourcepool"
	"github.com/pkg/errors"
)

// ResourceTag is a unique identifier of a resource claim
type ResourceTag struct {
	ResourceTag string `json:"resourceTag"`
}

type PoolLabel struct {
	PoolLabel string
}

// Pool is a resource provider
type Pool interface {
	LabeledPool
	ClaimResource() (*ent.Resource, error)
	FreeResource(RawResourceProps) error
	QueryResource(RawResourceProps) (*ent.Resource, error)
	QueryResources() (ent.Resources, error)
	Destroy() error
}

type LabeledPool interface {
	AddLabel(label PoolLabel) error
}

// SingletonPool provides only a single resource that can be reclaimed under various tags
type SetPool struct {
	*ent.ResourcePool

	ctx    context.Context
	client *ent.Client
}

type SingletonPool struct {
	SetPool
}

type RawResourceProps map[string]interface{}

// NewSingletonPool creates a brand new pool allocating DB entities in the process
func NewSingletonPool(
	ctx context.Context,
	client *ent.Client,
	resourceType *ent.ResourceType,
	propertyValues RawResourceProps,
	poolName string) (Pool, error) {

	pool, err := newPoolInner(ctx, client, resourceType, []RawResourceProps{propertyValues}, poolName)

	if err != nil {
		return nil, err
	}

	return &SingletonPool{SetPool{pool, ctx, client}}, nil
}

// NewSetPool creates a brand new pool allocating DB entities in the process
func NewSetPool(
	ctx context.Context,
	client *ent.Client,
	resourceType *ent.ResourceType,
	propertyValues []RawResourceProps,
	poolName string) (Pool, error) {

	// TODO check that propertyValues are unique

	pool, err := newPoolInner(ctx, client, resourceType, propertyValues, poolName)

	if err != nil {
		return nil, err
	}

	return &SetPool{pool, ctx, client}, nil
}

func newPoolInner(ctx context.Context,
	client *ent.Client,
	resourceType *ent.ResourceType,
	propertyValues []RawResourceProps,
	poolName string) (*ent.ResourcePool, error) {
	pool, err := client.ResourcePool.Create().
		SetName(poolName).
		SetPoolType("singleton").
		SetResourceType(resourceType).
		Save(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create new pool \"%s\". Error creating pool", poolName)
	}

	// Pre-create all resources
	for _, rawResourceProps := range propertyValues {
		// Parse & create the props
		var props ent.Properties
		if props, err = parseProps(ctx, client, resourceType, rawResourceProps); err != nil {
			return nil, errors.Wrapf(err, "Unable to create new pool \"%s\". Error parsing properties", poolName)
		}

		// Create pre-allocated resource
		_, err = client.Resource.Create().
			SetPool(pool).
			SetClaimed(false).
			AddProperties(props...).
			Save(ctx)

		if err != nil {
			return nil, errors.Wrapf(err, "Unable to create new pool \"%s\". Error creating resource", poolName)
		}
	}

	return pool, nil
}

// ExistingPool creates a new pool
func ExistingPool(
	ctx context.Context,
	client *ent.Client,
	poolName string) (Pool, error) {

	pool, err := client.ResourcePool.Query().
		Where(resourcePool.NameEQ(poolName)).
		Only(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot create pool from existing entity")
	}

	switch pool.PoolType {
	case resourcePool.PoolTypeSingleton:
		return &SingletonPool{SetPool{pool, ctx, client}}, nil
	case resourcePool.PoolTypeSet:
		return &SetPool{pool, ctx, client}, nil
	default:
		return nil, errors.Errorf("Unknown pool type \"%s\"", pool.PoolType)
	}
}

func parseProps(
	ctx context.Context,
	tx *ent.Client,
	resourceType *ent.ResourceType,
	propertyValues RawResourceProps) (ent.Properties, error) {

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

func compareProps(
	ctx context.Context,
	resourceType *ent.ResourceType,
	propertyValues RawResourceProps) ([]predicate.Property, error) {

	var predicates []predicate.Property
	for pN, pV := range propertyValues {
		pT, err := resourceType.QueryPropertyTypes().Where(propertyType.NameEQ(pN)).Only(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "Unknown property: \"%s\" for resource type: \"%s\"", pN, resourceType)
		}

		predicate := property.HasTypeWith(propertyType.ID(pT.ID))

		// TODO is there a better way of parsing individual types ? Reuse something from inv ?
		// TODO add additional types
		// TODO we have this switch in 2 places
		switch pT.Type {
		case "int":
			predicate = property.And(predicate, property.IntValEQ(pV.(int)))
		case "string":
			predicate = property.And(predicate, property.StringValEQ(pV.(string)))
		default:
			return nil, errors.Errorf("Unsupported property type \"%s\"", pT.Type)
		}

		predicates = append(predicates, predicate)
	}

	return predicates, nil
}

// Destroy removes the pool from DB if there are no more claims
func (pool SetPool) Destroy() error {
	// Check if there are no more claims
	claims, err := pool.QueryResources()
	if err != nil {
		return err
	}

	if len(claims) > 0 {
		return errors.Errorf("Unable to destroy pool \"%s\", there are claimed resources",
			pool.Name)
	}

	// Delete props
	resources, err := pool.FindResources().All(pool.ctx)
	if err != nil {
		return errors.Wrapf(err, "Cannot destroy pool \"%s\". Unable to cleanup resoruces", pool.Name)
	}
	for _, res := range resources {
		props, err := res.QueryProperties().All(pool.ctx)
		if err != nil {
			return errors.Wrapf(err, "Cannot destroy pool \"%s\". Unable to cleanup resoruces", pool.Name)
		}

		for _, prop := range props {
			pool.client.Property.DeleteOne(prop).Exec(pool.ctx)
		}
		if err != nil {
			return errors.Wrapf(err, "Cannot destroy pool \"%s\". Unable to cleanup resoruces", pool.Name)
		}
	}

	// Delete resources
	_, err = pool.client.Resource.Delete().Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).Exec(pool.ctx)
	if err != nil {
		return errors.Wrapf(err, "Cannot destroy pool \"%s\". Unable to cleanup resoruces", pool.Name)
	}

	// Delete pool itself
	err = pool.client.ResourcePool.DeleteOne(pool.ResourcePool).Exec(pool.ctx)
	if err != nil {
		return errors.Wrapf(err, "Cannot destroy pool \"%s\"", pool.Name)
	}

	return nil
}

func (pool SetPool) AddLabel(label PoolLabel) error {
	// TODO implement labeling
	return errors.Errorf("NOT IMPLEMENTED")
}

func (pool SetPool) ClaimResource() (*ent.Resource, error) {
	// Allocate new resource for this tag
	unclaimedRes, err := pool.queryUnclaimedResourceEager()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to find unclaimed resource in pool \"%s\"",
		pool.Name)
	}
	
	err = pool.client.Resource.UpdateOne(unclaimedRes).SetClaimed(true).Exec(pool.ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to claim a resource in pool \"%s\"", pool.Name)
	}
	return unclaimedRes, err
}

func (pool SingletonPool) ClaimResource() (*ent.Resource, error) {
	return pool.queryUnclaimedResourceEager()
}

func (pool SetPool) FreeResource(raw RawResourceProps) error {
	return pool.freeResourceInner(raw)
}

func (pool SingletonPool) FreeResource(raw RawResourceProps) error {
	return nil
}

func (pool SetPool) freeResourceInner(raw RawResourceProps) error {
	query, err := pool.FindResource(raw)
	if err != nil {
		return errors.Wrapf(err, "Unable to find resource in pool: \"%s\"", pool.Name)
	}
	res, err := query.
		WithProperties().
		Only(pool.ctx)

	if err != nil {
		return errors.Wrapf(err, "Unable to free a resource in pool \"%s\". Unable to find resource", pool.Name)
	}

	if res.Claimed == false {
		return errors.Wrapf(err, "Unable to free a resource in pool \"%s\". It has not been claimed", pool.Name)
	}

	pool.client.Resource.UpdateOne(res).SetClaimed(false).Exec(pool.ctx)
	if err != nil {
		return errors.Wrapf(err, "Unable to free a resource in pool \"%s\". Unable to unclaim", pool.Name)
	}

	return nil
}

func (pool SetPool) FindResource(raw RawResourceProps) (*ent.ResourceQuery, error) {
	propComparator, err := compareProps(pool.ctx, pool.QueryResourceType().OnlyX(pool.ctx), raw)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to find resource in pool: \"%s\"", pool.Name)
	}

	return pool.FindResources().
		Where(resource.HasPropertiesWith(propComparator...)), nil
}

func (pool SetPool) QueryResource(raw RawResourceProps) (*ent.Resource, error) {
	query, err := pool.FindResource(raw)
	if err != nil {
		return nil, err
	}
	return query.
		Where(resource.Claimed(true)).
		Only(pool.ctx)
}

// load eagerly with some edges, ready to be copied
func (pool SetPool) queryUnclaimedResourceEager() (*ent.Resource, error) {
	// Find first unclaimed
	resource, err := pool.FindResources().
		Where(resource.Claimed(false)).
		First(pool.ctx)

	// No more unclaimed
	if ent.IsNotFound(err) {
		return nil, errors.Wrapf(err, "No more free resources in the pool: \"%s\"", pool.Name)
	}

	return resource, err
}

func (pool SetPool) FindResources() *ent.ResourceQuery {
	return pool.client.Resource.Query().
		Where(resource.HasPoolWith(resourcePool.ID(pool.ID)))
}

func (pool SetPool) QueryResources() (ent.Resources, error) {
	resource, err := pool.FindResources().
		Where(resource.Claimed(true)).
		All(pool.ctx)

	return resource, err
}
