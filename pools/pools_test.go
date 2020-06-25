package pools

import (
	"context"
	"log"
	"testing"

	"github.com/marosmars/resourceManager/authz"
	"github.com/marosmars/resourceManager/authz/models"
	"github.com/marosmars/resourceManager/ent"
	_ "github.com/marosmars/resourceManager/ent/runtime"
	_ "github.com/mattn/go-sqlite3"
)

func getContext() context.Context {
	ctx := context.Background()
	ctx = authz.NewContext(ctx, &models.PermissionSettings{
		CanWrite:        true,
		WorkforcePolicy: authz.NewWorkforcePolicy(true, true)})
	return ctx
}

func openDb(ctx context.Context) *ent.Client {
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

func getResourceType(ctx context.Context, client *ent.Client) *ent.ResourceType {
	propType, _ := client.PropertyType.Create().
		SetName("vlan").
		SetType("int").
		SetIntVal(0).
		SetMandatory(true).
		Save(ctx)

	resType, _ := client.ResourceType.Create().
		SetName("vlan").
		AddPropertyTypes(propType).
		Save(ctx)

	return resType
}

func TestNewSingletonPool(t *testing.T) {
	ctx := getContext()
	client := openDb(ctx)
	defer client.Close()
	resType := getResourceType(ctx, client)

	pool, err := NewSingletonPool(ctx, client, resType, map[string]interface{}{
		"vlan": 44,
	}, "singleton")

	if err != nil {
		t.Fatal(err)
	}

	err = pool.Destroy()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClaimResoource(t *testing.T) {
	ctx := getContext()
	client := openDb(ctx)
	defer client.Close()
	resType := getResourceType(ctx, client)

	pool, _ := NewSingletonPool(ctx, client, resType, map[string]interface{}{
		"vlan": 44,
	}, "singleton")

	claim, err := pool.ClaimResource(Scope{"customer1"})
	if err != nil {
		t.Error(err)
	}
	t.Log(claim)
	
	claim2, err := pool.ClaimResource(Scope{"customer2"})
	if err != nil {
		t.Error(err)
	}
	t.Log(claim2)

	if len(claim.QueryProperties().AllX(ctx)) != 1 {
		t.Fatalf("Missing properties in resource claim: %s", claim)
	}
	if claim.QueryProperties().AllX(ctx)[0].IntVal != int(44) {
		t.Fatalf("Wrong property in resource claim: %s", claim)
	}
	if claim2.QueryProperties().AllX(ctx)[0].IntVal != int(44) {
		t.Fatalf("Wrong property in resource claim: %s", claim)
	}

	pool.FreeResource(Scope{"customer1"})
	pool.FreeResource(Scope{"customer2"})
}