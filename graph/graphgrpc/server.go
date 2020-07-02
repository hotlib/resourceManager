// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graphgrpc

import (
	"context"
	"database/sql"
	"fmt"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/marosmars/resourceManager/actions/executor"
	"github.com/marosmars/resourceManager/graph/graphgrpc/schema"
	"github.com/marosmars/resourceManager/grpc-middleware/sqltx"
	"github.com/marosmars/resourceManager/log"
	"github.com/marosmars/resourceManager/viewer"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func newServer(tenancy viewer.Tenancy, db *sql.DB, logger log.Logger, registry *executor.Registry) (*grpc.Server, func(), error) {
	grpc_zap.ReplaceGrpcLoggerV2(logger.Background())
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger.Background()),
			grpc_recovery.UnaryServerInterceptor(),
			sqltx.UnaryServerInterceptor(db),
		)),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)
	schema.RegisterTenantServiceServer(s,
		NewTenantService(func(ctx context.Context) ExecQueryer {
			return sqltx.FromContext(ctx)
		}),
	)

	reflection.Register(s)
	err := view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return nil, nil, fmt.Errorf("registering grpc views: %w", err)
	}
	return s, func() { view.Unregister(ocgrpc.DefaultServerViews...) }, nil
}
