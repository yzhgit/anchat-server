//go:build wireinject
// +build wireinject

package main

import (
	brokerpkg "flamingo/pkg/broker"
	confpkg "flamingo/pkg/config"
	dbpkg "flamingo/pkg/database"
	grpcpkg "flamingo/pkg/grpc"
	regpkg "flamingo/pkg/registry"

	"flamingo/app/rtc/internal/handler"
	"flamingo/app/rtc/internal/repository"
	"flamingo/app/rtc/internal/server"
	"flamingo/app/rtc/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// initApp init kratos application.
func initApp(logger log.Logger, bc confpkg.Bootstrap, lc confpkg.Livekit) (*kratos.App, func(), error) {
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
		newApp,
	))
}
