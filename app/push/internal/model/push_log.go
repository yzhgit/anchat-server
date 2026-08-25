package model

import "time"

// PushType represents business push category.
type PushType int16

const (
	PushTypeUnspecified    PushType = 0
	PushTypeMessageNew     PushType = 1
	PushTypeMessageMention PushType = 2
	PushTypeFriendRequest  PushType = 3
	PushTypeGroupInvited   PushType = 4
	PushTypeCallInvite     PushType = 5
)

// PushStatus represents push delivery status.
type PushStatus int16

const (
	PushStatusUnspecified PushStatus = 0
	PushStatusPending     PushStatus = 1
	PushStatusSent        PushStatus = 2
	PushStatusFailed      PushStatus = 3
)

// PushLog push log model
type PushLog struct {
	ID           int64      `gorm:"primaryKey;autoIncrement"`
	UserID       string     `gorm:"column:user_id;not null;index"`
	PushType     PushType   `gorm:"column:push_type;type:smallint;not null;default:0"`
	Title        string     `gorm:"column:title"`
	Content      string     `gorm:"column:content"`
	TargetCount  int        `gorm:"column:target_count;not null;default:0"`
	SuccessCount int        `gorm:"column:success_count;not null;default:0"`
	FailureCount int        `gorm:"column:failure_count;not null;default:0"`
	JPushMsgID   string     `gorm:"column:jpush_msg_id"`
	Status       PushStatus `gorm:"column:status;type:smallint;not null;default:1"` // 1-pending/2-sent/3-failed
	ErrorMsg     string     `gorm:"column:error_msg"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (PushLog) TableName() string {
	return "push_logs"
}
