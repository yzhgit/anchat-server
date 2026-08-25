package logging

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
)

// NewLogger builds a kratos logger wrapping the standard stdout logger with
// the fields every service logs (timestamp, service name/version, trace/span
// id, caller). It mirrors the logger init block used in each service's main.
func NewLogger(serviceName, version string) log.Logger {
	return log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"service.name", serviceName,
		"service.version", version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
		"caller", log.DefaultCaller,
	)
}
