package service

import (
	"context"
	"encoding/json"
	"time"

	conversationv1 "flamingo/api/conversation/v1"
	friendv1 "flamingo/api/friend/v1"
	groupv1 "flamingo/api/group/v1"
	messagev1 "flamingo/api/message/v1"

	confpkg "flamingo/pkg/config"

	"flamingo/app/message/internal/handler"
	"flamingo/app/message/internal/model"
	"flamingo/pkg/broker"
	"flamingo/pkg/errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

const (
	defaultAnchorWindowLimit = 20
	maxAnchorWindowLimit     = 100
)

// messageServiceImpl message service implementation
type messageServiceImpl struct {
	typingConfig        confpkg.Typing
	messageRepo         MessageRepository
	readReceiptRepo     ReadReceiptRepository
	sequenceRepo        SequenceRepository
	sendIdempotencyRepo SendIdempotencyRepository
	typingRepo          TypingRepository
	conversationClient  conversationv1.ConversationServiceClient
	friendClient        friendv1.FriendServiceClient
	groupClient         groupv1.GroupServiceClient
	broker              broker.Broker
	db                  *gorm.DB
	log                 *log.Helper
}

var _ handler.MessageService = (*messageServiceImpl)(nil)

// NewMessageService creates message service
func NewMessageService(
	typingConfig confpkg.Typing,
	messageRepo MessageRepository,
	readReceiptRepo ReadReceiptRepository,
	sequenceRepo SequenceRepository,
	sendIdempotencyRepo SendIdempotencyRepository,
	typingRepo TypingRepository,
	conversationClient conversationv1.ConversationServiceClient,
	friendClient friendv1.FriendServiceClient,
	groupClient groupv1.GroupServiceClient,
	broker broker.Broker,
	db *gorm.DB,
	logger log.Logger,
) handler.MessageService {
	// if typingConfig.DefaultTTL <= 0 {
	// 	typingConfig.DefaultTTL = 5
	// }
	// if typingConfig.MinTTL <= 0 {
	// 	typingConfig.MinTTL = 3
	// }
	// if typingConfig.MaxTTL <= 0 {
	// 	typingConfig.MaxTTL = 8
	// }
	// if typingConfig.EmitDebounce <= 0 {
	// 	typingConfig.EmitDebounce = 2
	// }
	// if typingConfig.MinTTL > typingConfig.MaxTTL {
	// 	typingConfig.MinTTL = typingConfig.MaxTTL
	// }

	return &messageServiceImpl{
		messageRepo:         messageRepo,
		readReceiptRepo:     readReceiptRepo,
		sequenceRepo:        sequenceRepo,
		sendIdempotencyRepo: sendIdempotencyRepo,
		typingRepo:          typingRepo,
		typingConfig:        typingConfig,
		conversationClient:  conversationClient,
		friendClient:        friendClient,
		groupClient:         groupClient,
		db:                  db,
		log:                 log.NewHelper(logger),
	}
}

// SendMessage sends a message
func (s *messageServiceImpl) SendMessage(ctx context.Context, userID string, req *messagev1.SendMessageRequest) (*messagev1.SendMessageResponse, error) {
	l := s.log.WithContext(ctx)
	conversation, err := s.authorizeSend(ctx, userID, req.ConversationId)
	if err != nil {
		return nil, err
	}

	localID := req.GetLocalId()
	if localID == "" {
		return nil, errors.NewBusiness(errors.CodeParamError, "local_id is required")
	}
	if s.sendIdempotencyRepo == nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "idempotency repo is not initialized")
	}

	var message *model.Message
	created := false

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		messageRepoTx := s.messageRepo.WithTx(tx)
		sequenceRepoTx := s.sequenceRepo.WithTx(tx)
		idempotencyRepoTx := s.sendIdempotencyRepo.WithTx(tx)

		if err := idempotencyRepoTx.CreateIfNotExists(ctx, &model.MessageSendIdempotency{
			SenderID:       userID,
			ConversationID: req.ConversationId,
			LocalID:        localID,
		}); err != nil {
			return err
		}

		idempotencyRecord, err := idempotencyRepoTx.GetForUpdateByKey(ctx, userID, req.ConversationId, localID)
		if err != nil {
			return err
		}

		// Idempotency hit: return existing message
		if idempotencyRecord.MessageID != "" {
			existing, err := messageRepoTx.GetByMessageID(ctx, idempotencyRecord.MessageID)
			if err != nil {
				return err
			}
			message = existing
			return nil
		}

		// Create new message (in same transaction as sequence allocation)
		sequence, err := sequenceRepoTx.IncrementAndGet(ctx, req.ConversationId)
		if err != nil {
			l.Errorw("msg", "Failed to increment sequence", "error", err)
			return errors.NewBusiness(errors.CodeSequenceGenerateFailed, "")
		}

		now := time.Now()
		newMessage := &model.Message{
			MessageID:        uuid.New().String(),
			ConversationID:   req.ConversationId,
			ConversationType: model.ConversationType(conversation.ConversationType),
			TargetID:         conversation.TargetId,
			SenderID:         userID,
			ContentType:      model.ContentType(req.ContentType),
			Content:          req.Content,
			Sequence:         sequence,
			Status:           model.MessageStatusNormal,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if conversation.AutoDeleteDuration > 0 {
			expireTime := now.Add(time.Duration(conversation.AutoDeleteDuration) * time.Second)
			newMessage.AutoDeleteExpireTime = &expireTime
			newMessage.ExpireTime = &expireTime
		}
		if conversation.BurnAfterReading > 0 {
			newMessage.BurnAfterReadingSeconds = conversation.BurnAfterReading
		}
		if req.ReplyTo != nil {
			replyToID := req.ReplyTo.MessageId
			newMessage.ReplyTo = &replyToID
		}
		if len(req.AtUsers) > 0 {
			newMessage.AtUsers = req.AtUsers
		}

		if err := messageRepoTx.Create(ctx, newMessage); err != nil {
			l.Errorw("msg", "Failed to create message", "error", err)
			return errors.NewBusiness(errors.CodeMessageSendFailed, "")
		}

		if err := idempotencyRepoTx.BindMessageID(ctx, userID, req.ConversationId, localID, newMessage.MessageID); err != nil {
			return err
		}

		message = newMessage
		created = true
		return nil
	})
	if err != nil {
		l.Errorw("msg", "Failed to send message in transaction", "error", err)
		if errors.IsBusiness(err) {
			return nil, err
		}
		return nil, errors.NewBusiness(errors.CodeMessageSendFailed, "")
	}

	if message == nil {
		return nil, errors.NewBusiness(errors.CodeMessageSendFailed, "")
	}

	if created {
		if err := s.publishNewMessageNotification(ctx, message, req.AtUsers); err != nil {
			l.Errorw("msg", "Failed to publish message event", "error", err)
		}

		if len(req.AtUsers) > 0 {
			if err := s.publishMentionNotification(message); err != nil {
				l.Errorw("msg", "Failed to publish mention event", "error", err)
			}
		}
	}

	pbMsg := s.modelToProto(message)
	return &messagev1.SendMessageResponse{
		Message: pbMsg,
	}, nil
}

