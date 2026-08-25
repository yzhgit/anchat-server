package service

import (
	"context"
	"time"

	"flamingo/app/message/internal/model"

	"gorm.io/gorm"
)

// MessageRepository message repository interface
type MessageRepository interface {
	Create(ctx context.Context, message *model.Message) error
	CreateBatch(ctx context.Context, messages []*model.Message) error
	GetByID(ctx context.Context, id int64) (*model.Message, error)
	GetByMessageID(ctx context.Context, messageID string) (*model.Message, error)
	GetByConversation(ctx context.Context, conversationID string, startSeq, endSeq int64, limit int, reverse bool) ([]*model.Message, error)
	GetLatestByConversation(ctx context.Context, conversationID string, limit int) ([]*model.Message, error)
	GetBySender(ctx context.Context, senderID string, limit, offset int) ([]*model.Message, error)
	UpdateStatus(ctx context.Context, messageID string, status model.MessageStatus) error
	Delete(ctx context.Context, messageID string) error
	CountByConversation(ctx context.Context, conversationID string) (int64, error)
	CountUnreadByConversation(ctx context.Context, conversationID string, lastReadSeq int64) (int64, error)
	SearchMessages(ctx context.Context, keyword string, conversationID *string, contentType *model.ContentType, limit, offset int) ([]*model.Message, int64, error)
	GetByReplyTo(ctx context.Context, replyToMessageID string) ([]*model.Message, error)
	// GetExpiredMessages retrieves expired messages (paginated)
	GetExpiredMessages(ctx context.Context, before time.Time, limit int) ([]*model.Message, error)
	// BatchUpdateStatus batch updates message status
	BatchUpdateStatus(ctx context.Context, messageIDs []string, status model.MessageStatus) error
	// GetExpiredMessageIDs retrieves expired message IDs (for event)
	GetExpiredMessageIDs(ctx context.Context, before time.Time, limit int) ([]string, error)
	WithTx(tx *gorm.DB) MessageRepository
}

// ReadReceiptRepository read receipt repository interface
type ReadReceiptRepository interface {
	Upsert(ctx context.Context, receipt *model.MessageReadReceipt) error
	GetByConversationAndUser(ctx context.Context, conversationID, userID string) (*model.MessageReadReceipt, error)
	GetByConversation(ctx context.Context, conversationID string) ([]*model.MessageReadReceipt, error)
	GetByUser(ctx context.Context, userID string) ([]*model.MessageReadReceipt, error)
	Delete(ctx context.Context, conversationID, userID string) error
	WithTx(tx *gorm.DB) ReadReceiptRepository
}

// SendIdempotencyRepository send idempotency repository interface
type SendIdempotencyRepository interface {
	CreateIfNotExists(ctx context.Context, rec *model.MessageSendIdempotency) error
	GetForUpdateByKey(ctx context.Context, senderID, conversationID, localID string) (*model.MessageSendIdempotency, error)
	BindMessageID(ctx context.Context, senderID, conversationID, localID, messageID string) error
	WithTx(tx *gorm.DB) SendIdempotencyRepository
}

// SequenceRepository sequence repository interface
type SequenceRepository interface {
	GetOrCreate(ctx context.Context, conversationID string) (*model.ConversationSequence, error)
	IncrementAndGet(ctx context.Context, conversationID string) (int64, error)
	GetCurrentSeq(ctx context.Context, conversationID string) (int64, error)
	Reset(ctx context.Context, conversationID string) error
	Delete(ctx context.Context, conversationID string) error
	WithTx(tx *gorm.DB) SequenceRepository
}

// TypingRepository typing status cache repository
type TypingRepository interface {
	SetState(ctx context.Context, conversationID, fromUserID string, ttl time.Duration) error
	ClearState(ctx context.Context, conversationID, fromUserID string) error
	AcquireEmitToken(ctx context.Context, conversationID, fromUserID string, ttl time.Duration) (bool, error)
}
