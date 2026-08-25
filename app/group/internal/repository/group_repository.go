package repository

import (
	"context"
	"time"

	"flamingo/app/group/internal/model"
	"flamingo/app/group/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// groupRepositoryImpl is the group repository implementation
type groupRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(db *gorm.DB, logger log.Logger) service.GroupRepository {
	return &groupRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates a group
func (r *groupRepositoryImpl) Create(ctx context.Context, group *model.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetByGroupID gets group by group ID
func (r *groupRepositoryImpl) GetByGroupID(ctx context.Context, groupID string) (*model.Group, error) {
	var group model.Group
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND status = ?", groupID, model.GroupStatusNormal).
		First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// Update updates a group
func (r *groupRepositoryImpl) Update(ctx context.Context, group *model.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// UpdateFields updates specified fields
func (r *groupRepositoryImpl) UpdateFields(ctx context.Context, groupID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&model.Group{}).
		Where("group_id = ? AND status = ?", groupID, model.GroupStatusNormal).
		Updates(updates).Error
}

// Delete deletes a group (soft delete, updates status to dissolved)
func (r *groupRepositoryImpl) Delete(ctx context.Context, groupID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Group{}).
		Where("group_id = ?", groupID).
		Updates(map[string]interface{}{
			"status":     model.GroupStatusDissolved,
			"updated_at": time.Now(),
		}).Error
}

// UpdateMemberCount updates member count (atomic operation)
func (r *groupRepositoryImpl) UpdateMemberCount(ctx context.Context, groupID string, delta int32) error {
	return r.db.WithContext(ctx).
		Model(&model.Group{}).
		Where("group_id = ?", groupID).
		Updates(map[string]interface{}{
			"member_count": gorm.Expr("member_count + ?", delta),
			"updated_at":   time.Now(),
		}).Error
}

// GetGroupsByOwner gets groups created by user
func (r *groupRepositoryImpl) GetGroupsByOwner(ctx context.Context, ownerID string) ([]*model.Group, error) {
	var groups []*model.Group
	err := r.db.WithContext(ctx).
		Where("owner_id = ? AND status = ?", ownerID, model.GroupStatusNormal).
		Order("created_at DESC").
		Find(&groups).Error
	return groups, err
}

// Search searches groups by name
func (r *groupRepositoryImpl) Search(ctx context.Context, keyword string, limit int) ([]*model.Group, error) {
	var groups []*model.Group
	err := r.db.WithContext(ctx).
		Where("name LIKE ? AND status = ?", "%"+keyword+"%", model.GroupStatusNormal).
		Order("member_count DESC").
		Limit(limit).
		Find(&groups).Error
	return groups, err
}

// WithTx uses transaction
func (r *groupRepositoryImpl) WithTx(tx *gorm.DB) service.GroupRepository {
	return &groupRepositoryImpl{db: tx}
}
