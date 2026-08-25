package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	groupv1 "flamingo/api/group/v1"
	messagev1 "flamingo/api/message/v1"
	userv1 "flamingo/api/user/v1"

	"flamingo/app/group/internal/handler"
	"flamingo/app/group/internal/model"
	"flamingo/pkg/broker"
	"flamingo/pkg/consts"
	"flamingo/pkg/errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

const (
	maxPinnedMessagesPerGroup        = 20
	systemPinnedMessageOperatorUser  = "system"
	pinnedMessagePreviewMaxRuneCount = 100
)

type pinnedMessageSnapshot struct {
	content     string
	contentType model.PinnedMessageContentType
	messageSeq  *int64
}

// groupServiceImpl represents the group service implementation
type groupServiceImpl struct {
	groupRepo       GroupRepository
	memberRepo      GroupMemberRepository
	settingRepo     GroupSettingRepository
	joinRequestRepo GroupJoinRequestRepository
	pinnedRepo      GroupPinnedMessageRepository
	qrcodeRepo      GroupQRCodeRepository
	messageService  messagev1.MessageServiceClient
	userService     userv1.UserServiceClient
	broker          broker.Broker
	db              *gorm.DB
	log             *log.Helper
}

var _ handler.GroupService = (*groupServiceImpl)(nil)

// NewGroupService creates a new group service
func NewGroupService(
	groupRepo GroupRepository,
	memberRepo GroupMemberRepository,
	settingRepo GroupSettingRepository,
	joinRequestRepo GroupJoinRequestRepository,
	pinnedRepo GroupPinnedMessageRepository,
	qrcodeRepo GroupQRCodeRepository,
	messageService messagev1.MessageServiceClient,
	userService userv1.UserServiceClient,
	broker broker.Broker,
	db *gorm.DB,
	logger log.Logger,
) handler.GroupService {
	return &groupServiceImpl{
		groupRepo:       groupRepo,
		memberRepo:      memberRepo,
		settingRepo:     settingRepo,
		joinRequestRepo: joinRequestRepo,
		pinnedRepo:      pinnedRepo,
		qrcodeRepo:      qrcodeRepo,
		messageService:  messageService,
		userService:     userService,
		broker:          broker,
		db:              db,
		log:             log.NewHelper(logger),
	}
}

// CreateGroup creates a new group
func (s *groupServiceImpl) CreateGroup(ctx context.Context, userID string, req *groupv1.CreateGroupRequest) (*groupv1.CreateGroupResponse, error) {
	l := s.log.WithContext(ctx)
	ownerID := userID

	// Validate: at least owner + 1 member required
	if len(req.MemberIds) == 0 {
		return nil, errors.NewBusiness(errors.CodeGroupMemberTooFew, "At least one member must be invited")
	}

	// Validate: member count cannot exceed max limit
	totalMembers := len(req.MemberIds) + 1 // +1 for owner
	if totalMembers > 500 {
		return nil, errors.NewBusiness(errors.CodeGroupMemberLimitReached, "Group member count exceeds limit")
	}

	// Generate unique group ID
	groupID := uuid.New().String()

	// Create group object
	group := &model.Group{
		GroupID:     groupID,
		Name:        req.Name,
		Avatar:      ptrStr(req.Avatar),
		OwnerID:     ownerID,
		MemberCount: int32(totalMembers),
		MaxMembers:  500,
		Status:      model.GroupStatusNormal,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create group with transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		groupRepoTx := s.groupRepo.WithTx(tx)
		memberRepoTx := s.memberRepo.WithTx(tx)
		settingRepoTx := s.settingRepo.WithTx(tx)

		// 1. Create group record
		if err := groupRepoTx.Create(ctx, group); err != nil {
			l.Errorw("msg", "Failed to create group", "error", err)
			return err
		}

		// 2. Create default group settings
		setting := model.DefaultGroupSetting(groupID)
		if err := settingRepoTx.Create(ctx, setting); err != nil {
			l.Errorw("msg", "Failed to create group settings", "error", err)
			return err
		}

		// 3. Add owner
		ownerMember := &model.GroupMember{
			GroupID:  groupID,
			UserID:   ownerID,
			Role:     model.GroupRoleOwner,
			JoinedAt: time.Now(),
		}
		if err := memberRepoTx.AddMember(ctx, ownerMember); err != nil {
			l.Errorw("msg", "Failed to add owner", "error", err)
			return err
		}

		// 4. Add initial members
		members := make([]*model.GroupMember, 0, len(req.MemberIds))
		now := time.Now()
		for _, memberID := range req.MemberIds {
			if memberID == ownerID {
				continue // skip owner
			}
			members = append(members, &model.GroupMember{
				GroupID:  groupID,
				UserID:   memberID,
				Role:     model.GroupRoleMember,
				JoinedAt: now,
			})
		}

		if len(members) > 0 {
			if err := memberRepoTx.AddMembers(ctx, members); err != nil {
				l.Errorw("msg", "Failed to add members", "error", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish member joined event (to all members)
	for _, memberID := range req.MemberIds {
		s.publishMemberJoinedNotification(groupID, memberID, ownerID)
	}

	// Build response
	resp := &groupv1.CreateGroupResponse{
		GroupId:     group.GroupID,
		Name:        group.Name,
		OwnerId:     group.OwnerID,
		MemberCount: group.MemberCount,
		CreatedAt:   timestamppb.New(group.CreatedAt),
	}
	if group.Avatar != "" {
		resp.Avatar = &group.Avatar
	}

	return resp, nil
}

// GetGroupInfo gets group information
func (s *groupServiceImpl) GetGroupInfo(ctx context.Context, req *groupv1.GetGroupInfoRequest) (*groupv1.GetGroupInfoResponse, error) {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	userID := req.GetUserId()

	// Get group info
	group, err := s.groupRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeGroupNotFound, "Group not found")
		}
		l.Errorw("msg", "Failed to get group", "error", err)
		return nil, err
	}

	// Check group status
	if !group.IsActive() {
		return nil, errors.NewBusiness(errors.CodeGroupDissolved, "Group has been dissolved")
	}

	// Get user's role in the group and display name
	displayName := group.Name
	if userID != "" {
		member, err := s.memberRepo.GetMember(ctx, groupID, userID)
		if err == nil && member.GroupRemark != "" {
			displayName = member.GroupRemark
		}
	}

	joinVerify := s.getGroupJoinVerify(ctx, groupID)

	return &groupv1.GetGroupInfoResponse{
		GroupId:      group.GroupID,
		Name:         group.Name,
		DisplayName:  displayName,
		Avatar:       group.Avatar,
		Announcement: group.Announcement,
		Description:  group.Description,
		OwnerId:      group.OwnerID,
		MemberCount:  group.MemberCount,
		MaxMembers:   group.MaxMembers,
		JoinVerify:   joinVerify,
		IsMuted:      group.IsMuted,
		Status:       int32(group.Status),
		CreatedAt:    timestamppb.New(group.CreatedAt),
		UpdatedAt:    timestamppb.New(group.UpdatedAt),
	}, nil
}

// UpdateGroup updates group information
func (s *groupServiceImpl) UpdateGroup(ctx context.Context, userID string, req *groupv1.UpdateGroupRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId

	// Permission check: owner and admin can modify; regular members need group settings to allow
	member, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}

	if !member.CanManageGroup() {
		settings, settingsErr := s.settingRepo.GetSettings(ctx, groupID)
		if settingsErr != nil && settingsErr != gorm.ErrRecordNotFound {
			return settingsErr
		}
		if settings == nil || !settings.AllowMemberModify {
			return errors.NewBusiness(errors.CodeNoAdminPermission, "No permission to update group info")
		}
	}

	// Build update fields
	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.Announcement != nil {
		updates["announcement"] = *req.Announcement
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) == 0 {
		return nil
	}

	// Update group
	if err := s.groupRepo.UpdateFields(ctx, groupID, updates); err != nil {
		l.Errorw("msg", "Failed to update group", "error", err)
		return err
	}

	// Publish group info updated event
	s.publishGroupInfoUpdatedNotification(groupID, userID, req)

	return nil
}

