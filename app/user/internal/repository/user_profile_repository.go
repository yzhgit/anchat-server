package repository

import (
	"context"

	"flamingo/app/user/internal/model"
	"flamingo/app/user/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// userProfileRepositoryImpl user profile repository implementation
type userProfileRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewUserProfileRepository creates user profile repository
func NewUserProfileRepository(db *gorm.DB, logger log.Logger) service.UserProfileRepository {
	return &userProfileRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates user profile
func (r *userProfileRepositoryImpl) Create(ctx context.Context, profile *model.UserProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

// GetByUserID retrieves profile by user ID
func (r *userProfileRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*model.UserProfile, error) {
	var profile model.UserProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// Update updates user profile
func (r *userProfileRepositoryImpl) Update(ctx context.Context, profile *model.UserProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

// UpdateQRCode updates QR code
func (r *userProfileRepositoryImpl) UpdateQRCode(ctx context.Context, userID, qrcodeURL string) error {
	return r.db.WithContext(ctx).
		Model(&model.UserProfile{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"qrcode_url":        qrcodeURL,
			"qrcode_updated_at": gorm.Expr("NOW()"),
		}).Error
}

// CheckNicknameExists checks if nickname exists
func (r *userProfileRepositoryImpl) CheckNicknameExists(ctx context.Context, nickname string, excludeUserID string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.UserProfile{}).Where("nickname = ?", nickname)

	if excludeUserID != "" {
		query = query.Where("user_id != ?", excludeUserID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// SearchByKeyword searches users by keyword
func (r *userProfileRepositoryImpl) SearchByKeyword(ctx context.Context, keyword string, limit, offset int) ([]*model.UserProfile, int64, error) {
	var profiles []*model.UserProfile
	var total int64

	// Search nickname
	query := r.db.WithContext(ctx).Model(&model.UserProfile{}).
		Where("nickname LIKE ?", "%"+keyword+"%")

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginated query
	if err := query.Limit(limit).Offset(offset).Find(&profiles).Error; err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
}

// WithTx returns a repository instance that uses the given transaction
func (r *userProfileRepositoryImpl) WithTx(tx *gorm.DB) service.UserProfileRepository {
	return &userProfileRepositoryImpl{db: tx, log: r.log}
}
