package repository

import (
	"context"

	"flamingo/app/friend/internal/model"
	"flamingo/app/friend/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// friendRequestRepositoryImpl is the friend request repository implementation
type friendRequestRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewFriendRequestRepository creates a new friend request repository
func NewFriendRequestRepository(db *gorm.DB, logger log.Logger) service.FriendRequestRepository {
	return &friendRequestRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates a friend request
func (r *friendRequestRepositoryImpl) Create(ctx context.Context, request *model.FriendRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

// GetByID retrieves a friend request by ID
func (r *friendRequestRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.FriendRequest, error) {
	var request model.FriendRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// GetByUserIDs retrieves the latest friend request by user IDs
func (r *friendRequestRepositoryImpl) GetByUserIDs(ctx context.Context, fromUserID, toUserID string) (*model.FriendRequest, error) {
	var request model.FriendRequest
	err := r.db.WithContext(ctx).
		Where("from_user_id = ? AND to_user_id = ?", fromUserID, toUserID).
		Order("created_at DESC").
		First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// GetPendingRequest retrieves pending friend request
func (r *friendRequestRepositoryImpl) GetPendingRequest(ctx context.Context, fromUserID, toUserID string) (*model.FriendRequest, error) {
	var request model.FriendRequest
	err := r.db.WithContext(ctx).
		Where("from_user_id = ? AND to_user_id = ? AND status = ?", fromUserID, toUserID, model.FriendRequestStatusPending).
		Order("created_at DESC").
		First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// GetReceivedRequests retrieves received friend requests
func (r *friendRequestRepositoryImpl) GetReceivedRequests(ctx context.Context, userID string) ([]*model.FriendRequest, error) {
	var requests []*model.FriendRequest
	err := r.db.WithContext(ctx).
		Where("to_user_id = ?", userID).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// GetSentRequests retrieves sent friend requests
func (r *friendRequestRepositoryImpl) GetSentRequests(ctx context.Context, userID string) ([]*model.FriendRequest, error) {
	var requests []*model.FriendRequest
	err := r.db.WithContext(ctx).
		Where("from_user_id = ?", userID).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// UpdateStatus updates request status
func (r *friendRequestRepositoryImpl) UpdateStatus(ctx context.Context, id int64, status model.FriendRequestStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.FriendRequest{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Update updates a friend request
func (r *friendRequestRepositoryImpl) Update(ctx context.Context, request *model.FriendRequest) error {
	return r.db.WithContext(ctx).Save(request).Error
}

// WithTx uses transaction
func (r *friendRequestRepositoryImpl) WithTx(tx *gorm.DB) service.FriendRequestRepository {
	return &friendRequestRepositoryImpl{db: tx}
}
