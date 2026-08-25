package repository

import (
	"context"

	"flamingo/app/user/internal/model"
	"flamingo/app/user/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// userSettingsRepositoryImpl user settings repository implementation
type userSettingsRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewUserSettingsRepository creates user settings repository
func NewUserSettingsRepository(db *gorm.DB, logger log.Logger) service.UserSettingsRepository {
	return &userSettingsRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates user settings
func (r *userSettingsRepositoryImpl) Create(ctx context.Context, settings *model.UserSettings) error {
	return r.db.WithContext(ctx).Create(settings).Error
}

// GetByUserID retrieves settings by user ID
func (r *userSettingsRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*model.UserSettings, error) {
	var settings model.UserSettings
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// Update updates user settings
func (r *userSettingsRepositoryImpl) Update(ctx context.Context, settings *model.UserSettings) error {
	return r.db.WithContext(ctx).Save(settings).Error
}

// WithTx returns a repository instance that uses the given transaction
func (r *userSettingsRepositoryImpl) WithTx(tx *gorm.DB) service.UserSettingsRepository {
	return &userSettingsRepositoryImpl{db: tx, log: r.log}
}
