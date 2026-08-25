package handler

import (
	"context"

	friendv1 "flamingo/api/friend/v1"

	empty "github.com/golang/protobuf/ptypes/empty"
)

// friendProxy implements FriendServiceHTTPServer by forwarding to the gRPC client.
type friendProxy struct {
	client friendv1.FriendServiceClient
}

func (p *friendProxy) GetFriendList(ctx context.Context, req *friendv1.GetFriendListRequest) (*friendv1.GetFriendListResponse, error) {
	return p.client.GetFriendList(ctx, req)
}

func (p *friendProxy) GetFriendRequests(ctx context.Context, req *friendv1.GetFriendRequestsRequest) (*friendv1.GetFriendRequestsResponse, error) {
	return p.client.GetFriendRequests(ctx, req)
}

func (p *friendProxy) SendFriendRequest(ctx context.Context, req *friendv1.SendFriendRequestRequest) (*empty.Empty, error) {
	return p.client.SendFriendRequest(ctx, req)
}

func (p *friendProxy) HandleFriendRequest(ctx context.Context, req *friendv1.HandleFriendRequestRequest) (*empty.Empty, error) {
	return p.client.HandleFriendRequest(ctx, req)
}

func (p *friendProxy) DeleteFriend(ctx context.Context, req *friendv1.DeleteFriendRequest) (*empty.Empty, error) {
	return p.client.DeleteFriend(ctx, req)
}

func (p *friendProxy) UpdateRemark(ctx context.Context, req *friendv1.UpdateRemarkRequest) (*empty.Empty, error) {
	return p.client.UpdateRemark(ctx, req)
}

func (p *friendProxy) GetBlacklist(ctx context.Context, req *friendv1.GetBlacklistRequest) (*friendv1.GetBlacklistResponse, error) {
	return p.client.GetBlacklist(ctx, req)
}

func (p *friendProxy) AddToBlacklist(ctx context.Context, req *friendv1.AddToBlacklistRequest) (*empty.Empty, error) {
	return p.client.AddToBlacklist(ctx, req)
}

func (p *friendProxy) RemoveFromBlacklist(ctx context.Context, req *friendv1.RemoveFromBlacklistRequest) (*empty.Empty, error) {
	return p.client.RemoveFromBlacklist(ctx, req)
}
