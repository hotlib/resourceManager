package pools

import (
    "context"

    "github.com/marosmars/resourceManager/ent"
    "github.com/pkg/errors"
)

// TxFunction is a WithTx lambda
type TxFunction func(tx *ent.Tx) (interface{}, error)

// WithTx function executes a lambda function within a tx
func WithTx(
    ctx context.Context,
    client *ent.Client,
    fn TxFunction) (interface{}, error) {
    tx, err := client.Tx(ctx)
    if err != nil {
        return nil, err
    }
    defer func() {
        if v := recover(); v != nil {
            tx.Rollback()
            panic(v)
        }
    }()
    retVal, err := fn(tx)
    if err != nil {
        if rerr := tx.Rollback(); rerr != nil {
            err = errors.Wrapf(err, "rolling back transaction: %v", rerr)
        }
        return nil, err
    }
    if err := tx.Commit(); err != nil {
        return nil, errors.Wrapf(err, "committing transaction: %v", err)
    }
    return retVal, nil
}
