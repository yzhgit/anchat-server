package service

import (
	"context"

	"flamingo/app/push/internal/model"
)

// PushTokenRow push token record read from user_push_tokens table
type PushTokenRow struct {
	UserID   string
	DeviceID string
	Token    string // JPush registration_id
	Platform int16  // 1-ios / 2-android
}

// PushLogRepository push log repository interface
type PushLogRepository interface {
	Create(ctx context.Context, log *model.PushLog) error
	GetTokensByUserID(ctx context.Context, userID string) ([]*PushTokenRow, error)
	GetTokensByUserIDs(ctx context.Context, userIDs []string) (map[string][]*PushTokenRow, error)
}
