package model

import (
	"time"
)

// UserSettings user settings model
type UserSettings struct {
	ID                    int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID                string    `gorm:"column:user_id;unique;not null" json:"userId"`
	NotificationEnabled   bool      `gorm:"column:notification_enabled;not null;default:true" json:"notificationEnabled"`
	SoundEnabled          bool      `gorm:"column:sound_enabled;not null;default:true" json:"soundEnabled"`
	VibrationEnabled      bool      `gorm:"column:vibration_enabled;not null;default:true" json:"vibrationEnabled"`
	MessagePreviewEnabled bool      `gorm:"column:message_preview_enabled;not null;default:true" json:"messagePreviewEnabled"`
	FriendVerifyRequired  bool      `gorm:"column:friend_verify_required;not null;default:true" json:"friendVerifyRequired"`
	SearchByPhone         bool      `gorm:"column:search_by_phone;not null;default:true" json:"searchByPhone"`
	SearchByID            bool      `gorm:"column:search_by_id;not null;default:true" json:"searchById"`
	Language              string    `gorm:"column:language;not null;default:'zh_CN'" json:"language"`
	CreatedAt             time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt             time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName returns table name
func (UserSettings) TableName() string {
	return "user_settings"
}
