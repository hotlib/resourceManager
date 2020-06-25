package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strconv"

	"github.com/marosmars/resourceManager/graph/graphql/generated"
	"github.com/marosmars/resourceManager/graph/graphql/model"
	"github.com/marosmars/resourceManager/pools"
)

func (r *mutationResolver) ClaimResource(ctx context.Context, input model.Scope) (*model.Resource, error) {
	resource, err := r.Pool.ClaimResource(pools.Scope{Scope: input.Scope})
	return &model.Resource{Config: resource.Scope}, err //TODO check resource == nil
}

func (r *mutationResolver) FreeResource(ctx context.Context, input model.Scope) (string, error) {
	err := r.Pool.FreeResource(pools.Scope{Scope: input.Scope})
	if err == nil {
		return "all ok", err
	}
	return err.Error(), err
}

func (r *queryResolver) QueryResource(ctx context.Context, input model.Scope) (*model.Resource, error) {
	queryResource, err := r.Pool.QueryResource(pools.Scope{Scope: input.Scope})

	if queryResource == nil {
		return &model.Resource{Config: "NOTHING!"}, err
	}

	return &model.Resource{Config: queryResource.Scope + ":" + strconv.Itoa(queryResource.ID)}, err
}

func (r *queryResolver) QueryResources(ctx context.Context) ([]*model.Resource, error) {
	queryResources, err := r.Pool.QueryResources()
	var result []*model.Resource
	for _, s := range queryResources {
		result = append(result, &model.Resource{Config: s.Scope + ":" + strconv.Itoa(s.ID)})
	}
	return result, err
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
