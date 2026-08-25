//go:build wireinject
// +build wireinject

package main

import (
	brokerpkg "flamingo/pkg/broker"
	cachepkg "flamingo/pkg/cache"
	confpkg "flamingo/pkg/config"
	dbpkg "flamingo/pkg/database"
	grpcpkg "flamingo/pkg/grpc"
	regpkg "flamingo/pkg/registry"

	"flamingo/app/message/internal/handler"
	"flamingo/app/message/internal/repository"
	"flamingo/app/message/internal/server"
	"flamingo/app/message/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// initApp init kratos application.
func initApp(logger log.Logger, bc confpkg.Bootstrap, tc confpkg.Typing) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		repository.ProviderSet,
		service.ProviderSet,
		handler.ProviderSet,
		confpkg.ProviderSet,
		regpkg.ProviderSet,
		grpcpkg.ProviderSet,
		dbpkg.ProviderSet,
		brokerpkg.ProviderSet,
		cachepkg.ProviderSet,
		newApp,
	))
}
