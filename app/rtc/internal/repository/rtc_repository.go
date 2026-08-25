package repository

import (
	"context"

	"flamingo/app/rtc/internal/model"
	"flamingo/app/rtc/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type callRepository struct {
	db  *gorm.DB
	log *log.Helper
}

func NewCallRepository(db *gorm.DB, logger log.Logger) service.CallRepository {
	return &callRepository{
		db:  db,
		log: log.NewHelper(logger),
	}
}

func (r *callRepository) CreateCallSession(ctx context.Context, session *model.CallSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *callRepository) GetCallSession(ctx context.Context, callID string) (*model.CallSession, error) {
	var session model.CallSession
	err := r.db.WithContext(ctx).Where("call_id = ?", callID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *callRepository) UpdateCallSession(ctx context.Context, session *model.CallSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *callRepository) ListCallLogs(ctx context.Context, userID string, page, pageSize int) ([]*model.CallSession, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var sessions []*model.CallSession
	var total int64

	query := r.db.WithContext(ctx).Model(&model.CallSession{}).
		Where("caller_id = ? OR callee_id = ?", userID, userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}
