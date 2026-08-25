//go:build wireinject
// +build wireinject

package main

import (
	brokerpkg "flamingo/pkg/broker"
	confpkg "flamingo/pkg/config"
	dbpkg "flamingo/pkg/database"
	regpkg "flamingo/pkg/registry"

	"flamingo/app/conversation/internal/handler"
	"flamingo/app/conversation/internal/repository"
	"flamingo/app/conversation/internal/server"
	"flamingo/app/conversation/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// initApp init kratos application.
func initApp(logger log.Logger, bc confpkg.Bootstrap) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		repository.ProviderSet,
		service.ProviderSet,
		handler.ProviderSet,
		confpkg.ProviderSet,
		regpkg.ProviderSet,
		dbpkg.ProviderSet,
		brokerpkg.ProviderSet,
		newApp,
	))
}
