package model

import "time"

// Group represents a group model
type Group struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID      string    `gorm:"column:group_id;not null;uniqueIndex" json:"groupId"`
	Name         string    `gorm:"column:name;not null;size:100" json:"name"`
	Avatar       string    `gorm:"column:avatar;size:255" json:"avatar"`
	Announcement string    `gorm:"column:announcement;type:text" json:"announcement"`
	Description  string    `gorm:"column:description;type:text" json:"description"`
	OwnerID      string    `gorm:"column:owner_id;not null" json:"ownerId"`
	MemberCount  int32     `gorm:"column:member_count;default:0" json:"memberCount"`
	MaxMembers   int32     `gorm:"column:max_members;default:500" json:"maxMembers"`
	IsMuted      bool      `gorm:"column:is_muted;default:false" json:"isMuted"`
	Status       int16     `gorm:"column:status;default:1" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName returns the table name
func (Group) TableName() string {
	return "groups"
}

// GroupStatus represents group status
const (
	GroupStatusDissolved = 0 // dissolved
	GroupStatusNormal    = 1 // normal
)

// IsActive checks if group is active
func (g *Group) IsActive() bool {
	return g.Status == GroupStatusNormal
}

// IsFull checks if group is full
func (g *Group) IsFull() bool {
	return g.MemberCount >= g.MaxMembers
}
