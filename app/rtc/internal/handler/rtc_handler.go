package handler

import (
	"context"

	rtcv1 "flamingo/api/rtc/v1"

	"flamingo/app/rtc/internal/model"
	"flamingo/pkg/errors"
	"flamingo/pkg/md"

	"github.com/go-kratos/kratos/v2/log"
	empty "github.com/golang/protobuf/ptypes/empty"
)

// RtcService is the real-time communication service interface
type RtcService interface {
	InitiateCall(ctx context.Context, callerID, calleeID string, callType model.CallType) (*rtcv1.CallSession, error)
	JoinCall(ctx context.Context, callID, userID string) (*rtcv1.CallSession, error)
	RejectCall(ctx context.Context, callID, userID string) error
	EndCall(ctx context.Context, callID, userID string) error
	GetCallSession(ctx context.Context, callID, userID string) (*rtcv1.CallSession, error)
	ListCallLogs(ctx context.Context, userID string, page, pageSize int) (*rtcv1.ListCallLogsResponse, error)
	CreateMeeting(ctx context.Context, creatorID, title, password string, maxParticipants int) (*rtcv1.Meeting, error)
	JoinMeeting(ctx context.Context, userID, roomID, password string) (*rtcv1.Meeting, error)
	EndMeeting(ctx context.Context, roomID, creatorID string) error
	GetMeeting(ctx context.Context, roomID string) (*rtcv1.Meeting, error)
	ListMeetings(ctx context.Context, page, pageSize int) (*rtcv1.ListMeetingsResponse, error)
}

type RtcHandler struct {
	rtcv1.UnimplementedRtcServiceServer
	svc RtcService
	log *log.Helper
}

// NewRtcHandler creates a gRPC handler
func NewRtcHandler(svc RtcService, logger log.Logger) *RtcHandler {
	return &RtcHandler{
		svc: svc,
		log: log.NewHelper(logger),
	}
}

func (s *RtcHandler) InitiateCall(ctx context.Context, req *rtcv1.InitiateCallRequest) (*rtcv1.CallSession, error) {
	callerID := md.MustGetUserID(ctx)
	if callerID == "" || req.CalleeId == "" {
		return nil, errors.BadRequest(ctx, "caller_id and callee_id are required")
	}
	if callerID == req.CalleeId {
		return nil, errors.BadRequest(ctx, "caller and callee cannot be the same")
	}
	callType := model.CallType(req.CallType)
	resp, err := s.svc.InitiateCall(ctx, callerID, req.CalleeId, callType)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *RtcHandler) JoinCall(ctx context.Context, req *rtcv1.JoinCallRequest) (*rtcv1.CallSession, error) {
	userID := md.MustGetUserID(ctx)
	if req.CallId == "" || userID == "" {
		return nil, errors.BadRequest(ctx, "call_id and user_id are required")
	}
	resp, err := s.svc.JoinCall(ctx, req.CallId, userID)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *RtcHandler) RejectCall(ctx context.Context, req *rtcv1.RejectCallRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.CallId == "" || userID == "" {
		return nil, errors.BadRequest(ctx, "call_id and user_id are required")
	}
	if err := s.svc.RejectCall(ctx, req.CallId, userID); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *RtcHandler) EndCall(ctx context.Context, req *rtcv1.EndCallRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.CallId == "" || userID == "" {
		return nil, errors.BadRequest(ctx, "call_id and user_id are required")
	}
	if err := s.svc.EndCall(ctx, req.CallId, userID); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *RtcHandler) GetCallSession(ctx context.Context, req *rtcv1.GetCallSessionRequest) (*rtcv1.CallSession, error) {
	userID := md.MustGetUserID(ctx)
	if req.CallId == "" || userID == "" {
		return nil, errors.BadRequest(ctx, "call_id and user_id are required")
	}
	resp, err := s.svc.GetCallSession(ctx, req.CallId, userID)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *RtcHandler) ListCallLogs(ctx context.Context, req *rtcv1.ListCallLogsRequest) (*rtcv1.ListCallLogsResponse, error) {
	userID := md.MustGetUserID(ctx)
	if userID == "" {
		return nil, errors.BadRequest(ctx, "user_id is required")
	}
	resp, err := s.svc.ListCallLogs(ctx, userID, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *RtcHandler) CreateMeeting(ctx context.Context, req *rtcv1.CreateMeetingRequest) (*rtcv1.Meeting, error) {
	creatorID := md.MustGetUserID(ctx)
	if creatorID == "" || req.Title == "" {
		return nil, errors.BadRequest(ctx, "creator_id and title are required")
	}
	password := ""
	if req.Password != nil {
		password = *req.Password
	}
	resp, err := s.svc.CreateMeeting(ctx, creatorID, req.Title, password, int(req.MaxParticipants))
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *RtcHandler) JoinMeeting(ctx context.Context, req *rtcv1.JoinMeetingRequest) (*rtcv1.Meeting, error) {
	userID := md.MustGetUserID(ctx)
	if userID == "" || req.RoomId == "" {
		return nil, errors.BadRequest(ctx, "user_id and room_id are required")
	}
	password := ""
	if req.Password != nil {
		password = *req.Password
	}
	resp, err := s.svc.JoinMeeting(ctx, userID, req.RoomId, password)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *RtcHandler) EndMeeting(ctx context.Context, req *rtcv1.EndMeetingRequest) (*empty.Empty, error) {
	creatorID := md.MustGetUserID(ctx)
	if req.RoomId == "" || creatorID == "" {
		return nil, errors.BadRequest(ctx, "room_id and creator_id are required")
	}
	if err := s.svc.EndMeeting(ctx, req.RoomId, creatorID); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *RtcHandler) GetMeeting(ctx context.Context, req *rtcv1.GetMeetingRequest) (*rtcv1.Meeting, error) {
	if req.RoomId == "" {
		return nil, errors.BadRequest(ctx, "room_id is required")
	}
	resp, err := s.svc.GetMeeting(ctx, req.RoomId)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *RtcHandler) ListMeetings(ctx context.Context, req *rtcv1.ListMeetingsRequest) (*rtcv1.ListMeetingsResponse, error) {
	resp, err := s.svc.ListMeetings(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}
