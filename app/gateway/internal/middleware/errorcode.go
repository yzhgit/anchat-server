package middleware

import (
	"context"
	"strconv"

	"flamingo/pkg/consts"
	pkggrpc "flamingo/pkg/grpc"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// ErrorCodeMiddleware extracts the business error code from gRPC errors
// returned by the handler chain and sets it as an HTTP response header (X-Error-Code).
// This allows the gateway to propagate business error codes to HTTP clients.
func ErrorCodeMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			resp, err := handler(ctx, req)
			if err == nil {
				return resp, nil
			}

			// Extract business error code from the wrapped error
			// (set by the gRPC client interceptor's trailing metadata).
			code := pkggrpc.ExtractBusinessCode(err)
			if code == 0 {
				return resp, err
			}

			// Set the business error code as an HTTP response header.
			if rw, ok := http.ResponseWriterFromServerContext(ctx); ok {
				rw.Header().Set(consts.ErrorBusinessCodeHeaderKey, strconv.Itoa(code))
			}

			return resp, err
		}
	}
}