// SendTyping sends typing status
func (s *messageServiceImpl) SendTyping(ctx context.Context, userID string, req *messagev1.SendTypingRequest) error {
	l := s.log.WithContext(ctx)
	if userID == "" || req.ConversationId == "" {
		return errors.NewBusiness(errors.CodeParamError, "from_user_id and conversation_id are required")
	}
	if s.typingRepo == nil {
		return errors.NewBusiness(errors.CodeInternalError, "typing repo is not initialized")
	}

	conversation, err := s.authorizeSend(ctx, userID, req.ConversationId)
	if err != nil {
		return err
	}
	if conversation.ConversationType != conversationv1.ConversationType_CONVERSATION_TYPE_SINGLE {
		return errors.NewBusiness(errors.CodeInvalidOperation, "typing is only supported in single conversation")
	}

	ttl, err := s.resolveTypingTTL(&req.TtlSeconds)
	if err != nil {
		return err
	}
	now := time.Now()

	if !req.Typing {
		if err := s.typingRepo.ClearState(ctx, req.ConversationId, userID); err != nil {
			l.Errorw("msg", "Failed to clear typing state", "conversationID", req.ConversationId, "fromUserID", userID, "error", err)
			return errors.NewBusiness(errors.CodeInternalError, "failed to clear typing state")
		}
		return s.publishTypingNotification(conversation.TargetId, userID, req, false, now)
	}

	if err := s.typingRepo.SetState(ctx, req.ConversationId, userID, ttl); err != nil {
		l.Errorw("msg", "Failed to persist typing state", "conversationID", req.ConversationId, "fromUserID", userID, "error", err)
		return errors.NewBusiness(errors.CodeInternalError, "failed to persist typing state")
	}

	emit, err := s.typingRepo.AcquireEmitToken(ctx, req.ConversationId, userID, s.typingConfig.EmitDebounce.AsDuration())
	if err != nil {
		l.Errorw("msg", "Failed to acquire typing emit token", "conversationID", req.ConversationId, "fromUserID", userID, "error", err)
		return errors.NewBusiness(errors.CodeInternalError, "failed to acquire typing emit token")
	}
	if !emit {
		return nil
	}

	return s.publishTypingNotification(conversation.TargetId, userID, req, true, now.Add(ttl))
}

func (s *messageServiceImpl) authorizeSend(ctx context.Context, senderID, conversationID string) (*conversationv1.Conversation, error) {
	if s.conversationClient == nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "conversation client is not initialized")
	}
	if senderID == "" || conversationID == "" {
		return nil, errors.NewBusiness(errors.CodeParamError, "sender_id and conversation_id are required")
	}

	conversation, err := s.conversationClient.GetConversation(ctx, &conversationv1.GetConversationRequest{
		ConversationId: conversationID,
	})
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeConversationNotFound, "conversation not found")
	}
	if conversation.ConversationType != conversationv1.ConversationType_CONVERSATION_TYPE_SINGLE &&
		conversation.ConversationType != conversationv1.ConversationType_CONVERSATION_TYPE_GROUP {
		return nil, errors.NewBusiness(errors.CodeParamError, "conversation_type must be single or group")
	}

	targetID := conversation.TargetId
	if targetID == "" {
		return nil, errors.NewBusiness(errors.CodeParamError, "target_id is required")
	}

	if conversation.ConversationType == conversationv1.ConversationType_CONVERSATION_TYPE_GROUP {
		if s.groupClient == nil {
			return nil, errors.NewBusiness(errors.CodeInternalError, "group client is not initialized")
		}
		memberResp, err := s.groupClient.IsMember(ctx, &groupv1.IsMemberRequest{
			GroupId: targetID,
			UserId:  senderID,
		})
		if err != nil {
			return nil, errors.NewBusiness(errors.CodeInternalError, "failed to verify group membership")
		}
		if !memberResp.IsMember {
			return nil, errors.NewBusiness(errors.CodeMessagePermissionDenied, "sender is not a group member")
		}
	}
	if conversation.ConversationType == conversationv1.ConversationType_CONVERSATION_TYPE_SINGLE {
		if s.friendClient == nil {
			return nil, errors.NewBusiness(errors.CodeInternalError, "friend client is not initialized")
		}
		blockedResp, err := s.friendClient.IsBlocked(ctx, &friendv1.IsBlockedRequest{
			UserId:       senderID,
			TargetUserId: targetID,
		})
		if err != nil {
			return nil, errors.NewBusiness(errors.CodeInternalError, "failed to verify blacklist")
		}
		if blockedResp.IsBlocked {
			return nil, errors.NewBusiness(errors.CodeUserBlocked, "user blocked")
		}
	}

	return conversation, nil
}

// GetMessages retrieves message list
func (s *messageServiceImpl) GetMessages(ctx context.Context, req *messagev1.GetMessagesRequest) (*messagev1.GetMessagesResponse, error) {
	l := s.log.WithContext(ctx)
	// Parameter validation
	if req.Limit <= 0 {
		req.Limit = 20 // default 20
	}
	if req.Limit > 100 {
		req.Limit = 100 // max 100
	}
	limit := int(req.Limit)

	messages, err := s.messageRepo.GetByConversation(ctx, req.ConversationId, 0, 0, limit+1, false)
	if err != nil {
		l.Errorw("msg", "Failed to get messages", "error", err)
		return nil, errors.NewBusiness(errors.CodeInternalError, "Failed to retrieve messages")
	}

	hasMore := false
	if len(messages) > limit {
		hasMore = true
		messages = messages[:limit]
	}

	// Convert to protobuf messages
	pbMessages := make([]*messagev1.Message, 0, len(messages))
	for _, msg := range messages {
		pbMsg := s.modelToProto(msg)
		pbMessages = append(pbMessages, pbMsg)
	}

	return &messagev1.GetMessagesResponse{
		Messages: pbMessages,
		HasMore:  hasMore,
	}, nil
}

