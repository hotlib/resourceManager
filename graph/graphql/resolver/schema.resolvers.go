package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"github.com/marosmars/resourceManager/ent"
	"github.com/marosmars/resourceManager/graph/graphql/generated"
	"github.com/marosmars/resourceManager/graph/graphql/model"
	"github.com/marosmars/resourceManager/pools"
	"strconv"
)

func (r *mutationResolver) ClaimResource(ctx context.Context, input model.Scope) (*ent.Resource, error) {
	return r.Pool.ClaimResource(pools.Scope{Scope: input.Scope})
}

func (r *mutationResolver) FreeResource(ctx context.Context, input model.Scope) (string, error) {
	err := r.Pool.FreeResource(pools.Scope{Scope: input.Scope})
	if err == nil {
		return "all ok", err
	}
	return err.Error(), err
}

func (r *queryResolver) QueryResource(ctx context.Context, input model.Scope) (*ent.Resource, error) {
	return r.Pool.QueryResource(pools.Scope{Scope: input.Scope})
}

func (r *queryResolver) QueryResources(ctx context.Context) ([]*ent.Resource, error) {
	return r.Pool.QueryResources()
}

func (r *resourceResolver) Config(ctx context.Context, obj *ent.Resource) (string, error) {
	return "Name:" + obj.Scope + " ID: " + strconv.Itoa(obj.ID), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Resource returns generated.ResourceResolver implementation.
func (r *Resolver) Resource() generated.ResourceResolver { return &resourceResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type resourceResolver struct{ *Resolver }
