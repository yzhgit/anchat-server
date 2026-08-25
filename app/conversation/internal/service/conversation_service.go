package service

import (
	"context"
	"fmt"
	"time"

	conversationv1 "flamingo/api/conversation/v1"

	"flamingo/app/conversation/internal/handler"
	"flamingo/app/conversation/internal/model"
	"flamingo/pkg/broker"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// conversationServiceImpl is the implementation of conversation service
type conversationServiceImpl struct {
	conversationRepo ConversationRepository
	broker           broker.Broker
	log              *log.Helper
}

var _ handler.ConversationService = (*conversationServiceImpl)(nil)

// NewConversationService creates a new conversation service
func NewConversationService(
	conversationRepo ConversationRepository,
	broker broker.Broker,
	logger log.Logger,
) handler.ConversationService {
	return &conversationServiceImpl{
		conversationRepo: conversationRepo,
		broker:           broker,
		log:              log.NewHelper(logger),
	}
}

// GetConversations retrieves the list of user conversations
func (s *conversationServiceImpl) GetConversations(ctx context.Context, userID string, req *conversationv1.GetConversationsRequest) (*conversationv1.GetConversationsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	var updatedBefore *time.Time
	if req.UpdatedBefore != nil {
		t := req.UpdatedBefore.AsTime()
		updatedBefore = &t
	}

	conversations, err := s.conversationRepo.ListByUser(ctx, userID, limit, updatedBefore)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	pbConversations := make([]*conversationv1.Conversation, 0, len(conversations))
	for _, c := range conversations {
		pbConversations = append(pbConversations, toProtoConversation(c))
	}

	return &conversationv1.GetConversationsResponse{
		Conversations: pbConversations,
	}, nil
}

// GetConversation retrieves a single conversation
func (s *conversationServiceImpl) GetConversation(ctx context.Context, userID, conversationID string) (*conversationv1.Conversation, error) {
	conversation, err := s.conversationRepo.GetByID(ctx, conversationID)
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("conversation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conversation.UserID != userID {
		return nil, fmt.Errorf("conversation not found")
	}
	return toProtoConversation(conversation), nil
}

// CreateOrUpdateConversation creates or updates a conversation (internal RPC, user_id from request).
// Uses a pure Upsert driven by a unique constraint on (user_id, conversation_type, target_id)
// to eliminate the check-then-create race: concurrent callers cannot both create.
func (s *conversationServiceImpl) CreateOrUpdateConversation(ctx context.Context, req *conversationv1.CreateOrUpdateConversationRequest) (*conversationv1.Conversation, error) {
	conversationType, err := parseConversationType(req.ConversationType)
	if err != nil {
		return nil, fmt.Errorf("invalid conversation type: %w", err)
	}

	var msgTime *time.Time
	if req.UpdatedAt != nil {
		t := req.UpdatedAt.AsTime()
		msgTime = &t
	}

	conversation := &model.Conversation{
		ConversationID:   uuid.New().String(),
		ConversationType: conversationType,
		UserID:           req.UserId,
		TargetID:         req.TargetId,
		LastMessageID:    req.FirstMessageId,
		LastMessageTime:  msgTime,
	}

	if upsertErr := s.conversationRepo.Upsert(ctx, conversation); upsertErr != nil {
		return nil, fmt.Errorf("failed to create or update conversation: %w", upsertErr)
	}

	// Refresh from DB to return the persisted state (the Upsert already holds the
	// unique constraint so the write itself is race-free). A failed read-after-write
	// (e.g. a concurrent delete) must not turn a successful Upsert into an error.
	latest, err := s.conversationRepo.GetByUserAndTarget(ctx, req.UserId, conversationType, req.TargetId)
	if err != nil {
		l := s.log.WithContext(ctx)
		l.Warnw("msg", "failed to refresh conversation after upsert (concurrent delete?)",
			"userID", req.UserId, "conversationType", conversationType, "targetID", req.TargetId, "error", err)
		return &conversationv1.Conversation{}, nil
	}
	return toProtoConversation(latest), nil
}

// DeleteConversation deletes a conversation and sends event
func (s *conversationServiceImpl) DeleteConversation(ctx context.Context, userID, conversationID string) error {
	l := s.log.WithContext(ctx)
	if err := s.conversationRepo.Delete(ctx, userID, conversationID); err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	// Publish conversation deletion event (multi-device sync)
	notif := broker.NewNotification(broker.TypeConversationDeleted, userID, broker.PriorityNormal).
		AddPayloadField("conversation_id", conversationID)
	if err := s.broker.PublishToUser(userID, notif); err != nil {
		l.Warnw("msg", "Failed to publish conversation deleted event", "userID", userID, "error", err)
	}

	return nil
}

// SetPinned sets pinned status and sends event
func (s *conversationServiceImpl) SetPinned(ctx context.Context, userID, conversationID string, pinned bool) error {
	l := s.log.WithContext(ctx)
	var pinTime *time.Time
	if pinned {
		t := time.Now()
		pinTime = &t
	}

	if err := s.conversationRepo.SetPinned(ctx, userID, conversationID, pinned, pinTime); err != nil {
		return fmt.Errorf("failed to set pinned: %w", err)
	}

	// Publish pinned status sync event (multi-device sync)
	notif := broker.NewNotification(broker.TypeConversationPinUpdated, userID, broker.PriorityNormal).
		AddPayloadField("conversation_id", conversationID).
		AddPayloadField("is_pinned", pinned)
	if err := s.broker.PublishToUser(userID, notif); err != nil {
		l.Warnw("msg", "Failed to publish conversation pin event", "userID", userID, "error", err)
	}

	return nil
}

// SetMuted sets muted status and sends event
func (s *conversationServiceImpl) SetMuted(ctx context.Context, userID, conversationID string, muted bool) error {
	l := s.log.WithContext(ctx)
	if err := s.conversationRepo.SetMuted(ctx, userID, conversationID, muted); err != nil {
		return fmt.Errorf("failed to set muted: %w", err)
	}

	// Publish muted setting sync event (multi-device sync)
	notif := broker.NewNotification(broker.TypeConversationMuteUpdated, userID, broker.PriorityNormal).
		AddPayloadField("conversation_id", conversationID).
		AddPayloadField("is_muted", muted)
	if err := s.broker.PublishToUser(userID, notif); err != nil {
		l.Warnw("msg", "Failed to publish conversation mute event", "userID", userID, "error", err)
	}

	return nil
}

// SetBurnAfterReading sets burn after reading duration and sends event
func (s *conversationServiceImpl) SetBurnAfterReading(ctx context.Context, userID, conversationID string, duration int32) error {
	l := s.log.WithContext(ctx)
	if err := s.conversationRepo.SetBurnAfterReading(ctx, userID, conversationID, duration); err != nil {
		return fmt.Errorf("failed to set burn after reading: %w", err)
	}

	// Publish burn after reading config change event (multi-device sync)
	notif := broker.NewNotification(broker.TypeConversationBurnUpdated, userID, broker.PriorityNormal).
		AddPayloadField("conversation_id", conversationID).
		AddPayloadField("burn_after_reading", duration)
	if err := s.broker.PublishToUser(userID, notif); err != nil {
		l.Warnw("msg", "Failed to publish conversation burn event", "userID", userID, "error", err)
	}

	return nil
}

// SetAutoDelete sets auto delete duration and sends event
func (s *conversationServiceImpl) SetAutoDelete(ctx context.Context, userID, conversationID string, duration int32) error {
	l := s.log.WithContext(ctx)
	if err := s.conversationRepo.SetAutoDelete(ctx, userID, conversationID, duration); err != nil {
		return fmt.Errorf("failed to set auto delete: %w", err)
	}

	// Publish auto delete config change event (multi-device sync)
	notif := broker.NewNotification(broker.TypeConversationAutoDeleteUpdated, userID, broker.PriorityNormal).
		AddPayloadField("conversation_id", conversationID).
		AddPayloadField("auto_delete_duration", duration)
	if err := s.broker.PublishToUser(userID, notif); err != nil {
		l.Warnw("msg", "Failed to publish conversation auto delete event", "userID", userID, "error", err)
	}

	return nil
}

// ClearUnread clears unread count and sends event
func (s *conversationServiceImpl) ClearUnread(ctx context.Context, userID, conversationID string) error {
	l := s.log.WithContext(ctx)
	if err := s.conversationRepo.ClearUnread(ctx, userID, conversationID); err != nil {
		return fmt.Errorf("failed to clear unread: %w", err)
	}

	// Get latest total unread count
	total, err := s.conversationRepo.SumUnread(ctx, userID)
	if err != nil {
		l.Warnw("msg", "Failed to get total unread after clear", "userID", userID, "error", err)
	}

	// Publish unread count update event (multi-device sync)
	notif := broker.NewNotification(broker.TypeConversationUnreadUpdated, userID, broker.PriorityNormal).
		AddPayloadField("conversation_id", conversationID).
		AddPayloadField("unread_count", 0).
		AddPayloadField("total_unread", total)
	if err := s.broker.PublishToUser(userID, notif); err != nil {
		l.Warnw("msg", "Failed to publish unread event", "userID", userID, "error", err)
	}

	return nil
}

// GetTotalUnread gets user's total unread count
func (s *conversationServiceImpl) GetTotalUnread(ctx context.Context, userID string) (int32, error) {
	total, err := s.conversationRepo.SumUnread(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get total unread: %w", err)
	}
	return total, nil
}

// IncrUnread increments unread count and sends event
func (s *conversationServiceImpl) IncrUnread(ctx context.Context, userID, conversationID string, count int64) error {
	l := s.log.WithContext(ctx)
	if err := s.conversationRepo.IncrUnread(ctx, userID, conversationID, int32(count)); err != nil {
		return fmt.Errorf("failed to incr conversation unread: %w", err)
	}

	// Publish unread count update event
	notif := broker.NewNotification(broker.TypeConversationUnreadUpdated, userID, broker.PriorityNormal).
		AddPayloadField("conversation_id", conversationID)
	if err := s.broker.PublishToUser(userID, notif); err != nil {
		l.Warnw("msg", "Failed to publish unread incr event", "userID", userID, "error", err)
	}

	return nil
}

// toProtoConversation converts model.Conversation to conversation.v1.Conversation protobuf
func toProtoConversation(c *model.Conversation) *conversationv1.Conversation {
	pb := &conversationv1.Conversation{
		ConversationId:     c.ConversationID,
		ConversationType:   conversationv1.ConversationType(c.ConversationType),
		TargetId:           c.TargetID,
		UnreadCount:        int32(c.UnreadCount),
		IsPinned:           c.IsPinned,
		IsMuted:            c.IsMuted,
		BurnAfterReading:   c.BurnAfterReading,
		AutoDeleteDuration: c.AutoDeleteDuration,
		CreatedAt:          timestamppb.New(c.CreatedAt),
		UpdatedAt:          timestamppb.New(c.UpdatedAt),
	}
	if c.LastMessageID != "" {
		pb.LastMessageId = &c.LastMessageID
	}
	if c.LastMessageContent != "" {
		pb.LastMessageContent = &c.LastMessageContent
	}
	if c.LastMessageTime != nil {
		pb.LastMessageTime = timestamppb.New(*c.LastMessageTime)
	}
	if c.PinTime != nil {
		pb.PinTime = timestamppb.New(*c.PinTime)
	}
	return pb
}

// parseConversationType converts a string to model.ConversationType
func parseConversationType(s string) (model.ConversationType, error) {
	switch s {
	case "CONVERSATION_TYPE_SINGLE", "single":
		return model.ConversationTypeSingle, nil
	case "CONVERSATION_TYPE_GROUP", "group":
		return model.ConversationTypeGroup, nil
	case "CONVERSATION_TYPE_SYSTEM", "system":
		return model.ConversationTypeSystem, nil
	default:
		return model.ConversationTypeUnspecified, fmt.Errorf("unknown conversation type: %s", s)
	}
}
