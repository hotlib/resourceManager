package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strconv"

	"github.com/marosmars/resourceManager/graph/generated"
	models "github.com/marosmars/resourceManager/graph/model"
	"github.com/marosmars/resourceManager/pools"
)

func (r *mutationResolver) ClaimResource(ctx context.Context, input models.Scope) (*models.Resource, error) {
	resource, err := r.Pool.ClaimResource(pools.Scope{Scope: input.Scope})
	return &models.Resource{Config: resource.Scope}, err //TODO check resource == nil
}

func (r *mutationResolver) FreeResource(ctx context.Context, input models.Scope) (string, error) {
	err := r.Pool.FreeResource(pools.Scope{Scope: input.Scope})
	if err == nil {
		return "all ok", err
	}
	return err.Error(), err 
}

func (r *queryResolver) QueryResource(ctx context.Context, input models.Scope) (*models.Resource, error) {
	queryResource, err := r.Pool.QueryResource(pools.Scope{Scope: input.Scope})

	if queryResource == nil {
		return & models.Resource{Config:"NOTHING!"}, err
	}
	
	return & models.Resource{Config:queryResource.Scope + ":" + strconv.Itoa(queryResource.ID)}, err
}

func (r *queryResolver) QueryResources(ctx context.Context) ([]*models.Resource, error) {
	queryResources, err := r.Pool.QueryResources()
	var result [] *models.Resource
	for _, s := range queryResources {
		result = append(result, & models.Resource{Config: s.Scope + ":" + strconv.Itoa(s.ID)})
	}
	return result, err
}

func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