// DisbandGroup dissolves a group
func (s *groupServiceImpl) DisbandGroup(ctx context.Context, userID string, req *groupv1.DissolveGroupRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId

	// Permission check: only owner can dissolve
	member, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}

	if !member.IsOwner() {
		return errors.NewBusiness(errors.CodeNoOwnerPermission, "Only the owner can dissolve the group")
	}

	// Soft delete group
	if err := s.groupRepo.Delete(ctx, groupID); err != nil {
		l.Errorw("msg", "Failed to dissolve group", "error", err)
		return err
	}

	// Get group info for event
	group, _ := s.groupRepo.GetByGroupID(ctx, groupID)
	groupName := ""
	if group != nil {
		groupName = group.Name
	}

	// Publish group disbanded event
	s.publishGroupDisbandedNotification(groupID, userID, groupName)

	return nil
}

// GetUserGroups gets list of groups user has joined
func (s *groupServiceImpl) GetUserGroups(ctx context.Context, req *groupv1.GetUserGroupsRequest) (*groupv1.GetUserGroupsResponse, error) {
	l := s.log.WithContext(ctx)
	userID := req.UserId
	lastUpdateTime := req.LastUpdateTime

	var members []*model.GroupMember
	var err error

	// Incremental sync
	if lastUpdateTime != nil && *lastUpdateTime > 0 {
		t := time.Unix(*lastUpdateTime, 0)
		members, err = s.memberRepo.GetUserGroupsByUpdateTime(ctx, userID, t)
	} else {
		members, err = s.memberRepo.GetUserGroups(ctx, userID)
	}

	if err != nil {
		l.Errorw("msg", "Failed to get user groups", "error", err)
		return nil, err
	}

	// Get group details
	groups := make([]*groupv1.GroupInfo, 0, len(members))
	for _, member := range members {
		group, err := s.groupRepo.GetByGroupID(ctx, member.GroupID)
		if err != nil {
			l.Warnw("msg", "Failed to get group info", "groupID", member.GroupID, "error", err)
			continue
		}

		if !group.IsActive() {
			continue // skip dissolved groups
		}

		displayName := group.Name
		if member.GroupRemark != "" {
			displayName = member.GroupRemark
		}

		groups = append(groups, &groupv1.GroupInfo{
			GroupId:     group.GroupID,
			Name:        group.Name,
			DisplayName: displayName,
			Avatar:      group.Avatar,
			MemberCount: group.MemberCount,
			UpdatedAt:   timestamppb.New(group.UpdatedAt),
		})
	}

	return &groupv1.GetUserGroupsResponse{
		Groups:     groups,
		Total:      int64(len(groups)),
		UpdateTime: time.Now().Unix(),
	}, nil
}

// GetGroupMembers gets list of group members
func (s *groupServiceImpl) GetGroupMembers(ctx context.Context, req *groupv1.GetGroupMembersRequest) (*groupv1.GetGroupMembersResponse, error) {
	l := s.log.WithContext(ctx)
	userID := req.UserId
	groupID := req.GroupId
	page := 1
	if req.Page != nil {
		page = int(*req.Page)
	}
	pageSize := 20
	if req.PageSize != nil {
		pageSize = int(*req.PageSize)
	}

	// Validate user is a group member
	isMember, err := s.memberRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
	}

	// Default pagination params
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// Get member list
	members, total, err := s.memberRepo.GetMembers(ctx, groupID, page, pageSize)
	if err != nil {
		l.Errorw("msg", "Failed to get group members", "error", err)
		return nil, err
	}

	// Convert to proto
	memberResponses := make([]*groupv1.GroupMember, 0, len(members))
	for _, m := range members {
		mutedUntil := (*timestamppb.Timestamp)(nil)
		if m.MutedUntil != nil {
			mutedUntil = timestamppb.New(*m.MutedUntil)
		}
		response := &groupv1.GroupMember{
			UserId:     m.UserID,
			Role:       groupv1.GroupRole(m.Role),
			JoinedAt:   timestamppb.New(m.JoinedAt),
			MutedUntil: mutedUntil,
		}
		if m.GroupNickname != "" {
			response.GroupNickname = &m.GroupNickname
		}
		memberResponses = append(memberResponses, response)
	}

	return &groupv1.GetGroupMembersResponse{
		Members: memberResponses,
		Total:   total,
	}, nil
}

// InviteMembers invites members to a group
func (s *groupServiceImpl) InviteMembers(ctx context.Context, userID string, req *groupv1.InviteMembersRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	inviteeIDs := req.InviteeIds

	// Get group info
	group, err := s.groupRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeGroupNotFound, "Group not found")
		}
		return err
	}

	// Check group status
	if !group.IsActive() {
		return errors.NewBusiness(errors.CodeGroupDissolved, "Group has been dissolved")
	}

	// Permission check: get inviter info
	inviter, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}

	// Check group settings: can regular members invite
	settings, err := s.settingRepo.GetSettings(ctx, groupID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if settings != nil && !settings.AllowMemberInvite && !inviter.CanManageGroup() {
		return errors.NewBusiness(errors.CodeNoAdminPermission, "Group settings do not allow regular members to invite")
	}
	joinVerify := true
	if settings != nil {
		joinVerify = settings.JoinVerify
	}

	// Validate: member count not at limit
	if group.IsFull() {
		return errors.NewBusiness(errors.CodeGroupMemberLimitReached, "Group member limit reached")
	}

	// Process each invitee
	for _, inviteeID := range inviteeIDs {
		// Check if already a member
		isMember, _ := s.memberRepo.IsMember(ctx, groupID, inviteeID)
		if isMember {
			continue // skip users who are already members
		}

		// If verification needed, create join request
		if joinVerify {
			request := &model.GroupJoinRequest{
				GroupID:   groupID,
				UserID:    inviteeID,
				InviterID: userID,
				Status:    model.JoinRequestStatusPending,
				CreatedAt: time.Now(),
			}
			if err := s.joinRequestRepo.Create(ctx, request); err != nil {
				l.Errorw("msg", "Failed to create join request", "error", err)
				continue
			}
		} else {
			// Directly add member
			err := s.db.Transaction(func(tx *gorm.DB) error {
				memberRepoTx := s.memberRepo.WithTx(tx)
				groupRepoTx := s.groupRepo.WithTx(tx)

				member := &model.GroupMember{
					GroupID:  groupID,
					UserID:   inviteeID,
					Role:     model.GroupRoleMember,
					JoinedAt: time.Now(),
				}
				if err := memberRepoTx.AddMember(ctx, member); err != nil {
					return err
				}

				// Update member count
				return groupRepoTx.UpdateMemberCount(ctx, groupID, 1)
			})

			if err != nil {
				l.Errorw("msg", "Failed to add member", "inviteeID", inviteeID, "error", err)
				continue
			}
		}
	}

	return nil
}

