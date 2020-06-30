package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"github.com/marosmars/resourceManager/ent"
	"github.com/marosmars/resourceManager/graph/graphql/generated"
	"github.com/marosmars/resourceManager/pools"
)

func (r *mutationResolver) ClaimResource(ctx context.Context, input pools.ResourceTag) (*ent.Resource, error) {
	return r.Pool.ClaimResource(input)
}

func (r *mutationResolver) FreeResource(ctx context.Context, input pools.ResourceTag) (string, error) {
	err := r.Pool.FreeResource(input)
	if err == nil {
		return "all ok", err
	}
	return err.Error(), err
}

func (r *propertyResolver) PropertyType(ctx context.Context, obj *ent.Property) (int, error) {
	propertyType, err := obj.QueryType().Only(ctx)
	return propertyType.ID, err //TODO ID is wrong
}

func (r *queryResolver) QueryResource(ctx context.Context, input pools.ResourceTag) (*ent.Resource, error) {
	return r.Pool.QueryResource(input)
}

func (r *queryResolver) QueryResources(ctx context.Context) ([]*ent.Resource, error) {
	return r.Pool.QueryResources()
}

func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

func (r *Resolver) Property() generated.PropertyResolver { return &propertyResolver{r} }

func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type propertyResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

