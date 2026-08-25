package repository

import (
	"context"
	"time"

	"flamingo/app/user/internal/model"
	"flamingo/app/user/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type verificationCodeRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

func NewVerificationCodeRepository(db *gorm.DB, logger log.Logger) service.VerificationCodeRepository {
	return &verificationCodeRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

func (r *verificationCodeRepositoryImpl) Create(ctx context.Context, code *model.VerificationCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *verificationCodeRepositoryImpl) GetByCodeID(ctx context.Context, codeID string) (*model.VerificationCode, error) {
	var code model.VerificationCode
	err := r.db.WithContext(ctx).
		Where("code_id = ?", codeID).
		First(&code).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *verificationCodeRepositoryImpl) GetLatestByTarget(ctx context.Context, target string, targetType model.VerificationTargetType, purpose model.VerificationPurpose) (*model.VerificationCode, error) {
	var code model.VerificationCode
	err := r.db.WithContext(ctx).
		Where("target = ? AND target_type = ? AND purpose = ?", target, targetType, purpose).
		Order("created_at DESC").
		First(&code).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *verificationCodeRepositoryImpl) UpdateStatus(ctx context.Context, codeID string, status model.VerificationCodeStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.VerificationCode{}).
		Where("code_id = ?", codeID).
		Update("status", status).Error
}

func (r *verificationCodeRepositoryImpl) UpdateVerifiedAt(ctx context.Context, codeID string, verifiedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.VerificationCode{}).
		Where("code_id = ?", codeID).
		Updates(map[string]interface{}{
			"status":      model.CodeStatusVerified,
			"verified_at": verifiedAt,
		}).Error
}

func (r *verificationCodeRepositoryImpl) IncrementAttempts(ctx context.Context, codeID string) error {
	return r.db.WithContext(ctx).
		Model(&model.VerificationCode{}).
		Where("code_id = ?", codeID).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1")).Error
}

func (r *verificationCodeRepositoryImpl) Delete(ctx context.Context, codeID string) error {
	return r.db.WithContext(ctx).
		Where("code_id = ?", codeID).
		Delete(&model.VerificationCode{}).Error
}

func (r *verificationCodeRepositoryImpl) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ? AND status = ?", time.Now(), model.CodeStatusPending).
		Delete(&model.VerificationCode{})
	return result.RowsAffected, result.Error
}

func (r *verificationCodeRepositoryImpl) WithTx(tx *gorm.DB) service.VerificationCodeRepository {
	return &verificationCodeRepositoryImpl{db: tx, log: r.log}
}
