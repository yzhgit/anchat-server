package service

import (
	"context"
	"time"

	friendv1 "flamingo/api/friend/v1"
	rtcv1 "flamingo/api/rtc/v1"

	"flamingo/app/rtc/internal/handler"
	"flamingo/app/rtc/internal/model"

	"flamingo/pkg/broker"
	confpkg "flamingo/pkg/config"
	"flamingo/pkg/crypto"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	livekitsdk "github.com/livekit/server-sdk-go/v2"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const tokenTTL = 2 * time.Hour

type rtcServiceImpl struct {
	conf         confpkg.Livekit
	roomService  *livekitsdk.RoomServiceClient
	friendClient friendv1.FriendServiceClient
	callRepo     CallRepository
	meetingRepo  MeetingRepository
	broker       broker.Broker
	log          *log.Helper
}

var _ handler.RtcService = (*rtcServiceImpl)(nil)

// NewRtcService creates an audio/video service
func NewRtcService(
	conf confpkg.Livekit,
	friendClient friendv1.FriendServiceClient,
	callRepo CallRepository,
	meetingRepo MeetingRepository,
	broker broker.Broker,
	logger log.Logger,
) handler.RtcService {
	roomService := livekitsdk.NewRoomServiceClient(conf.URL, conf.APIKey, conf.APISecret)
	return &rtcServiceImpl{
		conf:         conf,
		roomService:  roomService,
		friendClient: friendClient,
		callRepo:     callRepo,
		meetingRepo:  meetingRepo,
		broker:       broker,
		log:          log.NewHelper(logger),
	}
}

// ── Call Related ──────────────────────────────────────────────

func (s *rtcServiceImpl) InitiateCall(ctx context.Context, callerID, calleeID string, callType model.CallType) (*rtcv1.CallSession, error) {
	l := s.log.WithContext(ctx)
	if s.friendClient == nil {
		return nil, status.Error(codes.Internal, "friend client is not initialized")
	}

	blockedResp, err := s.friendClient.IsBlocked(ctx, &friendv1.IsBlockedRequest{
		UserId:       callerID,
		TargetUserId: calleeID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify blacklist: %v", err)
	}
	if blockedResp.IsBlocked {
		return nil, status.Error(codes.PermissionDenied, "user blocked")
	}

	callID := uuid.NewString()
	roomName := "call_" + callID

	// Create LiveKit Room (set EmptyTimeout to 5 minutes, waiting for callee to answer)
	emptyTimeout := uint32(300)
	_, err = s.roomService.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:         roomName,
		EmptyTimeout: emptyTimeout,
	})
	if err != nil {
		l.Errorw("msg", "InitiateCall: create livekit room failed", "error", err)
		return nil, status.Errorf(codes.Internal, "create room: %v", err)
	}

	// Generate caller token
	token, err := s.generateToken(roomName, callerID, true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate token: %v", err)
	}

	// Persist call session
	session := &model.CallSession{
		CallID:   callID,
		CallerID: callerID,
		CalleeID: calleeID,
		CallType: callType,
		Status:   model.CallStatusRinging,
		RoomName: roomName,
	}
	if err := s.callRepo.CreateCallSession(ctx, session); err != nil {
		l.Errorw("msg", "InitiateCall: save session failed", "error", err)
		return nil, status.Errorf(codes.Internal, "save session: %v", err)
	}

	// Notify callee
	notif := broker.NewNotification(broker.TypeLiveKitCallInvite, callerID, broker.PriorityHigh).
		AddPayloadField("call_id", callID).
		AddPayloadField("caller_id", callerID).
		AddPayloadField("call_type", int32(callType))
	if err := s.broker.PublishToUser(calleeID, notif); err != nil {
		l.Warnw("msg", "InitiateCall: notify callee failed", "error", err)
	}

	return &rtcv1.CallSession{
		CallId:   callID,
		RoomName: roomName,
		Token:    token,
	}, nil
}

