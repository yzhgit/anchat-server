package repository

import (
	"context"
	"time"

	"flamingo/app/group/internal/model"
	"flamingo/app/group/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// groupJoinRequestRepositoryImpl is the join request repository implementation
type groupJoinRequestRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewGroupJoinRequestRepository creates a new join request repository
func NewGroupJoinRequestRepository(db *gorm.DB, logger log.Logger) service.GroupJoinRequestRepository {
	return &groupJoinRequestRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates a join request
func (r *groupJoinRequestRepositoryImpl) Create(ctx context.Context, request *model.GroupJoinRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

// GetByID gets join request by ID
func (r *groupJoinRequestRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.GroupJoinRequest, error) {
	var request model.GroupJoinRequest
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// UpdateStatus updates request status
func (r *groupJoinRequestRepositoryImpl) UpdateStatus(ctx context.Context, id int64, status model.JoinRequestStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.GroupJoinRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// GetPendingRequestsByGroup gets pending requests for group
func (r *groupJoinRequestRepositoryImpl) GetPendingRequestsByGroup(ctx context.Context, groupID string) ([]*model.GroupJoinRequest, error) {
	var requests []*model.GroupJoinRequest
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND status = ?", groupID, model.JoinRequestStatusPending).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// GetPendingRequestsByUser gets pending requests for user
func (r *groupJoinRequestRepositoryImpl) GetPendingRequestsByUser(ctx context.Context, userID string) ([]*model.GroupJoinRequest, error) {
	var requests []*model.GroupJoinRequest
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.JoinRequestStatusPending).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// GetRequestsByGroup gets request list for group (filterable by status)
func (r *groupJoinRequestRepositoryImpl) GetRequestsByGroup(ctx context.Context, groupID string, status *model.JoinRequestStatus) ([]*model.GroupJoinRequest, error) {
	var requests []*model.GroupJoinRequest
	query := r.db.WithContext(ctx).Where("group_id = ?", groupID)

	if status != nil {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// GetExistingRequest gets existing pending request (prevents duplicate requests)
func (r *groupJoinRequestRepositoryImpl) GetExistingRequest(ctx context.Context, groupID, userID string) (*model.GroupJoinRequest, error) {
	var request model.GroupJoinRequest
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND status = ?", groupID, userID, model.JoinRequestStatusPending).
		First(&request).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &request, nil
}

// WithTx uses transaction
func (r *groupJoinRequestRepositoryImpl) WithTx(tx *gorm.DB) service.GroupJoinRequestRepository {
	return &groupJoinRequestRepositoryImpl{db: tx}
}
