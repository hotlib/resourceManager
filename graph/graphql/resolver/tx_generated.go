// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package resolver

import (
	"context"
	"fmt"

	"github.com/marosmars/resourceManager/ent"
	"github.com/marosmars/resourceManager/graph/graphql/generated"
	"github.com/marosmars/resourceManager/graph/graphql/model"
)

// txResolver wraps a mutation resolver and executes every mutation under a transaction.
type txResolver struct {
	generated.MutationResolver
}

func (tr txResolver) WithTransaction(ctx context.Context, f func(context.Context, generated.MutationResolver) error) error {
	tx, err := ent.FromContext(ctx).Tx(ctx)
	if err != nil {
		return fmt.Errorf("creating transaction: %w", err)
	}
	ctx = ent.NewTxContext(ctx, tx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	ctx = ent.NewContext(ctx, tx.Client())
	if err := f(ctx, tr.MutationResolver); err != nil {
		if r := tx.Rollback(); r != nil {
			err = fmt.Errorf("rolling back transaction: %v", r)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func (tr txResolver) ClaimResource(ctx context.Context, input model.Scope) (*model.Resource, error) {
	var result, zero *model.Resource
	if err := tr.WithTransaction(ctx, func(ctx context.Context, mr generated.MutationResolver) (err error) {
		result, err = mr.ClaimResource(ctx, input)
		return
	}); err != nil {
		return zero, err
	}
	return result, nil
}

func (tr txResolver) FreeResource(ctx context.Context, input model.Scope) (string, error) {
	var result, zero string
	if err := tr.WithTransaction(ctx, func(ctx context.Context, mr generated.MutationResolver) (err error) {
		result, err = mr.FreeResource(ctx, input)
		return
	}); err != nil {
		return zero, err
	}
	return result, nil
}