func (s *rtcServiceImpl) JoinCall(ctx context.Context, callID, userID string) (*rtcv1.CallSession, error) {
	l := s.log.WithContext(ctx)
	session, err := s.callRepo.GetCallSession(ctx, callID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "call session not found")
	}
	if session.Status != model.CallStatusRinging {
		return nil, status.Errorf(codes.FailedPrecondition, "call is not in ringing state: %d", session.Status)
	}
	if session.CalleeID != userID {
		return nil, status.Error(codes.PermissionDenied, "not the callee of this call")
	}

	// Generate callee token
	token, err := s.generateToken(session.RoomName, userID, false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate token: %v", err)
	}

	// Update session status
	now := time.Now()
	session.Status = model.CallStatusConnected
	session.ConnectedAt = &now
	if err := s.callRepo.UpdateCallSession(ctx, session); err != nil {
		l.Warnw("msg", "JoinCall: update session failed", "error", err)
	}

	// Notify caller
	notif := broker.NewNotification(broker.TypeLiveKitCallStatus, userID, broker.PriorityHigh).
		AddPayloadField("call_id", callID).
		AddPayloadField("status", int32(model.CallStatusConnected))
	if err := s.broker.PublishToUser(session.CallerID, notif); err != nil {
		l.Warnw("msg", "JoinCall: notify caller failed", "error", err)
	}

	return &rtcv1.CallSession{
		CallId:   callID,
		RoomName: session.RoomName,
		Token:    token,
	}, nil
}

func (s *rtcServiceImpl) RejectCall(ctx context.Context, callID, userID string) error {
	l := s.log.WithContext(ctx)
	session, err := s.callRepo.GetCallSession(ctx, callID)
	if err != nil {
		return status.Error(codes.NotFound, "call session not found")
	}
	if session.Status != model.CallStatusRinging {
		return status.Errorf(codes.FailedPrecondition, "call is not in ringing state: %d", session.Status)
	}
	if session.CalleeID != userID {
		return status.Error(codes.PermissionDenied, "not the callee of this call")
	}

	now := time.Now()
	session.Status = model.CallStatusRejected
	session.EndedAt = &now
	if err := s.callRepo.UpdateCallSession(ctx, session); err != nil {
		l.Warnw("msg", "RejectCall: update session failed", "error", err)
	}

	// Delete Room (no need to wait)
	go s.deleteRoom(session.RoomName)

	// Notify caller
	notif := broker.NewNotification(broker.TypeLiveKitCallRejected, userID, broker.PriorityHigh).
		AddPayloadField("call_id", callID).
		AddPayloadField("callee_id", userID)
	if err := s.broker.PublishToUser(session.CallerID, notif); err != nil {
		l.Warnw("msg", "RejectCall: notify caller failed", "error", err)
	}
	return nil
}

func (s *rtcServiceImpl) EndCall(ctx context.Context, callID, userID string) error {
	l := s.log.WithContext(ctx)
	session, err := s.callRepo.GetCallSession(ctx, callID)
	if err != nil {
		return status.Error(codes.NotFound, "call session not found")
	}
	if session.CallerID != userID && session.CalleeID != userID {
		return status.Error(codes.PermissionDenied, "not a participant of this call")
	}
	if session.Status == model.CallStatusEnded || session.Status == model.CallStatusRejected {
		return nil
	}

	now := time.Now()
	newStatus := model.CallStatusEnded
	if session.Status == model.CallStatusRinging {
		newStatus = model.CallStatusCancelled
	}
	session.Status = newStatus
	session.EndedAt = &now
	if session.ConnectedAt != nil {
		session.Duration = int(now.Sub(*session.ConnectedAt).Seconds())
	}
	if err := s.callRepo.UpdateCallSession(ctx, session); err != nil {
		l.Warnw("msg", "EndCall: update session failed", "error", err)
	}

	go s.deleteRoom(session.RoomName)

	// Notify peer
	targetID := session.CallerID
	if userID == session.CallerID {
		targetID = session.CalleeID
	}
	notif := broker.NewNotification(broker.TypeLiveKitCallStatus, userID, broker.PriorityHigh).
		AddPayloadField("call_id", callID).
		AddPayloadField("status", int32(newStatus))
	if err := s.broker.PublishToUser(targetID, notif); err != nil {
		l.Warnw("msg", "EndCall: notify peer failed", "error", err)
	}
	return nil
}