// GetMessagesBefore retrieves messages before anchor message
func (s *messageServiceImpl) GetMessagesBefore(ctx context.Context, userID string, req *messagev1.GetMessagesBeforeRequest) (*messagev1.GetMessagesBeforeResponse, error) {
	if req.ConversationId == "" || req.AnchorMessageId == "" {
		return nil, errors.NewBusiness(errors.CodeParamError, "conversation_id and anchor_message_id are required")
	}
	if err := s.ensureConversationAccessible(ctx, userID, req.ConversationId); err != nil {
		return nil, err
	}

	limit := normalizeAnchorWindowLimit(req.Limit)

	anchor, err := s.getAnchorMessage(ctx, req.ConversationId, req.AnchorMessageId)
	if err != nil {
		return nil, err
	}

	messages, hasMore, err := s.fetchBeforeMessages(ctx, req.ConversationId, anchor.Sequence, limit)
	if err != nil {
		return nil, err
	}

	return &messagev1.GetMessagesBeforeResponse{
		Messages: s.modelsToProto(messages),
		HasMore:  hasMore,
	}, nil
}

// GetMessagesAfter retrieves messages after anchor message
func (s *messageServiceImpl) GetMessagesAfter(ctx context.Context, userID string, req *messagev1.GetMessagesAfterRequest) (*messagev1.GetMessagesAfterResponse, error) {
	if req.ConversationId == "" || req.AnchorMessageId == "" {
		return nil, errors.NewBusiness(errors.CodeParamError, "conversation_id and anchor_message_id are required")
	}
	if err := s.ensureConversationAccessible(ctx, userID, req.ConversationId); err != nil {
		return nil, err
	}

	limit := normalizeAnchorWindowLimit(req.Limit)

	anchor, err := s.getAnchorMessage(ctx, req.ConversationId, req.AnchorMessageId)
	if err != nil {
		return nil, err
	}

	messages, hasMore, err := s.fetchAfterMessages(ctx, req.ConversationId, anchor.Sequence, limit)
	if err != nil {
		return nil, err
	}

	return &messagev1.GetMessagesAfterResponse{
		Messages: s.modelsToProto(messages),
		HasMore:  hasMore,
	}, nil
}

// GetMessagesAroundAnchor retrieves messages around anchor message
func (s *messageServiceImpl) GetMessagesAroundAnchor(ctx context.Context, userID string, req *messagev1.GetMessagesAroundAnchorRequest) (*messagev1.GetMessagesAroundAnchorResponse, error) {
	if req.ConversationId == "" || req.AnchorMessageId == "" {
		return nil, errors.NewBusiness(errors.CodeParamError, "conversation_id and anchor_message_id are required")
	}
	if err := s.ensureConversationAccessible(ctx, userID, req.ConversationId); err != nil {
		return nil, err
	}

	beforeLimit := normalizeAnchorWindowLimit(req.Before)
	afterLimit := normalizeAnchorWindowLimit(req.After)
	includeAnchor := true

	anchor, err := s.getAnchorMessage(ctx, req.ConversationId, req.AnchorMessageId)
	if err != nil {
		return nil, err
	}

	beforeMessages, hasMoreBefore, err := s.fetchBeforeMessages(ctx, req.ConversationId, anchor.Sequence, beforeLimit)
	if err != nil {
		return nil, err
	}
	afterMessages, hasMoreAfter, err := s.fetchAfterMessages(ctx, req.ConversationId, anchor.Sequence, afterLimit)
	if err != nil {
		return nil, err
	}

	allMessages := make([]*messagev1.Message, 0, len(beforeMessages)+len(afterMessages)+1)
	allMessages = append(allMessages, s.modelsToProto(beforeMessages)...)
	if includeAnchor {
		allMessages = append(allMessages, s.modelToProto(anchor))
	}
	allMessages = append(allMessages, s.modelsToProto(afterMessages)...)

	resp := &messagev1.GetMessagesAroundAnchorResponse{
		Messages:      allMessages,
		HasMoreBefore: hasMoreBefore,
		HasMoreAfter:  hasMoreAfter,
	}

	return resp, nil
}

// GetFirstUnreadAnchor retrieves first unread message anchor
func (s *messageServiceImpl) GetFirstUnreadAnchor(ctx context.Context, userID string, req *messagev1.GetFirstUnreadAnchorRequest) (*messagev1.GetFirstUnreadAnchorResponse, error) {
	l := s.log.WithContext(ctx)
	if req.ConversationId == "" {
		return nil, errors.NewBusiness(errors.CodeParamError, "conversation_id is required")
	}
	if err := s.ensureConversationAccessible(ctx, userID, req.ConversationId); err != nil {
		return nil, err
	}

	withContext := false
	beforeLimit := defaultAnchorWindowLimit
	afterLimit := defaultAnchorWindowLimit

	lastReadSeq := int64(0)
	receipt, err := s.readReceiptRepo.GetByConversationAndUser(ctx, req.ConversationId, userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		l.Errorw("msg", "Failed to get read receipt for first unread anchor", "error", err)
		return nil, errors.NewBusiness(errors.CodeInternalError, "Failed to get read receipt")
	}
	if receipt != nil {
		lastReadSeq = receipt.LastReadSeq
	}

	startSeq := lastReadSeq + 1
	unreadMessages, err := s.messageRepo.GetByConversation(ctx, req.ConversationId, startSeq, 0, 1, false)
	if err != nil {
		l.Errorw("msg", "Failed to get first unread anchor", "error", err)
		return nil, errors.NewBusiness(errors.CodeInternalError, "Failed to retrieve first unread anchor")
	}
	if len(unreadMessages) == 0 {
		return &messagev1.GetFirstUnreadAnchorResponse{}, nil
	}

	anchor := unreadMessages[0]
	resp := &messagev1.GetFirstUnreadAnchorResponse{
		FirstUnreadMessageId: &anchor.MessageID,
		UnreadCount:          int64(len(unreadMessages)),
		FirstUnreadMessage:   s.modelToProto(anchor),
	}

	if !withContext {
		return resp, nil
	}

	_ = beforeLimit
	_ = afterLimit

	return resp, nil
}

