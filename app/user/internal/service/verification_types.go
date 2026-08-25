package service

import (
	"time"

	"flamingo/app/user/internal/model"
)

// SendCodeRequest send verification code internal request.
type SendCodeRequest struct {
	Target     string
	TargetType model.VerificationTargetType
	Purpose    model.VerificationPurpose
	DeviceID   string
}

// SendCodeResponse send verification code internal response.
type SendCodeResponse struct {
	CodeID    string
	ExpiresIn int64
	Sent      bool
	Message   string
}

// VerifyCodeRequest verify code internal request.
type VerifyCodeRequest struct {
	Target     string
	TargetType model.VerificationTargetType
	Purpose    model.VerificationPurpose
	Code       string
}

// VerifyCodeResponse verify code internal response.
type VerifyCodeResponse struct {
	Valid   bool
	CodeID  string
	Message string
}

// CheckCodeStatusResponse check code status internal response.
type CheckCodeStatusResponse struct {
	Status    model.VerificationCodeStatus
	ExpiresAt time.Time
}
