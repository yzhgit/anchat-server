package handler

import (
	"context"

	friendv1 "flamingo/api/friend/v1"

	"flamingo/app/friend/internal/model"
	"flamingo/pkg/errors"
	"flamingo/pkg/md"

	"github.com/go-kratos/kratos/v2/log"
	empty "github.com/golang/protobuf/ptypes/empty"
)

// FriendService is the friend service interface
type FriendService interface {
	GetFriendList(ctx context.Context, userID string) (*friendv1.GetFriendListResponse, error)
	SendFriendRequest(ctx context.Context, userID string, req *friendv1.SendFriendRequestRequest) error
	HandleFriendRequest(ctx context.Context, userID string, req *friendv1.HandleFriendRequestRequest) error
	GetFriendRequests(ctx context.Context, userID string, requestType model.FriendRequestQueryType) (*friendv1.GetFriendRequestsResponse, error)
	DeleteFriend(ctx context.Context, userID, friendID string) error
	UpdateRemark(ctx context.Context, userID, friendID, remark string) error
	AddToBlacklist(ctx context.Context, userID, blockedUserID string) error
	RemoveFromBlacklist(ctx context.Context, userID, blockedUserID string) error
	GetBlacklist(ctx context.Context, userID string) (*friendv1.GetBlacklistResponse, error)
	IsFriend(ctx context.Context, userID, friendID string) (bool, error)
	IsBlocked(ctx context.Context, userID, targetUserID string) (bool, error)
	BatchCheckFriend(ctx context.Context, userID string, friendIDs []string) (map[string]bool, error)
}

type FriendHandler struct {
	friendv1.UnimplementedFriendServiceServer
	svc FriendService
	log *log.Helper
}

// NewFriendServer creates a new friend gRPC handler
func NewFriendHandler(svc FriendService, logger log.Logger) *FriendHandler {
	return &FriendHandler{
		svc: svc,
		log: log.NewHelper(logger),
	}
}

// GetFriendList retrieves the friend list
func (s *FriendHandler) GetFriendList(ctx context.Context, req *friendv1.GetFriendListRequest) (*friendv1.GetFriendListResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.GetFriendList(ctx, userID)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// SendFriendRequest sends a friend request
func (s *FriendHandler) SendFriendRequest(ctx context.Context, req *friendv1.SendFriendRequestRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.ToUserId == "" {
		return nil, errors.BadRequest(ctx, "to_user_id is required")
	}
	if userID == req.ToUserId {
		return nil, errors.BadRequest(ctx, "cannot add yourself as a friend")
	}
	source := model.FriendRequestSource(req.Source)
	if !source.IsValid() {
		return nil, errors.BadRequest(ctx, "invalid source")
	}

	err := s.svc.SendFriendRequest(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return &empty.Empty{}, nil
}

// HandleFriendRequest handles a friend request
func (s *FriendHandler) HandleFriendRequest(ctx context.Context, req *friendv1.HandleFriendRequestRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	action := model.FriendRequestAction(req.Action)
	if !action.IsValid() {
		return nil, errors.BadRequest(ctx, "invalid action")
	}

	err := s.svc.HandleFriendRequest(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return &empty.Empty{}, nil
}

// GetFriendRequests retrieves the friend request list
func (s *FriendHandler) GetFriendRequests(ctx context.Context, req *friendv1.GetFriendRequestsRequest) (*friendv1.GetFriendRequestsResponse, error) {
	userID := md.MustGetUserID(ctx)
	queryType := model.FriendRequestQueryType(req.RequestType)
	if queryType == model.FriendRequestQueryTypeUnknown {
		queryType = model.FriendRequestQueryTypeReceived
	}
	if !queryType.IsValid() {
		return nil, errors.BadRequest(ctx, "invalid request_type")
	}

	resp, err := s.svc.GetFriendRequests(ctx, userID, queryType)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// DeleteFriend deletes a friend
func (s *FriendHandler) DeleteFriend(ctx context.Context, req *friendv1.DeleteFriendRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.FriendId == "" {
		return nil, errors.BadRequest(ctx, "friend_id is required")
	}
	err := s.svc.DeleteFriend(ctx, userID, req.FriendId)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// UpdateRemark updates friend remark
func (s *FriendHandler) UpdateRemark(ctx context.Context, req *friendv1.UpdateRemarkRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.FriendId == "" {
		return nil, errors.BadRequest(ctx, "friend_id is required")
	}
	err := s.svc.UpdateRemark(ctx, userID, req.FriendId, req.Remark)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// AddToBlacklist adds user to blacklist
func (s *FriendHandler) AddToBlacklist(ctx context.Context, req *friendv1.AddToBlacklistRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.BlockedUserId == "" {
		return nil, errors.BadRequest(ctx, "blocked_user_id is required")
	}
	if userID == req.BlockedUserId {
		return nil, errors.BadRequest(ctx, "cannot block yourself")
	}
	err := s.svc.AddToBlacklist(ctx, userID, req.BlockedUserId)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// RemoveFromBlacklist removes user from blacklist
func (s *FriendHandler) RemoveFromBlacklist(ctx context.Context, req *friendv1.RemoveFromBlacklistRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.BlockedUserId == "" {
		return nil, errors.BadRequest(ctx, "blocked_user_id is required")
	}
	err := s.svc.RemoveFromBlacklist(ctx, userID, req.BlockedUserId)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// GetBlacklist retrieves the blacklist
func (s *FriendHandler) GetBlacklist(ctx context.Context, req *friendv1.GetBlacklistRequest) (*friendv1.GetBlacklistResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.GetBlacklist(ctx, userID)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// IsFriend checks if users are friends
func (s *FriendHandler) IsFriend(ctx context.Context, req *friendv1.IsFriendRequest) (*friendv1.IsFriendResponse, error) {
	isFriend, err := s.svc.IsFriend(ctx, req.UserId, req.TargetUserId)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &friendv1.IsFriendResponse{
		IsFriend: isFriend,
	}, nil
}

// IsBlocked checks if user is blocked
func (s *FriendHandler) IsBlocked(ctx context.Context, req *friendv1.IsBlockedRequest) (*friendv1.IsBlockedResponse, error) {
	isBlocked, err := s.svc.IsBlocked(ctx, req.UserId, req.TargetUserId)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &friendv1.IsBlockedResponse{
		IsBlocked: isBlocked,
	}, nil
}

// BatchCheckFriend batch checks friend relationships
func (s *FriendHandler) BatchCheckFriend(ctx context.Context, req *friendv1.BatchCheckFriendRequest) (*friendv1.BatchCheckFriendResponse, error) {
	results, err := s.svc.BatchCheckFriend(ctx, req.UserId, req.TargetUserIds)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &friendv1.BatchCheckFriendResponse{
		FriendshipStatus: results,
	}, nil
}
