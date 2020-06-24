package main

import (
	"context"
	"log"

	"github.com/marosmars/resourceManager/pools"
	"github.com/marosmars/resourceManager/authz"
	"github.com/marosmars/resourceManager/authz/models"
	"github.com/marosmars/resourceManager/ent"
	_ "github.com/marosmars/resourceManager/ent/runtime"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	// run the auto migration tool.
	ctx := context.Background()
	ctx = authz.NewContext(ctx, &models.PermissionSettings{
		CanWrite:        true,
		WorkforcePolicy: authz.NewWorkforcePolicy(true, true)})
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

	pool, _ := pools.NewSingletonPool(ctx, client, resType, map[string]interface{}{
		"vlan": 44,
	}, "singleton")
	defer pool.Destroy()

	claimScope := pools.Scope{"client1"}
	pool.ClaimResource(claimScope)

	log.Println(pool.QueryResource(claimScope))

	pool.FreeResource(claimScope)

	log.Println(pool.QueryResource(claimScope))
}
