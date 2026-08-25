package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	defaultRegistry = prometheus.NewRegistry()

	requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "anychat",
			Subsystem: "requests",
			Name:      "total",
			Help:      "Total number of requests",
		},
		[]string{"transport", "operation", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "anychat",
			Subsystem: "requests",
			Name:      "duration_seconds",
			Help:      "Request duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"transport", "operation"},
	)
)

func init() {
	defaultRegistry.MustRegister(requestTotal)
	defaultRegistry.MustRegister(requestDuration)
}

// DefaultRegistry returns the shared Prometheus registry used by this package.
func DefaultRegistry() prometheus.Gatherer {
	return defaultRegistry
}

// MetricsHandler returns an http.Handler that serves /metrics for Prometheus scraping.
func MetricsHandler() http.Handler {
	return promhttp.HandlerFor(defaultRegistry, promhttp.HandlerOpts{})
}

// Middleware returns a kratos middleware that records request counters and latency
// histograms. Labels: transport (grpc/http), operation (from transport context),
// status ("ok" or the kratos error code as a string).
func Middleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				kind  string
				op    string
				tr    transport.Transporter
				found bool
			)

			start := time.Now()

			if tr, found = transport.FromServerContext(ctx); found {
				kind = string(tr.Kind())
				op = tr.Operation()
			}

			reply, err := handler(ctx, req)

			status := "ok"
			if err != nil {
				code := errors.Code(err)
				if code != 0 {
					status = fmt.Sprintf("%d", code)
				} else {
					status = "error"
				}
			}

			requestTotal.WithLabelValues(kind, op, status).Inc()
			requestDuration.WithLabelValues(kind, op).Observe(float64(time.Since(start)) / 1e9)

			return reply, err
		}
	}
}
