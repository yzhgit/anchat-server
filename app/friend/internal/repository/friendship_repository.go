package repository

import (
	"context"
	"time"

	"flamingo/app/friend/internal/model"
	"flamingo/app/friend/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// friendshipRepositoryImpl is the friendship repository implementation
type friendshipRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewFriendshipRepository creates a new friendship repository
func NewFriendshipRepository(db *gorm.DB, logger log.Logger) service.FriendshipRepository {
	return &friendshipRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates a friendship
func (r *friendshipRepositoryImpl) Create(ctx context.Context, friendship *model.Friendship) error {
	return r.db.WithContext(ctx).Create(friendship).Error
}

// CreateBatch batch creates friendships (for bidirectional creation)
func (r *friendshipRepositoryImpl) CreateBatch(ctx context.Context, friendships []*model.Friendship) error {
	return r.db.WithContext(ctx).Create(friendships).Error
}

// GetByUserAndFriend retrieves a friendship by user ID and friend ID
func (r *friendshipRepositoryImpl) GetByUserAndFriend(ctx context.Context, userID, friendID string) (*model.Friendship, error) {
	var friendship model.Friendship
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND friend_id = ? AND status = ?", userID, friendID, model.FriendshipStatusNormal).
		First(&friendship).Error
	if err != nil {
		return nil, err
	}
	return &friendship, nil
}

// GetFriendList retrieves the friend list
func (r *friendshipRepositoryImpl) GetFriendList(ctx context.Context, userID string) ([]*model.Friendship, error) {
	var friendships []*model.Friendship
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.FriendshipStatusNormal).
		Order("updated_at DESC").
		Find(&friendships).Error
	return friendships, err
}

// GetFriendListByUpdateTime retrieves the friend list by update time (incremental sync)
func (r *friendshipRepositoryImpl) GetFriendListByUpdateTime(ctx context.Context, userID string, lastUpdateTime time.Time) ([]*model.Friendship, error) {
	var friendships []*model.Friendship
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND updated_at > ?", userID, lastUpdateTime).
		Order("updated_at DESC").
		Find(&friendships).Error
	return friendships, err
}

// Update updates a friendship
func (r *friendshipRepositoryImpl) Update(ctx context.Context, friendship *model.Friendship) error {
	return r.db.WithContext(ctx).Save(friendship).Error
}

// UpdateRemark updates remark
func (r *friendshipRepositoryImpl) UpdateRemark(ctx context.Context, userID, friendID, remark string) error {
	return r.db.WithContext(ctx).
		Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ? AND status = ?", userID, friendID, model.FriendshipStatusNormal).
		Update("remark", remark).Error
}

// Delete deletes a friendship (soft delete, updates status)
func (r *friendshipRepositoryImpl) Delete(ctx context.Context, userID, friendID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Update("status", model.FriendshipStatusDeleted).Error
}

// DeleteBidirectional deletes bidirectional friendship (must be called in transaction)
func (r *friendshipRepositoryImpl) DeleteBidirectional(ctx context.Context, userID, friendID string) error {
	// Delete A->B
	if err := r.Delete(ctx, userID, friendID); err != nil {
		return err
	}
	// Delete B->A
	return r.Delete(ctx, friendID, userID)
}

// IsFriend checks if users are friends
func (r *friendshipRepositoryImpl) IsFriend(ctx context.Context, userID, friendID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ? AND status = ?", userID, friendID, model.FriendshipStatusNormal).
		Count(&count).Error
	return count > 0, err
}

// WithTx uses transaction
func (r *friendshipRepositoryImpl) WithTx(tx *gorm.DB) service.FriendshipRepository {
	return &friendshipRepositoryImpl{db: tx}
}
