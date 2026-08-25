package handler

import (
	"context"

	pushv1 "flamingo/api/push/v1"

	"flamingo/app/push/internal/model"
	"flamingo/pkg/errors"

	"github.com/go-kratos/kratos/v2/log"
)

// PushService push service interface
type PushService interface {
	// SendPush sends push to specified user list
	SendPush(ctx context.Context, userIDs []string, title, content string, pushType model.PushType, extras map[string]string) (successCount, failureCount int, msgID string, err error)
}

type PushHandler struct {
	pushv1.UnimplementedPushServiceServer
	svc PushService
	log *log.Helper
}

// NewServer creates gRPC handler
func NewPushHandler(svc PushService, logger log.Logger) *PushHandler {
	return &PushHandler{
		svc: svc,
		log: log.NewHelper(logger),
	}
}

// SendPush sends push notification to specified user list
func (s *PushHandler) SendPush(ctx context.Context, req *pushv1.SendPushRequest) (*pushv1.SendPushResponse, error) {
	if len(req.UserIds) == 0 {
		return nil, errors.BadRequest(ctx, "user_ids is required")
	}
	if req.Title == "" {
		return nil, errors.BadRequest(ctx, "title is required")
	}

	successCount, failureCount, msgID, err := s.svc.SendPush(
		ctx,
		req.UserIds,
		req.Title,
		req.Content,
		model.PushType(req.PushType),
		req.Extras,
	)
	if err != nil {
		return nil, errors.Internal(ctx, err)
	}

	return &pushv1.SendPushResponse{
		SuccessCount: int32(successCount),
		FailureCount: int32(failureCount),
		MsgId:        msgID,
	}, nil
}
