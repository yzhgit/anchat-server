package grpc

import (
	"context"

	callingpb "github.com/anychat/server/api/proto/calling"
	commonpb "github.com/anychat/server/api/proto/common"
	"github.com/anychat/server/internal/calling/model"
	"github.com/anychat/server/internal/calling/service"
	"github.com/anychat/server/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the Calling gRPC server (implements CallingService interface)
type Server struct {
	callingpb.UnimplementedCallingServiceServer
	svc service.CallingService
}

// NewServer creates a gRPC server
func NewServer(service service.CallingService) *Server {
	return &Server{svc: service}
}

func (s *Server) InitiateCall(ctx context.Context, req *callingpb.InitiateCallRequest) (*callingpb.InitiateCallResponse, error) {
	if req.CallerId == "" || req.CalleeId == "" {
		return nil, status.Error(codes.InvalidArgument, "caller_id and callee_id are required")
	}
	if req.CallerId == req.CalleeId {
		return nil, status.Error(codes.InvalidArgument, "caller and callee cannot be the same")
	}
	callType := model.CallType(req.CallType)
	resp, err := s.svc.InitiateCall(ctx, req.CallerId, req.CalleeId, callType)
	if err != nil {
		logger.Error("InitiateCall gRPC failed", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (s *Server) JoinCall(ctx context.Context, req *callingpb.JoinCallRequest) (*callingpb.JoinCallResponse, error) {
	if req.CallId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "call_id and user_id are required")
	}
	return s.svc.JoinCall(ctx, req.CallId, req.UserId)
}

func (s *Server) RejectCall(ctx context.Context, req *callingpb.RejectCallRequest) (*commonpb.Empty, error) {
	if req.CallId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "call_id and user_id are required")
	}
	if err := s.svc.RejectCall(ctx, req.CallId, req.UserId); err != nil {
		return nil, err
	}
	return &commonpb.Empty{}, nil
}

func (s *Server) EndCall(ctx context.Context, req *callingpb.EndCallRequest) (*commonpb.Empty, error) {
	if req.CallId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "call_id and user_id are required")
	}
	if err := s.svc.EndCall(ctx, req.CallId, req.UserId); err != nil {
		return nil, err
	}
	return &commonpb.Empty{}, nil
}

func (s *Server) GetCallSession(ctx context.Context, req *callingpb.GetCallSessionRequest) (*callingpb.CallSession, error) {
	if req.CallId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "call_id and user_id are required")
	}
	return s.svc.GetCallSession(ctx, req.CallId, req.UserId)
}

func (s *Server) ListCallLogs(ctx context.Context, req *callingpb.ListCallLogsRequest) (*callingpb.ListCallLogsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	return s.svc.ListCallLogs(ctx, req.UserId, int(req.Page), int(req.PageSize))
}

func (s *Server) CreateMeeting(ctx context.Context, req *callingpb.CreateMeetingRequest) (*callingpb.CreateMeetingResponse, error) {
	if req.CreatorId == "" || req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "creator_id and title are required")
	}
	return s.svc.CreateMeeting(ctx, req.CreatorId, req.Title, req.Password, int(req.MaxParticipants))
}

func (s *Server) JoinMeeting(ctx context.Context, req *callingpb.JoinMeetingRequest) (*callingpb.JoinMeetingResponse, error) {
	if req.UserId == "" || req.RoomId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and room_id are required")
	}
	return s.svc.JoinMeeting(ctx, req.UserId, req.RoomId, req.Password)
}

func (s *Server) EndMeeting(ctx context.Context, req *callingpb.EndMeetingRequest) (*commonpb.Empty, error) {
	if req.RoomId == "" || req.CreatorId == "" {
		return nil, status.Error(codes.InvalidArgument, "room_id and creator_id are required")
	}
	if err := s.svc.EndMeeting(ctx, req.RoomId, req.CreatorId); err != nil {
		return nil, err
	}
	return &commonpb.Empty{}, nil
}

func (s *Server) GetMeeting(ctx context.Context, req *callingpb.GetMeetingRequest) (*callingpb.MeetingRoom, error) {
	if req.RoomId == "" {
		return nil, status.Error(codes.InvalidArgument, "room_id is required")
	}
	return s.svc.GetMeeting(ctx, req.RoomId)
}

func (s *Server) ListMeetings(ctx context.Context, req *callingpb.ListMeetingsRequest) (*callingpb.ListMeetingsResponse, error) {
	return s.svc.ListMeetings(ctx, int(req.Page), int(req.PageSize))
}
