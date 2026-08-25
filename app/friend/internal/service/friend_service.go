package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	conversationv1 "flamingo/api/conversation/v1"
	friendv1 "flamingo/api/friend/v1"
	userv1 "flamingo/api/user/v1"

	"flamingo/app/friend/internal/handler"
	"flamingo/app/friend/internal/model"
	"flamingo/pkg/broker"
	"flamingo/pkg/errors"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// friendServiceImpl is the friend service implementation
type friendServiceImpl struct {
	friendshipRepo     FriendshipRepository
	requestRepo        FriendRequestRepository
	blacklistRepo      BlacklistRepository
	userClient         userv1.UserServiceClient
	conversationClient conversationv1.ConversationServiceClient
	broker             broker.Broker
	db                 *gorm.DB
	log                *log.Helper
}

var _ handler.FriendService = (*friendServiceImpl)(nil)

func ptrString(s string) *string { return &s }

// NewFriendService creates a new friend service
func NewFriendService(
	friendshipRepo FriendshipRepository,
	requestRepo FriendRequestRepository,
	blacklistRepo BlacklistRepository,
	userClient userv1.UserServiceClient,
	conversationClient conversationv1.ConversationServiceClient,
	broker broker.Broker,
	db *gorm.DB,
	logger log.Logger,
) handler.FriendService {
	return &friendServiceImpl{
		friendshipRepo:     friendshipRepo,
		requestRepo:        requestRepo,
		blacklistRepo:      blacklistRepo,
		userClient:         userClient,
		conversationClient: conversationClient,
		broker:             broker,
		db:                 db,
		log:                log.NewHelper(logger),
	}
}

// GetFriendList retrieves the friend list
func (s *friendServiceImpl) GetFriendList(ctx context.Context, userID string) (*friendv1.GetFriendListResponse, error) {
	l := s.log.WithContext(ctx)
	friendships, err := s.friendshipRepo.GetFriendList(ctx, userID)

	if err != nil {
		l.Errorw("msg", "Failed to get friend list", "error", err)
		return nil, err
	}

	friends := make([]*friendv1.Friend, 0, len(friendships))
	for _, f := range friendships {
		friend := &friendv1.Friend{
			UserId: f.FriendID,
			Remark: f.Remark,
		}

		// Get user info (optional, failure does not affect overall result)
		if userInfo, err := s.getUserInfo(ctx, f.FriendID); err == nil {
			friend.Nickname = userInfo.Nickname
			friend.Avatar = userInfo.Avatar
		}

		friends = append(friends, friend)
	}

	return &friendv1.GetFriendListResponse{
		Friends: friends,
	}, nil
}

