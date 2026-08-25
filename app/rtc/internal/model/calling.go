package model

import "time"

// CallType represents one-to-one call type.
type CallType int16

const (
	CallTypeAudio CallType = 0
	CallTypeVideo CallType = 1
)

// CallStatus represents one-to-one call session status.
type CallStatus int16

const (
	CallStatusRinging   CallStatus = 0
	CallStatusConnected CallStatus = 1
	CallStatusEnded     CallStatus = 2
	CallStatusRejected  CallStatus = 3
	CallStatusMissed    CallStatus = 4
	CallStatusCancelled CallStatus = 5
)

// MeetingStatus represents meeting room status.
type MeetingStatus int16

const (
	MeetingStatusActive MeetingStatus = 0
	MeetingStatusEnded  MeetingStatus = 1
)

// CallSession represents a call session
type CallSession struct {
	ID          int64      `gorm:"primaryKey;autoIncrement"`
	CallID      string     `gorm:"column:call_id;uniqueIndex;not null"`
	CallerID    string     `gorm:"column:caller_id;not null;index"`
	CalleeID    string     `gorm:"column:callee_id;not null;index"`
	CallType    CallType   `gorm:"column:call_type;type:smallint;not null;default:0"` // 0-audio/1-video
	Status      CallStatus `gorm:"column:status;type:smallint;not null;default:0"`    // 0-ringing/1-connected/2-ended/3-rejected/4-missed/5-cancelled
	RoomName    string     `gorm:"column:room_name;not null"`
	StartedAt   time.Time  `gorm:"column:started_at;autoCreateTime"`
	ConnectedAt *time.Time `gorm:"column:connected_at"`
	EndedAt     *time.Time `gorm:"column:ended_at"`
	Duration    int        `gorm:"column:duration;not null;default:0"` // seconds
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (CallSession) TableName() string { return "call_sessions" }

// MeetingRoom represents a meeting room
type MeetingRoom struct {
	ID              int64         `gorm:"primaryKey;autoIncrement"`
	RoomID          string        `gorm:"column:room_id;uniqueIndex;not null"`
	CreatorID       string        `gorm:"column:creator_id;not null;index"`
	Title           string        `gorm:"column:title;not null"`
	RoomName        string        `gorm:"column:room_name;uniqueIndex;not null"`
	PasswordHash    string        `gorm:"column:password_hash"`
	MaxParticipants int           `gorm:"column:max_participants;not null;default:0"`
	Status          MeetingStatus `gorm:"column:status;type:smallint;not null;default:0"` // 0-active/1-ended
	StartedAt       time.Time     `gorm:"column:started_at;autoCreateTime"`
	EndedAt         *time.Time    `gorm:"column:ended_at"`
	CreatedAt       time.Time     `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time     `gorm:"column:updated_at;autoUpdateTime"`
}

func (MeetingRoom) TableName() string { return "meeting_rooms" }
