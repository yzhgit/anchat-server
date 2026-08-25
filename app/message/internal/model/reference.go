package model

import "time"

// MessageReference message reference relationship
type MessageReference struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MessageID        string    `gorm:"column:message_id;not null;uniqueIndex:uk_message_reply;index:idx_message" json:"messageId"`
	ReplyToMessageID string    `gorm:"column:reply_to_message_id;not null;uniqueIndex:uk_message_reply;index:idx_reply_to" json:"replyToMessageId"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

// TableName returns table name
func (MessageReference) TableName() string {
	return "message_references"
}
