package repository

import (
	"context"
	"time"

	"flamingo/app/conversation/internal/model"
	"flamingo/app/conversation/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// conversationRepositoryImpl is the conversation repository implementation
type conversationRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *gorm.DB, logger log.Logger) service.ConversationRepository {
	return &conversationRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Upsert creates or updates a conversation (updates last message info on conflict)
func (r *conversationRepositoryImpl) Upsert(ctx context.Context, conversation *model.Conversation) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "conversation_type"}, {Name: "target_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"last_message_id",
				"last_message_content",
				"last_message_time",
				"updated_at",
			}),
		}).
		Create(conversation).Error
}

// GetByID retrieves a conversation by conversation ID
func (r *conversationRepositoryImpl) GetByID(ctx context.Context, conversationID string) (*model.Conversation, error) {
	var conversation model.Conversation
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// GetByUserAndTarget retrieves a conversation by user ID and target ID
func (r *conversationRepositoryImpl) GetByUserAndTarget(ctx context.Context, userID string, conversationType model.ConversationType, targetID string) (*model.Conversation, error) {
	var conversation model.Conversation
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND conversation_type = ? AND target_id = ?", userID, conversationType, targetID).
		First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// ListByUser retrieves user conversation list (sorted by pinned + last message time)
func (r *conversationRepositoryImpl) ListByUser(ctx context.Context, userID string, limit int, updatedBefore *time.Time) ([]*model.Conversation, error) {
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if updatedBefore != nil {
		q = q.Where("updated_at < ?", updatedBefore)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var conversations []*model.Conversation
	err := q.Order("is_pinned DESC, COALESCE(last_message_time, created_at) DESC").
		Limit(limit).
		Find(&conversations).Error
	return conversations, err
}

// Delete deletes a conversation (only deletes conversation belonging to the user)
func (r *conversationRepositoryImpl) Delete(ctx context.Context, userID, conversationID string) error {
	return r.db.WithContext(ctx).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Delete(&model.Conversation{}).Error
}

// SetPinned sets pinned status
func (r *conversationRepositoryImpl) SetPinned(ctx context.Context, userID, conversationID string, pinned bool, pinTime *time.Time) error {
	updates := map[string]interface{}{
		"is_pinned": pinned,
		"pin_time":  pinTime,
	}
	return r.db.WithContext(ctx).Model(&model.Conversation{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Updates(updates).Error
}

// SetMuted sets muted status
func (r *conversationRepositoryImpl) SetMuted(ctx context.Context, userID, conversationID string, muted bool) error {
	return r.db.WithContext(ctx).Model(&model.Conversation{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("is_muted", muted).Error
}

// SetBurnAfterReading sets burn after reading duration
func (r *conversationRepositoryImpl) SetBurnAfterReading(ctx context.Context, userID, conversationID string, duration int32) error {
	return r.db.WithContext(ctx).Model(&model.Conversation{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("burn_after_reading", duration).Error
}

// SetAutoDelete sets auto delete duration
func (r *conversationRepositoryImpl) SetAutoDelete(ctx context.Context, userID, conversationID string, duration int32) error {
	return r.db.WithContext(ctx).Model(&model.Conversation{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("auto_delete_duration", duration).Error
}

// ClearUnread clears unread count
func (r *conversationRepositoryImpl) ClearUnread(ctx context.Context, userID, conversationID string) error {
	return r.db.WithContext(ctx).Model(&model.Conversation{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("unread_count", 0).Error
}

// IncrUnread increments unread count
func (r *conversationRepositoryImpl) IncrUnread(ctx context.Context, userID, conversationID string, count int32) error {
	return r.db.WithContext(ctx).Model(&model.Conversation{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		UpdateColumn("unread_count", gorm.Expr("unread_count + ?", count)).Error
}

// SumUnread counts all unread counts for user (muted conversations are not included)
func (r *conversationRepositoryImpl) SumUnread(ctx context.Context, userID string) (int32, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.Conversation{}).
		Where("user_id = ? AND is_muted = false", userID).
		Select("COALESCE(SUM(unread_count), 0)").
		Scan(&total).Error
	return int32(total), err
}

// WithTx returns a repository instance using transaction
func (r *conversationRepositoryImpl) WithTx(tx *gorm.DB) service.ConversationRepository {
	return &conversationRepositoryImpl{db: tx}
}