// GetMessageById retrieves message by ID
func (s *messageServiceImpl) GetMessageById(ctx context.Context, messageID string) (*messagev1.Message, error) {
	l := s.log.WithContext(ctx)
	message, err := s.messageRepo.GetByMessageID(ctx, messageID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeMessageNotFound, "")
		}
		l.Errorw("msg", "Failed to get message", "error", err)
		return nil, errors.NewBusiness(errors.CodeInternalError, "Failed to retrieve message")
	}

	return s.modelToProto(message), nil
}

// RecallMessage recalls a message
func (s *messageServiceImpl) RecallMessage(ctx context.Context, messageID, userID string) error {
	l := s.log.WithContext(ctx)
	// 1. Retrieve message
	message, err := s.messageRepo.GetByMessageID(ctx, messageID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeMessageNotFound, "")
		}
		return errors.NewBusiness(errors.CodeInternalError, "Failed to retrieve message")
	}

	// 2. Validate permission (can only recall own messages)
	if message.SenderID != userID {
		return errors.NewBusiness(errors.CodeMessagePermissionDenied, "Cannot recall other's message")
	}

	// 3. Check recall time limit (within 2 minutes)
	if time.Since(message.CreatedAt) > 2*time.Minute {
		return errors.NewBusiness(errors.CodeMessageRecallTimeLimit, "")
	}

	// 4. Update message status to recalled
	if err := s.messageRepo.UpdateStatus(ctx, messageID, model.MessageStatusRecall); err != nil {
		l.Errorw("msg", "Failed to recall message", "error", err)
		return errors.NewBusiness(errors.CodeMessageRecallFailed, "")
	}

	// 5. If message is pinned in group, auto unpin (failure only logs, does not block recall)
	if err := s.autoUnpinRecalledGroupMessage(ctx, message); err != nil {
		l.Warnw("msg", "Failed to auto-unpin recalled group message", "messageID", message.MessageID, "conversationID", message.ConversationID, "error", err)
	}

	// 6. Publish recall event
	if err := s.publishRecallNotification(ctx, message, userID); err != nil {
		l.Errorw("msg", "Failed to publish recall event", "error", err)
	}

	return nil
}

func (s *messageServiceImpl) autoUnpinRecalledGroupMessage(ctx context.Context, msg *model.Message) error {
	if s.groupClient == nil || msg.ConversationType != model.ConversationTypeGroup {
		return nil
	}

	groupID := msg.TargetID
	if groupID == "" {
		groupID = msg.ConversationID
	}
	if groupID == "" {
		return nil
	}

	_, err := s.groupClient.UnpinGroupMessage(ctx, &groupv1.UnpinGroupMessageRequest{
		GroupId:   groupID,
		MessageId: msg.MessageID,
	})
	return err
}

// DeleteMessage deletes a message
func (s *messageServiceImpl) DeleteMessage(ctx context.Context, messageID, userID string) error {
	l := s.log.WithContext(ctx)
	// 1. Retrieve message
	message, err := s.messageRepo.GetByMessageID(ctx, messageID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeMessageNotFound, "")
		}
		return errors.NewBusiness(errors.CodeInternalError, "Failed to retrieve message")
	}

	// 2. Validate permission (can only delete own messages)
	if message.SenderID != userID {
		return errors.NewBusiness(errors.CodeMessagePermissionDenied, "Cannot delete other's message")
	}

	// 3. Soft delete message
	if err := s.messageRepo.Delete(ctx, messageID); err != nil {
		l.Errorw("msg", "Failed to delete message", "error", err)
		return errors.NewBusiness(errors.CodeMessageDeleteFailed, "Failed to delete message")
	}

	// 4. Publish delete event so other devices/users remove the message from UI
	if err := s.publishDeleteNotification(ctx, message, userID); err != nil {
		l.Errorw("msg", "Failed to publish delete event", "error", err)
		return err
	}

	return nil
}

// publishDeleteNotification publishes delete event for other devices/users.
func (s *messageServiceImpl) publishDeleteNotification(ctx context.Context, msg *model.Message, operatorUserID string) error {
	l := s.log.WithContext(ctx)
	payload := map[string]interface{}{
		"message_id":        msg.MessageID,
		"conversation_id":   msg.ConversationID,
		"conversation_type": msg.ConversationType,
		"target_id":         msg.TargetID,
		"operator_user_id":  operatorUserID,
		"deleted_at":        time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeMessageDeleted,
		operatorUserID,
		broker.PriorityNormal,
	).WithPayload(payload)

	switch msg.ConversationType {
	case model.ConversationTypeSingle:
		receiverSet := map[string]struct{}{}
		if msg.SenderID != "" {
			receiverSet[msg.SenderID] = struct{}{}
		}
		if msg.TargetID != "" {
			receiverSet[msg.TargetID] = struct{}{}
		}
		if len(receiverSet) == 0 {
			l.Warnw("msg", "Skip delete event due to empty receivers", "messageID", msg.MessageID, "conversationID", msg.ConversationID)
			return nil
		}
		receiverIDs := make([]string, 0, len(receiverSet))
		for uid := range receiverSet {
			receiverIDs = append(receiverIDs, uid)
		}
		return s.broker.PublishToUsers(receiverIDs, notif)

	case model.ConversationTypeGroup:
		groupID := msg.TargetID
		if groupID == "" {
			groupID = msg.ConversationID
		}
		memberIDs, err := s.listGroupMemberIDs(ctx, operatorUserID, groupID, nil)
		if err != nil {
			return err
		}
		if len(memberIDs) == 0 {
			return nil
		}
		return s.broker.PublishToUsers(memberIDs, notif)
	}

	return nil
}