// RemoveMember removes a member from a group
func (s *groupServiceImpl) RemoveMember(ctx context.Context, userID string, req *groupv1.RemoveMemberRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	targetUserID := req.TargetUserId

	// Get operator info
	operator, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}

	// Get target member info
	target, err := s.memberRepo.GetMember(ctx, groupID, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "Target user is not a group member")
		}
		return err
	}

	// Permission check
	if target.IsOwner() {
		return errors.NewBusiness(errors.CodeCannotRemoveOwner, "Cannot remove group owner")
	}

	if !operator.CanRemoveMember(target.Role) {
		if target.Role == model.GroupRoleAdmin {
			return errors.NewBusiness(errors.CodeCannotRemoveAdmin, "Admins cannot remove other admins")
		}
		return errors.NewBusiness(errors.CodeNoAdminPermission, "No permission to remove member")
	}

	// Use transaction to delete member and update count
	err = s.db.Transaction(func(tx *gorm.DB) error {
		memberRepoTx := s.memberRepo.WithTx(tx)
		groupRepoTx := s.groupRepo.WithTx(tx)

		if err := memberRepoTx.RemoveMember(ctx, groupID, targetUserID); err != nil {
			return err
		}

		return groupRepoTx.UpdateMemberCount(ctx, groupID, -1)
	})

	if err != nil {
		l.Errorw("msg", "Failed to remove member", "error", err)
		return err
	}

	// Publish member left event
	s.publishMemberLeftNotification(groupID, targetUserID, userID, "removed_by_admin")

	return nil
}

// QuitGroup makes user quit a group
func (s *groupServiceImpl) QuitGroup(ctx context.Context, userID string, req *groupv1.QuitGroupRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId

	// Get member info
	member, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}

	// Owner cannot quit directly
	if member.IsOwner() {
		// Check if only owner remains
		count, err := s.memberRepo.GetMemberCount(ctx, groupID)
		if err != nil {
			return err
		}

		if count == 1 {
			// Auto dissolve group
			return s.DisbandGroup(ctx, userID, &groupv1.DissolveGroupRequest{
				GroupId: groupID,
			})
		}

		return errors.NewBusiness(errors.CodeCannotQuitOwnGroup, "Owner cannot quit group, please transfer ownership or dissolve group first")
	}

	// Use transaction to delete member and update count
	err = s.db.Transaction(func(tx *gorm.DB) error {
		memberRepoTx := s.memberRepo.WithTx(tx)
		groupRepoTx := s.groupRepo.WithTx(tx)

		if err := memberRepoTx.RemoveMember(ctx, groupID, userID); err != nil {
			return err
		}

		return groupRepoTx.UpdateMemberCount(ctx, groupID, -1)
	})

	if err != nil {
		l.Errorw("msg", "Failed to quit group", "error", err)
		return err
	}

	// Publish member left event
	s.publishMemberLeftNotification(groupID, userID, userID, "self_quit")

	return nil
}

// UpdateMemberRole updates member role
func (s *groupServiceImpl) UpdateMemberRole(ctx context.Context, userID string, req *groupv1.UpdateMemberRoleRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	targetUserID := req.TargetUserId

	// Permission check: only owner can set role
	operator, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}

	if !operator.IsOwner() {
		return errors.NewBusiness(errors.CodeNoOwnerPermission, "Only owner can set admin")
	}

	// Validate target member exists
	target, err := s.memberRepo.GetMember(ctx, groupID, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "Target user is not a group member")
		}
		return err
	}

	// Cannot modify owner role
	if target.IsOwner() {
		return errors.NewBusiness(errors.CodeNoOwnerPermission, "Cannot modify owner role")
	}

	// Convert proto role to model role
	newRole := model.GroupRole(req.Role)

	// Update role
	if err := s.memberRepo.UpdateRole(ctx, groupID, targetUserID, newRole); err != nil {
		l.Errorw("msg", "Failed to update member role", "error", err)
		return err
	}

	// Publish role changed event
	s.publishRoleChangedNotification(groupID, targetUserID, userID, target.Role, newRole)

	return nil
}

// UpdateMemberNickname updates member nickname in group
func (s *groupServiceImpl) UpdateMemberNickname(ctx context.Context, userID string, req *groupv1.UpdateMemberNicknameRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId

	// Validate user is group member
	isMember, err := s.memberRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
	}

	// Update nickname
	if err := s.memberRepo.UpdateNickname(ctx, groupID, userID, req.Nickname); err != nil {
		l.Errorw("msg", "Failed to update member nickname", "error", err)
		return err
	}

	return nil
}

// TransferOwnership transfers group ownership
func (s *groupServiceImpl) TransferOwnership(ctx context.Context, userID string, req *groupv1.TransferOwnershipRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	newOwnerID := req.NewOwnerId

	// Permission check: only owner can transfer
	operator, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}

	if !operator.IsOwner() {
		return errors.NewBusiness(errors.CodeNoOwnerPermission, "Only owner can transfer group")
	}

	// Validate new owner is group member
	newOwner, err := s.memberRepo.GetMember(ctx, groupID, newOwnerID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "New owner is not a group member")
		}
		return err
	}

	if newOwner.IsOwner() {
		return nil // already owner, no action needed
	}

	// Use transaction to transfer ownership
	err = s.db.Transaction(func(tx *gorm.DB) error {
		memberRepoTx := s.memberRepo.WithTx(tx)
		groupRepoTx := s.groupRepo.WithTx(tx)

		// 1. Update group owner_id
		if err := groupRepoTx.UpdateFields(ctx, groupID, map[string]any{"owner_id": newOwnerID}); err != nil {
			return err
		}

		// 2. Change original owner to admin
		if err := memberRepoTx.UpdateRole(ctx, groupID, userID, model.GroupRoleAdmin); err != nil {
			return err
		}

		// 3. Change new owner to owner
		return memberRepoTx.UpdateRole(ctx, groupID, newOwnerID, model.GroupRoleOwner)
	})

	if err != nil {
		l.Errorw("msg", "Failed to transfer ownership", "error", err)
		return err
	}

	return nil
}

