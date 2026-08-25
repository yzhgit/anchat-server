package handler

import (
	"context"
	"unicode/utf8"

	groupv1 "flamingo/api/group/v1"

	"flamingo/pkg/errors"
	"flamingo/pkg/md"

	"github.com/go-kratos/kratos/v2/log"
	empty "github.com/golang/protobuf/ptypes/empty"
)

// GroupService represents the group service interface
type GroupService interface {
	// Internal gRPC methods (called by other services)
	IsMember(ctx context.Context, req *groupv1.IsMemberRequest) (*groupv1.IsMemberResponse, error)
	GetGroupInfo(ctx context.Context, req *groupv1.GetGroupInfoRequest) (*groupv1.GetGroupInfoResponse, error)
	GetGroupMembers(ctx context.Context, req *groupv1.GetGroupMembersRequest) (*groupv1.GetGroupMembersResponse, error)
	GetUserGroups(ctx context.Context, req *groupv1.GetUserGroupsRequest) (*groupv1.GetUserGroupsResponse, error)

	// Group management
	CreateGroup(ctx context.Context, userID string, req *groupv1.CreateGroupRequest) (*groupv1.CreateGroupResponse, error)
	UpdateGroup(ctx context.Context, userID string, req *groupv1.UpdateGroupRequest) error
	DisbandGroup(ctx context.Context, userID string, req *groupv1.DissolveGroupRequest) error

	// Member management
	InviteMembers(ctx context.Context, userID string, req *groupv1.InviteMembersRequest) error
	RemoveMember(ctx context.Context, userID string, req *groupv1.RemoveMemberRequest) error
	QuitGroup(ctx context.Context, userID string, req *groupv1.QuitGroupRequest) error
	UpdateMemberRole(ctx context.Context, userID string, req *groupv1.UpdateMemberRoleRequest) error
	UpdateMemberNickname(ctx context.Context, userID string, req *groupv1.UpdateMemberNicknameRequest) error
	TransferOwnership(ctx context.Context, userID string, req *groupv1.TransferOwnershipRequest) error
	MuteMember(ctx context.Context, userID string, req *groupv1.MuteMemberRequest) error
	UnmuteMember(ctx context.Context, userID string, req *groupv1.UnmuteMemberRequest) error
	UpdateMemberRemark(ctx context.Context, userID string, req *groupv1.UpdateMemberRemarkRequest) error

	// Join requests
	JoinGroup(ctx context.Context, userID string, req *groupv1.JoinGroupRequest) (*groupv1.JoinGroupResponse, error)
	HandleJoinRequest(ctx context.Context, userID string, req *groupv1.HandleJoinRequestRequest) error
	GetJoinRequests(ctx context.Context, userID string, req *groupv1.GetJoinRequestsRequest) (*groupv1.GetJoinRequestsResponse, error)

	// Pinned messages
	PinGroupMessage(ctx context.Context, userID string, req *groupv1.PinGroupMessageRequest) error
	UnpinGroupMessage(ctx context.Context, userID string, req *groupv1.UnpinGroupMessageRequest) error
	GetPinnedMessages(ctx context.Context, userID string, req *groupv1.GetPinnedMessagesRequest) (*groupv1.GetPinnedMessagesResponse, error)
	SetGroupMute(ctx context.Context, userID string, req *groupv1.SetGroupMuteRequest) error

	// Group settings
	UpdateGroupSettings(ctx context.Context, userID string, req *groupv1.UpdateGroupSettingsRequest) error
	GetGroupSettings(ctx context.Context, req *groupv1.GetGroupSettingsRequest) (*groupv1.GetGroupSettingsResponse, error)

	// Group QR code
	GetGroupQRCode(ctx context.Context, userID string, req *groupv1.GetGroupQRCodeRequest) (*groupv1.GetGroupQRCodeResponse, error)
	RefreshGroupQRCode(ctx context.Context, req *groupv1.RefreshGroupQRCodeRequest) (*groupv1.GetGroupQRCodeResponse, error)
	GetGroupPreviewByQRCode(ctx context.Context, req *groupv1.GetGroupPreviewByQRCodeRequest) (*groupv1.GetGroupPreviewByQRCodeResponse, error)
	JoinGroupByQRCode(ctx context.Context, userID string, req *groupv1.JoinGroupByQRCodeRequest) (*groupv1.JoinGroupByQRCodeResponse, error)
}

