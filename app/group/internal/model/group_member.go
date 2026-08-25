package model

import "time"

// GroupRole represents group member role.
type GroupRole int16

// GroupRole values.
const (
	GroupRoleUnknown GroupRole = 0
	GroupRoleOwner   GroupRole = 1
	GroupRoleAdmin   GroupRole = 2
	GroupRoleMember  GroupRole = 3
)

// GroupMember represents a group member model
type GroupMember struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID       string     `gorm:"column:group_id;not null;uniqueIndex:uk_group_user" json:"groupId"`
	UserID        string     `gorm:"column:user_id;not null;uniqueIndex:uk_group_user" json:"userId"`
	GroupNickname string     `gorm:"column:group_nickname;size:50" json:"groupNickname"`
	GroupRemark   string     `gorm:"column:group_remark;size:20" json:"groupRemark"` // Remark for this group, only visible to self
	Role          GroupRole  `gorm:"column:role;type:smallint;not null;default:3" json:"role"`
	MutedUntil    *time.Time `gorm:"column:muted_until" json:"mutedUntil"`
	JoinedAt      time.Time  `gorm:"column:joined_at;not null;default:CURRENT_TIMESTAMP" json:"joinedAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName returns the table name
func (GroupMember) TableName() string {
	return "group_members"
}

var PermanentMutedUntil = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

// IsOwner checks if member is the owner
func (gm *GroupMember) IsOwner() bool {
	return gm.Role == GroupRoleOwner
}

// IsAdmin checks if member is admin (including owner)
func (gm *GroupMember) IsAdmin() bool {
	return gm.Role == GroupRoleOwner || gm.Role == GroupRoleAdmin
}

// CanManageGroup checks if member can manage the group (owner and admin)
func (gm *GroupMember) CanManageGroup() bool {
	return gm.IsAdmin()
}

// CanRemoveMember checks if member can remove target role
func (gm *GroupMember) CanRemoveMember(targetRole GroupRole) bool {
	if gm.IsOwner() {
		// Owner can remove anyone (except self)
		return true
	}
	if gm.Role == GroupRoleAdmin {
		// Admin can only remove regular members
		return targetRole == GroupRoleMember
	}
	return false
}

// CanMuteMember checks if member can mute target role
func (gm *GroupMember) CanMuteMember(targetRole GroupRole) bool {
	return gm.CanRemoveMember(targetRole)
}

// IsMutedNow checks if currently muted
func (gm *GroupMember) IsMutedNow() bool {
	return gm.MutedUntil != nil && gm.MutedUntil.After(time.Now())
}

// IsPermanentlyMuted checks if permanently muted
func (gm *GroupMember) IsPermanentlyMuted() bool {
	if gm.MutedUntil == nil {
		return false
	}
	return gm.MutedUntil.Equal(PermanentMutedUntil) || gm.MutedUntil.After(PermanentMutedUntil)
}