// MuteMember mutes a member
func (s *groupServiceImpl) MuteMember(ctx context.Context, userID string, req *groupv1.MuteMemberRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	targetUserID := req.TargetUserId

	operator, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}
	target, err := s.memberRepo.GetMember(ctx, groupID, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "Target user is not a group member")
		}
		return err
	}

	if target.IsOwner() {
		return errors.NewBusiness(errors.CodeNoAdminPermission, "Cannot mute group owner")
	}
	if !operator.CanMuteMember(target.Role) {
		return errors.NewBusiness(errors.CodeNoAdminPermission, "No permission to mute member")
	}

	var mutedUntil *time.Time
	switch req.Type {
	case groupv1.MuteType_MUTE_TYPE_PERMANENT:
		permanent := model.PermanentMutedUntil
		mutedUntil = &permanent
	case groupv1.MuteType_MUTE_TYPE_TEMPORARY:
		if req.DurationMinutes <= 0 {
			return errors.NewBusiness(errors.CodeParamError, "Temporary mute duration must be greater than 0")
		}
		t := time.Now().Add(time.Duration(req.DurationMinutes) * time.Minute)
		mutedUntil = &t
	default:
		return errors.NewBusiness(errors.CodeParamError, "Invalid mute type")
	}

	if err := s.memberRepo.UpdateMutedUntil(ctx, groupID, targetUserID, mutedUntil); err != nil {
		l.Errorw("msg", "Failed to mute member", "error", err)
		return err
	}

	s.publishMemberMutedNotification(groupID, userID, targetUserID, mutedUntil)
	return nil
}

// UnmuteMember unmutes a member
func (s *groupServiceImpl) UnmuteMember(ctx context.Context, userID string, req *groupv1.UnmuteMemberRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	targetUserID := req.TargetUserId

	operator, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}
	target, err := s.memberRepo.GetMember(ctx, groupID, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "Target user is not a group member")
		}
		return err
	}
	if target.IsOwner() {
		return errors.NewBusiness(errors.CodeNoAdminPermission, "Cannot operate on group owner")
	}
	if !operator.CanMuteMember(target.Role) {
		return errors.NewBusiness(errors.CodeNoAdminPermission, "No permission to unmute member")
	}
	if err := s.memberRepo.UpdateMutedUntil(ctx, groupID, targetUserID, nil); err != nil {
		l.Errorw("msg", "Failed to unmute member", "error", err)
		return err
	}
	s.publishMemberUnmutedNotification(groupID, userID, targetUserID)
	return nil
}

// JoinGroup joins a user to a group
func (s *groupServiceImpl) JoinGroup(ctx context.Context, userID string, req *groupv1.JoinGroupRequest) (*groupv1.JoinGroupResponse, error) {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId

	// Get group info
	group, err := s.groupRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeGroupNotFound, "Group not found")
		}
		return nil, err
	}

	// Check group status
	if !group.IsActive() {
		return nil, errors.NewBusiness(errors.CodeGroupDissolved, "Group has been dissolved")
	}

	// Check if already a member
	isMember, err := s.memberRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, errors.NewBusiness(errors.CodeAlreadyGroupMember, "You are already a group member")
	}

	// Check if already has pending request
	existingRequest, err := s.joinRequestRepo.GetExistingRequest(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if existingRequest != nil {
		return &groupv1.JoinGroupResponse{
			NeedVerify: true,
			RequestId:  &existingRequest.ID,
		}, nil
	}

	// Check if verification is needed
	joinVerify := s.getGroupJoinVerify(ctx, groupID)
	if joinVerify {
		// Create join request
		request := &model.GroupJoinRequest{
			GroupID:   groupID,
			UserID:    userID,
			Message:   ptrStr(req.Message),
			InviterID: ptrStr(req.InviterId),
			Status:    model.JoinRequestStatusPending,
			CreatedAt: time.Now(),
		}
		if err := s.joinRequestRepo.Create(ctx, request); err != nil {
			l.Errorw("msg", "Failed to create join request", "error", err)
			return nil, err
		}

		return &groupv1.JoinGroupResponse{
			NeedVerify: true,
			RequestId:  &request.ID,
		}, nil
	}

	// Directly join group
	err = s.db.Transaction(func(tx *gorm.DB) error {
		memberRepoTx := s.memberRepo.WithTx(tx)
		groupRepoTx := s.groupRepo.WithTx(tx)

		member := &model.GroupMember{
			GroupID:  groupID,
			UserID:   userID,
			Role:     model.GroupRoleMember,
			JoinedAt: time.Now(),
		}
		if err := memberRepoTx.AddMember(ctx, member); err != nil {
			return err
		}

		return groupRepoTx.UpdateMemberCount(ctx, groupID, 1)
	})

	if err != nil {
		l.Errorw("msg", "Failed to join group", "error", err)
		return nil, err
	}

	return &groupv1.JoinGroupResponse{
		NeedVerify: false,
	}, nil
}

// HandleJoinRequest handles a join request
func (s *groupServiceImpl) HandleJoinRequest(ctx context.Context, userID string, req *groupv1.HandleJoinRequestRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GetGroupId()
	requestID := req.GetRequestId()

	// Get request info
	request, err := s.joinRequestRepo.GetByID(ctx, requestID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeJoinRequestNotFound, "Join request not found")
		}
		return err
	}

	// Check request status
	if request.IsProcessed() {
		return errors.NewBusiness(errors.CodeJoinRequestProcessed, "Request already processed")
	}

	// Permission check: owner and admin can handle
	operator, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}

	if !operator.CanManageGroup() {
		return errors.NewBusiness(errors.CodeNoAdminPermission, "No permission to handle request")
	}

	// Handle request
	if req.Accept {
		// Accept request: use transaction
		err = s.db.Transaction(func(tx *gorm.DB) error {
			joinRequestRepoTx := s.joinRequestRepo.WithTx(tx)
			memberRepoTx := s.memberRepo.WithTx(tx)
			groupRepoTx := s.groupRepo.WithTx(tx)

			// 1. Update request status
			if err := joinRequestRepoTx.UpdateStatus(ctx, requestID, model.JoinRequestStatusAccepted); err != nil {
				return err
			}

			// 2. Add member
			member := &model.GroupMember{
				GroupID:  request.GroupID,
				UserID:   request.UserID,
				Role:     model.GroupRoleMember,
				JoinedAt: time.Now(),
			}
			if err := memberRepoTx.AddMember(ctx, member); err != nil {
				return err
			}

			// 3. Update member count
			return groupRepoTx.UpdateMemberCount(ctx, request.GroupID, 1)
		})

		if err != nil {
			l.Errorw("msg", "Failed to accept join request", "error", err)
			return err
		}
	} else {
		// Reject request
		if err := s.joinRequestRepo.UpdateStatus(ctx, requestID, model.JoinRequestStatusRejected); err != nil {
			l.Errorw("msg", "Failed to reject join request", "error", err)
			return err
		}
	}

	return nil
}