// MarkAsRead marks messages as read
func (s *messageServiceImpl) MarkAsRead(ctx context.Context, userID string, req *messagev1.MarkAsReadRequest) error {
	l := s.log.WithContext(ctx)
	if s.conversationClient == nil {
		return errors.NewBusiness(errors.CodeInternalError, "conversation client is not initialized")
	}
	if userID == "" || req.ConversationId == "" {
		return errors.NewBusiness(errors.CodeParamError, "user_id and conversation_id are required")
	}

	conversation, err := s.conversationClient.GetConversation(ctx, &conversationv1.GetConversationRequest{
		ConversationId: req.ConversationId,
	})
	if err != nil {
		return errors.NewBusiness(errors.CodeConversationNotFound, "conversation not found")
	}
	if conversation.ConversationType != conversationv1.ConversationType_CONVERSATION_TYPE_SINGLE &&
		conversation.ConversationType != conversationv1.ConversationType_CONVERSATION_TYPE_GROUP {
		return errors.NewBusiness(errors.CodeParamError, "conversation_type must be single or group")
	}

	effectiveReadSeq := int64(0)
	if req.LastReadSeq != nil {
		effectiveReadSeq = *req.LastReadSeq
	}

	existingReceipt, err := s.readReceiptRepo.GetByConversationAndUser(ctx, req.ConversationId, userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.NewBusiness(errors.CodeInternalError, "Failed to get read receipt")
	}
	if existingReceipt != nil && existingReceipt.LastReadSeq > effectiveReadSeq {
		effectiveReadSeq = existingReceipt.LastReadSeq
	}

	// Create or update read receipt
	receipt := &model.MessageReadReceipt{
		ConversationID:   req.ConversationId,
		ConversationType: model.ConversationType(conversation.ConversationType),
		TargetID:         conversation.TargetId,
		UserID:           userID,
		LastReadSeq:      effectiveReadSeq,
		ReadAt:           time.Now(),
	}

	if err := s.readReceiptRepo.Upsert(ctx, receipt); err != nil {
		return errors.NewBusiness(errors.CodeMarkReadFailed, "")
	}

	// Publish read receipt event (for single chat, notify the other party)
	if conversation.ConversationType == conversationv1.ConversationType_CONVERSATION_TYPE_SINGLE {
		if err := s.publishReadReceiptNotification(receipt); err != nil {
			l.Warnw("msg", "Failed to publish read receipt notification", "conversationID", req.ConversationId, "error", err)
		}
	}

	return nil
}

// MarkMessagesRead marks messages as read by message IDs
func (s *messageServiceImpl) MarkMessagesRead(ctx context.Context, userID string, req *messagev1.MarkMessagesReadRequest) (*messagev1.MarkMessagesReadResponse, error) {
	if userID == "" {
		return nil, errors.NewBusiness(errors.CodeParamError, "user_id is required")
	}
	if req.ConversationId == "" {
		return nil, errors.NewBusiness(errors.CodeParamError, "conversation_id is required")
	}
	if len(req.MessageIds) == 0 {
		return &messagev1.MarkMessagesReadResponse{}, nil
	}
	if err := s.ensureConversationAccessible(ctx, userID, req.ConversationId); err != nil {
		return nil, err
	}

	messageSet := make(map[string]struct{}, len(req.MessageIds))
	messageIDs := make([]string, 0, len(req.MessageIds))
	for _, id := range req.MessageIds {
		if id == "" {
			continue
		}
		if _, exists := messageSet[id]; exists {
			continue
		}
		messageSet[id] = struct{}{}
		messageIDs = append(messageIDs, id)
	}
	if len(messageIDs) == 0 {
		return &messagev1.MarkMessagesReadResponse{}, nil
	}

	var messages []*model.Message
	if err := s.db.WithContext(ctx).
		Model(&model.Message{}).
		Select("message_id", "sequence", "conversation_id", "status").
		Where("conversation_id = ? AND message_id IN ? AND status = ?", req.ConversationId, messageIDs, model.MessageStatusNormal).
		Find(&messages).Error; err != nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to mark messages as read")
	}

	maxSeq := int64(0)
	acceptedSet := make(map[string]struct{}, len(messages))
	for _, msg := range messages {
		acceptedSet[msg.MessageID] = struct{}{}
		if msg.Sequence >= maxSeq {
			maxSeq = msg.Sequence
		}
	}

	acceptedIDs := make([]string, 0, len(acceptedSet))
	for _, id := range messageIDs {
		if _, ok := acceptedSet[id]; ok {
			acceptedIDs = append(acceptedIDs, id)
		}
	}

	if len(acceptedIDs) == 0 {
		return &messagev1.MarkMessagesReadResponse{
			MarkedCount: 0,
			MessageIds:  acceptedIDs,
		}, nil
	}

	currentSeq := int64(0)
	receipt, err := s.readReceiptRepo.GetByConversationAndUser(ctx, req.ConversationId, userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to mark messages as read")
	}
	if receipt != nil {
		currentSeq = receipt.LastReadSeq
	}

	if maxSeq > currentSeq {
		markReq := &messagev1.MarkAsReadRequest{
			ConversationId: req.ConversationId,
			LastReadSeq:    &maxSeq,
		}
		if err := s.MarkAsRead(ctx, userID, markReq); err != nil {
			return nil, err
		}
	}

	return &messagev1.MarkMessagesReadResponse{
		MarkedCount: int32(len(acceptedIDs)),
		MessageIds:  acceptedIDs,
	}, nil
}

