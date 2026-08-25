package model

import (
	"time"
)

// ConversationType represents single or group message context.
type ConversationType int16

const (
	ConversationTypeUnspecified ConversationType = 0
	ConversationTypeSingle      ConversationType = 1
	ConversationTypeGroup       ConversationType = 2
)

// ContentType represents message payload category.
type ContentType int16

const (
	ContentTypeUnspecified ContentType = 0
	ContentTypeText        ContentType = 1
	ContentTypeImage       ContentType = 2
	ContentTypeVideo       ContentType = 3
	ContentTypeAudio       ContentType = 4
	ContentTypeFile        ContentType = 5
	ContentTypeLocation    ContentType = 6
	ContentTypeCard        ContentType = 7
)

// MessageStatus represents message lifecycle state.
type MessageStatus int16

// Message message model
type Message struct {
	ID                         int64            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MessageID                  string           `gorm:"column:message_id;not null;uniqueIndex" json:"messageId"`
	ConversationID             string           `gorm:"column:conversation_id;not null;index:idx_conversation_sequence" json:"conversationId"`
	ConversationType           ConversationType `gorm:"column:conversation_type;type:smallint;not null" json:"conversationType"` // 1-single/2-group
	TargetID                   string           `gorm:"column:target_id;not null;default:'';index:idx_messages_target_id" json:"targetId"`
	SenderID                   string           `gorm:"column:sender_id;not null;index:idx_sender_time" json:"senderId"`
	ContentType                ContentType      `gorm:"column:content_type;type:smallint;not null" json:"contentType"` // 1-text/2-image/3-video/4-audio/5-file/6-location/7-card
	Content                    string           `gorm:"column:content;type:jsonb;not null" json:"content"`
	Sequence                   int64            `gorm:"column:sequence;not null;uniqueIndex:uk_conversation_sequence" json:"sequence"`
	ReplyTo                    *string          `gorm:"column:reply_to" json:"replyTo,omitempty"`
	AtUsers                    []string         `gorm:"column:at_users;type:text[]" json:"atUsers,omitempty"`
	Status                     MessageStatus    `gorm:"column:status;type:smallint;default:0" json:"status"`                               // 0-normal 1-recalled 2-deleted
	BurnAfterReadingSeconds    int32            `gorm:"column:burn_after_reading_seconds;default:0" json:"burnAfterReadingSeconds"`        // burn-after-reading duration snapshot (seconds), 0 means not enabled
	AutoDeleteExpireTime       *time.Time       `gorm:"column:auto_delete_expire_time" json:"autoDeleteExpireTime,omitempty"`              // auto-delete policy expiration time
	BurnAfterReadingExpireTime *time.Time       `gorm:"column:burn_after_reading_expire_time" json:"burnAfterReadingExpireTime,omitempty"` // burn-after-reading policy expiration time
	ExpireTime                 *time.Time       `gorm:"column:expire_time;index:idx_expire_time" json:"expireTime,omitempty"`              // message expiration time, NULL means never expires
	CreatedAt                  time.Time        `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;index:idx_created_at" json:"createdAt"`
	UpdatedAt                  time.Time        `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName returns table name
func (Message) TableName() string {
	return "messages"
}

// MessageStatus message status
const (
	MessageStatusNormal  MessageStatus = 0 // normal
	MessageStatusRecall  MessageStatus = 1 // recalled
	MessageStatusDeleted MessageStatus = 2 // deleted
)

// IsNormal checks if message is normal
func (m *Message) IsNormal() bool {
	return m.Status == MessageStatusNormal
}

// IsRecalled checks if message is recalled
func (m *Message) IsRecalled() bool {
	return m.Status == MessageStatusRecall
}

// IsDeleted checks if message is deleted
func (m *Message) IsDeleted() bool {
	return m.Status == MessageStatusDeleted
}
