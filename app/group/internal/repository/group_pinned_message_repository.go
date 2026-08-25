package repository

import (
	"context"
	"time"

	"flamingo/app/group/internal/model"
	"flamingo/app/group/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type groupPinnedMessageRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

func NewGroupPinnedMessageRepository(db *gorm.DB, logger log.Logger) service.GroupPinnedMessageRepository {
	return &groupPinnedMessageRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

func (r *groupPinnedMessageRepositoryImpl) Upsert(ctx context.Context, pinned *model.GroupPinnedMessage) error {
	now := time.Now()
	updates := map[string]any{
		"pinned_by":    pinned.PinnedBy,
		"content":      pinned.Content,
		"content_type": pinned.ContentType,
		"message_seq":  pinned.MessageSeq,
		"created_at":   now,
		"updated_at":   now,
	}
	return r.db.WithContext(ctx).
		Model(&model.GroupPinnedMessage{}).
		Where("group_id = ? AND message_id = ?", pinned.GroupID, pinned.MessageID).
		Assign(updates).
		FirstOrCreate(pinned).Error
}

func (r *groupPinnedMessageRepositoryImpl) Delete(ctx context.Context, groupID, messageID string) (bool, error) {
	result := r.db.WithContext(ctx).
		Where("group_id = ? AND message_id = ?", groupID, messageID).
		Delete(&model.GroupPinnedMessage{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *groupPinnedMessageRepositoryImpl) ListByGroup(ctx context.Context, groupID string) ([]*model.GroupPinnedMessage, error) {
	var records []*model.GroupPinnedMessage
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

func (r *groupPinnedMessageRepositoryImpl) CountByGroup(ctx context.Context, groupID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.GroupPinnedMessage{}).
		Where("group_id = ?", groupID).
		Count(&total).Error
	return total, err
}

func (r *groupPinnedMessageRepositoryImpl) Exists(ctx context.Context, groupID, messageID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GroupPinnedMessage{}).
		Where("group_id = ? AND message_id = ?", groupID, messageID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *groupPinnedMessageRepositoryImpl) WithTx(tx *gorm.DB) service.GroupPinnedMessageRepository {
	return &groupPinnedMessageRepositoryImpl{db: tx}
}