// AckReadTriggers acknowledges burn-after-reading triggers
func (s *messageServiceImpl) AckReadTriggers(ctx context.Context, userID string, req *messagev1.AckReadTriggersRequest) (*messagev1.AckReadTriggersResponse, error) {
	l := s.log.WithContext(ctx)
	if userID == "" {
		return nil, errors.NewBusiness(errors.CodeParamError, "user_id is required")
	}
	if len(req.Events) == 0 {
		return &messagev1.AckReadTriggersResponse{}, nil
	}

	messageIDSet := make(map[string]struct{}, len(req.Events))
	messageIDs := make([]string, 0, len(req.Events))
	for _, event := range req.Events {
		if event.GetMessageId() == "" {
			continue
		}
		if _, exists := messageIDSet[event.GetMessageId()]; exists {
			continue
		}
		messageIDSet[event.GetMessageId()] = struct{}{}
		messageIDs = append(messageIDs, event.GetMessageId())
	}
	if len(messageIDs) == 0 {
		return &messagev1.AckReadTriggersResponse{}, nil
	}

	var candidates []*model.Message
	if err := s.db.WithContext(ctx).
		Model(&model.Message{}).
		Select("message_id", "sender_id", "burn_after_reading_seconds", "auto_delete_expire_time", "burn_after_reading_expire_time", "expire_time", "status").
		Where("message_id IN ? AND status = ? AND burn_after_reading_seconds > 0", messageIDs, model.MessageStatusNormal).
		Find(&candidates).Error; err != nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to ack read triggers")
	}

	now := time.Now()
	successIDs := make([]string, 0, len(candidates))
	ignoredSet := make(map[string]struct{}, len(messageIDs))
	for _, id := range messageIDs {
		ignoredSet[id] = struct{}{}
	}

	for _, msg := range candidates {
		// Sender cannot trigger their own burn-after-reading
		if msg.SenderID == userID {
			continue
		}

		successIDs = append(successIDs, msg.MessageID)
		delete(ignoredSet, msg.MessageID)

		burnExpire := now.Add(time.Duration(msg.BurnAfterReadingSeconds) * time.Second)

		updates := map[string]interface{}{
			"updated_at": now,
		}
		shouldUpdate := false

		if msg.BurnAfterReadingExpireTime == nil || burnExpire.Before(*msg.BurnAfterReadingExpireTime) {
			updates["burn_after_reading_expire_time"] = burnExpire
			shouldUpdate = true
		}

		finalExpire := burnExpire
		if msg.AutoDeleteExpireTime != nil && msg.AutoDeleteExpireTime.Before(finalExpire) {
			finalExpire = *msg.AutoDeleteExpireTime
		}
		if msg.ExpireTime == nil || finalExpire.Before(*msg.ExpireTime) {
			updates["expire_time"] = finalExpire
			shouldUpdate = true
		}

		if !shouldUpdate {
			continue
		}

		if err := s.db.WithContext(ctx).
			Model(&model.Message{}).
			Where("message_id = ? AND status = ? AND sender_id <> ? AND burn_after_reading_seconds > 0", msg.MessageID, model.MessageStatusNormal, userID).
			Updates(updates).Error; err != nil {
			l.Errorw("msg", "Failed to update burn expire time", "messageID", msg.MessageID, "error", err)
			return nil, errors.NewBusiness(errors.CodeInternalError, "failed to ack read triggers")
		}
	}

	return &messagev1.AckReadTriggersResponse{
		AckedCount: int32(len(successIDs)),
		MessageIds: successIDs,
	}, nil
}

// GetUnreadCount retrieves unread message count
func (s *messageServiceImpl) GetUnreadCount(ctx context.Context, conversationID, userID string, lastReadSeq *int64) (*messagev1.GetUnreadCountResponse, error) {
	if err := s.ensureConversationAccessible(ctx, userID, conversationID); err != nil {
		return nil, err
	}

	// If lastReadSeq not provided, get from read receipt
	var readSeq int64
	if lastReadSeq != nil {
		readSeq = *lastReadSeq
	} else {
		receipt, err := s.readReceiptRepo.GetByConversationAndUser(ctx, conversationID, userID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeInternalError, "Failed to get read receipt")
		}
		if receipt != nil {
			readSeq = receipt.LastReadSeq
		}
	}

	// Count unread
	unreadCount, err := s.messageRepo.CountUnreadByConversation(ctx, conversationID, readSeq)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeGetUnreadCountFailed, "")
	}

	return &messagev1.GetUnreadCountResponse{
		UnreadCount: unreadCount,
	}, nil
}

// GetReadReceipts retrieves read receipt list
func (s *messageServiceImpl) GetReadReceipts(ctx context.Context, conversationID, userID string) (*messagev1.GetReadReceiptsResponse, error) {
	if err := s.ensureConversationAccessible(ctx, userID, conversationID); err != nil {
		return nil, err
	}

	receipts, err := s.readReceiptRepo.GetByConversation(ctx, conversationID)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "Failed to retrieve read receipts")
	}

	pbReceipts := make([]*messagev1.ReadReceipt, 0, len(receipts))
	for _, receipt := range receipts {
		pbReceipt := &messagev1.ReadReceipt{
			UserId: receipt.UserID,
			ReadAt: timestamppb.New(receipt.ReadAt),
		}
		pbReceipts = append(pbReceipts, pbReceipt)
	}

	return &messagev1.GetReadReceiptsResponse{
		Receipts: pbReceipts,
	}, nil
}

// GetConversationSequence retrieves current conversation sequence
func (s *messageServiceImpl) GetConversationSequence(ctx context.Context, conversationID string) (int64, error) {
	seq, err := s.sequenceRepo.GetCurrentSeq(ctx, conversationID)
	if err != nil {
		return 0, errors.NewBusiness(errors.CodeInternalError, "Failed to get sequence")
	}
	return seq, nil
}

// SearchMessages searches messages
func (s *messageServiceImpl) SearchMessages(ctx context.Context, userID string, req *messagev1.SearchMessagesRequest) (*messagev1.SearchMessagesResponse, error) {
	// Parameter validation
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Search messages
	var conversationID *string
	if req.ConversationId != nil {
		conversationID = req.ConversationId
	}

	// Ensure the user has access to this conversation before searching.
	if conversationID != nil {
		if err := s.ensureConversationAccessible(ctx, userID, *conversationID); err != nil {
			return nil, err
		}
	}

	messages, total, err := s.messageRepo.SearchMessages(ctx, req.Keyword, conversationID, nil, int(req.PageSize), int(req.Page-1))
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeSearchMessageFailed, "")
	}

	// Convert to protobuf messages
	pbMessages := make([]*messagev1.Message, 0, len(messages))
	for _, msg := range messages {
		pbMsg := s.modelToProto(msg)
		pbMessages = append(pbMessages, pbMsg)
	}

	return &messagev1.SearchMessagesResponse{
		Messages: pbMessages,
		Total:    total,
	}, nil
}

func normalizeAnchorWindowLimit(limit int32) int {
	if limit <= 0 {
		return defaultAnchorWindowLimit
	}
	if limit > maxAnchorWindowLimit {
		return maxAnchorWindowLimit
	}
	return int(limit)
}

func (s *messageServiceImpl) getAnchorMessage(ctx context.Context, conversationID, anchorMessageID string) (*model.Message, error) {
	message, err := s.messageRepo.GetByMessageID(ctx, anchorMessageID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeMessageNotFound, "")
		}
		return nil, errors.NewBusiness(errors.CodeInternalError, "Failed to retrieve anchor message")
	}

	if message.ConversationID != conversationID || message.Status != model.MessageStatusNormal {
		return nil, errors.NewBusiness(errors.CodeMessageNotFound, "")
	}

	return message, nil
}

