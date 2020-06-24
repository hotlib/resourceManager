package main

import (
	"context"
	"github.com/marosmars/resourceManager/ent"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/marosmars/resourceManager/ent/runtime"
	"github.com/marosmars/resourceManager/graph"
	"github.com/marosmars/resourceManager/graph/generated"
	"github.com/marosmars/resourceManager/pools"
	_ "github.com/mattn/go-sqlite3"
)

const defaultPort = "8080"

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

//TODO 	defer client.Close() where??
func initDummyPool() pools.Pool {
	ctx := pools.GetContext()
	client := pools.OpenDb(ctx)
	resType := getResourceType(ctx, client)
	pool, _ := pools.NewSingletonPool(ctx, client, resType, map[string]interface{}{
		"vlan": 44,
	}, "singleton")
	return pool
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	var pool = initDummyPool()

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{Pool: pool}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
