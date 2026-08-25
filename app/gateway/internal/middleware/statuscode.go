package middleware

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapBusinessCodeToHTTP maps a business error code to an HTTP status code.
//
// Rules:
//  1. If biz code IS a standard HTTP code (400-599), use it directly.
//     Covers 401 (Unauthorized), 403 (Forbidden), 404 (NotFound).
//  2. Otherwise, map via the gRPC status code embedded in the error.
//  3. Default to 500 (Internal Server Error).
func mapBusinessCodeToHTTP(bizCode int, err error) int {
	if bizCode >= 400 && bizCode <= 599 {
		return bizCode
	}

	var grpcCode codes.Code
	if st, ok := status.FromError(err); ok {
		grpcCode = st.Code()
	}

	switch grpcCode {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
