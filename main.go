package main

import (
	"context"
	"fmt"
	"log"

	"github.com/marosmars/resourceManager/authz"
	"github.com/marosmars/resourceManager/authz/models"
	"github.com/marosmars/resourceManager/ent"
	"github.com/marosmars/resourceManager/ent/propertytype"
	"github.com/marosmars/resourceManager/ent/resourcetype"
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
	ctx = authz.NewContext(ctx, &models.PermissionSettings{CanWrite: true})
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	if _, err := createVlanType(ctx, client); err != nil {
		log.Fatalf("failed creating resource type: %v", err)
	}
	if _, err := createVlan(ctx, client); err != nil {
		log.Fatalf("failed creating resource: %v", err)
	}
	if _, err := queryRType(ctx, client); err != nil {
		log.Fatalf("failed querying resource type: %v", err)
	}
}

func createVlanType(ctx context.Context, client *ent.Client) (*ent.ResourceType, error) {
	p, err := client.PropertyType.Create().
		SetName("vlan").
		SetType("int").
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed creating property type: %v", err)
	}

	u, err := client.ResourceType.Create().
		SetName("Vlan").
		AddPropertyTypes(p).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed creating resource type: %v", err)
	}
	fmt.Println("resource type was created: ", u)
	return u, nil
}

func createVlan(ctx context.Context, client *ent.Client) (*ent.Resource, error) {
	resType, err := queryRType(ctx, client)
	propertyType, err := resType.QueryPropertyTypes().
		Where(propertytype.Name("vlan")).
		Only(ctx)

	client.Property.Create().
		SetIntVal(111).
		SetType(propertyType).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed creating property type: %v", err)
	}

	r, err := client.Resource.Create().
		SetName("vlan1").
		SetType(resType).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed creating property type: %v", err)
	}
	fmt.Println("resource was created: ", r)
	return r, nil
}

func queryRType(ctx context.Context, client *ent.Client) (*ent.ResourceType, error) {
	u, err := client.ResourceType.
		Query().
		Where(resourcetype.NameEQ("Vlan")).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying resource types: %v", err)
	}
	return u, nil
}
