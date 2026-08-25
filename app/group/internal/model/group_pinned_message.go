package model

import "time"

// PinnedMessageContentType mirrors message.v1.ContentType values. The
// content_type column stores the raw msg.GetContentType() value, so these
// constants match message.v1.ContentType 1:1.
type PinnedMessageContentType int32

const (
	PinnedMessageContentTypeUnspecified PinnedMessageContentType = 0
	PinnedMessageContentTypeText        PinnedMessageContentType = 1
	PinnedMessageContentTypeImage       PinnedMessageContentType = 2
	PinnedMessageContentTypeVideo       PinnedMessageContentType = 3
	PinnedMessageContentTypeAudio       PinnedMessageContentType = 4
	PinnedMessageContentTypeFile        PinnedMessageContentType = 5
	PinnedMessageContentTypeSticker     PinnedMessageContentType = 6
	PinnedMessageContentTypeEmoticon    PinnedMessageContentType = 7
	PinnedMessageContentTypeLocation    PinnedMessageContentType = 8
	PinnedMessageContentTypeRecall      PinnedMessageContentType = 9
	PinnedMessageContentTypeContact     PinnedMessageContentType = 10
	PinnedMessageContentTypeSystem      PinnedMessageContentType = 11
)

// GroupPinnedMessage represents a pinned message in a group
type GroupPinnedMessage struct {
	ID          int64                    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID     string                   `gorm:"column:group_id;not null;index:idx_group_pinned_messages_group_id" json:"groupId"`
	MessageID   string                   `gorm:"column:message_id;not null" json:"messageId"`
	MessageSeq  *int64                   `gorm:"column:message_seq" json:"messageSeq,omitempty"`
	PinnedBy    string                   `gorm:"column:pinned_by;not null" json:"pinnedBy"`
	Content     string                   `gorm:"column:content;type:text" json:"content"`
	ContentType PinnedMessageContentType `gorm:"column:content_type;type:smallint;not null;default:1" json:"contentType"`
	CreatedAt   time.Time                `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time                `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (GroupPinnedMessage) TableName() string {
	return "group_pinned_messages"
}