func (s *messageServiceImpl) fetchBeforeMessages(ctx context.Context, conversationID string, anchorSeq int64, limit int) ([]*model.Message, bool, error) {
	if anchorSeq <= 0 {
		return []*model.Message{}, false, nil
	}

	messages, err := s.messageRepo.GetByConversation(ctx, conversationID, 0, anchorSeq-1, limit+1, true)
	if err != nil {
		return nil, false, errors.NewBusiness(errors.CodeInternalError, "Failed to retrieve messages before anchor")
	}

	hasMore := false
	if len(messages) > limit {
		hasMore = true
		messages = messages[:limit]
	}
	reverseModelMessages(messages)

	return messages, hasMore, nil
}

func (s *messageServiceImpl) fetchAfterMessages(ctx context.Context, conversationID string, anchorSeq int64, limit int) ([]*model.Message, bool, error) {
	messages, err := s.messageRepo.GetByConversation(ctx, conversationID, anchorSeq+1, 0, limit+1, false)
	if err != nil {
		return nil, false, errors.NewBusiness(errors.CodeInternalError, "Failed to retrieve messages after anchor")
	}

	hasMore := false
	if len(messages) > limit {
		hasMore = true
		messages = messages[:limit]
	}

	return messages, hasMore, nil
}

func (s *messageServiceImpl) modelsToProto(messages []*model.Message) []*messagev1.Message {
	pbMessages := make([]*messagev1.Message, 0, len(messages))
	for _, msg := range messages {
		pbMessages = append(pbMessages, s.modelToProto(msg))
	}
	return pbMessages
}

func reverseModelMessages(messages []*model.Message) {
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
}

func (s *messageServiceImpl) ensureConversationAccessible(ctx context.Context, userID, conversationID string) error {
	if userID == "" || conversationID == "" {
		return errors.NewBusiness(errors.CodeParamError, "user_id and conversation_id are required")
	}
	if s.conversationClient == nil {
		return errors.NewBusiness(errors.CodeInternalError, "conversation client is not initialized")
	}
	if _, err := s.conversationClient.GetConversation(ctx, &conversationv1.GetConversationRequest{
		ConversationId: conversationID,
	}); err != nil {
		return errors.NewBusiness(errors.CodeConversationNotFound, "conversation not found")
	}
	return nil
}

func (s *messageServiceImpl) resolveTypingTTL(ttlSeconds *int32) (time.Duration, error) {
	var ttl time.Duration
	if ttlSeconds != nil {
		if *ttlSeconds <= 0 {
			return 0, errors.NewBusiness(errors.CodeParamError, "ttl_seconds must be positive")
		}
		ttl = time.Duration(*ttlSeconds) * time.Second
	} else if s.typingConfig.DefaultTTL != nil {
		ttl = s.typingConfig.DefaultTTL.AsDuration()
	} else {
		ttl = 5 * time.Second
	}

	minTTL := 3 * time.Second
	maxTTL := 8 * time.Second
	if s.typingConfig.MinTTL != nil {
		minTTL = s.typingConfig.MinTTL.AsDuration()
	}
	if s.typingConfig.MaxTTL != nil {
		maxTTL = s.typingConfig.MaxTTL.AsDuration()
	}
	if ttl < minTTL || ttl > maxTTL {
		return 0, errors.NewBusiness(errors.CodeParamError, "ttl_seconds is out of range")
	}
	return ttl, nil
}

func (s *messageServiceImpl) publishTypingNotification(toUserID, fromUserID string, req *messagev1.SendTypingRequest, typing bool, expireAt time.Time) error {
	if toUserID == "" {
		s.log.Warnw("msg", "Skip typing event due to empty target user", "conversationID", req.ConversationId, "fromUserID", fromUserID)
		return nil
	}

	payload := map[string]interface{}{
		"conversation_id": req.ConversationId,
		"from_user_id":    fromUserID,
		"typing":          typing,
		"timestamp":       time.Now().Unix(),
	}
	if typing {
		payload["expire_at"] = expireAt.Unix()
	}
	if deviceID := req.GetDeviceId(); deviceID != "" {
		payload["device_id"] = deviceID
	}

	notif := broker.NewNotification(
		broker.TypeMessageTyping,
		fromUserID,
		broker.PriorityLow,
	).WithPayload(payload)

	return s.broker.PublishToUser(toUserID, notif)
}

// modelToProto converts model to protobuf message
func (s *messageServiceImpl) modelToProto(msg *model.Message) *messagev1.Message {
	pbMsg := &messagev1.Message{
		MessageId:      msg.MessageID,
		ConversationId: msg.ConversationID,
		SenderId:       msg.SenderID,
		ContentType:    messagev1.ContentType(msg.ContentType),
		Content:        msg.Content,
		Seq:            msg.Sequence,
		Timestamp:      timestamppb.New(msg.CreatedAt),
	}

	if msg.ReplyTo != nil {
		pbMsg.ReplyTo = &messagev1.Message{
			MessageId: *msg.ReplyTo,
		}
	}

	if len(msg.AtUsers) > 0 {
		pbMsg.AtUsers = msg.AtUsers
	}

	return pbMsg
}

// publishNewMessageNotification publishes new message event
func (s *messageServiceImpl) publishNewMessageNotification(ctx context.Context, msg *model.Message, atUsers []string) error {
	l := s.log.WithContext(ctx)
	// Parse content to get message preview
	contentPreview := s.getContentPreview(msg.Content, msg.ContentType)

	payload := map[string]interface{}{
		"message_id":        msg.MessageID,
		"conversation_id":   msg.ConversationID,
		"conversation_type": msg.ConversationType,
		"target_id":         msg.TargetID,
		"from_user_id":      msg.SenderID,
		"content_type":      msg.ContentType,
		"content":           contentPreview,
		"sent_at":           msg.CreatedAt.Unix(),
		"seq":               msg.Sequence,
	}

	notif := broker.NewNotification(
		broker.TypeMessageNew,
		msg.SenderID,
		broker.PriorityNormal,
	).WithPayload(payload)

	// Determine push method based on conversation type
	switch msg.ConversationType {
	case model.ConversationTypeSingle:
		if msg.TargetID == "" {
			l.Warnw("msg", "Skip single message event due to empty target_id", "messageID", msg.MessageID, "conversationID", msg.ConversationID)
			return nil
		}
		return s.broker.PublishToUser(msg.TargetID, notif)
	case model.ConversationTypeGroup:
		groupID := msg.TargetID
		if groupID == "" {
			groupID = msg.ConversationID
		}
		excludedUserIDs := map[string]struct{}{msg.SenderID: {}}
		memberIDs, err := s.listGroupMemberIDs(ctx, msg.SenderID, groupID, excludedUserIDs)
		if err != nil {
			return err
		}
		if len(memberIDs) == 0 {
			return nil
		}
		return s.broker.PublishToUsers(memberIDs, notif)
	}

	return nil
}

