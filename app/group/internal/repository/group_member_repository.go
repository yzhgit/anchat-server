package repository

import (
	"context"
	"time"

	"flamingo/app/group/internal/model"
	"flamingo/app/group/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// groupMemberRepositoryImpl is the group member repository implementation
type groupMemberRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewGroupMemberRepository creates a new group member repository
func NewGroupMemberRepository(db *gorm.DB, logger log.Logger) service.GroupMemberRepository {
	return &groupMemberRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// AddMember adds a member
func (r *groupMemberRepositoryImpl) AddMember(ctx context.Context, member *model.GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// AddMembers adds multiple members
func (r *groupMemberRepositoryImpl) AddMembers(ctx context.Context, members []*model.GroupMember) error {
	return r.db.WithContext(ctx).Create(members).Error
}

// RemoveMember removes a member
func (r *groupMemberRepositoryImpl) RemoveMember(ctx context.Context, groupID, userID string) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&model.GroupMember{}).Error
}

// UpdateRole updates member role
func (r *groupMemberRepositoryImpl) UpdateRole(ctx context.Context, groupID, userID string, role model.GroupRole) error {
	return r.db.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Updates(map[string]interface{}{
			"role":       role,
			"updated_at": time.Now(),
		}).Error
}

// UpdateNickname updates group nickname
func (r *groupMemberRepositoryImpl) UpdateNickname(ctx context.Context, groupID, userID, nickname string) error {
	return r.db.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Updates(map[string]interface{}{
			"group_nickname": nickname,
			"updated_at":     time.Now(),
		}).Error
}

// UpdateRemark updates group remark (only visible to self)
func (r *groupMemberRepositoryImpl) UpdateRemark(ctx context.Context, groupID, userID, remark string) error {
	var remarkValue interface{} = remark
	if remark == "" {
		remarkValue = nil
	}

	return r.db.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Updates(map[string]interface{}{
			"group_remark": remarkValue,
			"updated_at":   time.Now(),
		}).Error
}

// UpdateMuted updates muted status
func (r *groupMemberRepositoryImpl) UpdateMutedUntil(ctx context.Context, groupID, userID string, mutedUntil *time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Updates(map[string]interface{}{
			"muted_until": mutedUntil,
			"updated_at":  time.Now(),
		}).Error
}

// GetMember gets member info
func (r *groupMemberRepositoryImpl) GetMember(ctx context.Context, groupID, userID string) (*model.GroupMember, error) {
	var member model.GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetMembers gets group member list (with pagination)
func (r *groupMemberRepositoryImpl) GetMembers(ctx context.Context, groupID string, page, pageSize int) ([]*model.GroupMember, int64, error) {
	var members []*model.GroupMember
	var total int64

	query := r.db.WithContext(ctx).Model(&model.GroupMember{}).Where("group_id = ?", groupID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginated query
	offset := (page - 1) * pageSize
	err := query.
		Order("joined_at ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&members).Error

	return members, total, err
}

// GetMembersByRole gets members by role
func (r *groupMemberRepositoryImpl) GetMembersByRole(ctx context.Context, groupID string, role model.GroupRole) ([]*model.GroupMember, error) {
	var members []*model.GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND role = ?", groupID, role).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

// GetMemberCount gets member count
func (r *groupMemberRepositoryImpl) GetMemberCount(ctx context.Context, groupID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ?", groupID).
		Count(&count).Error
	return count, err
}

// IsMember checks if user is a member
func (r *groupMemberRepositoryImpl) IsMember(ctx context.Context, groupID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error
	return count > 0, err
}

// GetUserGroups gets list of groups user joined
func (r *groupMemberRepositoryImpl) GetUserGroups(ctx context.Context, userID string) ([]*model.GroupMember, error) {
	var members []*model.GroupMember
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("joined_at DESC").
		Find(&members).Error
	return members, err
}

// GetUserGroupsByUpdateTime gets user groups by update time (incremental sync)
func (r *groupMemberRepositoryImpl) GetUserGroupsByUpdateTime(ctx context.Context, userID string, lastUpdateTime time.Time) ([]*model.GroupMember, error) {
	var members []*model.GroupMember
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND updated_at > ?", userID, lastUpdateTime).
		Order("updated_at DESC").
		Find(&members).Error
	return members, err
}

// WithTx uses transaction
func (r *groupMemberRepositoryImpl) WithTx(tx *gorm.DB) service.GroupMemberRepository {
	return &groupMemberRepositoryImpl{db: tx}
}
