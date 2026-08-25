package model

import "time"

// GroupQRCode represents a group QR code
type GroupQRCode struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID   string    `gorm:"column:group_id;not null;index" json:"groupId"`
	Token     string    `gorm:"column:token;not null;uniqueIndex" json:"token"`
	CreatedBy string    `gorm:"column:created_by;not null" json:"createdBy"`
	ExpireAt  time.Time `gorm:"column:expire_at;not null" json:"expireAt"`
	IsActive  bool      `gorm:"column:is_active;default:true" json:"isActive"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (GroupQRCode) TableName() string { return "group_qrcodes" }

// IsValid checks if QR code is valid (not expired and active)
func (q *GroupQRCode) IsValid() bool {
	return q.IsActive && q.ExpireAt.After(time.Now())
}

const DefaultQRCodeTTL = 7 * 24 * time.Hour

// QRCodeRenewThreshold auto-renew when within this duration of expiration
const QRCodeRenewThreshold = 24 * time.Hour
