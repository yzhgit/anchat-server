package model

import (
	"time"
)

// UserSession user session model
type UserSession struct {
	ID                    int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID                string    `gorm:"column:user_id;not null" json:"userId"`
	DeviceID              string    `gorm:"column:device_id;not null" json:"deviceId"`
	AccessToken           string    `gorm:"column:access_token;not null" json:"accessToken"`
	RefreshToken          string    `gorm:"column:refresh_token;not null" json:"refreshToken"`
	AccessTokenExpiresAt  time.Time `gorm:"column:access_token_expires_at;not null" json:"accessTokenExpiresAt"`
	RefreshTokenExpiresAt time.Time `gorm:"column:refresh_token_expires_at;not null" json:"refreshTokenExpiresAt"`
	CreatedAt             time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt             time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName returns table name
func (UserSession) TableName() string {
	return "user_sessions"
}

// IsAccessTokenExpired checks if access token is expired
func (s *UserSession) IsAccessTokenExpired() bool {
	return time.Now().After(s.AccessTokenExpiresAt)
}

// IsRefreshTokenExpired checks if refresh token is expired
func (s *UserSession) IsRefreshTokenExpired() bool {
	return time.Now().After(s.RefreshTokenExpiresAt)
}
