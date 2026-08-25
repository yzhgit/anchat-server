package handler

import (
	"context"

	messagev1 "flamingo/api/message/v1"

	empty "github.com/golang/protobuf/ptypes/empty"
)

// messageProxy implements MessageServiceHTTPServer by forwarding to the gRPC client.
type messageProxy struct {
	client messagev1.MessageServiceClient
}

func (p *messageProxy) SendMessage(ctx context.Context, req *messagev1.SendMessageRequest) (*messagev1.SendMessageResponse, error) {
	return p.client.SendMessage(ctx, req)
}

func (p *messageProxy) AckReadTriggers(ctx context.Context, req *messagev1.AckReadTriggersRequest) (*messagev1.AckReadTriggersResponse, error) {
	return p.client.AckReadTriggers(ctx, req)
}

func (p *messageProxy) DeleteMessage(ctx context.Context, req *messagev1.DeleteMessageRequest) (*empty.Empty, error) {
	return p.client.DeleteMessage(ctx, req)
}

func (p *messageProxy) GetConversationSequence(ctx context.Context, req *messagev1.GetConversationSequenceRequest) (*messagev1.GetConversationSequenceResponse, error) {
	return p.client.GetConversationSequence(ctx, req)
}

func (p *messageProxy) GetFirstUnreadAnchor(ctx context.Context, req *messagev1.GetFirstUnreadAnchorRequest) (*messagev1.GetFirstUnreadAnchorResponse, error) {
	return p.client.GetFirstUnreadAnchor(ctx, req)
}

func (p *messageProxy) GetMessageById(ctx context.Context, req *messagev1.GetMessageByIdRequest) (*messagev1.Message, error) {
	return p.client.GetMessageById(ctx, req)
}

func (p *messageProxy) GetMessages(ctx context.Context, req *messagev1.GetMessagesRequest) (*messagev1.GetMessagesResponse, error) {
	return p.client.GetMessages(ctx, req)
}

func (p *messageProxy) GetMessagesAfter(ctx context.Context, req *messagev1.GetMessagesAfterRequest) (*messagev1.GetMessagesAfterResponse, error) {
	return p.client.GetMessagesAfter(ctx, req)
}

func (p *messageProxy) GetMessagesAroundAnchor(ctx context.Context, req *messagev1.GetMessagesAroundAnchorRequest) (*messagev1.GetMessagesAroundAnchorResponse, error) {
	return p.client.GetMessagesAroundAnchor(ctx, req)
}

func (p *messageProxy) GetMessagesBefore(ctx context.Context, req *messagev1.GetMessagesBeforeRequest) (*messagev1.GetMessagesBeforeResponse, error) {
	return p.client.GetMessagesBefore(ctx, req)
}

func (p *messageProxy) GetReadReceipts(ctx context.Context, req *messagev1.GetReadReceiptsRequest) (*messagev1.GetReadReceiptsResponse, error) {
	return p.client.GetReadReceipts(ctx, req)
}

func (p *messageProxy) GetUnreadCount(ctx context.Context, req *messagev1.GetUnreadCountRequest) (*messagev1.GetUnreadCountResponse, error) {
	return p.client.GetUnreadCount(ctx, req)
}

func (p *messageProxy) MarkAsRead(ctx context.Context, req *messagev1.MarkAsReadRequest) (*empty.Empty, error) {
	return p.client.MarkAsRead(ctx, req)
}

func (p *messageProxy) MarkMessagesRead(ctx context.Context, req *messagev1.MarkMessagesReadRequest) (*messagev1.MarkMessagesReadResponse, error) {
	return p.client.MarkMessagesRead(ctx, req)
}

func (p *messageProxy) RecallMessage(ctx context.Context, req *messagev1.RecallMessageRequest) (*empty.Empty, error) {
	return p.client.RecallMessage(ctx, req)
}

func (p *messageProxy) SearchMessages(ctx context.Context, req *messagev1.SearchMessagesRequest) (*messagev1.SearchMessagesResponse, error) {
	return p.client.SearchMessages(ctx, req)
}

func (p *messageProxy) SendTyping(ctx context.Context, req *messagev1.SendTypingRequest) (*empty.Empty, error) {
	return p.client.SendTyping(ctx, req)
}
