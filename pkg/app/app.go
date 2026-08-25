// Package app provides the shared kratos application constructor and lifecycle
// plumbing that every service wires into its newApp.  It intentionally stays
// framework-level: no service-specific business logic lives here.
package app

import (
	"context"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// Lifecycle bundles the BeforeStart / AfterStop hooks that a service wants to
// run around app startup and shutdown.  The zero value is a no-op lifecycle
// and is what services without startup/shutdown work inject.
type Lifecycle struct {
	Start func(context.Context) error
	Stop  func(context.Context) error
}

// DefaultLifecycle returns the no-op lifecycle.  It is intended as a Wire
// provider so that services without startup/shutdown work satisfy newApp's
// Lifecycle parameter without having their own provider.
func DefaultLifecycle() Lifecycle { return Lifecycle{} }

// New constructs the kratos App for a service.  Nil gs or hs is silently
// omitted, allowing HTTP-only, gRPC-only, or no-transport services to use the
// same newApp signature as the common case (both present).
func New(name, version string, logger log.Logger, gs *grpc.Server, hs *http.Server, rr registry.Registrar, lc Lifecycle) *kratos.App {
	opts := []kratos.Option{
		kratos.Name(name),
		kratos.Version(version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
	}
	var servers []transport.Server
	if gs != nil {
		servers = append(servers, gs)
	}
	if hs != nil {
		servers = append(servers, hs)
	}
	if len(servers) > 0 {
		opts = append(opts, kratos.Server(servers...))
	}
	opts = append(opts, kratos.Registrar(rr))
	if lc.Start != nil {
		opts = append(opts, kratos.BeforeStart(lc.Start))
	}
	if lc.Stop != nil {
		opts = append(opts, kratos.AfterStop(lc.Stop))
	}
	return kratos.New(opts...)
}
