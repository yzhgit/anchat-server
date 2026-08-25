package service

import (
	"context"
	"flamingo/app/rtc/internal/model"
)

// CallRepository is the call session repository
type CallRepository interface {
	CreateCallSession(ctx context.Context, session *model.CallSession) error
	GetCallSession(ctx context.Context, callID string) (*model.CallSession, error)
	UpdateCallSession(ctx context.Context, session *model.CallSession) error
	ListCallLogs(ctx context.Context, userID string, page, pageSize int) ([]*model.CallSession, int64, error)
}

// MeetingRepository is the meeting room repository
type MeetingRepository interface {
	CreateMeeting(ctx context.Context, meeting *model.MeetingRoom) error
	GetMeetingByRoomID(ctx context.Context, roomID string) (*model.MeetingRoom, error)
	UpdateMeeting(ctx context.Context, meeting *model.MeetingRoom) error
	ListActiveMeetings(ctx context.Context, page, pageSize int) ([]*model.MeetingRoom, int64, error)
}
