// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build wireinject

package graphhttp

import (
	"net/http"

	"github.com/marosmars/resourceManager/actions/executor"
	"github.com/marosmars/resourceManager/log"
	"github.com/marosmars/resourceManager/mysql"
	"github.com/marosmars/resourceManager/server"
	"github.com/marosmars/resourceManager/server/xserver"
	"github.com/marosmars/resourceManager/telemetry"
	"github.com/marosmars/resourceManager/viewer"
	"go.opencensus.io/stats/view"

	"github.com/google/wire"
	"github.com/gorilla/mux"
	"gocloud.dev/server/health"
)

// Config defines the http server config.
type Config struct {
	Tenancy      viewer.Tenancy
	Logger       log.Logger
	Telemetry    *telemetry.Config
	HealthChecks []health.Checker
}

// NewServer creates a server from config.
func NewServer(cfg Config) (*server.Server, func(), error) {
	wire.Build(
		xserver.ServiceSet,
		provideViews,
		wire.FieldsOf(new(Config), "Logger", "Telemetry", "HealthChecks"),
		newRouterConfig,
		newRouter,
		wire.Bind(new(http.Handler), new(*mux.Router)),
	)
	return nil, nil, nil
}

func newRouterConfig(config Config) (cfg routerConfig, err error) {
	registry := executor.NewRegistry()
	cfg = routerConfig{logger: config.Logger}
	cfg.viewer.tenancy = config.Tenancy
	cfg.actions.registry = registry
	return cfg, nil
}

func provideViews() []*view.View {
	views := xserver.DefaultViews()
	views = append(views, mysql.DefaultViews...)
	return views
}