type GroupHandler struct {
	groupv1.UnimplementedGroupServiceServer
	svc GroupService
	log *log.Helper
}

// NewGroupServer creates a new group gRPC handler
func NewGroupHandler(svc GroupService, logger log.Logger) *GroupHandler {
	return &GroupHandler{
		svc: svc,
		log: log.NewHelper(logger),
	}
}

// ========== Internal service call interface implementation ==========

func (s *GroupHandler) IsMember(ctx context.Context, req *groupv1.IsMemberRequest) (*groupv1.IsMemberResponse, error) {
	resp, err := s.svc.IsMember(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) GetGroupInfo(ctx context.Context, req *groupv1.GetGroupInfoRequest) (*groupv1.GetGroupInfoResponse, error) {
	resp, err := s.svc.GetGroupInfo(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) GetGroupMembers(ctx context.Context, req *groupv1.GetGroupMembersRequest) (*groupv1.GetGroupMembersResponse, error) {
	resp, err := s.svc.GetGroupMembers(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) GetUserGroups(ctx context.Context, req *groupv1.GetUserGroupsRequest) (*groupv1.GetUserGroupsResponse, error) {
	resp, err := s.svc.GetUserGroups(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// ========== Gateway HTTP API call interface implementation ==========

func (s *GroupHandler) CreateGroup(ctx context.Context, req *groupv1.CreateGroupRequest) (*groupv1.CreateGroupResponse, error) {
	userID := md.MustGetUserID(ctx)
	if req.Name == "" {
		return nil, errors.BadRequest(ctx, "group name is required")
	}
	if len(req.MemberIds) == 0 {
		return nil, errors.BadRequest(ctx, "at least one member must be invited")
	}
	if len(req.MemberIds) > 499 {
		return nil, errors.BadRequest(ctx, "group member count exceeds limit")
	}
	resp, err := s.svc.CreateGroup(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) UpdateGroup(ctx context.Context, req *groupv1.UpdateGroupRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if req.Name == nil && req.Avatar == nil && req.Announcement == nil && req.Description == nil {
		return nil, errors.BadRequest(ctx, "at least one field must be updated")
	}
	if err := s.svc.UpdateGroup(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) DisbandGroup(ctx context.Context, req *groupv1.DissolveGroupRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if err := s.svc.DisbandGroup(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) InviteMembers(ctx context.Context, req *groupv1.InviteMembersRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if len(req.InviteeIds) == 0 {
		return nil, errors.BadRequest(ctx, "at least one invitee must be specified")
	}
	if err := s.svc.InviteMembers(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) RemoveMember(ctx context.Context, req *groupv1.RemoveMemberRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if req.TargetUserId == "" {
		return nil, errors.BadRequest(ctx, "target_user_id is required")
	}
	if err := s.svc.RemoveMember(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) QuitGroup(ctx context.Context, req *groupv1.QuitGroupRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if err := s.svc.QuitGroup(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) UpdateMemberRole(ctx context.Context, req *groupv1.UpdateMemberRoleRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if req.TargetUserId == "" {
		return nil, errors.BadRequest(ctx, "target_user_id is required")
	}
	if err := s.svc.UpdateMemberRole(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) UpdateMemberNickname(ctx context.Context, req *groupv1.UpdateMemberNicknameRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if utf8.RuneCountInString(req.Nickname) > 20 {
		return nil, errors.BadRequest(ctx, "nickname cannot exceed 20 characters")
	}
	if err := s.svc.UpdateMemberNickname(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) TransferOwnership(ctx context.Context, req *groupv1.TransferOwnershipRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if req.NewOwnerId == "" {
		return nil, errors.BadRequest(ctx, "new_owner_id is required")
	}
	if err := s.svc.TransferOwnership(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) JoinGroup(ctx context.Context, req *groupv1.JoinGroupRequest) (*groupv1.JoinGroupResponse, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	resp, err := s.svc.JoinGroup(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) HandleJoinRequest(ctx context.Context, req *groupv1.HandleJoinRequestRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if req.RequestId == 0 {
		return nil, errors.BadRequest(ctx, "request_id is required")
	}
	if err := s.svc.HandleJoinRequest(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) GetJoinRequests(ctx context.Context, req *groupv1.GetJoinRequestsRequest) (*groupv1.GetJoinRequestsResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.GetJoinRequests(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) PinGroupMessage(ctx context.Context, req *groupv1.PinGroupMessageRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if req.MessageId == "" {
		return nil, errors.BadRequest(ctx, "message_id is required")
	}
	if err := s.svc.PinGroupMessage(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) UnpinGroupMessage(ctx context.Context, req *groupv1.UnpinGroupMessageRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if req.MessageId == "" {
		return nil, errors.BadRequest(ctx, "message_id is required")
	}
	if err := s.svc.UnpinGroupMessage(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) GetPinnedMessages(ctx context.Context, req *groupv1.GetPinnedMessagesRequest) (*groupv1.GetPinnedMessagesResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.GetPinnedMessages(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) SetGroupMute(ctx context.Context, req *groupv1.SetGroupMuteRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if err := s.svc.SetGroupMute(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) MuteMember(ctx context.Context, req *groupv1.MuteMemberRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if req.TargetUserId == "" {
		return nil, errors.BadRequest(ctx, "target_user_id is required")
	}
	if err := s.svc.MuteMember(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) UnmuteMember(ctx context.Context, req *groupv1.UnmuteMemberRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if req.TargetUserId == "" {
		return nil, errors.BadRequest(ctx, "target_user_id is required")
	}
	if err := s.svc.UnmuteMember(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) UpdateGroupSettings(ctx context.Context, req *groupv1.UpdateGroupSettingsRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if err := s.svc.UpdateGroupSettings(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) GetGroupSettings(ctx context.Context, req *groupv1.GetGroupSettingsRequest) (*groupv1.GetGroupSettingsResponse, error) {
	resp, err := s.svc.GetGroupSettings(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) UpdateMemberRemark(ctx context.Context, req *groupv1.UpdateMemberRemarkRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.GroupId == "" {
		return nil, errors.BadRequest(ctx, "group_id is required")
	}
	if utf8.RuneCountInString(req.Remark) > 20 {
		return nil, errors.BadRequest(ctx, "remark cannot exceed 20 characters")
	}
	if err := s.svc.UpdateMemberRemark(ctx, userID, req); err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

func (s *GroupHandler) GetGroupQRCode(ctx context.Context, req *groupv1.GetGroupQRCodeRequest) (*groupv1.GetGroupQRCodeResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.GetGroupQRCode(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) RefreshGroupQRCode(ctx context.Context, req *groupv1.RefreshGroupQRCodeRequest) (*groupv1.GetGroupQRCodeResponse, error) {
	resp, err := s.svc.RefreshGroupQRCode(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) GetGroupPreviewByQRCode(ctx context.Context, req *groupv1.GetGroupPreviewByQRCodeRequest) (*groupv1.GetGroupPreviewByQRCodeResponse, error) {
	resp, err := s.svc.GetGroupPreviewByQRCode(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

func (s *GroupHandler) JoinGroupByQRCode(ctx context.Context, req *groupv1.JoinGroupByQRCodeRequest) (*groupv1.JoinGroupByQRCodeResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.JoinGroupByQRCode(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}
