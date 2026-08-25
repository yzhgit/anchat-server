package model

import "time"

// MessageReadReceipt message read receipt
type MessageReadReceipt struct {
	ID                int64            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ConversationID    string           `gorm:"column:conversation_id;not null;uniqueIndex:uk_conversation_user" json:"conversationId"`
	ConversationType  ConversationType `gorm:"column:conversation_type;type:smallint;not null" json:"conversationType"`
	TargetID          string           `gorm:"column:target_id;not null;default:''" json:"targetId"`
	UserID            string           `gorm:"column:user_id;not null;uniqueIndex:uk_conversation_user;index:idx_user" json:"userId"`
	LastReadSeq       int64            `gorm:"column:last_read_seq;not null" json:"lastReadSeq"`
	LastReadMessageID *string          `gorm:"column:last_read_message_id" json:"lastReadMessageId,omitempty"`
	ReadAt            time.Time        `gorm:"column:read_at;not null;default:CURRENT_TIMESTAMP" json:"readAt"`
}

// TableName returns table name
func (MessageReadReceipt) TableName() string {
	return "message_read_receipts"
}
