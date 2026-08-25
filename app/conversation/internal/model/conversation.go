package model

import (
	"time"
)

// ConversationType represents conversation category.
type ConversationType int16

const (
	ConversationTypeUnspecified ConversationType = 0
	ConversationTypeSingle      ConversationType = 1
	ConversationTypeGroup       ConversationType = 2
	ConversationTypeSystem      ConversationType = 3
)

// Conversation is the conversation model
type Conversation struct {
	ConversationID     string           `gorm:"column:conversation_id;primaryKey"`
	ConversationType   ConversationType `gorm:"column:conversation_type;type:smallint;not null"` // 1-single/2-group/3-system
	UserID             string           `gorm:"column:user_id;not null"`
	TargetID           string           `gorm:"column:target_id;not null"`
	LastMessageID      string           `gorm:"column:last_message_id"`
	LastMessageContent string           `gorm:"column:last_message_content"`
	LastMessageTime    *time.Time       `gorm:"column:last_message_time"`
	UnreadCount        int32            `gorm:"column:unread_count;default:0"`
	IsPinned           bool             `gorm:"column:is_pinned;default:false"`
	IsMuted            bool             `gorm:"column:is_muted;default:false"`
	PinTime            *time.Time       `gorm:"column:pin_time"`
	BurnAfterReading   int32            `gorm:"column:burn_after_reading;default:0"`   // burn after reading duration (seconds), 0 means not enabled
	AutoDeleteDuration int32            `gorm:"column:auto_delete_duration;default:0"` // auto delete duration (seconds), 0 means not enabled
	CreatedAt          time.Time        `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time        `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the table name
func (Conversation) TableName() string {
	return "conversations"
}