func (s *messageServiceImpl) listGroupMemberIDs(ctx context.Context, operatorUserID, groupID string, excludedUserIDs map[string]struct{}) ([]string, error) {
	if s.groupClient == nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "group client is not initialized")
	}

	const pageSizeValue int32 = 100
	pageValue := int32(1)
	pageSize := pageSizeValue
	memberSet := make(map[string]struct{})

	for {
		resp, err := s.groupClient.GetGroupMembers(ctx, &groupv1.GetGroupMembersRequest{
			GroupId:  groupID,
			UserId:   operatorUserID,
			Page:     &pageValue,
			PageSize: &pageSize,
		})
		if err != nil {
			return nil, errors.NewBusiness(errors.CodeInternalError, "failed to load group members")
		}

		if len(resp.Members) == 0 {
			break
		}

		for _, member := range resp.Members {
			if member.UserId == "" {
				continue
			}
			if _, excluded := excludedUserIDs[member.UserId]; excluded {
				continue
			}
			memberSet[member.UserId] = struct{}{}
		}

		if int64(pageValue)*int64(pageSizeValue) >= resp.Total {
			break
		}
		pageValue++
	}

	memberIDs := make([]string, 0, len(memberSet))
	for userID := range memberSet {
		memberIDs = append(memberIDs, userID)
	}
	return memberIDs, nil
}

// publishMentionNotification publishes @mention event
func (s *messageServiceImpl) publishMentionNotification(msg *model.Message) error {
	if len(msg.AtUsers) == 0 {
		return nil
	}

	l := s.log.WithContext(context.Background())

	contentPreview := s.getContentPreview(msg.Content, msg.ContentType)
	groupID := msg.TargetID
	if groupID == "" {
		groupID = msg.ConversationID
	}

	for _, userID := range msg.AtUsers {
		payload := map[string]interface{}{
			"message_id":   msg.MessageID,
			"group_id":     groupID,
			"from_user_id": msg.SenderID,
			"content":      contentPreview,
			"mention_type": "single",
			"sent_at":      msg.CreatedAt.Unix(),
		}

		notif := broker.NewNotification(
			broker.TypeMessageMentioned,
			msg.SenderID,
			broker.PriorityHigh,
		).WithPayload(payload)

		if err := s.broker.PublishToUser(userID, notif); err != nil {
			l.Warnw("msg", "Failed to publish @mention notification", "userID", userID, "error", err)
		}
	}

	return nil
}

// publishRecallNotification publishes recall event
func (s *messageServiceImpl) publishRecallNotification(ctx context.Context, msg *model.Message, operatorUserID string) error {
	l := s.log.WithContext(ctx)
	payload := map[string]interface{}{
		"message_id":        msg.MessageID,
		"conversation_id":   msg.ConversationID,
		"conversation_type": msg.ConversationType,
		"target_id":         msg.TargetID,
		"operator_user_id":  operatorUserID,
		"recalled_at":       time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeMessageRecalled,
		operatorUserID,
		broker.PriorityNormal,
	).WithPayload(payload)

	switch msg.ConversationType {
	case model.ConversationTypeSingle:
		receiverSet := map[string]struct{}{}
		if msg.SenderID != "" {
			receiverSet[msg.SenderID] = struct{}{}
		}
		if msg.TargetID != "" {
			receiverSet[msg.TargetID] = struct{}{}
		}
		if len(receiverSet) == 0 {
			l.Warnw("msg", "Skip recall event due to empty receivers", "messageID", msg.MessageID, "conversationID", msg.ConversationID)
			return nil
		}

		receiverIDs := make([]string, 0, len(receiverSet))
		for userID := range receiverSet {
			receiverIDs = append(receiverIDs, userID)
		}
		return s.broker.PublishToUsers(receiverIDs, notif)

	case model.ConversationTypeGroup:
		groupID := msg.TargetID
		if groupID == "" {
			groupID = msg.ConversationID
		}
		memberIDs, err := s.listGroupMemberIDs(ctx, operatorUserID, groupID, nil)
		if err != nil {
			return err
		}
		if len(memberIDs) == 0 {
			return nil
		}
		return s.broker.PublishToUsers(memberIDs, notif)
	}

	return nil
}

// publishReadReceiptNotification publishes read receipt event
func (s *messageServiceImpl) publishReadReceiptNotification(receipt *model.MessageReadReceipt) error {
	payload := map[string]interface{}{
		"conversation_id":   receipt.ConversationID,
		"conversation_type": receipt.ConversationType,
		"target_id":         receipt.TargetID,
		"reader_user_id":    receipt.UserID,
		"last_read_seq":     receipt.LastReadSeq,
		"read_at":           receipt.ReadAt.Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeMessageReadReceipt,
		receipt.UserID,
		broker.PriorityLow,
	).WithPayload(payload)

	// Single chat read receipt: notify the other party
	if receipt.ConversationType == model.ConversationTypeSingle {
		if receipt.TargetID == "" {
			s.log.Warnw("msg", "Skip read receipt event due to empty target_id", "conversationID", receipt.ConversationID, "readerUserID", receipt.UserID)
			return nil
		}
		return s.broker.PublishToUser(receipt.TargetID, notif)
	}

	return nil
}

// getContentPreview gets content preview
func (s *messageServiceImpl) getContentPreview(content string, contentType model.ContentType) string {
	switch contentType {
	case model.ContentTypeText:
		// Parse JSON to get text content
		var textContent struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(content), &textContent); err == nil {
			if len(textContent.Text) > 100 {
				return textContent.Text[:100] + "..."
			}
			return textContent.Text
		}
		return "[Text Message]"
	case model.ContentTypeImage:
		return "[Image]"
	case model.ContentTypeVideo:
		return "[Video]"
	case model.ContentTypeAudio:
		return "[Voice]"
	case model.ContentTypeFile:
		return "[File]"
	case model.ContentTypeLocation:
		return "[Location]"
	case model.ContentTypeCard:
		return "[Contact]"
	default:
		return "[Message]"
	}
}