// GetJoinRequests gets join request list
func (s *groupServiceImpl) GetJoinRequests(ctx context.Context, userID string, req *groupv1.GetJoinRequestsRequest) (*groupv1.GetJoinRequestsResponse, error) {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId

	// Permission check: owner and admin can view join requests
	operator, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return nil, err
	}
	if !operator.CanManageGroup() {
		return nil, errors.NewBusiness(errors.CodeNoAdminPermission, "No permission to view join requests")
	}

	// Get request list
	var status *model.JoinRequestStatus
	if req.Status != nil {
		s := model.JoinRequestStatus(*req.Status)
		status = &s
	}
	requests, err := s.joinRequestRepo.GetRequestsByGroup(ctx, groupID, status)
	if err != nil {
		l.Errorw("msg", "Failed to get join requests", "error", err)
		return nil, err
	}

	// Convert to proto
	requestResponses := make([]*groupv1.JoinRequest, 0, len(requests))
	for _, r := range requests {
		inviterID := (*string)(nil)
		if r.InviterID != "" {
			inviterID = &r.InviterID
		}
		message := (*string)(nil)
		if r.Message != "" {
			message = &r.Message
		}
		response := &groupv1.JoinRequest{
			Id:        r.ID,
			GroupId:   r.GroupID,
			UserId:    r.UserID,
			InviterId: inviterID,
			Message:   message,
			Status:    groupv1.JoinRequestStatus(r.Status),
			CreatedAt: timestamppb.New(r.CreatedAt),
		}
		requestResponses = append(requestResponses, response)
	}

	return &groupv1.GetJoinRequestsResponse{
		Requests: requestResponses,
		Total:    int64(len(requestResponses)),
	}, nil
}

// PinGroupMessage pins a message in group
func (s *groupServiceImpl) PinGroupMessage(ctx context.Context, userID string, req *groupv1.PinGroupMessageRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	messageID := req.MessageId

	group, err := s.groupRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeGroupNotFound, "Group not found")
		}
		return err
	}
	if !group.IsActive() {
		return errors.NewBusiness(errors.CodeGroupDissolved, "Group has been dissolved")
	}

	member, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}
	if !member.CanManageGroup() {
		return errors.NewBusiness(errors.CodeNoAdminPermission, "No permission to pin message")
	}

	exists, err := s.pinnedRepo.Exists(ctx, groupID, messageID)
	if err != nil {
		l.Errorw("msg", "Failed to check pinned message existence", "error", err)
		return err
	}
	if !exists {
		total, err := s.pinnedRepo.CountByGroup(ctx, groupID)
		if err != nil {
			l.Errorw("msg", "Failed to count pinned messages", "error", err)
			return err
		}
		if total >= maxPinnedMessagesPerGroup {
			return errors.NewBusiness(errors.CodeGroupPinnedLimitExceeded, "Pinned message limit reached, please unpin some first")
		}
	}

	snapshot, err := s.getPinnedMessageSnapshot(ctx, groupID, messageID)
	if err != nil {
		return err
	}

	now := time.Now()
	record := &model.GroupPinnedMessage{
		GroupID:     groupID,
		MessageID:   messageID,
		MessageSeq:  snapshot.messageSeq,
		PinnedBy:    userID,
		Content:     snapshot.content,
		ContentType: snapshot.contentType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.pinnedRepo.Upsert(ctx, record); err != nil {
		l.Errorw("msg", "Failed to pin group message", "error", err)
		return err
	}

	s.publishGroupMessagePinnedNotification(groupID, userID, messageID)
	return nil
}

// UnpinGroupMessage unpins a message in group
func (s *groupServiceImpl) UnpinGroupMessage(ctx context.Context, userID string, req *groupv1.UnpinGroupMessageRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	messageID := req.MessageId

	if userID != "" {
		group, err := s.groupRepo.GetByGroupID(ctx, groupID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.NewBusiness(errors.CodeGroupNotFound, "Group not found")
			}
			return err
		}
		if !group.IsActive() {
			return errors.NewBusiness(errors.CodeGroupDissolved, "Group has been dissolved")
		}

		member, err := s.memberRepo.GetMember(ctx, groupID, userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
			}
			return err
		}
		if !member.CanManageGroup() {
			return errors.NewBusiness(errors.CodeNoAdminPermission, "No permission to unpin")
		}
	}

	deleted, err := s.pinnedRepo.Delete(ctx, groupID, messageID)
	if err != nil {
		l.Errorw("msg", "Failed to unpin group message", "error", err)
		return err
	}

	if deleted {
		operatorID := userID
		if operatorID == "" {
			operatorID = systemPinnedMessageOperatorUser
		}
		s.publishGroupMessageUnpinnedNotification(groupID, operatorID, messageID)
	}
	return nil
}

// GetPinnedMessages gets pinned messages
func (s *groupServiceImpl) GetPinnedMessages(ctx context.Context, userID string, req *groupv1.GetPinnedMessagesRequest) (*groupv1.GetPinnedMessagesResponse, error) {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId

	group, err := s.groupRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeGroupNotFound, "Group not found")
		}
		return nil, err
	}
	if !group.IsActive() {
		return nil, errors.NewBusiness(errors.CodeGroupDissolved, "Group has been dissolved")
	}

	isMember, err := s.memberRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
	}

	records, err := s.pinnedRepo.ListByGroup(ctx, groupID)
	if err != nil {
		l.Errorw("msg", "Failed to list pinned messages", "error", err)
		return nil, err
	}

	messages := make([]*groupv1.PinnedMessage, 0, len(records))
	var version int64
	for _, item := range records {
		contentType := (*messagev1.ContentType)(nil)
		if item.ContentType != model.PinnedMessageContentTypeUnspecified {
			ct := messagev1.ContentType(item.ContentType)
			contentType = &ct
		}
		msg := &groupv1.PinnedMessage{
			MessageId:   item.MessageID,
			Content:     item.Content,
			PinnedBy:    item.PinnedBy,
			PinnedAt:    item.CreatedAt.Unix(),
			ContentType: contentType,
			MessageSeq:  item.MessageSeq,
		}
		messages = append(messages, msg)

		updatedAt := item.UpdatedAt
		if updatedAt.IsZero() {
			updatedAt = item.CreatedAt
		}
		if ts := updatedAt.Unix(); ts > version {
			version = ts
		}
	}

	resp := &groupv1.GetPinnedMessagesResponse{
		Messages: messages,
		Total:    int32(len(messages)),
		Version:  version,
	}
	if len(messages) > 0 {
		resp.TopMessage = messages[0]
	}
	return resp, nil
}

// SetGroupMute sets group-wide mute
func (s *groupServiceImpl) SetGroupMute(ctx context.Context, userID string, req *groupv1.SetGroupMuteRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	enabled := req.Enabled

	group, err := s.groupRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeGroupNotFound, "Group not found")
		}
		return err
	}
	if !group.IsActive() {
		return errors.NewBusiness(errors.CodeGroupDissolved, "Group has been dissolved")
	}

	member, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}
	if !member.CanManageGroup() {
		return errors.NewBusiness(errors.CodeNoAdminPermission, "No permission to set group mute")
	}

	if err := s.groupRepo.UpdateFields(ctx, groupID, map[string]any{"is_muted": enabled}); err != nil {
		l.Errorw("msg", "Failed to set group mute", "error", err)
		return err
	}

	messageText := "Group mute disabled"
	if enabled {
		messageText = "Group mute enabled"
	}
	if err := s.sendGroupSystemMessage(ctx, groupID, userID, messageText); err != nil {
		l.Warnw("msg", "Failed to send group mute system message", "groupID", groupID, "error", err)
	}

	s.publishGroupMutedNotification(groupID, userID, enabled)
	return nil
}

