package pools

import (
	"context"

	"github.com/marosmars/resourceManager/ent"
	"github.com/marosmars/resourceManager/ent/predicate"
	property "github.com/marosmars/resourceManager/ent/property"
	propertyType "github.com/marosmars/resourceManager/ent/propertytype"
	resourcePool "github.com/marosmars/resourceManager/ent/resourcepool"
	"github.com/pkg/errors"
)

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

type SetPool struct {
	*ent.ResourcePool

	ctx    context.Context
	client *ent.Client
}

type SingletonPool struct {
	SetPool
}

type RawResourceProps map[string]interface{}

func newPoolInner(ctx context.Context,
	client *ent.Client,
	resourceType *ent.ResourceType,
	propertyValues []RawResourceProps,
	poolName string,
	poolType resourcePool.PoolType) (*ent.ResourcePool, error) {
	pool, err := client.ResourcePool.Create().
		SetName(poolName).
		SetPoolType(poolType).
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

	return existingPool(ctx, client, pool)
}

func ExistingPoolFromId(
	ctx context.Context,
	client *ent.Client,
	poolId int) (Pool, error) {

	pool, err := client.ResourcePool.Query().
		Where(resourcePool.ID(poolId)).
		Only(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot create pool from existing entity")
	}

	return existingPool(ctx, client, pool)
}

func existingPool(
	ctx context.Context,
	client *ent.Client,
	pool *ent.ResourcePool) (Pool, error) {

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

