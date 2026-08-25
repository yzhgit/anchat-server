package repository

import (
	"context"

	"flamingo/app/message/internal/model"
	"flamingo/app/message/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// sequenceRepositoryImpl sequence repository implementation
type sequenceRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewSequenceRepository creates sequence repository
func NewSequenceRepository(db *gorm.DB, logger log.Logger) service.SequenceRepository {
	return &sequenceRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// GetOrCreate gets or creates sequence record
func (r *sequenceRepositoryImpl) GetOrCreate(ctx context.Context, conversationID string) (*model.ConversationSequence, error) {
	var seq model.ConversationSequence

	// Try to get
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		First(&seq).Error

	if err == nil {
		return &seq, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create if not exists
	seq = model.ConversationSequence{
		ConversationID: conversationID,
		CurrentSeq:     0,
	}

	err = r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&seq).Error

	if err != nil {
		// If concurrent create fails, try to get again
		err = r.db.WithContext(ctx).
			Where("conversation_id = ?", conversationID).
			First(&seq).Error
	}

	return &seq, err
}

// IncrementAndGet increments and gets sequence (atomic operation)
func (r *sequenceRepositoryImpl) IncrementAndGet(ctx context.Context, conversationID string) (int64, error) {
	// Ensure record exists
	if _, err := r.GetOrCreate(ctx, conversationID); err != nil {
		return 0, err
	}

	// Atomic increment
	result := r.db.WithContext(ctx).
		Model(&model.ConversationSequence{}).
		Where("conversation_id = ?", conversationID).
		UpdateColumn("current_seq", gorm.Expr("current_seq + 1"))

	if result.Error != nil {
		return 0, result.Error
	}

	// Get updated value
	var seq model.ConversationSequence
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		First(&seq).Error

	return seq.CurrentSeq, err
}

// GetCurrentSeq gets current sequence
func (r *sequenceRepositoryImpl) GetCurrentSeq(ctx context.Context, conversationID string) (int64, error) {
	var seq model.ConversationSequence
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		First(&seq).Error

	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return seq.CurrentSeq, nil
}

// Reset resets sequence
func (r *sequenceRepositoryImpl) Reset(ctx context.Context, conversationID string) error {
	return r.db.WithContext(ctx).
		Model(&model.ConversationSequence{}).
		Where("conversation_id = ?", conversationID).
		Update("current_seq", 0).Error
}

// Delete deletes sequence record
func (r *sequenceRepositoryImpl) Delete(ctx context.Context, conversationID string) error {
	return r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Delete(&model.ConversationSequence{}).Error
}

// WithTx uses transaction
func (r *sequenceRepositoryImpl) WithTx(tx *gorm.DB) service.SequenceRepository {
	return &sequenceRepositoryImpl{db: tx}
}
