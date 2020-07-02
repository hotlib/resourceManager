// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build wireinject

package main

import (
	"context"
	"fmt"

	"github.com/marosmars/resourceManager/graph/graphgrpc"
	"github.com/marosmars/resourceManager/ent"
	"github.com/marosmars/resourceManager/graph/graphhttp"
	"github.com/marosmars/resourceManager/log"
	"github.com/marosmars/resourceManager/mysql"
	"github.com/marosmars/resourceManager/server"
	"github.com/marosmars/resourceManager/viewer"

	"github.com/google/wire"
	"gocloud.dev/server/health"
	"google.golang.org/grpc"
)

func newApplication(ctx context.Context, flags *cliFlags) (*application, func(), error) {
	wire.Build(
		wire.FieldsOf(new(*cliFlags),
			"MySQLConfig",
			"LogConfig",
			"TelemetryConfig",
			"TenancyConfig",
		),
		log.Provider,
		newApp,
		newTenancy,
		newHealthChecks,
		newMySQLTenancy,
		mysql.Provider,
		graphhttp.NewServer,
		wire.Struct(new(graphhttp.Config), "*"),
		graphgrpc.NewServer,
		wire.Struct(new(graphgrpc.Config), "*"),
	)
	return nil, nil, nil
}

func newApp(logger log.Logger, httpServer *server.Server, grpcServer *grpc.Server, flags *cliFlags) *application {
	var app application
	app.Logger = logger.Background()
	app.http.Server = httpServer
	app.http.addr = flags.HTTPAddress.String()
	app.grpc.Server = grpcServer
	app.grpc.addr = flags.GRPCAddress.String()
	return &app
}

func newTenancy(tenancy *viewer.MySQLTenancy) (viewer.Tenancy, error) {
	initFunc := func(*ent.Client) {
		// NOOP
	}
	return viewer.NewCacheTenancy(tenancy, initFunc), nil
}

func newHealthChecks(tenancy *viewer.MySQLTenancy) []health.Checker {
	return []health.Checker{tenancy}
}

func newMySQLTenancy(config mysql.Config, tenancyConfig viewer.Config, logger log.Logger) (*viewer.MySQLTenancy, error) {
	tenancy, err := viewer.NewMySQLTenancy(config.String(), tenancyConfig.TenantMaxConn)
	if err != nil {
		return nil, fmt.Errorf("creating mysql tenancy: %w", err)
	}
	tenancy.SetLogger(logger)
	mysql.SetLogger(logger)
	return tenancy, nil
}
