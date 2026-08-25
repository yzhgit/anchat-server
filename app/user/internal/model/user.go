package model

import (
	"time"

	"gorm.io/gorm"
)

// User user model
type User struct {
	ID           string         `gorm:"column:id;primaryKey" json:"id"`
	Phone        *string        `gorm:"column:phone;unique" json:"phone"`
	Email        *string        `gorm:"column:email;unique" json:"email"`
	PasswordHash string         `gorm:"column:password_hash;not null" json:"-"`
	Status       int            `gorm:"column:status;not null;default:1" json:"status"` // 1-normal, 2-disabled
	CreatedAt    time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName returns table name
func (User) TableName() string {
	return "users"
}

// UserStatus user status constants
const (
	UserStatusNormal   = 1 // normal
	UserStatusDisabled = 2 // disabled
)

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusNormal
}
