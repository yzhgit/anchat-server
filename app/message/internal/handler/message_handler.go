package handler

import (
	"context"

	messagev1 "flamingo/api/message/v1"

	"flamingo/pkg/errors"
	"flamingo/pkg/md"

	"github.com/go-kratos/kratos/v2/log"
	empty "github.com/golang/protobuf/ptypes/empty"
)

// MessageService message service interface
type MessageService interface {
	SendMessage(ctx context.Context, userID string, req *messagev1.SendMessageRequest) (*messagev1.SendMessageResponse, error)
	SendTyping(ctx context.Context, userID string, req *messagev1.SendTypingRequest) error
	GetMessages(ctx context.Context, req *messagev1.GetMessagesRequest) (*messagev1.GetMessagesResponse, error)
	GetMessagesBefore(ctx context.Context, userID string, req *messagev1.GetMessagesBeforeRequest) (*messagev1.GetMessagesBeforeResponse, error)
	GetMessagesAfter(ctx context.Context, userID string, req *messagev1.GetMessagesAfterRequest) (*messagev1.GetMessagesAfterResponse, error)
	GetMessagesAroundAnchor(ctx context.Context, userID string, req *messagev1.GetMessagesAroundAnchorRequest) (*messagev1.GetMessagesAroundAnchorResponse, error)
	GetFirstUnreadAnchor(ctx context.Context, userID string, req *messagev1.GetFirstUnreadAnchorRequest) (*messagev1.GetFirstUnreadAnchorResponse, error)
	GetMessageById(ctx context.Context, messageID string) (*messagev1.Message, error)
	RecallMessage(ctx context.Context, messageID, userID string) error
	DeleteMessage(ctx context.Context, messageID, userID string) error
	MarkAsRead(ctx context.Context, userID string, req *messagev1.MarkAsReadRequest) error
	MarkMessagesRead(ctx context.Context, userID string, req *messagev1.MarkMessagesReadRequest) (*messagev1.MarkMessagesReadResponse, error)
	AckReadTriggers(ctx context.Context, userID string, req *messagev1.AckReadTriggersRequest) (*messagev1.AckReadTriggersResponse, error)
	GetUnreadCount(ctx context.Context, conversationID, userID string, lastReadSeq *int64) (*messagev1.GetUnreadCountResponse, error)
	GetReadReceipts(ctx context.Context, conversationID, userID string) (*messagev1.GetReadReceiptsResponse, error)
	GetConversationSequence(ctx context.Context, conversationID string) (int64, error)
	SearchMessages(ctx context.Context, userID string, req *messagev1.SearchMessagesRequest) (*messagev1.SearchMessagesResponse, error)
}

type MessageHandler struct {
	messagev1.UnimplementedMessageServiceServer
	svc MessageService
	log *log.Helper
}

// NewMessageHandler creates a new message gRPC handler
func NewMessageHandler(svc MessageService, logger log.Logger) *MessageHandler {
	return &MessageHandler{
		svc: svc,
		log: log.NewHelper(logger),
	}
}

