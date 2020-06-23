package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/marosmars/resourceManager/graph/generated"
	models "github.com/marosmars/resourceManager/graph/model"
)

var dummyData = models.Testdata{
	ID:   "100",
	Attr: "input",
}

func (r *mutationResolver) CreateTestdata(ctx context.Context, input string) (*models.Testdata, error) {
	return &dummyData, nil
}

func (r *queryResolver) Testdata(ctx context.Context) ([]*models.Testdata, error) {
	var emptyArray []*models.Testdata
	var result = append(emptyArray, &dummyData)
	return result, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }