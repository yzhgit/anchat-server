package md

import (
	"context"

	"flamingo/pkg/consts"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GetUserID extracts the operator user_id from incoming gRPC metadata.
// Returns the user_id and whether it was found.
func GetUserID(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	vals := md.Get(consts.OperatorUserIDMetadataKey)
	if len(vals) == 0 {
		return "", false
	}
	return vals[0], true
}

// RequireUserID extracts the operator user_id from incoming gRPC metadata.
// Returns an error if the user_id is missing.
func RequireUserID(ctx context.Context) (string, error) {
	if userID, ok := GetUserID(ctx); ok {
		return userID, nil
	}
	return "", status.Error(codes.Unauthenticated, "missing user_id metadata")
}

// MustGetUserID extracts the operator user_id from incoming gRPC metadata.
// Panics if the user_id is missing. Use only when the caller guarantees the metadata is present.
func MustGetUserID(ctx context.Context) string {
	userID, err := RequireUserID(ctx)
	if err != nil {
		panic(err)
	}
	return userID
}

// AppendUserIDToContext appends the user_id to outgoing gRPC metadata.
// Use this when a service needs to forward the operator identity to another service.
func AppendUserIDToContext(ctx context.Context, userID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, consts.OperatorUserIDMetadataKey, userID)
}
