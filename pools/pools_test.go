package pools

import (
	"context"
	"log"
	"reflect"
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

	claim1, err := pool.ClaimResource(Scope{"customer1"})
	if err != nil {
		t.Error(err)
	}
	t.Log(claim1)
	if claim1.Scope != "customer1" {
		t.Fatalf("Wrong scope in %s", claim1)
	}

	claim2, err := pool.ClaimResource(Scope{"customer2"})
	if err != nil {
		t.Error(err)
	}
	t.Log(claim2)
	if claim2.Scope != "customer2" {
		t.Fatalf("Wrong scope in %s", claim2)
	}

	entityPool := pool.(*SingletonPool).ResourcePool
	if claim1.QueryPool().OnlyX(ctx).ID != entityPool.ID {
		t.Fatalf("Wrong resource pool set expected: %s but was: %s",
			entityPool, claim1.QueryPool().OnlyX(ctx))
	}
	if claim2.QueryPool().OnlyX(ctx).ID != entityPool.ID {
		t.Fatalf("Wrong resource pool set expected %s but was: %s",
			entityPool, claim2.QueryPool().OnlyX(ctx))
	}

	claimProps1 := claim1.QueryProperties().AllX(ctx)
	claimProps2 := claim2.QueryProperties().AllX(ctx)
	if len(claimProps1) != 1 {
		t.Fatalf("Missing properties in resource claim: %s", claim1)
	}
	if claimProps1[0].IntVal != int(44) {
		t.Fatalf("Wrong property in resource claim: %s", claim1)
	}

	if claimProps2[0].IntVal != int(44) {
		t.Fatalf("Wrong property in resource claim: %s", claim1)
	}

	assertDb(ctx, client, t, 1, 1, 1, 3, 3)

	pool.FreeResource(Scope{"customer1"})
	pool.FreeResource(Scope{"customer2"})

	assertDb(ctx, client, t, 1, 1, 1, 1, 1)

	pool.Destroy()

	assertDb(ctx, client, t, 1, 1, 0, 0, 0)
}

func assertDb(ctx context.Context, client *ent.Client, t *testing.T, count ...int) {
	assertInstancesInDb(client.PropertyType.Query().AllX(ctx), count[0], t)
	assertInstancesInDb(client.ResourceType.Query().AllX(ctx), count[1], t)
	assertInstancesInDb(client.ResourcePool.Query().AllX(ctx), count[2], t)
	assertInstancesInDb(client.Property.Query().AllX(ctx), count[3], t)
	assertInstancesInDb(client.Resource.Query().AllX(ctx), count[4], t)
}

func assertInstancesInDb(instances interface{}, expected int, t *testing.T) {
	slice := reflect.ValueOf(instances)
	if slice.Kind() != reflect.Slice {
		t.Fatalf("%s is not a slice, cannot assert length", instances)
	}

	if slice.Len() != expected {
		t.Fatalf("%d different properties expected, got: %s", expected, slice)
	}
}
