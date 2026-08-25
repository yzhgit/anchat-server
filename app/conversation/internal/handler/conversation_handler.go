package handler

import (
	"context"

	conversationv1 "flamingo/api/conversation/v1"

	"flamingo/pkg/errors"
	"flamingo/pkg/md"

	"github.com/go-kratos/kratos/v2/log"
	empty "github.com/golang/protobuf/ptypes/empty"
)

// ConversationService is the interface for conversation service
type ConversationService interface {
	GetConversations(ctx context.Context, userID string, req *conversationv1.GetConversationsRequest) (*conversationv1.GetConversationsResponse, error)
	GetConversation(ctx context.Context, userID, conversationID string) (*conversationv1.Conversation, error)
	CreateOrUpdateConversation(ctx context.Context, req *conversationv1.CreateOrUpdateConversationRequest) (*conversationv1.Conversation, error)
	DeleteConversation(ctx context.Context, userID, conversationID string) error
	SetPinned(ctx context.Context, userID, conversationID string, pinned bool) error
	SetMuted(ctx context.Context, userID, conversationID string, muted bool) error
	SetBurnAfterReading(ctx context.Context, userID, conversationID string, duration int32) error
	SetAutoDelete(ctx context.Context, userID, conversationID string, duration int32) error
	ClearUnread(ctx context.Context, userID, conversationID string) error
	GetTotalUnread(ctx context.Context, userID string) (int32, error)
	IncrUnread(ctx context.Context, userID, conversationID string, count int64) error
}

type ConversationHandler struct {
	conversationv1.UnimplementedConversationServiceServer
	svc ConversationService
	log *log.Helper
}

// NewConversationHandler creates conversation gRPC handler
func NewConversationHandler(svc ConversationService, logger log.Logger) *ConversationHandler {
	return &ConversationHandler{
		svc: svc,
		log: log.NewHelper(logger),
	}
}

// GetConversations retrieves the list of user conversations
func (s *ConversationHandler) GetConversations(ctx context.Context, req *conversationv1.GetConversationsRequest) (*conversationv1.GetConversationsResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.GetConversations(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// GetConversation retrieves a single conversation
func (s *ConversationHandler) GetConversation(ctx context.Context, req *conversationv1.GetConversationRequest) (*conversationv1.Conversation, error) {
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	conversation, err := s.svc.GetConversation(ctx, md.MustGetUserID(ctx), req.ConversationId)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return conversation, nil
}

// CreateOrUpdateConversation creates or updates a conversation (internal RPC, user_id from request)
func (s *ConversationHandler) CreateOrUpdateConversation(ctx context.Context, req *conversationv1.CreateOrUpdateConversationRequest) (*conversationv1.Conversation, error) {
	if req.UserId == "" || req.TargetId == "" || req.ConversationType == "" {
		return nil, errors.BadRequest(ctx, "user_id, target_id and conversation_type are required")
	}
	conversation, err := s.svc.CreateOrUpdateConversation(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return conversation, nil
}

// DeleteConversation deletes a conversation
func (s *ConversationHandler) DeleteConversation(ctx context.Context, req *conversationv1.DeleteConversationRequest) (*empty.Empty, error) {
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if err := s.svc.DeleteConversation(ctx, md.MustGetUserID(ctx), req.ConversationId); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// SetPinned sets pinned status
func (s *ConversationHandler) SetPinned(ctx context.Context, req *conversationv1.SetPinnedRequest) (*empty.Empty, error) {
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if err := s.svc.SetPinned(ctx, md.MustGetUserID(ctx), req.ConversationId, req.Pinned); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// SetMuted sets muted status
func (s *ConversationHandler) SetMuted(ctx context.Context, req *conversationv1.SetMutedRequest) (*empty.Empty, error) {
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if err := s.svc.SetMuted(ctx, md.MustGetUserID(ctx), req.ConversationId, req.Muted); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// ClearUnread clears unread count
func (s *ConversationHandler) ClearUnread(ctx context.Context, req *conversationv1.ClearUnreadRequest) (*empty.Empty, error) {
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if err := s.svc.ClearUnread(ctx, md.MustGetUserID(ctx), req.ConversationId); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// GetTotalUnread gets total unread count
func (s *ConversationHandler) GetTotalUnread(ctx context.Context, req *conversationv1.GetTotalUnreadRequest) (*conversationv1.GetTotalUnreadResponse, error) {
	total, err := s.svc.GetTotalUnread(ctx, md.MustGetUserID(ctx))
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &conversationv1.GetTotalUnreadResponse{Total: int64(total)}, nil
}

// IncrUnread increments unread count (internal RPC, user_id from request)
func (s *ConversationHandler) IncrUnread(ctx context.Context, req *conversationv1.IncrUnreadRequest) (*empty.Empty, error) {
	if req.UserId == "" || req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "user_id and conversation_id are required")
	}
	if err := s.svc.IncrUnread(ctx, req.UserId, req.ConversationId, req.Incr); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// SetBurnAfterReading sets burn after reading
func (s *ConversationHandler) SetBurnAfterReading(ctx context.Context, req *conversationv1.SetBurnAfterReadingRequest) (*empty.Empty, error) {
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if err := s.svc.SetBurnAfterReading(ctx, md.MustGetUserID(ctx), req.ConversationId, req.Duration); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// SetAutoDelete sets auto delete
func (s *ConversationHandler) SetAutoDelete(ctx context.Context, req *conversationv1.SetAutoDeleteRequest) (*empty.Empty, error) {
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if err := s.svc.SetAutoDelete(ctx, md.MustGetUserID(ctx), req.ConversationId, req.Duration); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}
