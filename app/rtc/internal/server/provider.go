package server

import (
	"flamingo/pkg/app"

	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, app.DefaultLifecycle)
