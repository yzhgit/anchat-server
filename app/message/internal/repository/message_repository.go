package repository

import (
	"context"
	"time"

	"flamingo/app/message/internal/model"
	"flamingo/app/message/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// messageRepositoryImpl message repository implementation
type messageRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewMessageRepository creates message repository
func NewMessageRepository(db *gorm.DB, logger log.Logger) service.MessageRepository {
	return &messageRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates a message
func (r *messageRepositoryImpl) Create(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// CreateBatch creates messages in batch
func (r *messageRepositoryImpl) CreateBatch(ctx context.Context, messages []*model.Message) error {
	return r.db.WithContext(ctx).Create(messages).Error
}

// GetByID retrieves message by ID
func (r *messageRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	var message model.Message
	err := r.db.WithContext(ctx).
		Where("id = ? AND status != ?", id, model.MessageStatusDeleted).
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetByMessageID retrieves message by message ID
func (r *messageRepositoryImpl) GetByMessageID(ctx context.Context, messageID string) (*model.Message, error) {
	var message model.Message
	err := r.db.WithContext(ctx).
		Where("message_id = ? AND status != ?", messageID, model.MessageStatusDeleted).
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetByConversation retrieves conversation messages (supports sequence range query)
func (r *messageRepositoryImpl) GetByConversation(ctx context.Context, conversationID string, startSeq, endSeq int64, limit int, reverse bool) ([]*model.Message, error) {
	var messages []*model.Message
	query := r.db.WithContext(ctx).
		Where("conversation_id = ? AND status = ?", conversationID, model.MessageStatusNormal)

	if startSeq > 0 {
		query = query.Where("sequence >= ?", startSeq)
	}
	if endSeq > 0 {
		query = query.Where("sequence <= ?", endSeq)
	}

	if reverse {
		query = query.Order("sequence DESC")
	} else {
		query = query.Order("sequence ASC")
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&messages).Error
	return messages, err
}

// GetLatestByConversation retrieves latest messages of conversation
func (r *messageRepositoryImpl) GetLatestByConversation(ctx context.Context, conversationID string, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND status = ?", conversationID, model.MessageStatusNormal).
		Order("sequence DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// GetBySender retrieves messages by sender
func (r *messageRepositoryImpl) GetBySender(ctx context.Context, senderID string, limit, offset int) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).
		Where("sender_id = ? AND status = ?", senderID, model.MessageStatusNormal).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

// UpdateStatus updates message status
func (r *messageRepositoryImpl) UpdateStatus(ctx context.Context, messageID string, status model.MessageStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("message_id = ?", messageID).
		Update("status", status).Error
}

// Delete deletes a message (soft delete)
func (r *messageRepositoryImpl) Delete(ctx context.Context, messageID string) error {
	return r.UpdateStatus(ctx, messageID, model.MessageStatusDeleted)
}

// CountByConversation counts messages in conversation
func (r *messageRepositoryImpl) CountByConversation(ctx context.Context, conversationID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("conversation_id = ? AND status = ?", conversationID, model.MessageStatusNormal).
		Count(&count).Error
	return count, err
}

// CountUnreadByConversation counts unread messages in conversation
func (r *messageRepositoryImpl) CountUnreadByConversation(ctx context.Context, conversationID string, lastReadSeq int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("conversation_id = ? AND sequence > ? AND status = ?", conversationID, lastReadSeq, model.MessageStatusNormal).
		Count(&count).Error
	return count, err
}

// SearchMessages searches messages
func (r *messageRepositoryImpl) SearchMessages(ctx context.Context, keyword string, conversationID *string, contentType *model.ContentType, limit, offset int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("status = ?", model.MessageStatusNormal)

	// Keyword search (using JSONB contains query)
	if keyword != "" {
		query = query.Where("content::text ILIKE ?", "%"+keyword+"%")
	}

	// Conversation filter
	if conversationID != nil && *conversationID != "" {
		query = query.Where("conversation_id = ?", *conversationID)
	}

	// Content type filter
	if contentType != nil && *contentType != model.ContentTypeUnspecified {
		query = query.Where("content_type = ?", *contentType)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginated query
	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	return messages, total, err
}

// GetByReplyTo retrieves all messages replying to a message
func (r *messageRepositoryImpl) GetByReplyTo(ctx context.Context, replyToMessageID string) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).
		Where("reply_to = ? AND status = ?", replyToMessageID, model.MessageStatusNormal).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

// GetExpiredMessages retrieves expired messages (paginated)
func (r *messageRepositoryImpl) GetExpiredMessages(ctx context.Context, before time.Time, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).
		Where("expire_time IS NOT NULL AND expire_time <= ? AND status = ?", before, model.MessageStatusNormal).
		Order("expire_time ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// BatchUpdateStatus batch updates message status
func (r *messageRepositoryImpl) BatchUpdateStatus(ctx context.Context, messageIDs []string, status model.MessageStatus) error {
	if len(messageIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("message_id IN ?", messageIDs).
		Update("status", status).Error
}

// GetExpiredMessageIDs retrieves expired message IDs (for event)
func (r *messageRepositoryImpl) GetExpiredMessageIDs(ctx context.Context, before time.Time, limit int) ([]string, error) {
	var messageIDs []string
	err := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("expire_time IS NOT NULL AND expire_time <= ? AND status = ?", before, model.MessageStatusNormal).
		Order("expire_time ASC").
		Limit(limit).
		Pluck("message_id", &messageIDs).Error
	return messageIDs, err
}

// WithTx uses transaction
func (r *messageRepositoryImpl) WithTx(tx *gorm.DB) service.MessageRepository {
	return &messageRepositoryImpl{db: tx}
}
