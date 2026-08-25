package middleware

import (
	"context"

	"flamingo/pkg/auth"
	"flamingo/pkg/consts"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"google.golang.org/grpc/metadata"
)

// UserIDMiddleware injects the user_id from JWT claims into gRPC metadata
func UserIDMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 1. Extract claims from context
			if claims, ok := jwt.FromContext(ctx); ok {
				// Assert to your custom claims type
				if customClaims, ok := claims.(*auth.Claims); ok && customClaims.UserID != "" {
					// 2. Write user_id into gRPC metadata
					ctx = metadata.AppendToOutgoingContext(ctx, consts.OperatorUserIDMetadataKey, customClaims.UserID)
				}
			}
			// 3. Continue to the next handler
			return handler(ctx, req)
		}
	}
}