func (s *rtcServiceImpl) GetCallSession(ctx context.Context, callID, userID string) (*rtcv1.CallSession, error) {
	session, err := s.callRepo.GetCallSession(ctx, callID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "call session not found")
	}
	if session.CallerID != userID && session.CalleeID != userID {
		return nil, status.Error(codes.PermissionDenied, "not a participant of this call")
	}
	return toProtoCallSession(session), nil
}

func (s *rtcServiceImpl) ListCallLogs(ctx context.Context, userID string, page, pageSize int) (*rtcv1.ListCallLogsResponse, error) {
	sessions, total, err := s.callRepo.ListCallLogs(ctx, userID, page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list call logs: %v", err)
	}
	pbLogs := make([]*rtcv1.CallLog, len(sessions))
	for i, s := range sessions {
		pbLogs[i] = toProtoCallLog(s)
	}
	return &rtcv1.ListCallLogsResponse{CallLogs: pbLogs, Total: total}, nil
}

// ── Meeting room related ────────────────────────────────────────────

func (s *rtcServiceImpl) CreateMeeting(ctx context.Context, creatorID, title, password string, maxParticipants int) (*rtcv1.Meeting, error) {
	l := s.log.WithContext(ctx)
	roomID := uuid.NewString()
	roomName := "meeting_" + roomID

	_, err := s.roomService.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		MaxParticipants: uint32(maxParticipants),
	})
	if err != nil {
		l.Errorw("msg", "CreateMeeting: create livekit room failed", "error", err)
		return nil, status.Errorf(codes.Internal, "create room: %v", err)
	}

	var passwordHash string
	if password != "" {
		pwHash, err := crypto.HashPassword(password)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "hash meeting password: %v", err)
		}
		passwordHash = pwHash
	}

	meeting := &model.MeetingRoom{
		RoomID:          roomID,
		CreatorID:       creatorID,
		Title:           title,
		RoomName:        roomName,
		PasswordHash:    passwordHash,
		MaxParticipants: maxParticipants,
		Status:          model.MeetingStatusActive,
	}
	if err := s.meetingRepo.CreateMeeting(ctx, meeting); err != nil {
		return nil, status.Errorf(codes.Internal, "save meeting: %v", err)
	}

	// Creator has RoomAdmin permission
	token, err := s.generateToken(roomName, creatorID, true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate token: %v", err)
	}

	return toProtoMeeting(meeting, token), nil
}

func (s *rtcServiceImpl) JoinMeeting(ctx context.Context, userID, roomID, password string) (*rtcv1.Meeting, error) {
	meeting, err := s.meetingRepo.GetMeetingByRoomID(ctx, roomID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "meeting not found")
	}
	if meeting.Status != model.MeetingStatusActive {
		return nil, status.Error(codes.FailedPrecondition, "meeting has ended")
	}
	if meeting.PasswordHash != "" && !crypto.CheckPassword(password, meeting.PasswordHash) {
		return nil, status.Error(codes.PermissionDenied, "incorrect password")
	}

	isAdmin := meeting.CreatorID == userID
	token, err := s.generateToken(meeting.RoomName, userID, isAdmin)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate token: %v", err)
	}
	return toProtoMeeting(meeting, token), nil
}

func (s *rtcServiceImpl) EndMeeting(ctx context.Context, roomID, creatorID string) error {
	l := s.log.WithContext(ctx)
	meeting, err := s.meetingRepo.GetMeetingByRoomID(ctx, roomID)
	if err != nil {
		return status.Error(codes.NotFound, "meeting not found")
	}
	if meeting.CreatorID != creatorID {
		return status.Error(codes.PermissionDenied, "only creator can end the meeting")
	}
	if meeting.Status == model.MeetingStatusEnded {
		return nil
	}

	now := time.Now()
	meeting.Status = model.MeetingStatusEnded
	meeting.EndedAt = &now
	if err := s.meetingRepo.UpdateMeeting(ctx, meeting); err != nil {
		l.Warnw("msg", "EndMeeting: update meeting failed", "error", err)
	}

	go s.deleteRoom(meeting.RoomName)
	return nil
}

