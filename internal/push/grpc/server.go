package grpc

import (
	"context"

	pushpb "github.com/anychat/server/api/proto/push"
	"github.com/anychat/server/internal/push/model"
	"github.com/anychat/server/internal/push/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server Push gRPC server
type Server struct {
	pushpb.UnimplementedPushServiceServer
	svc service.PushService
}

// NewServer creates gRPC server
func NewServer(service service.PushService) *Server {
	return &Server{svc: service}
}

// SendPush sends push notification to specified user list
func (s *Server) SendPush(ctx context.Context, req *pushpb.SendPushRequest) (*pushpb.SendPushResponse, error) {
	if len(req.UserIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_ids is required")
	}
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	successCount, failureCount, msgID, err := s.svc.SendPush(
		ctx,
		req.UserIds,
		req.Title,
		req.Content,
		model.PushType(req.PushType),
		req.Extras,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pushpb.SendPushResponse{
		SuccessCount: int32(successCount),
		FailureCount: int32(failureCount),
		MsgId:        msgID,
	}, nil
}
