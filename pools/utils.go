package pools

import (
    "context"
    "github.com/marosmars/resourceManager/authz"
    "github.com/marosmars/resourceManager/authz/models"
    "log"

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

func GetContext() context.Context {
    ctx := context.Background()
    ctx = authz.NewContext(ctx, &models.PermissionSettings{
        CanWrite:        true,
        WorkforcePolicy: authz.NewWorkforcePolicy(true, true)})
    return ctx
}

func OpenDb(ctx context.Context) *ent.Client {
    client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
    if err != nil {
        log.Fatalf("failed opening connection to sqlite: %v", err)
    }
    // run the auto migration tool.
    if err := client.Schema.Create(ctx); err != nil {
        log.Fatalf("failed creating schema resources: %v", err)
    }

    return client
}