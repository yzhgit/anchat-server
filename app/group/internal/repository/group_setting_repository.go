package repository

import (
	"context"
	"time"

	"flamingo/app/group/internal/model"
	"flamingo/app/group/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// groupSettingRepositoryImpl is the group settings repository implementation
type groupSettingRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewGroupSettingRepository creates a new group settings repository
func NewGroupSettingRepository(db *gorm.DB, logger log.Logger) service.GroupSettingRepository {
	return &groupSettingRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates group settings
func (r *groupSettingRepositoryImpl) Create(ctx context.Context, setting *model.GroupSetting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

// GetSettings gets group settings
func (r *groupSettingRepositoryImpl) GetSettings(ctx context.Context, groupID string) (*model.GroupSetting, error) {
	var setting model.GroupSetting
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// UpdateSettings updates group settings
func (r *groupSettingRepositoryImpl) UpdateSettings(ctx context.Context, groupID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&model.GroupSetting{}).
		Where("group_id = ?", groupID).
		Updates(updates).Error
}

// Delete deletes group settings
func (r *groupSettingRepositoryImpl) Delete(ctx context.Context, groupID string) error {
	return r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Delete(&model.GroupSetting{}).Error
}

// WithTx uses transaction
func (r *groupSettingRepositoryImpl) WithTx(tx *gorm.DB) service.GroupSettingRepository {
	return &groupSettingRepositoryImpl{db: tx}
}
