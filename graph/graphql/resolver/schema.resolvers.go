package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/marosmars/resourceManager/ent"
	"github.com/marosmars/resourceManager/ent/resourcepool"
	"github.com/marosmars/resourceManager/graph/graphql/generated"
	p "github.com/marosmars/resourceManager/pools"
)

func (r *mutationResolver) ClaimResource(ctx context.Context, poolName string) (*ent.Resource, error) {
	pool, err := p.ExistingPool(ctx, r.ClientFrom(ctx), poolName)
	if err != nil {
		return nil, err
	}

	return pool.ClaimResource()
}

func (r *mutationResolver) FreeResource(ctx context.Context, input map[string]interface{}, poolName string) (string, error) {
	pool, err := p.ExistingPool(ctx, r.ClientFrom(ctx), poolName)
	if err != nil {
		return err.Error(), err
	}
	err = pool.FreeResource(input)
	if err == nil {
		return "all ok", nil
	}

	return err.Error(), err
}

func (r *mutationResolver) CreatePool(ctx context.Context, poolType *resourcepool.PoolType, resourceTypeID int, poolName string, poolValues []map[string]interface{}, allocationScript string) (*ent.ResourcePool, error) {
	var client = r.ClientFrom(ctx)

	resType, _ := client.ResourceType.Get(ctx, resourceTypeID)

	//TODO we support int, but we always get int64 instead of int
	for i, v := range poolValues {
		for k, val := range v {
			fmt.Printf("key[%s] value[%s]\n", k, v)
			if reflect.TypeOf(val).String() == "int64" {
				poolValues[i][k] = int(val.(int64))
			}
		}
	}

	var rawProps []p.RawResourceProps

	for _, v := range poolValues {
		rawProps = append(rawProps, v)
	}

	if resourcepool.PoolTypeSet == *poolType {
		_, rp, err := p.NewSetPoolWithMeta(ctx, client, resType, rawProps, poolName)
		return rp, err
	} else if resourcepool.PoolTypeSingleton == *poolType {
		if len(rawProps) > 0 {
			_, rp, err := p.NewSingletonPoolWithMeta(ctx, client, resType, rawProps[0], poolName)
			return rp, err
		} else {
			//TODO logging missing rawProps parameter
		}
	}

	//TODO something went wrong, log or smth
	return nil, nil
}

func (r *mutationResolver) DeleteResourcePool(ctx context.Context, resourcePoolID int) (string, error) {
	client := r.ClientFrom(ctx)
	pool, err := p.ExistingPoolFromId(ctx, client, resourcePoolID)

	if err != nil {
		return "not ok", err
	}

	//TODO can you delete a pool which has allocated resources??
	allocatedResources, err2 := pool.QueryResources()
	if len(allocatedResources) > 0 || err2 != nil {
		return "not ok", errors.New("resource pool has allocated resources, deallocate those first")
	}

	return "ok", client.ResourcePool.DeleteOneID(resourcePoolID).Exec(ctx)
}

func (r *mutationResolver) CreateResourceType(ctx context.Context, resourceName string, resourceProperties map[string]interface{}) (*ent.ResourceType, error) {
	var client = r.ClientFrom(ctx)
	//TODO property and resource name the same?
	//TODO check error
	var propType, _ = p.CreatePropertyType(ctx, client, resourceName, resourceProperties["type"], resourceProperties["init"])

	resType, _ := client.ResourceType.Create().
		SetName(resourceName).
		AddPropertyTypes(propType).
		Save(ctx)

	return resType, nil
}

func (r *mutationResolver) DeleteResourceType(ctx context.Context, resourceTypeID int) (string, error) {
	client := r.ClientFrom(ctx)
	resourceType, err := client.ResourceType.Get(ctx, resourceTypeID)

	if err != nil {
		return "not ok", err
	}

	pools, err2 := client.ResourceType.QueryPools(resourceType).All(ctx)

	if err2 != nil {
		return "not ok", err2
	}

	if len(pools) > 0 {
		return "not ok", errors.New("resourceType has pools, can't delete (delete resource pools first)")
	}

	return "ok", client.ResourcePool.DeleteOneID(resourceTypeID).Exec(ctx)
}