// SendMessage sends a message
func (h *MessageHandler) SendMessage(ctx context.Context, req *messagev1.SendMessageRequest) (*messagev1.SendMessageResponse, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "SendMessage called",
		"senderId", userID,
		"conversationId", req.ConversationId,
		"contentType", int32(req.ContentType))

	// Parameter validation
	if userID == "" {
		return nil, errors.BadRequest(ctx, "sender_id is required")
	}
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if req.ContentType == messagev1.ContentType_CONTENT_TYPE_UNSPECIFIED {
		return nil, errors.BadRequest(ctx, "content_type is required")
	}
	if req.Content == "" {
		return nil, errors.BadRequest(ctx, "content is required")
	}
	if req.GetLocalId() == "" {
		return nil, errors.BadRequest(ctx, "local_id is required")
	}

	resp, err := h.svc.SendMessage(ctx, userID, req)
	if err != nil {
		l.Errorw("msg", "Failed to send message", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GetMessages retrieves message list
func (h *MessageHandler) GetMessages(ctx context.Context, req *messagev1.GetMessagesRequest) (*messagev1.GetMessagesResponse, error) {
	l := h.log.WithContext(ctx)
	l.Infow("msg", "GetMessages called",
		"conversationId", req.ConversationId,
		"limit", req.Limit)

	// Parameter validation
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}

	resp, err := h.svc.GetMessages(ctx, req)
	if err != nil {
		l.Errorw("msg", "Failed to get messages", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GetMessagesBefore retrieves messages before anchor message
func (h *MessageHandler) GetMessagesBefore(ctx context.Context, req *messagev1.GetMessagesBeforeRequest) (*messagev1.GetMessagesBeforeResponse, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "GetMessagesBefore called",
		"conversationId", req.ConversationId,
		"anchorMessageId", req.AnchorMessageId,
		"userId", userID,
		"limit", req.Limit)

	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if req.AnchorMessageId == "" {
		return nil, errors.BadRequest(ctx, "anchor_message_id is required")
	}
	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}

	resp, err := h.svc.GetMessagesBefore(ctx, userID, req)
	if err != nil {
		l.Errorw("msg", "Failed to get messages before anchor", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GetMessagesAfter retrieves messages after anchor message
func (h *MessageHandler) GetMessagesAfter(ctx context.Context, req *messagev1.GetMessagesAfterRequest) (*messagev1.GetMessagesAfterResponse, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "GetMessagesAfter called",
		"conversationId", req.ConversationId,
		"anchorMessageId", req.AnchorMessageId,
		"userId", userID,
		"limit", req.Limit)

	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if req.AnchorMessageId == "" {
		return nil, errors.BadRequest(ctx, "anchor_message_id is required")
	}
	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}

	resp, err := h.svc.GetMessagesAfter(ctx, userID, req)
	if err != nil {
		l.Errorw("msg", "Failed to get messages after anchor", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GetMessagesAroundAnchor retrieves messages around anchor message
func (h *MessageHandler) GetMessagesAroundAnchor(ctx context.Context, req *messagev1.GetMessagesAroundAnchorRequest) (*messagev1.GetMessagesAroundAnchorResponse, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "GetMessagesAroundAnchor called",
		"conversationId", req.ConversationId,
		"anchorMessageId", req.AnchorMessageId,
		"userId", userID,
		"before", req.Before,
		"after", req.After)

	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if req.AnchorMessageId == "" {
		return nil, errors.BadRequest(ctx, "anchor_message_id is required")
	}
	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}

	resp, err := h.svc.GetMessagesAroundAnchor(ctx, userID, req)
	if err != nil {
		l.Errorw("msg", "Failed to get messages around anchor", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GetFirstUnreadAnchor retrieves first unread message anchor
func (h *MessageHandler) GetFirstUnreadAnchor(ctx context.Context, req *messagev1.GetFirstUnreadAnchorRequest) (*messagev1.GetFirstUnreadAnchorResponse, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "GetFirstUnreadAnchor called",
		"conversationId", req.ConversationId,
		"userId", userID)

	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}

	resp, err := h.svc.GetFirstUnreadAnchor(ctx, userID, req)
	if err != nil {
		l.Errorw("msg", "Failed to get first unread anchor", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GetMessageById retrieves message by ID
func (h *MessageHandler) GetMessageById(ctx context.Context, req *messagev1.GetMessageByIdRequest) (*messagev1.Message, error) {
	l := h.log.WithContext(ctx)
	l.Infow("msg", "GetMessageById called", "messageId", req.MessageId)

	// Parameter validation
	if req.MessageId == "" {
		return nil, errors.BadRequest(ctx, "message_id is required")
	}

	resp, err := h.svc.GetMessageById(ctx, req.MessageId)
	if err != nil {
		l.Errorw("msg", "Failed to get message", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// RecallMessage recalls a message
func (h *MessageHandler) RecallMessage(ctx context.Context, req *messagev1.RecallMessageRequest) (*empty.Empty, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "RecallMessage called",
		"messageId", req.MessageId,
		"userId", userID)

	// Parameter validation
	if req.MessageId == "" {
		return nil, errors.BadRequest(ctx, "message_id is required")
	}
	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}

	err := h.svc.RecallMessage(ctx, req.MessageId, userID)
	if err != nil {
		l.Errorw("msg", "Failed to recall message", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return &empty.Empty{}, nil
}

// DeleteMessage deletes a message
func (h *MessageHandler) DeleteMessage(ctx context.Context, req *messagev1.DeleteMessageRequest) (*empty.Empty, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "DeleteMessage called",
		"messageId", req.MessageId,
		"userId", userID)

	// Parameter validation
	if req.MessageId == "" {
		return nil, errors.BadRequest(ctx, "message_id is required")
	}
	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}

	err := h.svc.DeleteMessage(ctx, req.MessageId, userID)
	if err != nil {
		l.Errorw("msg", "Failed to delete message", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return &empty.Empty{}, nil
}

// MarkAsRead marks messages as read
func (h *MessageHandler) MarkAsRead(ctx context.Context, req *messagev1.MarkAsReadRequest) (*empty.Empty, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "MarkAsRead called",
		"conversationId", req.ConversationId,
		"userId", userID,
		"lastReadSeq", req.LastReadSeq)

	// Parameter validation
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}

	err := h.svc.MarkAsRead(ctx, userID, req)
	if err != nil {
		l.Errorw("msg", "Failed to mark as read", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return &empty.Empty{}, nil
}

// MarkMessagesRead marks messages as read by message IDs
func (h *MessageHandler) MarkMessagesRead(ctx context.Context, req *messagev1.MarkMessagesReadRequest) (*messagev1.MarkMessagesReadResponse, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "MarkMessagesRead called",
		"conversationId", req.ConversationId,
		"userId", userID,
		"messageCount", len(req.MessageIds))

	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}
	if len(req.MessageIds) == 0 {
		return nil, errors.BadRequest(ctx, "message_ids is required")
	}

	resp, err := h.svc.MarkMessagesRead(ctx, userID, req)
	if err != nil {
		l.Errorw("msg", "Failed to mark messages as read", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// AckReadTriggers acknowledges burn-after-reading triggers
func (h *MessageHandler) AckReadTriggers(ctx context.Context, req *messagev1.AckReadTriggersRequest) (*messagev1.AckReadTriggersResponse, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "AckReadTriggers called",
		"userId", userID,
		"eventCount", len(req.Events))

	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}
	if len(req.Events) == 0 {
		return nil, errors.BadRequest(ctx, "events is required")
	}

	resp, err := h.svc.AckReadTriggers(ctx, userID, req)
	if err != nil {
		l.Errorw("msg", "Failed to ack read triggers", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GetUnreadCount retrieves unread message count
func (h *MessageHandler) GetUnreadCount(ctx context.Context, req *messagev1.GetUnreadCountRequest) (*messagev1.GetUnreadCountResponse, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "GetUnreadCount called",
		"conversationId", req.ConversationId,
		"userId", userID)

	// Parameter validation
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}
	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}

	resp, err := h.svc.GetUnreadCount(ctx, req.ConversationId, userID, req.LastReadSeq)
	if err != nil {
		l.Errorw("msg", "Failed to get unread count", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GetReadReceipts retrieves read receipts
func (h *MessageHandler) GetReadReceipts(ctx context.Context, req *messagev1.GetReadReceiptsRequest) (*messagev1.GetReadReceiptsResponse, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "GetReadReceipts called",
		"conversationId", req.ConversationId,
		"userId", userID)

	// Parameter validation
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}

	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}

	resp, err := h.svc.GetReadReceipts(ctx, req.ConversationId, userID)
	if err != nil {
		l.Errorw("msg", "Failed to get read receipts", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GetConversationSequence retrieves conversation sequence
func (h *MessageHandler) GetConversationSequence(ctx context.Context, req *messagev1.GetConversationSequenceRequest) (*messagev1.GetConversationSequenceResponse, error) {
	l := h.log.WithContext(ctx)
	l.Infow("msg", "GetConversationSequence called", "conversationId", req.ConversationId)

	// Parameter validation
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}

	seq, err := h.svc.GetConversationSequence(ctx, req.ConversationId)
	if err != nil {
		l.Errorw("msg", "Failed to get conversation sequence", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return &messagev1.GetConversationSequenceResponse{CurrentSeq: seq}, nil
}

// SearchMessages searches messages
func (h *MessageHandler) SearchMessages(ctx context.Context, req *messagev1.SearchMessagesRequest) (*messagev1.SearchMessagesResponse, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "SearchMessages called",
		"userId", userID,
		"keyword", req.Keyword)

	// Parameter validation
	if userID == "" {
		return nil, errors.BadRequest(ctx, "x-user-id metadata is required")
	}
	if req.Keyword == "" {
		return nil, errors.BadRequest(ctx, "keyword is required")
	}
	if req.ConversationId == nil || *req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}

	resp, err := h.svc.SearchMessages(ctx, userID, req)
	if err != nil {
		l.Errorw("msg", "Failed to search messages", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// SendTyping sends typing status
func (h *MessageHandler) SendTyping(ctx context.Context, req *messagev1.SendTypingRequest) (*empty.Empty, error) {
	l := h.log.WithContext(ctx)
	userID := md.MustGetUserID(ctx)
	l.Infow("msg", "SendTyping called",
		"fromUserId", userID,
		"conversationId", req.ConversationId,
		"typing", req.Typing)

	if userID == "" {
		return nil, errors.BadRequest(ctx, "from_user_id is required")
	}
	if req.ConversationId == "" {
		return nil, errors.BadRequest(ctx, "conversation_id is required")
	}

	if err := h.svc.SendTyping(ctx, userID, req); err != nil {
		l.Errorw("msg", "Failed to send typing status", "error", err)
		return nil, errors.ConvertError(ctx, err)
	}

	return &empty.Empty{}, nil
}
