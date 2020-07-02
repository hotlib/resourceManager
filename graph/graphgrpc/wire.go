// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build wireinject

package graphgrpc

import (
	"database/sql"

	"github.com/marosmars/resourceManager/actions/executor"
	"github.com/marosmars/resourceManager/log"
	"github.com/marosmars/resourceManager/viewer"

	"github.com/google/wire"
	"google.golang.org/grpc"
)

// Config defines the grpc server config.
type Config struct {
	DB      *sql.DB
	Logger  log.Logger
	Tenancy viewer.Tenancy
}

// NewServer creates a server from config.
func NewServer(cfg Config) (*grpc.Server, func(), error) {
	wire.Build(
		wire.FieldsOf(new(Config), "Tenancy", "DB", "Logger"),
		newActionsRegistry,
		newServer,
	)
	return nil, nil, nil
}

func newActionsRegistry() *executor.Registry {
	registry := executor.NewRegistry()
	return registry
}
