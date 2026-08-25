package handler

import (
	"context"

	conversationv1 "flamingo/api/conversation/v1"

	empty "github.com/golang/protobuf/ptypes/empty"
)

// conversationProxy implements ConversationServiceHTTPServer by forwarding to the gRPC client.
type conversationProxy struct {
	client conversationv1.ConversationServiceClient
}

func (p *conversationProxy) GetConversations(ctx context.Context, req *conversationv1.GetConversationsRequest) (*conversationv1.GetConversationsResponse, error) {
	return p.client.GetConversations(ctx, req)
}

func (p *conversationProxy) GetConversation(ctx context.Context, req *conversationv1.GetConversationRequest) (*conversationv1.Conversation, error) {
	return p.client.GetConversation(ctx, req)
}

func (p *conversationProxy) ClearUnread(ctx context.Context, req *conversationv1.ClearUnreadRequest) (*empty.Empty, error) {
	return p.client.ClearUnread(ctx, req)
}

func (p *conversationProxy) GetTotalUnread(ctx context.Context, req *conversationv1.GetTotalUnreadRequest) (*conversationv1.GetTotalUnreadResponse, error) {
	return p.client.GetTotalUnread(ctx, req)
}

func (p *conversationProxy) DeleteConversation(ctx context.Context, req *conversationv1.DeleteConversationRequest) (*empty.Empty, error) {
	return p.client.DeleteConversation(ctx, req)
}

func (p *conversationProxy) SetPinned(ctx context.Context, req *conversationv1.SetPinnedRequest) (*empty.Empty, error) {
	return p.client.SetPinned(ctx, req)
}

func (p *conversationProxy) SetMuted(ctx context.Context, req *conversationv1.SetMutedRequest) (*empty.Empty, error) {
	return p.client.SetMuted(ctx, req)
}

func (p *conversationProxy) SetBurnAfterReading(ctx context.Context, req *conversationv1.SetBurnAfterReadingRequest) (*empty.Empty, error) {
	return p.client.SetBurnAfterReading(ctx, req)
}

func (p *conversationProxy) SetAutoDelete(ctx context.Context, req *conversationv1.SetAutoDeleteRequest) (*empty.Empty, error) {
	return p.client.SetAutoDelete(ctx, req)
}