func (s *rtcServiceImpl) GetMeeting(ctx context.Context, roomID string) (*rtcv1.Meeting, error) {
	meeting, err := s.meetingRepo.GetMeetingByRoomID(ctx, roomID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "meeting not found")
	}
	return toProtoMeeting(meeting, ""), nil
}

func (s *rtcServiceImpl) ListMeetings(ctx context.Context, page, pageSize int) (*rtcv1.ListMeetingsResponse, error) {
	meetings, total, err := s.meetingRepo.ListActiveMeetings(ctx, page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list meetings: %v", err)
	}
	pbMeetings := make([]*rtcv1.Meeting, len(meetings))
	for i, m := range meetings {
		pbMeetings[i] = toProtoMeeting(m, "")
	}
	return &rtcv1.ListMeetingsResponse{Meetings: pbMeetings, Total: total}, nil
}

// ── Internal helpers ──────────────────────────────────────────────

// generateToken generates LiveKit JWT
// isAdmin=true grants RoomAdmin permission (meeting creator/call initiator)
func (s *rtcServiceImpl) generateToken(roomName, identity string, isAdmin bool) (string, error) {
	at := auth.NewAccessToken(s.conf.APIKey, s.conf.APISecret)
	grant := &auth.VideoGrant{
		RoomJoin:  true,
		Room:      roomName,
		RoomAdmin: isAdmin,
	}
	at.SetVideoGrant(grant).
		SetIdentity(identity).
		SetValidFor(tokenTTL)
	return at.ToJWT()
}

func (s *rtcServiceImpl) deleteRoom(roomName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s.roomService.DeleteRoom(ctx, &livekit.DeleteRoomRequest{Room: roomName}); err != nil {
		s.log.Warnw("msg", "delete room failed", "room", roomName, "error", err)
	}
}

// ── Proto conversion ────────────────────────────────────────────

func toProtoCallSession(s *model.CallSession) *rtcv1.CallSession {
	pb := &rtcv1.CallSession{
		CallId:   s.CallID,
		CallerId: s.CallerID,
		CalleeId: s.CalleeID,
		RoomName: s.RoomName,
	}

	pb.CallType = rtcv1.CallType(s.CallType)
	pb.Status = rtcv1.CallStatus(s.Status)

	if !s.StartedAt.IsZero() {
		pb.StartedAt = timestamppb.New(s.StartedAt)
	}
	if s.EndedAt != nil {
		pb.EndedAt = timestamppb.New(*s.EndedAt)
	}
	pb.Duration = int64(s.Duration)
	return pb
}

func toProtoCallLog(s *model.CallSession) *rtcv1.CallLog {
	return &rtcv1.CallLog{
		CallId:   s.CallID,
		TargetId: s.CalleeID,
		CallType: rtcv1.CallType(s.CallType),
		Status:   rtcv1.CallStatus(s.Status),
		Duration: int64(s.Duration),
	}
}

func toProtoMeeting(m *model.MeetingRoom, token string) *rtcv1.Meeting {
	pb := &rtcv1.Meeting{
		RoomId:              m.RoomID,
		Title:               m.Title,
		CreatorId:           m.CreatorID,
		MaxParticipants:     int32(m.MaxParticipants),
		CurrentParticipants: int32(0),
		IsActive:            m.Status == model.MeetingStatusActive,
	}
	pb.HasPassword = m.PasswordHash != ""
	if !m.StartedAt.IsZero() {
		pb.StartedAt = timestamppb.New(m.StartedAt)
	}
	if m.EndedAt != nil {
		pb.EndedAt = timestamppb.New(*m.EndedAt)
	}
	if token != "" {
		pb.Token = token
	}
	return pb
}
