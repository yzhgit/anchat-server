package service

import (
	"context"
	"time"

	"flamingo/app/user/internal/model"

	"gorm.io/gorm"
)

// UserProfileRepository user profile repository interface
type UserProfileRepository interface {
	Create(ctx context.Context, profile *model.UserProfile) error
	GetByUserID(ctx context.Context, userID string) (*model.UserProfile, error)
	Update(ctx context.Context, profile *model.UserProfile) error
	UpdateQRCode(ctx context.Context, userID, qrcodeURL string) error
	CheckNicknameExists(ctx context.Context, nickname string, excludeUserID string) (bool, error)
	SearchByKeyword(ctx context.Context, keyword string, limit, offset int) ([]*model.UserProfile, int64, error)
	WithTx(tx *gorm.DB) UserProfileRepository
}

// UserPushTokenRepository push token repository interface
type UserPushTokenRepository interface {
	CreateOrUpdate(ctx context.Context, token *model.UserPushToken) error
	GetByUserID(ctx context.Context, userID string) ([]*model.UserPushToken, error)
	GetByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*model.UserPushToken, error)
	Delete(ctx context.Context, userID, deviceID string) error
	WithTx(tx *gorm.DB) UserPushTokenRepository
}

// UserQRCodeRepository user QR code repository interface
type UserQRCodeRepository interface {
	Create(ctx context.Context, qrcode *model.UserQRCode) error
	GetByToken(ctx context.Context, token string) (*model.UserQRCode, error)
	GetLatestByUserID(ctx context.Context, userID string) (*model.UserQRCode, error)
	DeleteExpired(ctx context.Context) error
	WithTx(tx *gorm.DB) UserQRCodeRepository
}

// UserSettingsRepository user settings repository interface
type UserSettingsRepository interface {
	Create(ctx context.Context, settings *model.UserSettings) error
	GetByUserID(ctx context.Context, userID string) (*model.UserSettings, error)
	Update(ctx context.Context, settings *model.UserSettings) error
	WithTx(tx *gorm.DB) UserSettingsRepository
}

// UserRepository user repository interface.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByAccount(ctx context.Context, account string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdatePhone(ctx context.Context, userID string, phone *string) error
	UpdateEmail(ctx context.Context, userID string, email *string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateStatus(ctx context.Context, userID string, status int) error
	WithTx(tx *gorm.DB) UserRepository
}

// UserDeviceRepository user device repository interface.
type UserDeviceRepository interface {
	Create(ctx context.Context, device *model.UserDevice) error
	GetByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*model.UserDevice, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.UserDevice, error)
	GetByUserIDAndDeviceType(ctx context.Context, userID string, deviceType model.DeviceType) ([]*model.UserDevice, error)
	Update(ctx context.Context, device *model.UserDevice) error
	UpdateLastLogin(ctx context.Context, userID, deviceID, ip string) error
	WithTx(tx *gorm.DB) UserDeviceRepository
}

// UserSessionRepository user session repository interface.
type UserSessionRepository interface {
	Create(ctx context.Context, session *model.UserSession) error
	GetByAccessToken(ctx context.Context, accessToken string) (*model.UserSession, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*model.UserSession, error)
	GetByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*model.UserSession, error)
	Update(ctx context.Context, session *model.UserSession) error
	DeleteByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteByUserIDExceptDeviceID(ctx context.Context, userID, deviceID string) error
	WithTx(tx *gorm.DB) UserSessionRepository
}

// VerificationCodeRepository verification code repository interface.
type VerificationCodeRepository interface {
	Create(ctx context.Context, code *model.VerificationCode) error
	GetByCodeID(ctx context.Context, codeID string) (*model.VerificationCode, error)
	GetLatestByTarget(ctx context.Context, target string, targetType model.VerificationTargetType, purpose model.VerificationPurpose) (*model.VerificationCode, error)
	UpdateStatus(ctx context.Context, codeID string, status model.VerificationCodeStatus) error
	UpdateVerifiedAt(ctx context.Context, codeID string, verifiedAt time.Time) error
	IncrementAttempts(ctx context.Context, codeID string) error
	Delete(ctx context.Context, codeID string) error
	DeleteExpired(ctx context.Context) (int64, error)
	WithTx(tx *gorm.DB) VerificationCodeRepository
}

// VerificationTemplateRepository verification template repository interface.
type VerificationTemplateRepository interface {
	GetByPurpose(ctx context.Context, purpose model.VerificationPurpose) (*model.VerificationTemplate, error)
	GetActive(ctx context.Context) ([]*model.VerificationTemplate, error)
	Update(ctx context.Context, template *model.VerificationTemplate) error
}
