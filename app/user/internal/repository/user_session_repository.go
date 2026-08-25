package repository

import (
	"context"

	"flamingo/app/user/internal/model"
	"flamingo/app/user/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// userSessionRepositoryImpl user session repository implementation
type userSessionRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewUserSessionRepository creates user session repository
func NewUserSessionRepository(db *gorm.DB, logger log.Logger) service.UserSessionRepository {
	return &userSessionRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates session
func (r *userSessionRepositoryImpl) Create(ctx context.Context, session *model.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByAccessToken gets session by access token
func (r *userSessionRepositoryImpl) GetByAccessToken(ctx context.Context, accessToken string) (*model.UserSession, error) {
	var session model.UserSession
	err := r.db.WithContext(ctx).
		Where("access_token = ?", accessToken).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByRefreshToken gets session by refresh token
func (r *userSessionRepositoryImpl) GetByRefreshToken(ctx context.Context, refreshToken string) (*model.UserSession, error) {
	var session model.UserSession
	err := r.db.WithContext(ctx).
		Where("refresh_token = ?", refreshToken).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByUserIDAndDeviceID gets session by user ID and device ID
func (r *userSessionRepositoryImpl) GetByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*model.UserSession, error) {
	var session model.UserSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update updates session
func (r *userSessionRepositoryImpl) Update(ctx context.Context, session *model.UserSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// DeleteByUserIDAndDeviceID deletes session for specified device
func (r *userSessionRepositoryImpl) DeleteByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		Delete(&model.UserSession{}).Error
}

// DeleteByUserID deletes all sessions for user
func (r *userSessionRepositoryImpl) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.UserSession{}).Error
}

// DeleteByUserIDExceptDeviceID deletes all sessions except current device
func (r *userSessionRepositoryImpl) DeleteByUserIDExceptDeviceID(ctx context.Context, userID, deviceID string) error {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if deviceID != "" {
		query = query.Where("device_id != ?", deviceID)
	}
	return query.Delete(&model.UserSession{}).Error
}

// WithTx returns a repository instance that uses the given transaction
func (r *userSessionRepositoryImpl) WithTx(tx *gorm.DB) service.UserSessionRepository {
	return &userSessionRepositoryImpl{db: tx, log: r.log}
}
