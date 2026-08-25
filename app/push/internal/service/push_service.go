package service

import (
	"context"

	"flamingo/app/push/internal/handler"
	"flamingo/app/push/internal/model"
	"flamingo/pkg/jpush"

	"github.com/go-kratos/kratos/v2/log"
)

type pushServiceImpl struct {
	repo        PushLogRepository
	jpushClient *jpush.Client
	log         *log.Helper
}

var _ handler.PushService = (*pushServiceImpl)(nil)

// NewPushService creates push service
func NewPushService(
	repo PushLogRepository,
	jpushClient *jpush.Client,
	logger log.Logger,
) handler.PushService {
	return &pushServiceImpl{
		repo:        repo,
		jpushClient: jpushClient,
		log:         log.NewHelper(logger),
	}
}

// SendPush sends push notification to multiple users
func (s *pushServiceImpl) SendPush(
	ctx context.Context,
	userIDs []string,
	title, content string,
	pushType model.PushType,
	extras map[string]string,
) (successCount, failureCount int, msgID string, err error) {
	if len(userIDs) == 0 {
		return 0, 0, "", nil
	}

	// Batch query push tokens
	tokenMap, err := s.repo.GetTokensByUserIDs(ctx, userIDs)
	if err != nil {
		return 0, len(userIDs), "", err
	}

	// Collect all registration_id
	var regIDs []string
	for _, rows := range tokenMap {
		for _, row := range rows {
			if row.Token != "" {
				regIDs = append(regIDs, row.Token)
			}
		}
	}

	if len(regIDs) == 0 {
		// All users have no push tokens (not registered JPush or no device)
		return 0, 0, "", nil
	}

	// Call JPush REST API
	result, pushErr := s.jpushClient.PushToRegistrationIDs(regIDs, title, content, extras)
	if pushErr != nil {
		failureCount = len(regIDs)
		s.logPush(ctx, userIDs[0], pushType, title, content, len(regIDs), 0, failureCount, "", model.PushStatusFailed, pushErr.Error())
		return 0, failureCount, "", pushErr
	}

	successCount = len(regIDs)
	s.logPush(ctx, userIDs[0], pushType, title, content, len(regIDs), successCount, 0, result.MsgID, model.PushStatusSent, "")

	return successCount, 0, result.MsgID, nil
}

// logPush asynchronously writes push log
func (s *pushServiceImpl) logPush(ctx context.Context, userID string, pushType model.PushType, title, content string,
	targetCount, successCount, failureCount int,
	jpushMsgID string, status model.PushStatus, errMsg string,
) {
	log := &model.PushLog{
		UserID:       userID,
		PushType:     pushType,
		Title:        title,
		Content:      content,
		TargetCount:  targetCount,
		SuccessCount: successCount,
		FailureCount: failureCount,
		JPushMsgID:   jpushMsgID,
		Status:       status,
		ErrorMsg:     errMsg,
	}
	if err := s.repo.Create(ctx, log); err != nil {
		s.log.Warnw("msg", "Failed to save push log", "error", err)
	}
}
