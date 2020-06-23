package main

import (
	"context"
	"log"

	"github.com/marosmars/resourceManager/authz"
	"github.com/marosmars/resourceManager/authz/models"
	"github.com/marosmars/resourceManager/ent"
	resource "github.com/marosmars/resourceManager/ent/resource"
	resourcePool "github.com/marosmars/resourceManager/ent/resourcepool"
	_ "github.com/marosmars/resourceManager/ent/runtime"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func main() {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	// run the auto migration tool.
	ctx := context.Background()
	ctx = authz.NewContext(ctx, &models.PermissionSettings{CanWrite: true})
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	propType, err := client.PropertyType.Create().
		SetName("vlan").
		SetType("int").
		SetIntVal(0).
		Save(ctx)
	defer client.PropertyType.DeleteOne(propType).Exec(ctx)

	resType, _ := client.ResourceType.Create().
		SetName("vlan").
		AddPropertyTypes(propType).
		Save(ctx)
	defer client.ResourceType.DeleteOne(resType).Exec(ctx)

	pool, _ := client.ResourcePool.Create().
		SetName("singleton_vlan_11").
		SetPoolType("singleton").
		SetResourceType(resType).
		Save(ctx)
	defer client.ResourcePool.DeleteOne(pool).Exec(ctx)

	// Claim / allocate resource
	singletonPool{pool}.claimResource(ctx, client, "client1")

	// Query resource
	log.Println(singletonPool{pool}.queryResource(ctx, client, "client1"))

	// Delete resource
	singletonPool{pool}.freeResource(ctx, client, "client1")

	// Query resource
	log.Println(singletonPool{pool}.queryResource(ctx, client, "client1"))
}

// Extension functions for singleton type pool entity
type singletonPool struct {
	*ent.ResourcePool
}

func (pool singletonPool) claimResource(ctx context.Context, client *ent.Client, scope string) (*ent.Resource, error) {
	claim, err := WithTx(ctx, client, func(tx *ent.Tx) (interface{}, error) {
		return pool.claimResourceInner(ctx, tx.Client(), scope)
	})

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to claim a resource")
	}

	return claim.(*ent.Resource), nil
}

func (pool singletonPool) claimResourceInner(
	ctx context.Context,
	client *ent.Client,
	scope string) (*ent.Resource, error) {
	res, err := pool.queryResource(ctx, client, scope)

	// Resource exists for the scope, return the same one
	if res != nil {
		return res, nil
	}

	// Allocate new resource for this scope
	if ent.IsNotFound(err) {
		return client.Resource.Create().
			SetPool(pool.ResourcePool).
			SetScope(scope).
			Save(ctx)

		// TODO properties
	}

	return res, err
}

func (pool singletonPool) freeResource(ctx context.Context, client *ent.Client, scope string) {
	client.Resource.Delete().
		Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).
		Where(resource.ScopeEQ(scope)).
		Exec(ctx)
}

func (pool singletonPool) queryResource(
	ctx context.Context,
	client *ent.Client,
	scope string) (*ent.Resource, error) {
	resource, err := client.Resource.Query().
		Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).
		Where(resource.ScopeEQ(scope)).
		Only(ctx)

	return resource, err
}

type txFunction func(tx *ent.Tx) (interface{}, error)

// WithTx function executes a lambda function within a tx
func WithTx(
	ctx context.Context,
	client *ent.Client,
	fn txFunction) (interface{}, error) {
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
