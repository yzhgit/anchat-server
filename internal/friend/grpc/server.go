package grpc

import (
	"context"

	commonpb "github.com/anychat/server/api/proto/common"
	friendpb "github.com/anychat/server/api/proto/friend"
	"github.com/anychat/server/internal/friend/dto"
	"github.com/anychat/server/internal/friend/model"
	"github.com/anychat/server/internal/friend/service"
	"github.com/anychat/server/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FriendServer is the friend gRPC server
type FriendServer struct {
	friendpb.UnimplementedFriendServiceServer
	svc service.FriendService
}

// NewFriendServer creates a new friend gRPC server
func NewFriendServer(service service.FriendService) *FriendServer {
	return &FriendServer{
		svc: service,
	}
}

// GetFriendList retrieves the friend list
func (s *FriendServer) GetFriendList(ctx context.Context, req *friendpb.GetFriendListRequest) (*friendpb.GetFriendListResponse, error) {
	// Call service layer
	resp, err := s.svc.GetFriendList(ctx, req.UserId, req.LastUpdateTime)
	if err != nil {
		return nil, convertError(err)
	}

	// DTO -> Proto conversion
	friends := make([]*friendpb.Friend, 0, len(resp.Friends))
	for _, f := range resp.Friends {
		friend := &friendpb.Friend{
			UserId:    f.UserID,
			Remark:    f.Remark,
			CreatedAt: timestamppb.New(f.CreatedAt),
			UpdatedAt: timestamppb.New(f.UpdatedAt),
		}
		if f.UserInfo != nil {
			friend.UserInfo = &commonpb.UserInfo{
				UserId:   f.UserInfo.UserID,
				Nickname: f.UserInfo.Nickname,
				Avatar:   f.UserInfo.Avatar,
			}
		}
		friends = append(friends, friend)
	}

	return &friendpb.GetFriendListResponse{
		Friends: friends,
		Total:   resp.Total,
	}, nil
}

// SendFriendRequest sends a friend request
func (s *FriendServer) SendFriendRequest(ctx context.Context, req *friendpb.SendFriendRequestRequest) (*friendpb.SendFriendRequestResponse, error) {
	source := model.FriendRequestSource(req.Source)
	if !source.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid source")
	}

	// Proto -> DTO conversion
	dtoReq := &dto.SendFriendRequestRequest{
		UserID:  req.ToUserId,
		Message: req.Message,
		Source:  source,
	}

	// Call service layer
	resp, err := s.svc.SendFriendRequest(ctx, req.FromUserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &friendpb.SendFriendRequestResponse{
		RequestId:    resp.RequestID,
		AutoAccepted: resp.AutoAccepted,
	}, nil
}

// HandleFriendRequest handles a friend request
func (s *FriendServer) HandleFriendRequest(ctx context.Context, req *friendpb.HandleFriendRequestRequest) (*commonpb.Empty, error) {
	action := model.FriendRequestAction(req.Action)
	if !action.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid action")
	}

	// Proto -> DTO conversion
	dtoReq := &dto.HandleFriendRequestRequest{
		Action: int16(action),
	}

	// Call service layer
	err := s.svc.HandleFriendRequest(ctx, req.UserId, req.RequestId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// GetFriendRequests retrieves the friend request list
func (s *FriendServer) GetFriendRequests(ctx context.Context, req *friendpb.GetFriendRequestsRequest) (*friendpb.GetFriendRequestsResponse, error) {
	queryType := model.FriendRequestQueryType(req.RequestType)
	if queryType == model.FriendRequestQueryTypeUnknown {
		queryType = model.FriendRequestQueryTypeReceived
	}
	if !queryType.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid request_type")
	}

	// Call service layer
	resp, err := s.svc.GetFriendRequests(ctx, req.UserId, queryType)
	if err != nil {
		return nil, convertError(err)
	}

	// DTO -> Proto conversion
	requests := make([]*friendpb.FriendRequest, 0, len(resp.Requests))
	for _, r := range resp.Requests {
		request := &friendpb.FriendRequest{
			Id:         r.ID,
			FromUserId: r.FromUserID,
			ToUserId:   r.ToUserID,
			Message:    r.Message,
			Source:     friendpb.FriendRequestSource(r.Source),
			Status:     friendpb.FriendRequestStatus(r.Status),
			CreatedAt:  timestamppb.New(r.CreatedAt),
		}
		if r.FromUserInfo != nil {
			request.FromUserInfo = &commonpb.UserInfo{
				UserId:   r.FromUserInfo.UserID,
				Nickname: r.FromUserInfo.Nickname,
				Avatar:   r.FromUserInfo.Avatar,
			}
		}
		requests = append(requests, request)
	}

	return &friendpb.GetFriendRequestsResponse{
		Requests: requests,
		Total:    resp.Total,
	}, nil
}

// DeleteFriend deletes a friend
func (s *FriendServer) DeleteFriend(ctx context.Context, req *friendpb.DeleteFriendRequest) (*commonpb.Empty, error) {
	err := s.svc.DeleteFriend(ctx, req.UserId, req.FriendId)
	if err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

// UpdateRemark updates friend remark
func (s *FriendServer) UpdateRemark(ctx context.Context, req *friendpb.UpdateRemarkRequest) (*commonpb.Empty, error) {
	dtoReq := &dto.UpdateRemarkRequest{
		Remark: req.Remark,
	}

	err := s.svc.UpdateRemark(ctx, req.UserId, req.FriendId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

// AddToBlacklist adds user to blacklist
func (s *FriendServer) AddToBlacklist(ctx context.Context, req *friendpb.AddToBlacklistRequest) (*commonpb.Empty, error) {
	dtoReq := &dto.AddToBlacklistRequest{
		UserId: req.BlockedUserId,
	}

	err := s.svc.AddToBlacklist(ctx, req.UserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

// RemoveFromBlacklist removes user from blacklist
func (s *FriendServer) RemoveFromBlacklist(ctx context.Context, req *friendpb.RemoveFromBlacklistRequest) (*commonpb.Empty, error) {
	err := s.svc.RemoveFromBlacklist(ctx, req.UserId, req.BlockedUserId)
	if err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

// GetBlacklist retrieves the blacklist
func (s *FriendServer) GetBlacklist(ctx context.Context, req *friendpb.GetBlacklistRequest) (*friendpb.GetBlacklistResponse, error) {
	resp, err := s.svc.GetBlacklist(ctx, req.UserId)
	if err != nil {
		return nil, convertError(err)
	}

	// DTO -> Proto conversion
	items := make([]*friendpb.BlacklistItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		pbItem := &friendpb.BlacklistItem{
			Id:            item.ID,
			UserId:        item.UserID,
			BlockedUserId: item.BlockedUserID,
			CreatedAt:     timestamppb.New(item.CreatedAt),
		}
		if item.BlockedUserInfo != nil {
			pbItem.BlockedUserInfo = &commonpb.UserInfo{
				UserId:   item.BlockedUserInfo.UserID,
				Nickname: item.BlockedUserInfo.Nickname,
				Avatar:   item.BlockedUserInfo.Avatar,
			}
		}
		items = append(items, pbItem)
	}

	return &friendpb.GetBlacklistResponse{
		Items: items,
		Total: resp.Total,
	}, nil
}

// IsFriend checks if users are friends
func (s *FriendServer) IsFriend(ctx context.Context, req *friendpb.IsFriendRequest) (*friendpb.IsFriendResponse, error) {
	isFriend, err := s.svc.IsFriend(ctx, req.UserId, req.FriendId)
	if err != nil {
		return nil, convertError(err)
	}
	return &friendpb.IsFriendResponse{
		IsFriend: isFriend,
	}, nil
}

// IsBlocked checks if user is blocked
func (s *FriendServer) IsBlocked(ctx context.Context, req *friendpb.IsBlockedRequest) (*friendpb.IsBlockedResponse, error) {
	isBlocked, err := s.svc.IsBlocked(ctx, req.UserId, req.TargetUserId)
	if err != nil {
		return nil, convertError(err)
	}
	return &friendpb.IsBlockedResponse{
		IsBlocked: isBlocked,
	}, nil
}

// BatchCheckFriend batch checks friend relationships
func (s *FriendServer) BatchCheckFriend(ctx context.Context, req *friendpb.BatchCheckFriendRequest) (*friendpb.BatchCheckFriendResponse, error) {
	results, err := s.svc.BatchCheckFriend(ctx, req.UserId, req.FriendIds)
	if err != nil {
		return nil, convertError(err)
	}
	return &friendpb.BatchCheckFriendResponse{
		Results: results,
	}, nil
}

// convertError converts business errors to gRPC errors
func convertError(err error) error {
	if bizErr, ok := err.(*errors.Business); ok {
		switch bizErr.Code {
		case errors.CodeParamError:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeCannotAddSelf:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeUserNotFound:
			return status.Error(codes.NotFound, bizErr.Message)
		case errors.CodeAlreadyFriend:
			return status.Error(codes.AlreadyExists, bizErr.Message)
		case errors.CodeAlreadyInBlacklist:
			return status.Error(codes.AlreadyExists, bizErr.Message)
		case errors.CodeRequestExists:
			return status.Error(codes.AlreadyExists, bizErr.Message)
		case errors.CodeUserBlocked:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeRequestNotFound:
			return status.Error(codes.NotFound, bizErr.Message)
		case errors.CodePermissionDenied:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeNotFriend:
			return status.Error(codes.FailedPrecondition, bizErr.Message)
		case errors.CodeRequestProcessed:
			return status.Error(codes.FailedPrecondition, bizErr.Message)
		case errors.CodeNotInBlacklist:
			return status.Error(codes.FailedPrecondition, bizErr.Message)
		default:
			return status.Error(codes.Internal, bizErr.Message)
		}
	}
	return status.Error(codes.Internal, err.Error())
}
