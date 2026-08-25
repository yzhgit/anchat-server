package model

import "time"

// MessageSendIdempotency message send idempotency key
type MessageSendIdempotency struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SenderID       string    `gorm:"column:sender_id;not null;uniqueIndex:uk_sender_conversation_local" json:"senderId"`
	ConversationID string    `gorm:"column:conversation_id;not null;uniqueIndex:uk_sender_conversation_local" json:"conversationId"`
	LocalID        string    `gorm:"column:local_id;not null;uniqueIndex:uk_sender_conversation_local" json:"localId"`
	MessageID      string    `gorm:"column:message_id;not null;default:'';index:idx_message_idempotency_message_id" json:"messageId"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName returns table name
func (MessageSendIdempotency) TableName() string {
	return "message_send_idempotency"
}
