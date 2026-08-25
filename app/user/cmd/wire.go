//go:build wireinject
// +build wireinject

package main

import (
	authpkg "flamingo/pkg/auth"
	brokerpkg "flamingo/pkg/broker"
	cachepkg "flamingo/pkg/cache"
	confpkg "flamingo/pkg/config"
	dbpkg "flamingo/pkg/database"
	grpcpkg "flamingo/pkg/grpc"
	regpkg "flamingo/pkg/registry"
	senderpkg "flamingo/pkg/sender"

	"flamingo/app/user/internal/handler"
	"flamingo/app/user/internal/repository"
	"flamingo/app/user/internal/server"
	"flamingo/app/user/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// initApp init kratos application.
func initApp(logger log.Logger, bc confpkg.Bootstrap, vc confpkg.Verify) (*kratos.App, func(), error) {
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
		authpkg.ProviderSet,
		senderpkg.ProviderSet,
		newApp,
	))
}