// UpdateGroupSettings updates group settings
func (s *groupServiceImpl) UpdateGroupSettings(ctx context.Context, userID string, req *groupv1.UpdateGroupSettingsRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId

	// Permission check: only owner and admin can update settings
	member, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return err
	}

	if !member.CanManageGroup() {
		return errors.NewBusiness(errors.CodeNoAdminPermission, "No permission to update group settings")
	}

	// Build update fields
	updates := make(map[string]any)
	if req.JoinVerify != nil {
		updates["join_verify"] = *req.JoinVerify
	}
	if req.AllowMemberInvite != nil {
		updates["allow_member_invite"] = *req.AllowMemberInvite
	}
	if req.AllowViewHistory != nil {
		updates["allow_view_history"] = *req.AllowViewHistory
	}
	if req.AllowAddFriend != nil {
		updates["allow_add_friend"] = *req.AllowAddFriend
	}
	if req.AllowMemberModify != nil {
		updates["allow_member_modify"] = *req.AllowMemberModify
	}

	if len(updates) == 0 {
		return nil
	}

	// Update settings
	if err := s.settingRepo.UpdateSettings(ctx, groupID, updates); err != nil {
		l.Errorw("msg", "Failed to update group settings", "error", err)
		return err
	}

	s.publishGroupSettingsUpdatedNotification(groupID, userID, updates)
	return nil
}

// GetGroupSettings gets group settings
func (s *groupServiceImpl) GetGroupSettings(ctx context.Context, req *groupv1.GetGroupSettingsRequest) (*groupv1.GetGroupSettingsResponse, error) {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId

	settings, err := s.settingRepo.GetSettings(ctx, groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default settings
			return &groupv1.GetGroupSettingsResponse{
				GroupId:           groupID,
				JoinVerify:        true,
				AllowMemberInvite: true,
				AllowViewHistory:  true,
				AllowAddFriend:    true,
				AllowMemberModify: false,
			}, nil
		}
		l.Errorw("msg", "Failed to get group settings", "error", err)
		return nil, err
	}

	return &groupv1.GetGroupSettingsResponse{
		GroupId:           settings.GroupID,
		JoinVerify:        settings.JoinVerify,
		AllowMemberInvite: settings.AllowMemberInvite,
		AllowViewHistory:  settings.AllowViewHistory,
		AllowAddFriend:    settings.AllowAddFriend,
		AllowMemberModify: settings.AllowMemberModify,
	}, nil
}

// IsMember checks if user is group member (called by other services)
func (s *groupServiceImpl) IsMember(ctx context.Context, req *groupv1.IsMemberRequest) (*groupv1.IsMemberResponse, error) {
	groupID := req.GroupId
	userID := req.UserId

	member, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &groupv1.IsMemberResponse{IsMember: false}, nil
		}
		return nil, err
	}

	return &groupv1.IsMemberResponse{
		IsMember: true,
		Role:     groupv1.GroupRole(member.Role),
	}, nil
}

// UpdateMemberRemark sets/clears group remark (only visible to self)
func (s *groupServiceImpl) UpdateMemberRemark(ctx context.Context, userID string, req *groupv1.UpdateMemberRemarkRequest) error {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId
	remark := req.Remark

	if utf8.RuneCountInString(remark) > 20 {
		return errors.NewBusiness(errors.CodeParamError, "Remark cannot exceed 20 characters")
	}

	isMember, err := s.memberRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
	}

	if err := s.memberRepo.UpdateRemark(ctx, groupID, userID, remark); err != nil {
		l.Errorw("msg", "Failed to update member remark", "error", err)
		return err
	}
	return nil
}

// GetGroupQRCode gets group QR code (returns if valid, auto-renews if expiring, creates if not exists)
func (s *groupServiceImpl) GetGroupQRCode(ctx context.Context, userID string, req *groupv1.GetGroupQRCodeRequest) (*groupv1.GetGroupQRCodeResponse, error) {
	l := s.log.WithContext(ctx)
	groupID := req.GroupId

	isMember, err := s.memberRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
	}

	qr, err := s.qrcodeRepo.GetActiveByGroupID(ctx, groupID)
	if err == nil && qr.IsValid() {
		// Auto-renew when less than 1 day until expiration
		if time.Until(qr.ExpireAt) < model.QRCodeRenewThreshold {
			newExpire := time.Now().Add(model.DefaultQRCodeTTL)
			if renewErr := s.qrcodeRepo.UpdateExpireAt(ctx, qr.Token, newExpire); renewErr != nil {
				l.Warnw("msg", "Failed to renew qrcode", "error", renewErr)
			} else {
				qr.ExpireAt = newExpire
			}
		}
		return buildQRCodeResponse(qr), nil
	}

	// Create new QR code
	return s.createNewQRCode(ctx, userID, groupID)
}

// RefreshGroupQRCode refreshes group QR code (invalidates old one, owner/admin only)
func (s *groupServiceImpl) RefreshGroupQRCode(ctx context.Context, req *groupv1.RefreshGroupQRCodeRequest) (*groupv1.GetGroupQRCodeResponse, error) {
	l := s.log.WithContext(ctx)
	userID := req.UserId
	groupID := req.GroupId

	member, err := s.memberRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeNotGroupMember, "You are not a group member")
		}
		return nil, err
	}
	if !member.CanManageGroup() {
		return nil, errors.NewBusiness(errors.CodeNoAdminPermission, "Only owner or admin can refresh QR code")
	}

	// Invalidate old QR code
	if err := s.qrcodeRepo.InvalidateByGroupID(ctx, groupID); err != nil {
		l.Errorw("msg", "Failed to invalidate old qrcode", "error", err)
		return nil, err
	}

	resp, err := s.createNewQRCode(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	s.publishGroupQRCodeRefreshedNotification(groupID, userID, resp.Token, resp.ExpireAt)
	return resp, nil
}

// GetGroupPreviewByQRCode gets group preview by QR code token
func (s *groupServiceImpl) GetGroupPreviewByQRCode(ctx context.Context, req *groupv1.GetGroupPreviewByQRCodeRequest) (*groupv1.GetGroupPreviewByQRCodeResponse, error) {
	token := req.Token

	qr, err := s.qrcodeRepo.GetByToken(ctx, token)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeGroupQRInvalid, "QR code invalid")
		}
		return nil, err
	}

	if !qr.IsValid() {
		return nil, errors.NewBusiness(errors.CodeGroupQRExpired, "QR code expired, please contact group member to get new one")
	}

	group, err := s.groupRepo.GetByGroupID(ctx, qr.GroupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeGroupNotFound, "Group not found")
		}
		return nil, err
	}
	if !group.IsActive() {
		return nil, errors.NewBusiness(errors.CodeGroupDissolved, "Group has been dissolved")
	}

	return &groupv1.GetGroupPreviewByQRCodeResponse{
		GroupId:     group.GroupID,
		Name:        group.Name,
		Avatar:      group.Avatar,
		MemberCount: group.MemberCount,
		NeedVerify:  s.getGroupJoinVerify(ctx, group.GroupID),
	}, nil
}

