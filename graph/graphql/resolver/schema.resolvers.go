package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"reflect"
	"strings"

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

func (r *mutationResolver) CreateResourceType(ctx context.Context, resourceName string, resourceProperties map[string]interface{}) (*ent.ResourceType, error) {
	var client = r.ClientFrom(ctx)
	var resourceTypeName = fmt.Sprintf("%v", resourceProperties["type"])
	prop := client.PropertyType.Create().
		SetName(resourceName). //TODO property and resource name the same?
		SetType(resourceTypeName).
		SetMandatory(true)

	//TODO we support int, but we always get int64 instead of int
	if reflect.TypeOf(resourceProperties["init"]).String() == "int64" {
		resourceProperties["init"] = int(resourceProperties["init"].(int64))
	}

	in := []reflect.Value{reflect.ValueOf(resourceProperties["init"])}
	reflect.ValueOf(prop).MethodByName("Set" + strings.Title(resourceTypeName) + "Val").Call(in)

	var propType, _ = prop.Save(ctx)

	resType, _ := client.ResourceType.Create().
		SetName(resourceName).
		AddPropertyTypes(propType).
		Save(ctx)

	return resType, nil
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

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
