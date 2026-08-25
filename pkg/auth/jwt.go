package auth

import (
	"fmt"
	"time"

	"flamingo/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewJWTManager)

// Claims JWT claims
type Claims struct {
	UserID     string `json:"userId"`
	DeviceID   string `json:"deviceId"`
	DeviceType int16  `json:"deviceType"`
	TokenType  string `json:"tokenType"` // access, refresh
	jwt.RegisteredClaims
}

var _ jwt.Claims = (*Claims)(nil)

// Manager JWT manager
type JWTManager struct {
	conf config.Auth
}

// NewJWTManager creates a new JWT manager.
func NewJWTManager(conf config.Auth) (*JWTManager, error) {
	if conf.Secret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	if conf.AccessTokenExpire.AsDuration() <= 0 {
		return nil, fmt.Errorf("jwt access_token_expire must be positive")
	}
	if conf.RefreshTokenExpire.AsDuration() <= 0 {
		return nil, fmt.Errorf("jwt refresh_token_expire must be positive")
	}
	return &JWTManager{
		conf: conf,
	}, nil
}

// GenerateAccessToken generates an access token
func (m *JWTManager) GenerateAccessToken(userID, deviceID string, deviceType int16) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     userID,
		DeviceID:   deviceID,
		DeviceType: deviceType,
		TokenType:  "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.conf.AccessTokenExpire.AsDuration())),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.conf.Secret))
}

// GenerateRefreshToken generates a refresh token
func (m *JWTManager) GenerateRefreshToken(userID, deviceID string, deviceType int16) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     userID,
		DeviceID:   deviceID,
		DeviceType: deviceType,
		TokenType:  "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.conf.RefreshTokenExpire.AsDuration())),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.conf.Secret))
}

// ParseToken parses a token
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.conf.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateAccessToken validates an access token
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}
