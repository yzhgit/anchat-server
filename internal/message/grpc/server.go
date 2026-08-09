package grpc

import (
	"context"
	stderrors "errors"

	commonpb "github.com/anychat/server/api/proto/common"
	messagepb "github.com/anychat/server/api/proto/message"
	"github.com/anychat/server/internal/message/service"
	pkgerrors "github.com/anychat/server/pkg/errors"
	"github.com/anychat/server/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const operatorUserIDMetadataKey = "x-user-id"

// Server Message gRPC server
type Server struct {
	messagepb.UnimplementedMessageServiceServer
	svc service.MessageService
}

// NewServer creates gRPC server
func NewServer(service service.MessageService) *Server {
	return &Server{
		svc: service,
	}
}

// SendMessage sends a message
func (s *Server) SendMessage(ctx context.Context, req *messagepb.SendMessageRequest) (*messagepb.SendMessageResponse, error) {
	// Parameter validation
	if req.SenderId == "" {
		return nil, status.Error(codes.InvalidArgument, "sender_id is required")
	}
	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}
	if req.ContentType == messagepb.ContentType_CONTENT_TYPE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "content_type is required")
	}
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}
	if req.GetLocalId() == "" {
		return nil, status.Error(codes.InvalidArgument, "local_id is required")
	}

	resp, err := s.svc.SendMessage(ctx, req)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

// SendTyping sends typing status
func (s *Server) SendTyping(ctx context.Context, req *messagepb.SendTypingRequest) (*commonpb.Empty, error) {
	logger.Info("SendTyping called",
		zap.String("fromUserId", req.FromUserId),
		zap.String("conversationId", req.ConversationId),
		zap.Bool("typing", req.Typing))

	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}
	if req.FromUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "from_user_id is required")
	}

	if err := s.svc.SendTyping(ctx, req); err != nil {
		return nil, toStatusError(err)
	}

	return &commonpb.Empty{}, nil
}

// GetMessages retrieves message list
func (s *Server) GetMessages(ctx context.Context, req *messagepb.GetMessagesRequest) (*messagepb.GetMessagesResponse, error) {
	logger.Info("GetMessages called",
		zap.String("conversationId", req.ConversationId),
		zap.Int32("limit", req.Limit))

	// Parameter validation
	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}

	resp, err := s.svc.GetMessages(ctx, req)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

// GetMessagesBefore retrieves messages before anchor message
func (s *Server) GetMessagesBefore(ctx context.Context, req *messagepb.GetMessagesBeforeRequest) (*messagepb.GetMessagesBeforeResponse, error) {
	operatorUserID := getOperatorUserID(ctx)

	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}
	if req.AnchorMessageId == "" {
		return nil, status.Error(codes.InvalidArgument, "anchor_message_id is required")
	}
	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}

	resp, err := s.svc.GetMessagesBefore(ctx, operatorUserID, req)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

// GetMessagesAfter retrieves messages after anchor message
func (s *Server) GetMessagesAfter(ctx context.Context, req *messagepb.GetMessagesAfterRequest) (*messagepb.GetMessagesAfterResponse, error) {
	operatorUserID := getOperatorUserID(ctx)

	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}
	if req.AnchorMessageId == "" {
		return nil, status.Error(codes.InvalidArgument, "anchor_message_id is required")
	}
	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}

	resp, err := s.svc.GetMessagesAfter(ctx, operatorUserID, req)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

// GetMessagesAroundAnchor retrieves messages around anchor message
func (s *Server) GetMessagesAroundAnchor(ctx context.Context, req *messagepb.GetMessagesAroundAnchorRequest) (*messagepb.GetMessagesAroundAnchorResponse, error) {
	operatorUserID := getOperatorUserID(ctx)

	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}
	if req.AnchorMessageId == "" {
		return nil, status.Error(codes.InvalidArgument, "anchor_message_id is required")
	}
	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}

	resp, err := s.svc.GetMessagesAroundAnchor(ctx, operatorUserID, req)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

// GetFirstUnreadAnchor retrieves first unread message anchor
func (s *Server) GetFirstUnreadAnchor(ctx context.Context, req *messagepb.GetFirstUnreadAnchorRequest) (*messagepb.GetFirstUnreadAnchorResponse, error) {
	operatorUserID := getOperatorUserID(ctx)

	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}
	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}

	resp, err := s.svc.GetFirstUnreadAnchor(ctx, operatorUserID, req)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

// GetMessageById retrieves message by ID
func (s *Server) GetMessageById(ctx context.Context, req *messagepb.GetMessageByIdRequest) (*messagepb.Message, error) {
	// Parameter validation
	if req.MessageId == "" {
		return nil, status.Error(codes.InvalidArgument, "message_id is required")
	}

	msg, err := s.svc.GetMessageById(ctx, req.MessageId)
	if err != nil {
		return nil, toStatusError(err)
	}

	return msg, nil
}

// RecallMessage recalls a message
func (s *Server) RecallMessage(ctx context.Context, req *messagepb.RecallMessageRequest) (*commonpb.Empty, error) {
	operatorUserID := getOperatorUserID(ctx)

	// Parameter validation
	if req.MessageId == "" {
		return nil, status.Error(codes.InvalidArgument, "message_id is required")
	}
	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}

	err := s.svc.RecallMessage(ctx, req.MessageId, operatorUserID)
	if err != nil {
		return nil, toStatusError(err)
	}

	return &commonpb.Empty{}, nil
}