func (r *mutationResolver) UpdateResourceTypeName(ctx context.Context, resourceTypeID int, resourceName string) (*ent.ResourceType, error) {
	var client = r.ClientFrom(ctx)
	return client.ResourceType.UpdateOneID(resourceTypeID).SetName(resourceName).Save(ctx)
}

func (r *mutationResolver) AddResourceTypeProperty(ctx context.Context, resourceTypeID int, resourceProperties map[string]interface{}) (*ent.ResourceType, error) {
	var client = r.ClientFrom(ctx)

	exist, resourceType := p.CheckIfPoolsExist(ctx, client, resourceTypeID)

	if exist {
		//TODO add annoying GO error handling
		return nil, nil
	}

	propertyType, err := p.CreatePropertyType(ctx, client, resourceType.Name, resourceProperties["type"], resourceProperties["init"])

	if err != nil {
		//TODO add annoying GO error handling
		return nil, err
	}

	return client.ResourceType.UpdateOneID(resourceTypeID).AddPropertyTypeIDs(propertyType.ID).Save(ctx)
}

func (r *mutationResolver) AddExistingPropertyToResourceType(ctx context.Context, resourceTypeID int, propertyTypeID int) (int, error) {
	var client = r.ClientFrom(ctx)
	return propertyTypeID, client.ResourceType.UpdateOneID(resourceTypeID).AddPropertyTypeIDs(propertyTypeID).Exec(ctx)
}

func (r *mutationResolver) RemoveResourceTypeProperty(ctx context.Context, resourceTypeID int, propertyTypeID int) (*ent.ResourceType, error) {
	var client = r.ClientFrom(ctx)
	exist, _ := p.CheckIfPoolsExist(ctx, client, resourceTypeID)

	if exist {
		//TODO add annoying GO error handling
		return nil, nil
	}

	resourceType, err := client.ResourceType.UpdateOneID(resourceTypeID).RemovePropertyTypeIDs(propertyTypeID).Save(ctx)

	if err == nil {
		//TODO annoying erro handlign dome smth XXX
		err2 := client.PropertyType.DeleteOneID(propertyTypeID).Exec(ctx)
		return resourceType, err2
	}

	return resourceType, err
}

func (r *mutationResolver) CreatePropertyType(ctx context.Context, propertyName string, typeProperties map[string]interface{}) (*ent.PropertyType, error) {
	var client = r.ClientFrom(ctx)
	return p.CreatePropertyType(ctx, client, propertyName, typeProperties["type"], typeProperties["init"])
}

func (r *mutationResolver) UpdatePropertyType(ctx context.Context, propertyTypeID int, propertyName string, typeProperties map[string]interface{}) (bool, error) {
	var client = r.ClientFrom(ctx)
	if !p.HasPropertyTypeExistingProperties(ctx, client, propertyTypeID) {
		return true, p.UpdatePropertyType(ctx, client, propertyTypeID, propertyName, typeProperties["type"], typeProperties["init"])
	}

	return false, nil //TODO fill out error and log
}

func (r *mutationResolver) DeletePropertyType(ctx context.Context, propertyTypeID int) (bool, error) {
	var client = r.ClientFrom(ctx)
	if !p.HasPropertyTypeExistingProperties(ctx, client, propertyTypeID) {
		return true, client.PropertyType.DeleteOneID(propertyTypeID).Exec(ctx)
	}

	return false, nil //TODO fill out error and log
}

func (r *queryResolver) QueryResource(ctx context.Context, input map[string]interface{}, poolName string) (*ent.Resource, error) {
	pool, err := p.ExistingPool(ctx, r.ClientFrom(ctx), poolName)
	if err != nil {
		return nil, err
	}
	return pool.QueryResource(input)
}

func (r *queryResolver) QueryResources(ctx context.Context, poolName string) ([]*ent.Resource, error) {
	pool, err := p.ExistingPool(ctx, r.ClientFrom(ctx), poolName)
	if err != nil {
		return nil, err
	}
	return pool.QueryResources()
}

func (r *queryResolver) QueryResourceTypes(ctx context.Context) ([]*ent.ResourceType, error) {
	client := r.ClientFrom(ctx)
	return client.ResourceType.Query().All(ctx)
}

func (r *queryResolver) QueryResourcePools(ctx context.Context) ([]*ent.ResourcePool, error) {
	client := r.ClientFrom(ctx)
	return client.ResourcePool.Query().All(ctx)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
