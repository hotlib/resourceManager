package main

import (
    "context"
    "log"

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
    SingletonPool{pool, ctx, client}.claimResource(Scope{"client1"})

    // Query resource
    log.Println(SingletonPool{pool, ctx, client}.queryResource(Scope{"client1"}))

    // Delete resource
    SingletonPool{pool, ctx, client}.freeResource(Scope{"client1"})

    // Query resource
    log.Println(SingletonPool{pool, ctx, client}.queryResource(Scope{"client1"}))
}