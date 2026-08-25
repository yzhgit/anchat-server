package repository

import (
	"context"

	"flamingo/app/message/internal/model"
	"flamingo/app/message/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type sendIdempotencyRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewSendIdempotencyRepository creates send idempotency repository
func NewSendIdempotencyRepository(db *gorm.DB, logger log.Logger) service.SendIdempotencyRepository {
	return &sendIdempotencyRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// CreateIfNotExists creates idempotency record (ignores if exists)
func (r *sendIdempotencyRepositoryImpl) CreateIfNotExists(ctx context.Context, rec *model.MessageSendIdempotency) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(rec).Error
}

// GetForUpdateByKey queries by key and acquires row lock
func (r *sendIdempotencyRepositoryImpl) GetForUpdateByKey(ctx context.Context, senderID, conversationID, localID string) (*model.MessageSendIdempotency, error) {
	var rec model.MessageSendIdempotency
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("sender_id = ? AND conversation_id = ? AND local_id = ?", senderID, conversationID, localID).
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// BindMessageID binds actual message ID
func (r *sendIdempotencyRepositoryImpl) BindMessageID(ctx context.Context, senderID, conversationID, localID, messageID string) error {
	return r.db.WithContext(ctx).
		Model(&model.MessageSendIdempotency{}).
		Where("sender_id = ? AND conversation_id = ? AND local_id = ?", senderID, conversationID, localID).
		Update("message_id", messageID).Error
}

// WithTx uses transaction
func (r *sendIdempotencyRepositoryImpl) WithTx(tx *gorm.DB) service.SendIdempotencyRepository {
	return &sendIdempotencyRepositoryImpl{db: tx}
}
