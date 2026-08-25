package service

import (
	"context"
	"time"

	"flamingo/app/friend/internal/model"

	"gorm.io/gorm"
)

// BlacklistRepository is the blacklist repository interface
type BlacklistRepository interface {
	Create(ctx context.Context, blacklist *model.Blacklist) error
	GetByUserAndBlocked(ctx context.Context, userID, blockedUserID string) (*model.Blacklist, error)
	GetBlacklist(ctx context.Context, userID string) ([]*model.Blacklist, error)
	Delete(ctx context.Context, userID, blockedUserID string) error
	IsBlocked(ctx context.Context, userID, targetUserID string) (bool, error)
	WithTx(tx *gorm.DB) BlacklistRepository
}

// FriendRequestRepository is the friend request repository interface
type FriendRequestRepository interface {
	Create(ctx context.Context, request *model.FriendRequest) error
	GetByID(ctx context.Context, id int64) (*model.FriendRequest, error)
	GetByUserIDs(ctx context.Context, fromUserID, toUserID string) (*model.FriendRequest, error)
	GetPendingRequest(ctx context.Context, fromUserID, toUserID string) (*model.FriendRequest, error)
	GetReceivedRequests(ctx context.Context, userID string) ([]*model.FriendRequest, error)
	GetSentRequests(ctx context.Context, userID string) ([]*model.FriendRequest, error)
	UpdateStatus(ctx context.Context, id int64, status model.FriendRequestStatus) error
	Update(ctx context.Context, request *model.FriendRequest) error
	WithTx(tx *gorm.DB) FriendRequestRepository
}

// FriendshipRepository is the friendship repository interface
type FriendshipRepository interface {
	Create(ctx context.Context, friendship *model.Friendship) error
	CreateBatch(ctx context.Context, friendships []*model.Friendship) error
	GetByUserAndFriend(ctx context.Context, userID, friendID string) (*model.Friendship, error)
	GetFriendList(ctx context.Context, userID string) ([]*model.Friendship, error)
	GetFriendListByUpdateTime(ctx context.Context, userID string, lastUpdateTime time.Time) ([]*model.Friendship, error)
	Update(ctx context.Context, friendship *model.Friendship) error
	UpdateRemark(ctx context.Context, userID, friendID, remark string) error
	Delete(ctx context.Context, userID, friendID string) error
	DeleteBidirectional(ctx context.Context, userID, friendID string) error
	IsFriend(ctx context.Context, userID, friendID string) (bool, error)
	WithTx(tx *gorm.DB) FriendshipRepository
}
