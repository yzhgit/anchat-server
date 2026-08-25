package server

import (
	"flamingo/app/gateway/internal/handler"
	"flamingo/pkg/app"

	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for the gateway HTTP server.
var ProviderSet = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
	handler.ProviderSet,
	app.DefaultLifecycle,
)
