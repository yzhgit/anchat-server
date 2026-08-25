package repository

import (
	"context"

	"flamingo/app/user/internal/model"
	"flamingo/app/user/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// userPushTokenRepositoryImpl push token repository implementation
type userPushTokenRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewUserPushTokenRepository creates push token repository
func NewUserPushTokenRepository(db *gorm.DB, logger log.Logger) service.UserPushTokenRepository {
	return &userPushTokenRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// CreateOrUpdate creates or updates push token
func (r *userPushTokenRepositoryImpl) CreateOrUpdate(ctx context.Context, token *model.UserPushToken) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND device_id = ?", token.UserID, token.DeviceID).
		Assign(token).
		FirstOrCreate(token).Error
}

// GetByUserID retrieves all push tokens for user
func (r *userPushTokenRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*model.UserPushToken, error) {
	var tokens []*model.UserPushToken
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// GetByUserIDAndDeviceID retrieves push token by user ID and device ID
func (r *userPushTokenRepositoryImpl) GetByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*model.UserPushToken, error) {
	var token model.UserPushToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Delete deletes push token
func (r *userPushTokenRepositoryImpl) Delete(ctx context.Context, userID, deviceID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		Delete(&model.UserPushToken{}).Error
}

// WithTx returns a repository instance that uses the given transaction
func (r *userPushTokenRepositoryImpl) WithTx(tx *gorm.DB) service.UserPushTokenRepository {
	return &userPushTokenRepositoryImpl{db: tx, log: r.log}
}
