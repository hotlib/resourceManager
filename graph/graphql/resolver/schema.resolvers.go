package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/marosmars/resourceManager/ent"
	"github.com/marosmars/resourceManager/graph/graphql/generated"
)

func (r *mutationResolver) ClaimResource(ctx context.Context) (*ent.Resource, error) {
	return r.Pool.ClaimResource()
}

func (r *mutationResolver) FreeResource(ctx context.Context, input map[string]interface{}) (string, error) {
	err := r.Pool.FreeResource(input)
	if err == nil {
		return "all ok", err
	}
	return err.Error(), err
}

func (r *queryResolver) QueryResource(ctx context.Context, input map[string]interface{}) (*ent.Resource, error) {
	return r.Pool.QueryResource(input)
}

func (r *queryResolver) QueryResources(ctx context.Context) ([]*ent.Resource, error) {
	return r.Pool.QueryResources()
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

