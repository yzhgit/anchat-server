package grpc

import (
	"context"
	"strconv"

	"flamingo/pkg/consts"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// errorCodeCarrier embeds a business error code into a gRPC status error.
// It implements GRPCStatus() so that status.FromError continues to work,
// while also exposing the business code for downstream consumers.
type errorCodeCarrier struct {
	businessCode int
	err          error
}

func (e *errorCodeCarrier) Error() string {
	return e.err.Error()
}

func (e *errorCodeCarrier) Unwrap() error {
	return e.err
}

// GRPCStatus ensures status.FromError returns the original gRPC status.
func (e *errorCodeCarrier) GRPCStatus() *status.Status {
	st, _ := status.FromError(e.err)
	return st
}

// BusinessCode returns the business error code, or 0 if none.
func (e *errorCodeCarrier) BusinessCode() int {
	return e.businessCode
}

// ErrorCodeClient is a gRPC unary client interceptor that captures the
// business error code from trailing metadata (x-error-code) and wraps
// the error so downstream code can extract it via ExtractBusinessCode.
func ErrorCodeClient() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var trailer metadata.MD
		trailerOpt := grpc.Trailer(&trailer)
		err := invoker(ctx, method, req, reply, cc, append(opts, trailerOpt)...)
		if err == nil {
			return nil
		}

		// Only process gRPC status errors.
		if _, ok := status.FromError(err); !ok {
			return err
		}

		// Extract business error code from trailing metadata.
		codes := trailer.Get(consts.ErrorBusinessCodeMetadataKey)
		if len(codes) == 0 {
			return err
		}

		bizCode, _ := strconv.Atoi(codes[0])
		if bizCode == 0 {
			return err
		}

		return &errorCodeCarrier{businessCode: bizCode, err: err}
	}
}

// ExtractBusinessCode extracts the business error code from an error,
// returning 0 if none is found.
func ExtractBusinessCode(err error) int {
	for e := err; e != nil; e = unwrap(e) {
		if c, ok := e.(*errorCodeCarrier); ok {
			return c.BusinessCode()
		}
	}
	return 0
}

func unwrap(err error) error {
	type unwrapper interface{ Unwrap() error }
	if u, ok := err.(unwrapper); ok {
		return u.Unwrap()
	}
	return nil
}
