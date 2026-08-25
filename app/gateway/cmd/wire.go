//go:build wireinject
// +build wireinject

package main

import (
	confpkg "flamingo/pkg/config"
	grpcpkg "flamingo/pkg/grpc"
	regpkg "flamingo/pkg/registry"

	"flamingo/app/gateway/internal/server"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// initApp initializes the gateway Kratos application via Wire.
func initApp(logger log.Logger, bc confpkg.Bootstrap) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		confpkg.ProviderSet,
		regpkg.ProviderSet,
		grpcpkg.ProviderSet,
		newApp,
	))
}
