package main

import (
	"context"
	"fmt"

	"github.com/marosmars/resourceManager/graph/graphgrpc/schema"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func main() {
	// TODO this is for testing purposes only
	CreateTenant("FBC")
}

// Create a new tenant in RM
func CreateTenant(tenant string) {
	conn, err := grpc.Dial("localhost:8885", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Cannot open gRPC: %s", err)
		// TODO error
	}
	
	defer conn.Close()

	client := schema.NewTenantServiceClient(conn)

	newTenant, err := client.Create(context.Background(), &wrapperspb.StringValue{Value: tenant})
	if err != nil {
		fmt.Errorf("Cannot create a tenant: %s", err)
	}

	fmt.Printf("Tenant created successfully: %s", newTenant)
}
