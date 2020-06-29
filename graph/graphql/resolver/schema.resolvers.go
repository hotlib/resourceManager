package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/marosmars/resourceManager/ent"
	"github.com/marosmars/resourceManager/graph/graphql/generated"
	"github.com/marosmars/resourceManager/graph/graphql/model"
	"github.com/marosmars/resourceManager/pools"
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

func (r *propertyResolver) PropertyType(ctx context.Context, obj *ent.Property) (int, error) {
	//TODO get property-type for Property
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) QueryResource(ctx context.Context, input model.Scope) (*ent.Resource, error) {
	return r.Pool.QueryResource(pools.Scope{Scope: input.Scope})
}

func (r *queryResolver) QueryResources(ctx context.Context) ([]*ent.Resource, error) {
	return r.Pool.QueryResources()
}

func (r *resourceEdgesResolver) Properties(ctx context.Context, obj *ent.ResourceEdges) (*ent.Property, error) {
	//TODO get properties for ResourceEdge
	panic(fmt.Errorf("not implemented"))
}

func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

func (r *Resolver) Property() generated.PropertyResolver { return &propertyResolver{r} }

func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

func (r *Resolver) ResourceEdges() generated.ResourceEdgesResolver { return &resourceEdgesResolver{r} }

type mutationResolver struct{ *Resolver }
type propertyResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type resourceEdgesResolver struct{ *Resolver }

