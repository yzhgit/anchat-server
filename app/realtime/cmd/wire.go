//go:build wireinject
// +build wireinject

package main

import (
	authpkg "flamingo/pkg/auth"
	brokerpkg "flamingo/pkg/broker"
	confpkg "flamingo/pkg/config"
	grpcpkg "flamingo/pkg/grpc"
	regpkg "flamingo/pkg/registry"

	"flamingo/app/realtime/internal/handler"
	"flamingo/app/realtime/internal/server"
	"flamingo/app/realtime/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// initApp init kratos application.
func initApp(logger log.Logger, bc confpkg.Bootstrap) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		handler.ProviderSet,
		service.ProviderSet,
		confpkg.ProviderSet,
		regpkg.ProviderSet,
		grpcpkg.ProviderSet,
		brokerpkg.ProviderSet,
		authpkg.ProviderSet,
		newApp,
	))
}
