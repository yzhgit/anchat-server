package model

import (
	"time"
)

// UserProfile user profile model
type UserProfile struct {
	ID              int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID          string     `gorm:"column:user_id;unique;not null" json:"userId"`
	Nickname        string     `gorm:"column:nickname;not null" json:"nickname"`
	Avatar          string     `gorm:"column:avatar" json:"avatar"`
	Signature       string     `gorm:"column:signature" json:"signature"`
	Gender          int        `gorm:"column:gender;not null;default:0" json:"gender"` // 0-unknown 1-male 2-female
	Birthday        *time.Time `gorm:"column:birthday" json:"birthday"`
	Region          string     `gorm:"column:region" json:"region"`
	QRCodeURL       string     `gorm:"column:qrcode_url" json:"qrcodeUrl"`
	QRCodeUpdatedAt *time.Time `gorm:"column:qrcode_updated_at" json:"qrcodeUpdatedAt"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName returns table name
func (UserProfile) TableName() string {
	return "user_profiles"
}

// Gender gender constants
const (
	GenderUnknown = 0 // unknown
	GenderMale    = 1 // male
	GenderFemale  = 2 // female
)
