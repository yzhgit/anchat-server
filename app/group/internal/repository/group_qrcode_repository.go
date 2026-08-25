package repository

import (
	"context"
	"time"

	"flamingo/app/group/internal/model"
	"flamingo/app/group/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type groupQRCodeRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

func NewGroupQRCodeRepository(db *gorm.DB, logger log.Logger) service.GroupQRCodeRepository {
	return &groupQRCodeRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates a QR code record
func (r *groupQRCodeRepositoryImpl) Create(ctx context.Context, qrcode *model.GroupQRCode) error {
	return r.db.WithContext(ctx).Create(qrcode).Error
}

// GetActiveByGroupID gets current active QR code for group
func (r *groupQRCodeRepositoryImpl) GetActiveByGroupID(ctx context.Context, groupID string) (*model.GroupQRCode, error) {
	var qr model.GroupQRCode
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND is_active = true", groupID).
		Order("created_at DESC").
		First(&qr).Error
	if err != nil {
		return nil, err
	}
	return &qr, nil
}

// GetByToken gets QR code by token
func (r *groupQRCodeRepositoryImpl) GetByToken(ctx context.Context, token string) (*model.GroupQRCode, error) {
	var qr model.GroupQRCode
	err := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&qr).Error
	if err != nil {
		return nil, err
	}
	return &qr, nil
}

// InvalidateByGroupID marks all active QR codes for group as invalid
func (r *groupQRCodeRepositoryImpl) InvalidateByGroupID(ctx context.Context, groupID string) error {
	return r.db.WithContext(ctx).
		Model(&model.GroupQRCode{}).
		Where("group_id = ? AND is_active = true", groupID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error
}

// UpdateExpireAt renews QR code expiration
func (r *groupQRCodeRepositoryImpl) UpdateExpireAt(ctx context.Context, token string, expireAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.GroupQRCode{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"expire_at":  expireAt,
			"updated_at": time.Now(),
		}).Error
}

func (r *groupQRCodeRepositoryImpl) WithTx(tx *gorm.DB) service.GroupQRCodeRepository {
	return &groupQRCodeRepositoryImpl{db: tx}
}
