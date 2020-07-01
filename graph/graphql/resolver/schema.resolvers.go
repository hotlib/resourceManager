package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/marosmars/resourceManager/ent"
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