// JoinGroupByQRCode joins group by QR code
func (s *groupServiceImpl) JoinGroupByQRCode(ctx context.Context, userID string, req *groupv1.JoinGroupByQRCodeRequest) (*groupv1.JoinGroupByQRCodeResponse, error) {
	l := s.log.WithContext(ctx)
	token := req.Token

	qr, err := s.qrcodeRepo.GetByToken(ctx, token)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeGroupQRInvalid, "QR code invalid")
		}
		return nil, err
	}

	if !qr.IsValid() {
		return nil, errors.NewBusiness(errors.CodeGroupQRExpired, "QR code expired, please contact group member to get new one")
	}

	groupID := qr.GroupID

	group, err := s.groupRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeGroupNotFound, "Group not found")
		}
		return nil, err
	}
	if !group.IsActive() {
		return nil, errors.NewBusiness(errors.CodeGroupDissolved, "Group has been dissolved")
	}
	if group.IsFull() {
		return nil, errors.NewBusiness(errors.CodeGroupMemberLimitReached, "Group member limit reached")
	}

	isMember, err := s.memberRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, errors.NewBusiness(errors.CodeAlreadyGroupMember, "You are already a group member")
	}

	// Reuse validation logic from JoinGroup
	joinVerify := s.getGroupJoinVerify(ctx, groupID)
	if joinVerify {
		// Check if already has pending request
		existingReq, _ := s.joinRequestRepo.GetExistingRequest(ctx, groupID, userID)
		if existingReq != nil {
			return &groupv1.JoinGroupByQRCodeResponse{
				Joined:     false,
				GroupId:    groupID,
				NeedVerify: true,
				RequestId:  &existingReq.ID,
			}, nil
		}

		request := &model.GroupJoinRequest{
			GroupID:   groupID,
			UserID:    userID,
			Status:    model.JoinRequestStatusPending,
			CreatedAt: time.Now(),
		}
		if err := s.joinRequestRepo.Create(ctx, request); err != nil {
			return nil, err
		}
		s.publishGroupJoinRequestedNotification(groupID, userID, request.ID, "qrcode")
		return &groupv1.JoinGroupByQRCodeResponse{
			Joined:     false,
			GroupId:    groupID,
			NeedVerify: true,
			RequestId:  &request.ID,
		}, nil
	}

	// Directly join
	err = s.db.Transaction(func(tx *gorm.DB) error {
		memberRepoTx := s.memberRepo.WithTx(tx)
		groupRepoTx := s.groupRepo.WithTx(tx)

		if err := memberRepoTx.AddMember(ctx, &model.GroupMember{
			GroupID:  groupID,
			UserID:   userID,
			Role:     model.GroupRoleMember,
			JoinedAt: time.Now(),
		}); err != nil {
			return err
		}
		return groupRepoTx.UpdateMemberCount(ctx, groupID, 1)
	})
	if err != nil {
		l.Errorw("msg", "Failed to join group by qrcode", "error", err)
		return nil, err
	}

	s.publishMemberJoinedNotification(groupID, userID, qr.CreatedBy)

	return &groupv1.JoinGroupByQRCodeResponse{
		Joined:     true,
		GroupId:    groupID,
		NeedVerify: false,
	}, nil
}

// createNewQRCode generates and persists a new QR code record
func (s *groupServiceImpl) createNewQRCode(ctx context.Context, userID, groupID string) (*groupv1.GetGroupQRCodeResponse, error) {
	l := s.log.WithContext(ctx)
	qr := &model.GroupQRCode{
		GroupID:   groupID,
		Token:     uuid.NewString(),
		CreatedBy: userID,
		ExpireAt:  time.Now().Add(model.DefaultQRCodeTTL),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.qrcodeRepo.Create(ctx, qr); err != nil {
		l.Errorw("msg", "Failed to create group qrcode", "error", err)
		return nil, err
	}
	return buildQRCodeResponse(qr), nil
}

func buildQRCodeResponse(qr *model.GroupQRCode) *groupv1.GetGroupQRCodeResponse {
	return &groupv1.GetGroupQRCodeResponse{
		Token:    qr.Token,
		DeepLink: fmt.Sprintf("anychat://join/group?token=%s", qr.Token),
		ExpireAt: qr.ExpireAt.Unix(),
	}
}

func (s *groupServiceImpl) getGroupJoinVerify(ctx context.Context, groupID string) bool {
	settings, err := s.settingRepo.GetSettings(ctx, groupID)
	if err != nil {
		return true
	}
	return settings.JoinVerify
}

func (s *groupServiceImpl) getPinnedMessageSnapshot(ctx context.Context, groupID, messageID string) (*pinnedMessageSnapshot, error) {
	l := s.log.WithContext(ctx)
	if s.messageService == nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "Message service unavailable")
	}

	msg, err := s.messageService.GetMessageById(ctx, &messagev1.GetMessageByIdRequest{
		MessageId: messageID,
	})
	if err != nil {
		l.Errorw("msg", "Failed to load pinned message content", "messageId", messageID, "error", err)
		return nil, errors.NewBusiness(errors.CodeMessageNotFound, "Message not found")
	}

	if msg.GetConversationId() != groupID {
		return nil, errors.NewBusiness(errors.CodeMessageNotInGroup, "Message does not belong to this group")
	}

	contentType := model.PinnedMessageContentType(msg.GetContentType())
	if contentType == model.PinnedMessageContentTypeUnspecified {
		contentType = model.PinnedMessageContentTypeText
	}

	var seq *int64
	if messageSeq := msg.GetSeq(); messageSeq > 0 {
		seq = &messageSeq
	}

	return &pinnedMessageSnapshot{
		content:     buildPinnedMessagePreview(msg.GetContent(), contentType),
		contentType: contentType,
		messageSeq:  seq,
	}, nil
}

func buildPinnedMessagePreview(content string, contentType model.PinnedMessageContentType) string {
	switch contentType {
	case model.PinnedMessageContentTypeText:
		var textContent struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(content), &textContent); err == nil {
			text := strings.TrimSpace(textContent.Text)
			if text != "" {
				return truncateWithEllipsis(text, pinnedMessagePreviewMaxRuneCount)
			}
		}
		rawText := strings.TrimSpace(content)
		if rawText != "" {
			return truncateWithEllipsis(rawText, pinnedMessagePreviewMaxRuneCount)
		}
		return "[Text]"
	case model.PinnedMessageContentTypeImage:
		return "[Image]"
	case model.PinnedMessageContentTypeVideo:
		return "[Video]"
	case model.PinnedMessageContentTypeAudio:
		return "[Voice]"
	case model.PinnedMessageContentTypeFile:
		return "[File]"
	case model.PinnedMessageContentTypeLocation:
		return "[Location]"
	case model.PinnedMessageContentTypeContact:
		return "[Contact]"
	case model.PinnedMessageContentTypeSticker:
		return "[Sticker]"
	case model.PinnedMessageContentTypeEmoticon:
		return "[Emoticon]"
	case model.PinnedMessageContentTypeRecall:
		return "[Recall]"
	case model.PinnedMessageContentTypeSystem:
		return "[System]"
	default:
		return "[Message]"
	}
}

func truncateWithEllipsis(text string, maxRunes int) string {
	if maxRunes <= 0 || utf8.RuneCountInString(text) <= maxRunes {
		return text
	}
	runes := []rune(text)
	return string(runes[:maxRunes]) + "..."
}

