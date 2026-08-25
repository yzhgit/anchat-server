//go:build wireinject
// +build wireinject

package main

import (
	brokerpkg "flamingo/pkg/broker"
	confpkg "flamingo/pkg/config"
	dbpkg "flamingo/pkg/database"
	jpushpkg "flamingo/pkg/jpush"
	regpkg "flamingo/pkg/registry"

	"flamingo/app/push/internal/handler"
	"flamingo/app/push/internal/repository"
	"flamingo/app/push/internal/server"
	"flamingo/app/push/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// initApp init kratos application.
func initApp(logger log.Logger, bc confpkg.Bootstrap, tc confpkg.JPush) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		repository.ProviderSet,
		service.ProviderSet,
		handler.ProviderSet,
		confpkg.ProviderSet,
		regpkg.ProviderSet,
		dbpkg.ProviderSet,
		brokerpkg.ProviderSet,
		jpushpkg.ProviderSet,
		newApp,
	))
}
