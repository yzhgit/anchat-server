package repository

import (
	"context"

	"flamingo/app/push/internal/model"
	"flamingo/app/push/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type pushLogRepository struct {
	db  *gorm.DB
	log *log.Helper
}

// NewPushLogRepository creates push log repository
func NewPushLogRepository(db *gorm.DB, logger log.Logger) service.PushLogRepository {
	return &pushLogRepository{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates push log
func (r *pushLogRepository) Create(ctx context.Context, log *model.PushLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetTokensByUserID retrieves all push tokens for specified user
func (r *pushLogRepository) GetTokensByUserID(ctx context.Context, userID string) ([]*service.PushTokenRow, error) {
	var rows []*service.PushTokenRow
	err := r.db.WithContext(ctx).Raw(
		`SELECT user_id, device_id, push_token AS token, platform
		   FROM user_push_tokens
		  WHERE user_id = ?`, userID,
	).Scan(&rows).Error
	return rows, err
}

// GetTokensByUserIDs retrieves push tokens for multiple users in batch
func (r *pushLogRepository) GetTokensByUserIDs(ctx context.Context, userIDs []string) (map[string][]*service.PushTokenRow, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	var rows []*service.PushTokenRow
	err := r.db.WithContext(ctx).Raw(
		`SELECT user_id, device_id, push_token AS token, platform
		   FROM user_push_tokens
		  WHERE user_id IN ?`, userIDs,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*service.PushTokenRow, len(userIDs))
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], row)
	}
	return result, nil
}
