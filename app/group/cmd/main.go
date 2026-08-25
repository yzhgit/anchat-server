package main

import (
	"flag"

	apppkg "flamingo/pkg/app"
	confpkg "flamingo/pkg/config"
	constspkg "flamingo/pkg/consts"
	logpkg "flamingo/pkg/logging"
	otelpkg "flamingo/pkg/otel"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	Name    = constspkg.GroupServiceName
	Version string
)

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, rr *consul.Registry, lc apppkg.Lifecycle) *kratos.App {
	return apppkg.New(Name, Version, logger, gs, hs, rr, lc)
}

func main() {
	var flagconf string
	flag.StringVar(&flagconf, "config", "./", "config file path")
	flag.Parse()

	logger := logpkg.NewLogger(Name, Version)

	var bc confpkg.Bootstrap
	if err := confpkg.LoadConfig(flagconf, &bc); err != nil {
		panic(err)
	}

	shutdown := otelpkg.InitTracer(Name, bc.Trace.Endpoint)
	defer shutdown()

	app, cleanup, err := initApp(logger, bc)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
