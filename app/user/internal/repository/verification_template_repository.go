package repository

import (
	"context"

	"flamingo/app/user/internal/model"
	"flamingo/app/user/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type verificationTemplateRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

func NewVerificationTemplateRepository(db *gorm.DB, logger log.Logger) service.VerificationTemplateRepository {
	return &verificationTemplateRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

func (r *verificationTemplateRepositoryImpl) GetByPurpose(ctx context.Context, purpose model.VerificationPurpose) (*model.VerificationTemplate, error) {
	var template model.VerificationTemplate
	err := r.db.WithContext(ctx).
		Where("purpose = ? AND is_active = ?", purpose, true).
		First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *verificationTemplateRepositoryImpl) GetActive(ctx context.Context) ([]*model.VerificationTemplate, error) {
	var templates []*model.VerificationTemplate
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&templates).Error
	return templates, err
}

func (r *verificationTemplateRepositoryImpl) Update(ctx context.Context, template *model.VerificationTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}
