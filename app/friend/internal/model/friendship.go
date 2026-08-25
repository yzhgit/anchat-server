package model

import "time"

// Friendship is the friendship model
type Friendship struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"column:user_id;not null;uniqueIndex:uk_user_friend" json:"userId"`
	FriendID  string    `gorm:"column:friend_id;not null;uniqueIndex:uk_user_friend" json:"friendId"`
	Remark    string    `gorm:"column:remark;size:50" json:"remark"`
	Status    int16     `gorm:"column:status;default:1" json:"status"` // 0-deleted 1-normal
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName is the table name
func (Friendship) TableName() string {
	return "friendships"
}

// FriendshipStatus is the friendship status constant
const (
	FriendshipStatusDeleted = 0 // deleted
	FriendshipStatusNormal  = 1 // normal
)

// IsActive returns true if the friendship is active
func (f *Friendship) IsActive() bool {
	return f.Status == FriendshipStatusNormal
}
