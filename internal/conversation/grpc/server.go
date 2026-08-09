package grpc

import (
	"context"

	commonpb "github.com/anychat/server/api/proto/common"
	conversationpb "github.com/anychat/server/api/proto/conversation"
	"github.com/anychat/server/internal/conversation/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the Conversation gRPC server
type Server struct {
	conversationpb.UnimplementedConversationServiceServer
	svc service.ConversationService
}

// NewServer creates a new gRPC server
func NewServer(service service.ConversationService) *Server {
	return &Server{svc: service}
}

// GetConversations retrieves the list of user conversations
func (s *Server) GetConversations(ctx context.Context, req *conversationpb.GetConversationsRequest) (*conversationpb.GetConversationsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	resp, err := s.svc.GetConversations(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

// GetConversation retrieves a single conversation
func (s *Server) GetConversation(ctx context.Context, req *conversationpb.GetConversationRequest) (*conversationpb.Conversation, error) {
	if req.UserId == "" || req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and conversation_id are required")
	}
	conversation, err := s.svc.GetConversation(ctx, req.UserId, req.ConversationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return conversation, nil
}

// CreateOrUpdateConversation creates or updates a conversation
func (s *Server) CreateOrUpdateConversation(ctx context.Context, req *conversationpb.CreateOrUpdateConversationRequest) (*conversationpb.Conversation, error) {
	if req.UserId == "" || req.TargetId == "" || req.ConversationType == conversationpb.ConversationType_CONVERSATION_TYPE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "user_id, target_id and conversation_type are required")
	}
	conversation, err := s.svc.CreateOrUpdateConversation(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return conversation, nil
}

// DeleteConversation deletes a conversation
func (s *Server) DeleteConversation(ctx context.Context, req *conversationpb.DeleteConversationRequest) (*commonpb.Empty, error) {
	if req.UserId == "" || req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and conversation_id are required")
	}
	if err := s.svc.DeleteConversation(ctx, req.UserId, req.ConversationId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.Empty{}, nil
}

// SetPinned sets pinned status
func (s *Server) SetPinned(ctx context.Context, req *conversationpb.SetPinnedRequest) (*commonpb.Empty, error) {
	if req.UserId == "" || req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and conversation_id are required")
	}
	if err := s.svc.SetPinned(ctx, req.UserId, req.ConversationId, req.Pinned); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.Empty{}, nil
}

// SetMuted sets muted status
func (s *Server) SetMuted(ctx context.Context, req *conversationpb.SetMutedRequest) (*commonpb.Empty, error) {
	if req.UserId == "" || req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and conversation_id are required")
	}
	if err := s.svc.SetMuted(ctx, req.UserId, req.ConversationId, req.Muted); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.Empty{}, nil
}

// ClearUnread clears unread count
func (s *Server) ClearUnread(ctx context.Context, req *conversationpb.ClearUnreadRequest) (*commonpb.Empty, error) {
	if req.UserId == "" || req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and conversation_id are required")
	}
	if err := s.svc.ClearUnread(ctx, req.UserId, req.ConversationId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.Empty{}, nil
}

// GetTotalUnread gets total unread count
func (s *Server) GetTotalUnread(ctx context.Context, req *conversationpb.GetTotalUnreadRequest) (*conversationpb.GetTotalUnreadResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	total, err := s.svc.GetTotalUnread(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &conversationpb.GetTotalUnreadResponse{TotalUnread: total}, nil
}

// IncrUnread increments unread count
func (s *Server) IncrUnread(ctx context.Context, req *conversationpb.IncrUnreadRequest) (*commonpb.Empty, error) {
	if req.UserId == "" || req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and conversation_id are required")
	}
	if err := s.svc.IncrUnread(ctx, req.UserId, req.ConversationId, req.Count); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.Empty{}, nil
}

// SetBurnAfterReading sets burn after reading
func (s *Server) SetBurnAfterReading(ctx context.Context, req *conversationpb.SetBurnAfterReadingRequest) (*commonpb.Empty, error) {
	if req.UserId == "" || req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and conversation_id are required")
	}
	if err := s.svc.SetBurnAfterReading(ctx, req.UserId, req.ConversationId, req.Duration); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.Empty{}, nil
}

// SetAutoDelete sets auto delete
func (s *Server) SetAutoDelete(ctx context.Context, req *conversationpb.SetAutoDeleteRequest) (*commonpb.Empty, error) {
	if req.UserId == "" || req.ConversationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and conversation_id are required")
	}
	if err := s.svc.SetAutoDelete(ctx, req.UserId, req.ConversationId, req.Duration); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.Empty{}, nil
}
