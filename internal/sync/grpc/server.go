package grpc

import (
	"context"

	syncpb "github.com/anychat/server/api/proto/sync"
	"github.com/anychat/server/internal/sync/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server Sync gRPC server
type Server struct {
	syncpb.UnimplementedSyncServiceServer
	svc service.SyncService
}

// NewServer creates gRPC server
func NewServer(service service.SyncService) *Server {
	return &Server{svc: service}
}

// Sync full/incremental sync
func (s *Server) Sync(ctx context.Context, req *syncpb.SyncRequest) (*syncpb.SyncResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	resp, err := s.svc.Sync(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

// SyncMessages message backfill
func (s *Server) SyncMessages(ctx context.Context, req *syncpb.SyncMessagesRequest) (*syncpb.SyncMessagesResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	resp, err := s.svc.SyncMessages(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}