// DeleteMessage deletes a message
func (s *Server) DeleteMessage(ctx context.Context, req *messagepb.DeleteMessageRequest) (*commonpb.Empty, error) {
	operatorUserID := getOperatorUserID(ctx)

	// Parameter validation
	if req.MessageId == "" {
		return nil, status.Error(codes.InvalidArgument, "message_id is required")
	}
	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}

	err := s.svc.DeleteMessage(ctx, req.MessageId, operatorUserID)
	if err != nil {
		return nil, toStatusError(err)
	}

	return &commonpb.Empty{}, nil
}

func getOperatorUserID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(operatorUserIDMetadataKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// MarkAsRead marks messages as read
func (s *Server) MarkAsRead(ctx context.Context, req *messagepb.MarkAsReadRequest) (*commonpb.Empty, error) {
	operatorUserID := getOperatorUserID(ctx)

	// Parameter validation
	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}
	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}

	err := s.svc.MarkAsRead(ctx, operatorUserID, req)
	if err != nil {
		return nil, toStatusError(err)
	}

	return &commonpb.Empty{}, nil
}

// MarkMessagesRead marks messages as read by message IDs
func (s *Server) MarkMessagesRead(ctx context.Context, req *messagepb.MarkMessagesReadRequest) (*messagepb.MarkMessagesReadResponse, error) {
	operatorUserID := getOperatorUserID(ctx)

	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}
	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}
	if len(req.MessageIds) == 0 {
		return &messagepb.MarkMessagesReadResponse{}, nil
	}

	resp, err := s.svc.MarkMessagesRead(ctx, operatorUserID, req)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

// AckReadTriggers acknowledges burn-after-reading triggers
func (s *Server) AckReadTriggers(ctx context.Context, req *messagepb.AckReadTriggersRequest) (*messagepb.AckReadTriggersResponse, error) {
	operatorUserID := getOperatorUserID(ctx)

	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}
	if len(req.Events) == 0 {
		return &messagepb.AckReadTriggersResponse{}, nil
	}

	resp, err := s.svc.AckReadTriggers(ctx, operatorUserID, req)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

// GetUnreadCount retrieves unread message count
func (s *Server) GetUnreadCount(ctx context.Context, req *messagepb.GetUnreadCountRequest) (*messagepb.GetUnreadCountResponse, error) {
	operatorUserID := getOperatorUserID(ctx)

	// Parameter validation
	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}
	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}

	resp, err := s.svc.GetUnreadCount(ctx, req.ConversationId, operatorUserID, req.LastReadSeq)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

// GetReadReceipts retrieves read receipts
func (s *Server) GetReadReceipts(ctx context.Context, req *messagepb.GetReadReceiptsRequest) (*messagepb.GetReadReceiptsResponse, error) {
	operatorUserID := getOperatorUserID(ctx)

	// Parameter validation
	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}

	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}

	resp, err := s.svc.GetReadReceipts(ctx, req.ConversationId, operatorUserID)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

// GetConversationSequence retrieves conversation sequence
func (s *Server) GetConversationSequence(ctx context.Context, req *messagepb.GetConversationSequenceRequest) (*messagepb.GetConversationSequenceResponse, error) {
	// Parameter validation
	if req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}

	seq, err := s.svc.GetConversationSequence(ctx, req.ConversationId)
	if err != nil {
		return nil, toStatusError(err)
	}

	return &messagepb.GetConversationSequenceResponse{
		CurrentSeq: seq,
	}, nil
}

// SearchMessages searches messages
func (s *Server) SearchMessages(ctx context.Context, req *messagepb.SearchMessagesRequest) (*messagepb.SearchMessagesResponse, error) {
	operatorUserID := getOperatorUserID(ctx)

	// Parameter validation
	if operatorUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "x-user-id metadata is required")
	}
	if req.Keyword == "" {
		return nil, status.Error(codes.InvalidArgument, "keyword is required")
	}
	if req.ConversationId == nil || *req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}

	resp, err := s.svc.SearchMessages(ctx, operatorUserID, req)
	if err != nil {
		return nil, toStatusError(err)
	}

	return resp, nil
}

func toStatusError(err error) error {
	var bizErr *pkgerrors.Business
	if !stderrors.As(err, &bizErr) {
		return status.Error(codes.Internal, err.Error())
	}

	switch bizErr.Code {
	case pkgerrors.CodeParamError:
		return status.Error(codes.InvalidArgument, bizErr.Message)
	case pkgerrors.CodeConversationNotFound, pkgerrors.CodeMessageNotFound:
		return status.Error(codes.NotFound, bizErr.Message)
	case pkgerrors.CodeMessagePermissionDenied:
		return status.Error(codes.PermissionDenied, bizErr.Message)
	case pkgerrors.CodeUserBlocked:
		return status.Error(codes.PermissionDenied, bizErr.Message)
	case pkgerrors.CodeInvalidOperation:
		return status.Error(codes.FailedPrecondition, bizErr.Message)
	default:
		return status.Error(codes.Internal, bizErr.Message)
	}
}