func (s *groupServiceImpl) sendGroupSystemMessage(ctx context.Context, groupID, operatorID, text string) error {
	if s.messageService == nil {
		return nil
	}

	content, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}

	ctx = metadata.AppendToOutgoingContext(ctx, consts.OperatorUserIDMetadataKey, operatorID)
	_, err = s.messageService.SendMessage(ctx, &messagev1.SendMessageRequest{
		ConversationId: groupID,
		ContentType:    messagev1.ContentType_CONTENT_TYPE_TEXT,
		Content:        string(content),
		LocalId:        uuid.NewString(),
	})
	return err
}

// publishMemberJoinedNotification publishes member joined event
func (s *groupServiceImpl) publishMemberJoinedNotification(groupID, userID, inviterID string) {
	payload := map[string]interface{}{
		"group_id":        groupID,
		"user_id":         userID,
		"inviter_user_id": inviterID,
		"joined_at":       time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupMemberJoined,
		inviterID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish member joined event", "groupId", groupID, "error", err)
	}
}

// publishGroupJoinRequestedNotification publishes join request event
func (s *groupServiceImpl) publishGroupJoinRequestedNotification(groupID, userID string, requestID int64, source string) {
	payload := map[string]interface{}{
		"group_id":   groupID,
		"user_id":    userID,
		"request_id": requestID,
		"source":     source,
		"created_at": time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupJoinRequested,
		userID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish group join requested event", "groupId", groupID, "requestId", requestID, "error", err)
	}
}

// publishGroupQRCodeRefreshedNotification publishes QR code refreshed event
func (s *groupServiceImpl) publishGroupQRCodeRefreshedNotification(groupID, operatorID, token string, expireAt int64) {
	payload := map[string]interface{}{
		"group_id":         groupID,
		"operator_user_id": operatorID,
		"token":            token,
		"expire_at":        expireAt,
		"refreshed_at":     time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupQRCodeRefreshed,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish group qrcode refreshed event", "groupId", groupID, "error", err)
	}
}

// publishMemberLeftNotification publishes member left event
func (s *groupServiceImpl) publishMemberLeftNotification(groupID, userID, operatorID, reason string) {
	payload := map[string]interface{}{
		"group_id":         groupID,
		"user_id":          userID,
		"operator_user_id": operatorID,
		"reason":           reason,
		"left_at":          time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupMemberLeft,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish member left event", "groupId", groupID, "error", err)
	}
}

// publishGroupInfoUpdatedNotification publishes group info updated event
func (s *groupServiceImpl) publishGroupInfoUpdatedNotification(groupID, operatorID string, req *groupv1.UpdateGroupRequest) {
	updatedFields := []string{}
	payload := map[string]interface{}{
		"group_id":         groupID,
		"operator_user_id": operatorID,
		"updated_at":       time.Now().Unix(),
	}

	if req.Name != nil {
		updatedFields = append(updatedFields, "name")
		payload["group_name"] = *req.Name
	}
	if req.Avatar != nil {
		updatedFields = append(updatedFields, "avatar")
		payload["group_avatar"] = *req.Avatar
	}
	if req.Announcement != nil {
		updatedFields = append(updatedFields, "announcement")
		payload["announcement"] = *req.Announcement
	}
	if req.Description != nil {
		updatedFields = append(updatedFields, "description")
		payload["description"] = *req.Description
	}

	payload["updated_fields"] = updatedFields

	notif := broker.NewNotification(
		broker.TypeGroupInfoUpdated,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish group info updated event", "groupId", groupID, "error", err)
	}
}

// publishRoleChangedNotification publishes role changed event
func (s *groupServiceImpl) publishRoleChangedNotification(
	groupID, userID, operatorID string,
	oldRole, newRole model.GroupRole,
) {
	payload := map[string]interface{}{
		"group_id":         groupID,
		"user_id":          userID,
		"old_role":         oldRole,
		"new_role":         newRole,
		"operator_user_id": operatorID,
		"changed_at":       time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupRoleChanged,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish role changed event", "groupId", groupID, "error", err)
	}
}

func (s *groupServiceImpl) publishMemberMutedNotification(groupID, operatorID, targetUserID string, mutedUntil *time.Time) {
	var mutedUntilUnix int64
	if mutedUntil != nil {
		mutedUntilUnix = mutedUntil.Unix()
	}

	payload := map[string]interface{}{
		"group_id":         groupID,
		"target_user_id":   targetUserID,
		"operator_user_id": operatorID,
		"muted_until":      mutedUntilUnix,
		"updated_at":       time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupMemberMuted,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish member muted event", "groupId", groupID, "error", err)
	}
}

func (s *groupServiceImpl) publishMemberUnmutedNotification(groupID, operatorID, targetUserID string) {
	payload := map[string]interface{}{
		"group_id":         groupID,
		"target_user_id":   targetUserID,
		"operator_user_id": operatorID,
		"updated_at":       time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupMemberUnmuted,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish member unmuted event", "groupId", groupID, "error", err)
	}
}

// publishGroupMutedNotification publishes group muted event
func (s *groupServiceImpl) publishGroupMutedNotification(groupID, operatorID string, enabled bool) {
	payload := map[string]interface{}{
		"group_id":         groupID,
		"operator_user_id": operatorID,
		"enabled":          enabled,
		"updated_at":       time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupMuted,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish group muted event", "groupId", groupID, "error", err)
	}
}

func (s *groupServiceImpl) publishGroupSettingsUpdatedNotification(groupID, operatorID string, updates map[string]any) {
	updatedFields := make([]string, 0, len(updates))
	for key := range updates {
		updatedFields = append(updatedFields, key)
	}

	payload := map[string]interface{}{
		"group_id":         groupID,
		"operator_user_id": operatorID,
		"updated_fields":   updatedFields,
		"updated_at":       time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupSettingsUpdated,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish group settings updated event", "groupId", groupID, "error", err)
	}
}

func (s *groupServiceImpl) publishGroupMessagePinnedNotification(groupID, operatorID, messageID string) {
	payload := map[string]interface{}{
		"group_id":         groupID,
		"operator_user_id": operatorID,
		"message_id":       messageID,
		"pinned_at":        time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupMessagePinned,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish group message pinned event", "groupId", groupID, "error", err)
	}
}

func (s *groupServiceImpl) publishGroupMessageUnpinnedNotification(groupID, operatorID, messageID string) {
	payload := map[string]interface{}{
		"group_id":         groupID,
		"operator_user_id": operatorID,
		"message_id":       messageID,
		"unpinned_at":      time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupMessageUnpinned,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish group message unpinned event", "groupId", groupID, "error", err)
	}
}

func (s *groupServiceImpl) publishGroupDisbandedNotification(groupID, operatorID, groupName string) {
	payload := map[string]interface{}{
		"group_id":         groupID,
		"operator_user_id": operatorID,
		"group_name":       groupName,
		"disbanded_at":     time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeGroupDisbanded,
		operatorID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToGroup(groupID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish group disbanded event", "groupId", groupID, "error", err)
	}
}

// ptrStr returns a pointer to the given string value
func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