// SendFriendRequest sends a friend request
func (s *friendServiceImpl) SendFriendRequest(ctx context.Context, userID string, req *friendv1.SendFriendRequestRequest) error {
	l := s.log.WithContext(ctx)
	fromUserID := userID
	toUserID := req.ToUserId

	// Check blacklist
	isBlocked, err := s.blacklistRepo.IsBlocked(ctx, fromUserID, toUserID)
	if err != nil {
		return err
	}
	if isBlocked {
		return errors.NewBusiness(errors.CodeUserBlocked, "")
	}

	// Check if already a friend
	isFriend, err := s.friendshipRepo.IsFriend(ctx, fromUserID, toUserID)
	if err != nil {
		return err
	}
	if isFriend {
		return errors.NewBusiness(errors.CodeAlreadyFriend, "")
	}

	// Check if there is a pending request
	existingReq, err := s.requestRepo.GetPendingRequest(ctx, fromUserID, toUserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if existingReq != nil {
		return errors.NewBusiness(errors.CodeRequestExists, "")
	}

	// Get recipient settings, check if verification is required.
	// Pass toUserID explicitly — without it the user service would return
	// the sender's settings (from metadata), bypassing the recipient's
	// friend-verification preference.
	settingsResp, err := s.userClient.GetSettings(ctx, &userv1.GetSettingsRequest{
		UserId: ptrString(toUserID),
	})
	if err != nil {
		l.Errorw("msg", "Failed to get user settings", "error", err)
		return err
	}
	friendVerifyRequired := settingsResp.FriendVerifyRequired

	// Use transaction to handle request creation and possible auto-accept
	var autoAccepted bool
	err = s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// Create friend request
		status := model.FriendRequestStatusPending
		if !friendVerifyRequired {
			status = model.FriendRequestStatusAccepted
		}

		friendRequest := &model.FriendRequest{
			FromUserID: fromUserID,
			ToUserID:   toUserID,
			Message:    req.Message,
			Source:     model.FriendRequestSource(req.Source),
			Status:     status,
		}

		requestRepoTx := s.requestRepo.WithTx(tx)
		if err := requestRepoTx.Create(ctx, friendRequest); err != nil {
			l.Errorw("msg", "Failed to create friend request", "error", err)
			return err
		}

		// If recipient does not require verification, auto accept
		if !friendVerifyRequired {
			autoAccepted = true

			// Create bidirectional friendship
			friendships := []*model.Friendship{
				{
					UserID:    fromUserID,
					FriendID:  toUserID,
					Status:    model.FriendshipStatusNormal,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					UserID:    toUserID,
					FriendID:  fromUserID,
					Status:    model.FriendshipStatusNormal,
					CreatedAt: now,
					UpdatedAt: now,
				},
			}

			friendshipRepoTx := s.friendshipRepo.WithTx(tx)
			if err := friendshipRepoTx.CreateBatch(ctx, friendships); err != nil {
				l.Errorw("msg", "Failed to create friendship", "error", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Get newly created request ID
	createdReq, err := s.requestRepo.GetByUserIDs(ctx, fromUserID, toUserID)
	if err != nil {
		l.Errorw("msg", "Failed to get created friend request", "error", err)
		return err
	}
	requestID := createdReq.ID

	// Based on auto accept status, send different notifications
	if autoAccepted {
		s.createFriendConversation(ctx, fromUserID, toUserID)
		s.createFriendConversation(ctx, toUserID, fromUserID)

		s.publishFriendAddedNotification(fromUserID, toUserID)
		s.publishFriendAddedNotification(toUserID, fromUserID)
	} else {
		s.publishFriendRequestNotification(createdReq)
	}

	// Log the request ID for debugging
	l.Infow("msg", "SendFriendRequest completed",
		"request_id", requestID,
		"from", fromUserID,
		"to", toUserID,
		"auto_accepted", autoAccepted)

	return nil
}

// HandleFriendRequest handles a friend request
func (s *friendServiceImpl) HandleFriendRequest(ctx context.Context, userID string, req *friendv1.HandleFriendRequestRequest) error {
	l := s.log.WithContext(ctx)
	// Parse request ID from string
	var requestID int64
	if req.RequestId != "" {
		requestID, _ = strconv.ParseInt(req.RequestId, 10, 64)
	}

	// Get request record
	friendRequest, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeRequestNotFound, "")
		}
		return err
	}

	// Validate permission: only recipient can handle
	if friendRequest.ToUserID != userID {
		return errors.NewBusiness(errors.CodePermissionDenied, "")
	}

	// Check request status
	if !friendRequest.IsPending() {
		return errors.NewBusiness(errors.CodeRequestProcessed, "")
	}

	action := model.FriendRequestAction(req.Action)
	if !action.IsValid() {
		return errors.NewBusiness(errors.CodeParamError, "invalid action")
	}

	// Handle request
	if action == model.FriendRequestActionAccept {
		// Use transaction: update request status + create bidirectional friendship
		err = s.db.Transaction(func(tx *gorm.DB) error {
			// Update request status
			requestRepoTx := s.requestRepo.WithTx(tx)
			if err := requestRepoTx.UpdateStatus(ctx, requestID, model.FriendRequestStatusAccepted); err != nil {
				return err
			}

			// Create bidirectional friendship
			now := time.Now()
			friendships := []*model.Friendship{
				{
					UserID:    userID,
					FriendID:  friendRequest.FromUserID,
					Status:    model.FriendshipStatusNormal,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					UserID:    friendRequest.FromUserID,
					FriendID:  userID,
					Status:    model.FriendshipStatusNormal,
					CreatedAt: now,
					UpdatedAt: now,
				},
			}

			friendshipRepoTx := s.friendshipRepo.WithTx(tx)
			return friendshipRepoTx.CreateBatch(ctx, friendships)
		})

		if err != nil {
			l.Errorw("msg", "Failed to accept friend request", "error", err)
			return err
		}

		// Publish friend request accepted event
		s.publishFriendRequestHandledNotification(friendRequest, model.FriendRequestStatusAccepted)
	} else if action == model.FriendRequestActionReject {
		if err := s.requestRepo.UpdateStatus(ctx, requestID, model.FriendRequestStatusRejected); err != nil {
			l.Errorw("msg", "Failed to reject friend request", "error", err)
			return err
		}

		// Publish friend request rejected event
		s.publishFriendRequestHandledNotification(friendRequest, model.FriendRequestStatusRejected)
	}

	return nil
}

// GetFriendRequests retrieves the friend request list
func (s *friendServiceImpl) GetFriendRequests(ctx context.Context, userID string, requestType model.FriendRequestQueryType) (*friendv1.GetFriendRequestsResponse, error) {
	l := s.log.WithContext(ctx)
	var requests []*model.FriendRequest
	var err error

	if requestType == model.FriendRequestQueryTypeSent {
		requests, err = s.requestRepo.GetSentRequests(ctx, userID)
	} else {
		requests, err = s.requestRepo.GetReceivedRequests(ctx, userID)
	}

	if err != nil {
		l.Errorw("msg", "Failed to get friend requests", "error", err)
		return nil, err
	}

	friendRequests := make([]*friendv1.FriendRequest, 0, len(requests))
	for _, r := range requests {
		fr := &friendv1.FriendRequest{
			Id:         fmt.Sprintf("%d", r.ID),
			FromUserId: r.FromUserID,
			ToUserId:   r.ToUserID,
			Message:    r.Message,
			Status:     friendv1.FriendRequestStatus(r.Status),
			CreatedAt:  timestamppb.New(r.CreatedAt),
			UpdatedAt:  timestamppb.New(r.UpdatedAt),
		}

		// Get requester info
		if userInfo, err := s.getUserInfo(ctx, r.FromUserID); err == nil {
			fr.FromNickname = userInfo.Nickname
			fr.FromAvatar = userInfo.Avatar
		}

		friendRequests = append(friendRequests, fr)
	}

	return &friendv1.GetFriendRequestsResponse{
		Requests: friendRequests,
	}, nil
}

// DeleteFriend deletes a friend
func (s *friendServiceImpl) DeleteFriend(ctx context.Context, userID, friendID string) error {
	l := s.log.WithContext(ctx)
	// Check if is a friend
	isFriend, err := s.friendshipRepo.IsFriend(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if !isFriend {
		return errors.NewBusiness(errors.CodeNotFriend, "")
	}

	// Use transaction to delete bidirectional relationship
	err = s.db.Transaction(func(tx *gorm.DB) error {
		friendshipRepoTx := s.friendshipRepo.WithTx(tx)
		return friendshipRepoTx.DeleteBidirectional(ctx, userID, friendID)
	})

	if err != nil {
		l.Errorw("msg", "Failed to delete friend", "error", err)
		return err
	}

	// Publish friend deleted event
	s.publishFriendDeletedNotification(userID, friendID)

	return nil
}

// UpdateRemark updates friend remark
func (s *friendServiceImpl) UpdateRemark(ctx context.Context, userID, friendID, remark string) error {
	l := s.log.WithContext(ctx)
	// Check if is a friend
	isFriend, err := s.friendshipRepo.IsFriend(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if !isFriend {
		return errors.NewBusiness(errors.CodeNotFriend, "")
	}

	if err := s.friendshipRepo.UpdateRemark(ctx, userID, friendID, remark); err != nil {
		l.Errorw("msg", "Failed to update remark", "error", err)
		return err
	}

	// Publish remark updated event (multi-device sync)
	s.publishRemarkUpdatedNotification(userID, friendID, remark)

	return nil
}

// AddToBlacklist adds user to blacklist
func (s *friendServiceImpl) AddToBlacklist(ctx context.Context, userID, blockedUserID string) error {
	l := s.log.WithContext(ctx)
	var removedFriend bool
	err := s.db.Transaction(func(tx *gorm.DB) error {
		blacklistRepoTx := s.blacklistRepo.WithTx(tx)
		friendshipRepoTx := s.friendshipRepo.WithTx(tx)

		// Check if already in blacklist
		existing, err := blacklistRepoTx.GetByUserAndBlocked(ctx, userID, blockedUserID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if existing != nil {
			return errors.NewBusiness(errors.CodeAlreadyInBlacklist, "")
		}

		// Create blacklist record
		blacklist := &model.Blacklist{
			UserID:        userID,
			BlockedUserID: blockedUserID,
		}
		if err := blacklistRepoTx.Create(ctx, blacklist); err != nil {
			l.Errorw("msg", "Failed to add to blacklist", "error", err)
			return err
		}

		// If both are friends, block will automatically remove bidirectional friendship
		isFriend, err := friendshipRepoTx.IsFriend(ctx, userID, blockedUserID)
		if err != nil {
			return err
		}
		if isFriend {
			if err := friendshipRepoTx.DeleteBidirectional(ctx, userID, blockedUserID); err != nil {
				return err
			}
			removedFriend = true
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Publish blacklist changed event
	s.publishBlacklistChangedNotification(userID, blockedUserID, "add")

	// Trigger update on the removed friend's side
	if removedFriend {
		s.publishFriendDeletedNotification(userID, blockedUserID)
	}

	return nil
}

// RemoveFromBlacklist removes user from blacklist
func (s *friendServiceImpl) RemoveFromBlacklist(ctx context.Context, userID, blockedUserID string) error {
	l := s.log.WithContext(ctx)
	// Check if in blacklist
	existing, err := s.blacklistRepo.GetByUserAndBlocked(ctx, userID, blockedUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeNotInBlacklist, "")
		}
		return err
	}
	if existing == nil {
		return errors.NewBusiness(errors.CodeNotInBlacklist, "")
	}

	if err := s.blacklistRepo.Delete(ctx, userID, blockedUserID); err != nil {
		l.Errorw("msg", "Failed to remove from blacklist", "error", err)
		return err
	}

	// Publish blacklist changed event
	s.publishBlacklistChangedNotification(userID, blockedUserID, "remove")

	return nil
}

// GetBlacklist retrieves the blacklist
func (s *friendServiceImpl) GetBlacklist(ctx context.Context, userID string) (*friendv1.GetBlacklistResponse, error) {
	l := s.log.WithContext(ctx)
	blacklist, err := s.blacklistRepo.GetBlacklist(ctx, userID)
	if err != nil {
		l.Errorw("msg", "Failed to get blacklist", "error", err)
		return nil, err
	}

	items := make([]*friendv1.BlockedUser, 0, len(blacklist))
	for _, b := range blacklist {
		item := &friendv1.BlockedUser{
			UserId: b.BlockedUserID,
		}

		// Get blocked user info
		if userInfo, err := s.getUserInfo(ctx, b.BlockedUserID); err == nil {
			item.Nickname = userInfo.Nickname
			item.Avatar = userInfo.Avatar
		}

		items = append(items, item)
	}

	return &friendv1.GetBlacklistResponse{
		BlockedUsers: items,
	}, nil
}

// IsFriend checks if users are friends
func (s *friendServiceImpl) IsFriend(ctx context.Context, userID, friendID string) (bool, error) {
	return s.friendshipRepo.IsFriend(ctx, userID, friendID)
}

// IsBlocked checks if user is blocked
func (s *friendServiceImpl) IsBlocked(ctx context.Context, userID, targetUserID string) (bool, error) {
	return s.blacklistRepo.IsBlocked(ctx, userID, targetUserID)
}

// BatchCheckFriend batch checks friend relationships
func (s *friendServiceImpl) BatchCheckFriend(ctx context.Context, userID string, friendIDs []string) (map[string]bool, error) {
	l := s.log.WithContext(ctx)
	results := make(map[string]bool, len(friendIDs))

	for _, friendID := range friendIDs {
		isFriend, err := s.friendshipRepo.IsFriend(ctx, userID, friendID)
		if err != nil {
			l.Errorw("msg", "Failed to check friend", "friendID", friendID, "error", err)
			results[friendID] = false
			continue
		}
		results[friendID] = isFriend
	}

	return results, nil
}

// getUserInfo gets user info (internal helper method)
func (s *friendServiceImpl) getUserInfo(ctx context.Context, userID string) (*userv1.UserInfo, error) {
	resp, err := s.userClient.GetUserInfo(ctx, &userv1.GetUserInfoRequest{
		TargetUserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	userInfo := &userv1.UserInfo{
		UserId:   userID,
		Nickname: resp.Nickname,
		Avatar:   resp.Avatar,
	}

	return userInfo, nil
}

// publishFriendRequestNotification publishes friend request event
func (s *friendServiceImpl) publishFriendRequestNotification(req *model.FriendRequest) {
	payload := map[string]interface{}{
		"request_id":   req.ID,
		"from_user_id": req.FromUserID,
		"message":      req.Message,
		"source":       req.Source,
		"created_at":   req.CreatedAt.Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeFriendRequest,
		req.FromUserID,
		broker.PriorityHigh,
	).WithPayload(payload)

	if err := s.broker.PublishToUser(req.ToUserID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish friend request event", "toUserId", req.ToUserID, "error", err)
	}
}

// publishFriendRequestHandledNotification publishes friend request handled event
func (s *friendServiceImpl) publishFriendRequestHandledNotification(req *model.FriendRequest, status model.FriendRequestStatus) {
	payload := map[string]interface{}{
		"request_id": req.ID,
		"to_user_id": req.ToUserID,
		"status":     status,
		"handled_at": time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeFriendRequestHandled,
		req.ToUserID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToUser(req.FromUserID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish friend request handled event", "fromUserId", req.FromUserID, "error", err)
	}
}

// publishFriendDeletedNotification publishes friend deleted event
func (s *friendServiceImpl) publishFriendDeletedNotification(userID, friendID string) {
	payload := map[string]interface{}{
		"friend_user_id": userID,
		"deleted_at":     time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeFriendDeleted,
		userID,
		broker.PriorityNormal,
	).WithPayload(payload)

	// Notify the deleted friend
	if err := s.broker.PublishToUser(friendID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish friend deleted event", "friendId", friendID, "error", err)
	}
}

// publishRemarkUpdatedNotification publishes remark updated event (multi-device sync)
func (s *friendServiceImpl) publishRemarkUpdatedNotification(userID, friendID, remark string) {
	payload := map[string]interface{}{
		"friend_user_id": friendID,
		"remark":         remark,
		"updated_at":     time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeFriendRemarkUpdated,
		userID,
		broker.PriorityLow,
	).WithPayload(payload)

	// Push to user's other devices (multi-device sync)
	if err := s.broker.PublishToUser(userID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish remark updated event", "userId", userID, "error", err)
	}
}

// publishBlacklistChangedNotification publishes blacklist changed event
func (s *friendServiceImpl) publishBlacklistChangedNotification(userID, targetUserID, action string) {
	payload := map[string]interface{}{
		"target_user_id": targetUserID,
		"action":         action,
		"changed_at":     time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeBlacklistChanged,
		userID,
		broker.PriorityNormal,
	).WithPayload(payload)

	// Push to user's other devices (multi-device sync)
	if err := s.broker.PublishToUser(userID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish blacklist changed event", "userId", userID, "error", err)
	}
}

// createFriendConversation creates friend conversation
func (s *friendServiceImpl) createFriendConversation(ctx context.Context, userID, friendID string) {
	l := s.log.WithContext(ctx)
	if s.conversationClient == nil {
		return
	}

	_, err := s.conversationClient.CreateOrUpdateConversation(ctx, &conversationv1.CreateOrUpdateConversationRequest{
		UserId:           userID,
		ConversationType: "single",
		TargetId:         friendID,
	})
	if err != nil {
		l.Errorw("msg", "Failed to create friend conversation", "userId", userID, "friendId", friendID, "error", err)
	}
}

// publishFriendAddedNotification publishes friend added event (auto accepted)
func (s *friendServiceImpl) publishFriendAddedNotification(userID, addedByUserID string) {
	payload := map[string]interface{}{
		"added_by_user_id": addedByUserID,
		"created_at":       time.Now().Unix(),
	}

	notif := broker.NewNotification(
		broker.TypeFriendAdded,
		userID,
		broker.PriorityNormal,
	).WithPayload(payload)

	if err := s.broker.PublishToUser(userID, notif); err != nil {
		s.log.Errorw("msg", "Failed to publish friend added event", "userId", userID, "addedByUserId", addedByUserID, "error", err)
	}
}
