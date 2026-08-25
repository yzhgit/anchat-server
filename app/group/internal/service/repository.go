package service

import (
	"context"
	"time"

	"flamingo/app/group/internal/model"

	"gorm.io/gorm"
)

// GroupJoinRequestRepository defines the join request repository interface
type GroupJoinRequestRepository interface {
	Create(ctx context.Context, request *model.GroupJoinRequest) error
	GetByID(ctx context.Context, id int64) (*model.GroupJoinRequest, error)
	UpdateStatus(ctx context.Context, id int64, status model.JoinRequestStatus) error
	GetPendingRequestsByGroup(ctx context.Context, groupID string) ([]*model.GroupJoinRequest, error)
	GetPendingRequestsByUser(ctx context.Context, userID string) ([]*model.GroupJoinRequest, error)
	GetRequestsByGroup(ctx context.Context, groupID string, status *model.JoinRequestStatus) ([]*model.GroupJoinRequest, error)
	GetExistingRequest(ctx context.Context, groupID, userID string) (*model.GroupJoinRequest, error)
	WithTx(tx *gorm.DB) GroupJoinRequestRepository
}

// GroupMemberRepository defines the group member repository interface
type GroupMemberRepository interface {
	AddMember(ctx context.Context, member *model.GroupMember) error
	AddMembers(ctx context.Context, members []*model.GroupMember) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	UpdateRole(ctx context.Context, groupID, userID string, role model.GroupRole) error
	UpdateNickname(ctx context.Context, groupID, userID, nickname string) error
	UpdateRemark(ctx context.Context, groupID, userID, remark string) error
	UpdateMutedUntil(ctx context.Context, groupID, userID string, mutedUntil *time.Time) error
	GetMember(ctx context.Context, groupID, userID string) (*model.GroupMember, error)
	GetMembers(ctx context.Context, groupID string, page, pageSize int) ([]*model.GroupMember, int64, error)
	GetMembersByRole(ctx context.Context, groupID string, role model.GroupRole) ([]*model.GroupMember, error)
	GetMemberCount(ctx context.Context, groupID string) (int64, error)
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
	GetUserGroups(ctx context.Context, userID string) ([]*model.GroupMember, error)
	GetUserGroupsByUpdateTime(ctx context.Context, userID string, lastUpdateTime time.Time) ([]*model.GroupMember, error)
	WithTx(tx *gorm.DB) GroupMemberRepository
}

type GroupPinnedMessageRepository interface {
	Upsert(ctx context.Context, pinned *model.GroupPinnedMessage) error
	Delete(ctx context.Context, groupID, messageID string) (bool, error)
	ListByGroup(ctx context.Context, groupID string) ([]*model.GroupPinnedMessage, error)
	CountByGroup(ctx context.Context, groupID string) (int64, error)
	Exists(ctx context.Context, groupID, messageID string) (bool, error)
	WithTx(tx *gorm.DB) GroupPinnedMessageRepository
}

// GroupQRCodeRepository defines the group QR code repository interface
type GroupQRCodeRepository interface {
	Create(ctx context.Context, qrcode *model.GroupQRCode) error
	GetActiveByGroupID(ctx context.Context, groupID string) (*model.GroupQRCode, error)
	GetByToken(ctx context.Context, token string) (*model.GroupQRCode, error)
	InvalidateByGroupID(ctx context.Context, groupID string) error
	UpdateExpireAt(ctx context.Context, token string, expireAt time.Time) error
	WithTx(tx *gorm.DB) GroupQRCodeRepository
}

// GroupRepository defines the group repository interface
type GroupRepository interface {
	Create(ctx context.Context, group *model.Group) error
	GetByGroupID(ctx context.Context, groupID string) (*model.Group, error)
	Update(ctx context.Context, group *model.Group) error
	UpdateFields(ctx context.Context, groupID string, updates map[string]interface{}) error
	Delete(ctx context.Context, groupID string) error
	UpdateMemberCount(ctx context.Context, groupID string, delta int32) error
	GetGroupsByOwner(ctx context.Context, ownerID string) ([]*model.Group, error)
	Search(ctx context.Context, keyword string, limit int) ([]*model.Group, error)
	WithTx(tx *gorm.DB) GroupRepository
}

// GroupSettingRepository defines the group settings repository interface
type GroupSettingRepository interface {
	Create(ctx context.Context, setting *model.GroupSetting) error
	GetSettings(ctx context.Context, groupID string) (*model.GroupSetting, error)
	UpdateSettings(ctx context.Context, groupID string, updates map[string]interface{}) error
	Delete(ctx context.Context, groupID string) error
	WithTx(tx *gorm.DB) GroupSettingRepository
}
