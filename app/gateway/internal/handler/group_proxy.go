package handler

import (
	"context"

	groupv1 "flamingo/api/group/v1"

	empty "github.com/golang/protobuf/ptypes/empty"
)

// groupProxy implements GroupServiceHTTPServer by forwarding to the gRPC client.
type groupProxy struct {
	client groupv1.GroupServiceClient
}

func (p *groupProxy) CreateGroup(ctx context.Context, req *groupv1.CreateGroupRequest) (*groupv1.CreateGroupResponse, error) {
	return p.client.CreateGroup(ctx, req)
}

func (p *groupProxy) JoinGroupByQRCode(ctx context.Context, req *groupv1.JoinGroupByQRCodeRequest) (*groupv1.JoinGroupByQRCodeResponse, error) {
	return p.client.JoinGroupByQRCode(ctx, req)
}

func (p *groupProxy) GetGroupInfo(ctx context.Context, req *groupv1.GetGroupInfoRequest) (*groupv1.GetGroupInfoResponse, error) {
	return p.client.GetGroupInfo(ctx, req)
}

func (p *groupProxy) UpdateGroup(ctx context.Context, req *groupv1.UpdateGroupRequest) (*empty.Empty, error) {
	return p.client.UpdateGroup(ctx, req)
}

func (p *groupProxy) DisbandGroup(ctx context.Context, req *groupv1.DissolveGroupRequest) (*empty.Empty, error) {
	return p.client.DisbandGroup(ctx, req)
}

func (p *groupProxy) GetUserGroups(ctx context.Context, req *groupv1.GetUserGroupsRequest) (*groupv1.GetUserGroupsResponse, error) {
	return p.client.GetUserGroups(ctx, req)
}

func (p *groupProxy) GetGroupMembers(ctx context.Context, req *groupv1.GetGroupMembersRequest) (*groupv1.GetGroupMembersResponse, error) {
	return p.client.GetGroupMembers(ctx, req)
}

func (p *groupProxy) InviteMembers(ctx context.Context, req *groupv1.InviteMembersRequest) (*empty.Empty, error) {
	return p.client.InviteMembers(ctx, req)
}

func (p *groupProxy) RemoveMember(ctx context.Context, req *groupv1.RemoveMemberRequest) (*empty.Empty, error) {
	return p.client.RemoveMember(ctx, req)
}

func (p *groupProxy) MuteMember(ctx context.Context, req *groupv1.MuteMemberRequest) (*empty.Empty, error) {
	return p.client.MuteMember(ctx, req)
}

func (p *groupProxy) UnmuteMember(ctx context.Context, req *groupv1.UnmuteMemberRequest) (*empty.Empty, error) {
	return p.client.UnmuteMember(ctx, req)
}

func (p *groupProxy) UpdateMemberRole(ctx context.Context, req *groupv1.UpdateMemberRoleRequest) (*empty.Empty, error) {
	return p.client.UpdateMemberRole(ctx, req)
}

func (p *groupProxy) UpdateMemberNickname(ctx context.Context, req *groupv1.UpdateMemberNicknameRequest) (*empty.Empty, error) {
	return p.client.UpdateMemberNickname(ctx, req)
}

func (p *groupProxy) UpdateMemberRemark(ctx context.Context, req *groupv1.UpdateMemberRemarkRequest) (*empty.Empty, error) {
	return p.client.UpdateMemberRemark(ctx, req)
}

func (p *groupProxy) QuitGroup(ctx context.Context, req *groupv1.QuitGroupRequest) (*empty.Empty, error) {
	return p.client.QuitGroup(ctx, req)
}

func (p *groupProxy) TransferOwnership(ctx context.Context, req *groupv1.TransferOwnershipRequest) (*empty.Empty, error) {
	return p.client.TransferOwnership(ctx, req)
}

func (p *groupProxy) GetGroupSettings(ctx context.Context, req *groupv1.GetGroupSettingsRequest) (*groupv1.GetGroupSettingsResponse, error) {
	return p.client.GetGroupSettings(ctx, req)
}

func (p *groupProxy) UpdateGroupSettings(ctx context.Context, req *groupv1.UpdateGroupSettingsRequest) (*empty.Empty, error) {
	return p.client.UpdateGroupSettings(ctx, req)
}

func (p *groupProxy) SetGroupMute(ctx context.Context, req *groupv1.SetGroupMuteRequest) (*empty.Empty, error) {
	return p.client.SetGroupMute(ctx, req)
}

func (p *groupProxy) PinGroupMessage(ctx context.Context, req *groupv1.PinGroupMessageRequest) (*empty.Empty, error) {
	return p.client.PinGroupMessage(ctx, req)
}

func (p *groupProxy) UnpinGroupMessage(ctx context.Context, req *groupv1.UnpinGroupMessageRequest) (*empty.Empty, error) {
	return p.client.UnpinGroupMessage(ctx, req)
}

func (p *groupProxy) GetPinnedMessages(ctx context.Context, req *groupv1.GetPinnedMessagesRequest) (*groupv1.GetPinnedMessagesResponse, error) {
	return p.client.GetPinnedMessages(ctx, req)
}

func (p *groupProxy) GetGroupQRCode(ctx context.Context, req *groupv1.GetGroupQRCodeRequest) (*groupv1.GetGroupQRCodeResponse, error) {
	return p.client.GetGroupQRCode(ctx, req)
}

func (p *groupProxy) RefreshGroupQRCode(ctx context.Context, req *groupv1.RefreshGroupQRCodeRequest) (*groupv1.GetGroupQRCodeResponse, error) {
	return p.client.RefreshGroupQRCode(ctx, req)
}

func (p *groupProxy) JoinGroup(ctx context.Context, req *groupv1.JoinGroupRequest) (*groupv1.JoinGroupResponse, error) {
	return p.client.JoinGroup(ctx, req)
}

func (p *groupProxy) GetJoinRequests(ctx context.Context, req *groupv1.GetJoinRequestsRequest) (*groupv1.GetJoinRequestsResponse, error) {
	return p.client.GetJoinRequests(ctx, req)
}

func (p *groupProxy) HandleJoinRequest(ctx context.Context, req *groupv1.HandleJoinRequestRequest) (*empty.Empty, error) {
	return p.client.HandleJoinRequest(ctx, req)
}
