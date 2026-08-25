package server

import (
	confpkg "flamingo/pkg/config"
	"flamingo/pkg/metrics"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer creates a lightweight HTTP server that only exposes /metrics
// for Prometheus scraping. The listen address comes from server.http.addr
// in config.yaml; override per-service via the HTTP_ADDR env var.
func NewHTTPServer(c confpkg.Server) *http.Server {
	opts := []http.ServerOption{}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)
	srv.Handle("/metrics", metrics.MetricsHandler())
	return srv
}
