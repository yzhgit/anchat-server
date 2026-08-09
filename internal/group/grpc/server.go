package grpc

import (
	"context"

	commonpb "github.com/anychat/server/api/proto/common"
	grouppb "github.com/anychat/server/api/proto/group"
	"github.com/anychat/server/internal/group/dto"
	"github.com/anychat/server/internal/group/model"
	"github.com/anychat/server/internal/group/service"
	"github.com/anychat/server/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GroupServer represents the group gRPC server
type GroupServer struct {
	grouppb.UnimplementedGroupServiceServer
	svc service.GroupService
}

// NewGroupServer creates a new group gRPC server
func NewGroupServer(service service.GroupService) *GroupServer {
	return &GroupServer{
		svc: service,
	}
}

// ========== Internal service call interface implementation ==========

// GetGroupInfo gets group information
func (s *GroupServer) GetGroupInfo(ctx context.Context, req *grouppb.GetGroupInfoRequest) (*grouppb.GetGroupInfoResponse, error) {
	userID := ""
	if req.UserId != nil {
		userID = *req.UserId
	}
	resp, err := s.svc.GetGroupInfo(ctx, userID, req.GroupId)
	if err != nil {
		return nil, convertError(err)
	}

	return &grouppb.GetGroupInfoResponse{
		GroupId:      resp.GroupID,
		Name:         resp.Name,
		DisplayName:  resp.DisplayName,
		Avatar:       resp.Avatar,
		Announcement: resp.Announcement,
		Description:  resp.Description,
		OwnerId:      resp.OwnerID,
		MemberCount:  resp.MemberCount,
		MaxMembers:   resp.MaxMembers,
		JoinVerify:   resp.JoinVerify,
		IsMuted:      resp.IsMuted,
		Status:       int32(1), // active
		CreatedAt:    timestamppb.New(resp.CreatedAt),
		UpdatedAt:    timestamppb.New(resp.UpdatedAt),
	}, nil
}

// GetGroupMembers gets group member list
func (s *GroupServer) GetGroupMembers(ctx context.Context, req *grouppb.GetGroupMembersRequest) (*grouppb.GetGroupMembersResponse, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	resp, err := s.svc.GetGroupMembers(ctx, req.UserId, req.GroupId, page, pageSize)
	if err != nil {
		return nil, convertError(err)
	}

	// DTO -> Proto conversion
	members := make([]*grouppb.GroupMember, 0, len(resp.Members))
	for _, m := range resp.Members {
		member := &grouppb.GroupMember{
			UserId:   m.UserID,
			Role:     grouppb.GroupRole(m.Role),
			JoinedAt: timestamppb.New(m.JoinedAt),
		}
		if m.MutedUntil != nil {
			member.MutedUntil = timestamppb.New(*m.MutedUntil)
		}
		if m.GroupNickname != nil {
			member.GroupNickname = m.GroupNickname
		}
		if m.UserInfo != nil {
			member.UserInfo = &commonpb.UserInfo{
				UserId:   m.UserInfo.UserID,
				Nickname: m.UserInfo.Nickname,
				Avatar:   m.UserInfo.Avatar,
			}
		}
		members = append(members, member)
	}

	return &grouppb.GetGroupMembersResponse{
		Members: members,
		Total:   resp.Total,
	}, nil
}

// IsMember checks if user is group member
func (s *GroupServer) IsMember(ctx context.Context, req *grouppb.IsMemberRequest) (*grouppb.IsMemberResponse, error) {
	isMember, role, err := s.svc.IsMember(ctx, req.GroupId, req.UserId)
	if err != nil {
		return nil, convertError(err)
	}

	return &grouppb.IsMemberResponse{
		IsMember: isMember,
		Role:     grouppb.GroupRole(role),
	}, nil
}

// GetUserGroups gets list of groups user joined
func (s *GroupServer) GetUserGroups(ctx context.Context, req *grouppb.GetUserGroupsRequest) (*grouppb.GetUserGroupsResponse, error) {
	var lastUpdateTime *int64
	if req.LastUpdateTime != nil {
		lastUpdateTime = req.LastUpdateTime
	}

	resp, err := s.svc.GetUserGroups(ctx, req.UserId, lastUpdateTime)
	if err != nil {
		return nil, convertError(err)
	}

	// DTO -> Proto conversion
	groups := make([]*grouppb.GroupInfo, 0, len(resp.Groups))
	for _, g := range resp.Groups {
		groups = append(groups, &grouppb.GroupInfo{
			GroupId:     g.GroupID,
			Name:        g.Name,
			DisplayName: g.DisplayName,
			Avatar:      g.Avatar,
			MemberCount: g.MemberCount,
			UpdatedAt:   timestamppb.New(g.UpdatedAt),
		})
	}

	return &grouppb.GetUserGroupsResponse{
		Groups:     groups,
		UpdateTime: resp.UpdateTime,
		Total:      resp.Total,
	}, nil
}

// ========== Gateway HTTP API call interface implementation ==========

// CreateGroup creates a new group
func (s *GroupServer) CreateGroup(ctx context.Context, req *grouppb.CreateGroupRequest) (*grouppb.CreateGroupResponse, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.CreateGroupRequest{
		Name:      req.Name,
		MemberIDs: req.MemberIds,
	}
	if req.Avatar != nil {
		dtoReq.Avatar = *req.Avatar
	}

	// Call service layer
	resp, err := s.svc.CreateGroup(ctx, req.UserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &grouppb.CreateGroupResponse{
		GroupId:     resp.GroupID,
		Name:        resp.Name,
		Avatar:      &resp.Avatar,
		OwnerId:     resp.OwnerID,
		MemberCount: resp.MemberCount,
		CreatedAt:   timestamppb.New(resp.CreatedAt),
	}, nil
}

// UpdateGroup updates group information
func (s *GroupServer) UpdateGroup(ctx context.Context, req *grouppb.UpdateGroupRequest) (*commonpb.Empty, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.UpdateGroupRequest{
		Name:         req.Name,
		Avatar:       req.Avatar,
		Announcement: req.Announcement,
		Description:  req.Description,
	}

	// Call service layer
	err := s.svc.UpdateGroup(ctx, req.UserId, req.GroupId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// DissolveGroup dissolves a group
func (s *GroupServer) DissolveGroup(ctx context.Context, req *grouppb.DissolveGroupRequest) (*commonpb.Empty, error) {
	err := s.svc.DissolveGroup(ctx, req.UserId, req.GroupId)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// InviteMembers invites members
func (s *GroupServer) InviteMembers(ctx context.Context, req *grouppb.InviteMembersRequest) (*commonpb.Empty, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.InviteMembersRequest{
		UserIDs: req.InviteeIds,
	}

	// Call service layer
	err := s.svc.InviteMembers(ctx, req.UserId, req.GroupId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// RemoveMember removes a member
func (s *GroupServer) RemoveMember(ctx context.Context, req *grouppb.RemoveMemberRequest) (*commonpb.Empty, error) {
	err := s.svc.RemoveMember(ctx, req.UserId, req.GroupId, req.TargetUserId)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// QuitGroup quits a group
func (s *GroupServer) QuitGroup(ctx context.Context, req *grouppb.QuitGroupRequest) (*commonpb.Empty, error) {
	err := s.svc.QuitGroup(ctx, req.UserId, req.GroupId)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// UpdateMemberRole updates member role
func (s *GroupServer) UpdateMemberRole(ctx context.Context, req *grouppb.UpdateMemberRoleRequest) (*commonpb.Empty, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.UpdateMemberRoleRequest{
		Role: model.GroupRole(req.Role),
	}

	// Call service layer
	err := s.svc.UpdateMemberRole(ctx, req.UserId, req.GroupId, req.TargetUserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// UpdateMemberNickname updates member nickname
func (s *GroupServer) UpdateMemberNickname(ctx context.Context, req *grouppb.UpdateMemberNicknameRequest) (*commonpb.Empty, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.UpdateMemberNicknameRequest{
		Nickname: req.Nickname,
	}

	// Call service layer
	err := s.svc.UpdateMemberNickname(ctx, req.UserId, req.GroupId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// TransferOwnership transfers ownership
func (s *GroupServer) TransferOwnership(ctx context.Context, req *grouppb.TransferOwnershipRequest) (*commonpb.Empty, error) {
	err := s.svc.TransferOwnership(ctx, req.UserId, req.GroupId, req.NewOwnerId)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// JoinGroup joins a group
func (s *GroupServer) JoinGroup(ctx context.Context, req *grouppb.JoinGroupRequest) (*grouppb.JoinGroupResponse, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.JoinGroupRequest{}
	if req.Message != nil {
		dtoReq.Message = *req.Message
	}

	// Call service layer
	resp, err := s.svc.JoinGroup(ctx, req.UserId, req.GroupId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	result := &grouppb.JoinGroupResponse{
		NeedVerify: resp.NeedVerify,
	}
	if resp.RequestID != nil {
		result.RequestId = resp.RequestID
	}

	return result, nil
}

// HandleJoinRequest handles join request
func (s *GroupServer) HandleJoinRequest(ctx context.Context, req *grouppb.HandleJoinRequestRequest) (*commonpb.Empty, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.HandleJoinRequestRequest{
		Accept: req.Accept,
	}

	// Call service layer
	err := s.svc.HandleJoinRequest(ctx, req.UserId, req.RequestId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// GetJoinRequests gets join request list
func (s *GroupServer) GetJoinRequests(ctx context.Context, req *grouppb.GetJoinRequestsRequest) (*grouppb.GetJoinRequestsResponse, error) {
	var status *model.JoinRequestStatus
	if req.Status != nil {
		v := model.JoinRequestStatus(*req.Status)
		status = &v
	}

	resp, err := s.svc.GetJoinRequests(ctx, req.UserId, req.GroupId, status)
	if err != nil {
		return nil, convertError(err)
	}

	// DTO -> Proto conversion
	requests := make([]*grouppb.JoinRequest, 0, len(resp.Requests))
	for _, r := range resp.Requests {
		request := &grouppb.JoinRequest{
			Id:        r.ID,
			GroupId:   r.GroupID,
			UserId:    r.UserID,
			Message:   &r.Message,
			Status:    grouppb.JoinRequestStatus(r.Status),
			CreatedAt: timestamppb.New(r.CreatedAt),
		}
		if r.InviterID != nil {
			request.InviterId = r.InviterID
		}
		if r.UserInfo != nil {
			request.UserInfo = &commonpb.UserInfo{
				UserId:   r.UserInfo.UserID,
				Nickname: r.UserInfo.Nickname,
				Avatar:   r.UserInfo.Avatar,
			}
		}
		requests = append(requests, request)
	}

	return &grouppb.GetJoinRequestsResponse{
		Requests: requests,
		Total:    resp.Total,
	}, nil
}

// PinGroupMessage pins a message
func (s *GroupServer) PinGroupMessage(ctx context.Context, req *grouppb.PinGroupMessageRequest) (*commonpb.Empty, error) {
	if err := s.svc.PinGroupMessage(ctx, req.UserId, req.GroupId, req.MessageId); err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

// UnpinGroupMessage unpins a message
func (s *GroupServer) UnpinGroupMessage(ctx context.Context, req *grouppb.UnpinGroupMessageRequest) (*commonpb.Empty, error) {
	if err := s.svc.UnpinGroupMessage(ctx, req.UserId, req.GroupId, req.MessageId); err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

// GetPinnedMessages gets pinned message list
func (s *GroupServer) GetPinnedMessages(ctx context.Context, req *grouppb.GetPinnedMessagesRequest) (*grouppb.GetPinnedMessagesResponse, error) {
	resp, err := s.svc.GetPinnedMessages(ctx, req.UserId, req.GroupId)
	if err != nil {
		return nil, convertError(err)
	}

	items := make([]*grouppb.PinnedMessage, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		item := &grouppb.PinnedMessage{
			MessageId: m.MessageID,
			Content:   m.Content,
			PinnedBy:  m.PinnedBy,
			PinnedAt:  m.PinnedAt,
		}
		if m.ContentType != model.PinnedMessageContentTypeUnspecified {
			contentType := grouppb.MessageContentType(m.ContentType)
			item.ContentType = &contentType
		}
		if m.MessageSeq != nil {
			item.MessageSeq = m.MessageSeq
		}
		items = append(items, item)
	}

	var topMessage *grouppb.PinnedMessage
	if resp.TopMessage != nil {
		topMessage = &grouppb.PinnedMessage{
			MessageId: resp.TopMessage.MessageID,
			Content:   resp.TopMessage.Content,
			PinnedBy:  resp.TopMessage.PinnedBy,
			PinnedAt:  resp.TopMessage.PinnedAt,
		}
		if resp.TopMessage.ContentType != model.PinnedMessageContentTypeUnspecified {
			contentType := grouppb.MessageContentType(resp.TopMessage.ContentType)
			topMessage.ContentType = &contentType
		}
		if resp.TopMessage.MessageSeq != nil {
			topMessage.MessageSeq = resp.TopMessage.MessageSeq
		}
	}

	return &grouppb.GetPinnedMessagesResponse{
		Messages:   items,
		Total:      resp.Total,
		TopMessage: topMessage,
		Version:    resp.Version,
	}, nil
}

// SetGroupMute sets group mute
func (s *GroupServer) SetGroupMute(ctx context.Context, req *grouppb.SetGroupMuteRequest) (*commonpb.Empty, error) {
	if err := s.svc.SetGroupMute(ctx, req.UserId, req.GroupId, req.Enabled); err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

// MuteMember mutes a member
func (s *GroupServer) MuteMember(ctx context.Context, req *grouppb.MuteMemberRequest) (*commonpb.Empty, error) {
	dtoReq := &dto.MuteMemberRequest{
		Type:            int16(req.Type),
		DurationMinutes: req.DurationMinutes,
	}

	if err := s.svc.MuteMember(ctx, req.UserId, req.GroupId, req.TargetUserId, dtoReq); err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

// UnmuteMember unmutes a member
func (s *GroupServer) UnmuteMember(ctx context.Context, req *grouppb.UnmuteMemberRequest) (*commonpb.Empty, error) {
	if err := s.svc.UnmuteMember(ctx, req.UserId, req.GroupId, req.TargetUserId); err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

// UpdateGroupSettings updates group settings
func (s *GroupServer) UpdateGroupSettings(ctx context.Context, req *grouppb.UpdateGroupSettingsRequest) (*commonpb.Empty, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.UpdateGroupSettingsRequest{
		JoinVerify:        req.JoinVerify,
		AllowMemberInvite: req.AllowMemberInvite,
		AllowViewHistory:  req.AllowViewHistory,
		AllowAddFriend:    req.AllowAddFriend,
		AllowMemberModify: req.AllowMemberModify,
	}

	// Call service layer
	err := s.svc.UpdateGroupSettings(ctx, req.UserId, req.GroupId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// GetGroupSettings gets group settings
func (s *GroupServer) GetGroupSettings(ctx context.Context, req *grouppb.GetGroupSettingsRequest) (*grouppb.GetGroupSettingsResponse, error) {
	resp, err := s.svc.GetGroupSettings(ctx, req.GroupId)
	if err != nil {
		return nil, convertError(err)
	}

	return &grouppb.GetGroupSettingsResponse{
		GroupId:           resp.GroupID,
		JoinVerify:        resp.JoinVerify,
		AllowMemberInvite: resp.AllowMemberInvite,
		AllowViewHistory:  resp.AllowViewHistory,
		AllowAddFriend:    resp.AllowAddFriend,
		AllowMemberModify: resp.AllowMemberModify,
	}, nil
}

// UpdateMemberRemark sets/clears remark
func (s *GroupServer) UpdateMemberRemark(ctx context.Context, req *grouppb.UpdateMemberRemarkRequest) (*commonpb.Empty, error) {
	dtoReq := &dto.UpdateMemberRemarkRequest{Remark: req.Remark}
	if err := s.svc.UpdateMemberRemark(ctx, req.UserId, req.GroupId, dtoReq); err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

// GetGroupQRCode gets group QR code
func (s *GroupServer) GetGroupQRCode(ctx context.Context, req *grouppb.GetGroupQRCodeRequest) (*grouppb.GetGroupQRCodeResponse, error) {
	resp, err := s.svc.GetGroupQRCode(ctx, req.UserId, req.GroupId)
	if err != nil {
		return nil, convertError(err)
	}
	return &grouppb.GetGroupQRCodeResponse{
		Token:    resp.Token,
		DeepLink: resp.DeepLink,
		ExpireAt: resp.ExpireAt,
	}, nil
}

// RefreshGroupQRCode refreshes QR code
func (s *GroupServer) RefreshGroupQRCode(ctx context.Context, req *grouppb.RefreshGroupQRCodeRequest) (*grouppb.GetGroupQRCodeResponse, error) {
	resp, err := s.svc.RefreshGroupQRCode(ctx, req.UserId, req.GroupId)
	if err != nil {
		return nil, convertError(err)
	}
	return &grouppb.GetGroupQRCodeResponse{
		Token:    resp.Token,
		DeepLink: resp.DeepLink,
		ExpireAt: resp.ExpireAt,
	}, nil
}

// GetGroupPreviewByQRCode gets group preview by QR code
func (s *GroupServer) GetGroupPreviewByQRCode(ctx context.Context, req *grouppb.GetGroupPreviewByQRCodeRequest) (*grouppb.GetGroupPreviewByQRCodeResponse, error) {
	resp, err := s.svc.GetGroupPreviewByQRCode(ctx, req.Token)
	if err != nil {
		return nil, convertError(err)
	}
	return &grouppb.GetGroupPreviewByQRCodeResponse{
		GroupId:     resp.GroupID,
		Name:        resp.Name,
		Avatar:      resp.Avatar,
		MemberCount: resp.MemberCount,
		NeedVerify:  resp.NeedVerify,
	}, nil
}

// JoinGroupByQRCode joins group by QR code
func (s *GroupServer) JoinGroupByQRCode(ctx context.Context, req *grouppb.JoinGroupByQRCodeRequest) (*grouppb.JoinGroupByQRCodeResponse, error) {
	resp, err := s.svc.JoinGroupByQRCode(ctx, req.UserId, req.Token)
	if err != nil {
		return nil, convertError(err)
	}
	result := &grouppb.JoinGroupByQRCodeResponse{
		Joined:     resp.Joined,
		GroupId:    resp.GroupID,
		NeedVerify: resp.NeedVerify,
	}
	if resp.RequestID != nil {
		result.RequestId = resp.RequestID
	}
	return result, nil
}

// convertError converts business errors to gRPC errors
func convertError(err error) error {
	if bizErr, ok := err.(*errors.Business); ok {
		switch bizErr.Code {
		case errors.CodeGroupNotFound:
			return status.Error(codes.NotFound, bizErr.Message)
		case errors.CodeGroupDissolved:
			return status.Error(codes.FailedPrecondition, bizErr.Message)
		case errors.CodeGroupMemberTooFew:
			return status.Error(codes.FailedPrecondition, bizErr.Message)
		case errors.CodeGroupMemberLimitReached:
			return status.Error(codes.ResourceExhausted, bizErr.Message)
		case errors.CodeNotGroupMember:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeAlreadyGroupMember:
			return status.Error(codes.AlreadyExists, bizErr.Message)
		case errors.CodeNoOwnerPermission:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeNoAdminPermission:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeCannotRemoveOwner:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeCannotRemoveAdmin:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeGroupNameSensitive:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeAnnouncementSensitive:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeJoinRequestNotFound:
			return status.Error(codes.NotFound, bizErr.Message)
		case errors.CodeJoinRequestProcessed:
			return status.Error(codes.FailedPrecondition, bizErr.Message)
		case errors.CodeMemberMuted:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeCannotQuitOwnGroup:
			return status.Error(codes.FailedPrecondition, bizErr.Message)
		case errors.CodeGroupQRExpired:
			return status.Error(codes.FailedPrecondition, bizErr.Message)
		case errors.CodeGroupQRInvalid:
			return status.Error(codes.NotFound, bizErr.Message)
		case errors.CodeGroupPinnedLimitExceeded:
			return status.Error(codes.ResourceExhausted, bizErr.Message)
		case errors.CodeMessageNotFound:
			return status.Error(codes.NotFound, bizErr.Message)
		case errors.CodeMessageNotInGroup:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeParamError:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		default:
			return status.Error(codes.Internal, bizErr.Message)
		}
	}
	return status.Error(codes.Internal, err.Error())
}
