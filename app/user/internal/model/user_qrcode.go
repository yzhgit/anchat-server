package model

import (
	"time"
)

// UserQRCode user QR code model
type UserQRCode struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      string    `gorm:"column:user_id;not null" json:"userId"`
	QRCodeToken string    `gorm:"column:qrcode_token;unique;not null" json:"qrcodeToken"`
	QRCodeURL   string    `gorm:"column:qrcode_url;not null" json:"qrcodeUrl"`
	ExpiresAt   time.Time `gorm:"column:expires_at;not null" json:"expiresAt"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
}

// TableName returns table name
func (UserQRCode) TableName() string {
	return "user_qrcodes"
}

// IsExpired checks if QR code is expired
func (q *UserQRCode) IsExpired() bool {
	return time.Now().After(q.ExpiresAt)
}
