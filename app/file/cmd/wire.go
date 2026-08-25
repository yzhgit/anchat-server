//go:build wireinject
// +build wireinject

package main

import (
	confpkg "flamingo/pkg/config"
	dbpkg "flamingo/pkg/database"
	oss "flamingo/pkg/oss"
	regpkg "flamingo/pkg/registry"

	"flamingo/app/file/internal/handler"
	"flamingo/app/file/internal/repository"
	"flamingo/app/file/internal/server"
	"flamingo/app/file/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// initApp init kratos application.
func initApp(logger log.Logger, bc confpkg.Bootstrap, mc confpkg.Minio) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		repository.ProviderSet,
		service.ProviderSet,
		handler.ProviderSet,
		confpkg.ProviderSet,
		regpkg.ProviderSet,
		dbpkg.ProviderSet,
		oss.ProviderSet,
		newApp,
	))
}
