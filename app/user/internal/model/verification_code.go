package model

import (
	"fmt"
	"strings"
	"time"
)

// VerificationTargetType is the verification target type enum.
type VerificationTargetType int16

const (
	TargetTypeUnknown VerificationTargetType = 0
	TargetTypeSMS     VerificationTargetType = 1
	TargetTypeEmail   VerificationTargetType = 2
)

var targetTypeValueToString = map[VerificationTargetType]string{
	TargetTypeSMS:   "sms",
	TargetTypeEmail: "email",
}

func (t VerificationTargetType) String() string {
	if s, ok := targetTypeValueToString[t]; ok {
		return s
	}
	return "unknown"
}

func (t VerificationTargetType) IsValid() bool {
	return t == TargetTypeSMS || t == TargetTypeEmail
}

func ParseVerificationTargetType(value string) (VerificationTargetType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "sms":
		return TargetTypeSMS, nil
	case "email":
		return TargetTypeEmail, nil
	default:
		return TargetTypeUnknown, fmt.Errorf("unsupported verification target type: %s", value)
	}
}

// VerificationPurpose is the verification purpose enum.
type VerificationPurpose int16

const (
	PurposeUnknown       VerificationPurpose = 0
	PurposeRegister      VerificationPurpose = 1
	PurposeLogin         VerificationPurpose = 2
	PurposeResetPassword VerificationPurpose = 3
	PurposeBindPhone     VerificationPurpose = 4
	PurposeChangePhone   VerificationPurpose = 5
	PurposeBindEmail     VerificationPurpose = 6
	PurposeChangeEmail   VerificationPurpose = 7
)

var purposeValueToString = map[VerificationPurpose]string{
	PurposeRegister:      "register",
	PurposeLogin:         "login",
	PurposeResetPassword: "reset_password",
	PurposeBindPhone:     "bind_phone",
	PurposeChangePhone:   "change_phone",
	PurposeBindEmail:     "bind_email",
	PurposeChangeEmail:   "change_email",
}

func (p VerificationPurpose) String() string {
	if s, ok := purposeValueToString[p]; ok {
		return s
	}
	return "unknown"
}

func (p VerificationPurpose) IsValid() bool {
	return p >= PurposeRegister && p <= PurposeChangeEmail
}

func ParseVerificationPurpose(value string) (VerificationPurpose, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "register":
		return PurposeRegister, nil
	case "login":
		return PurposeLogin, nil
	case "reset_password":
		return PurposeResetPassword, nil
	case "bind_phone":
		return PurposeBindPhone, nil
	case "change_phone":
		return PurposeChangePhone, nil
	case "bind_email":
		return PurposeBindEmail, nil
	case "change_email":
		return PurposeChangeEmail, nil
	default:
		return PurposeUnknown, fmt.Errorf("unsupported verification purpose: %s", value)
	}
}

// VerificationCodeStatus is the verification code status enum.
type VerificationCodeStatus int16

const (
	CodeStatusUnknown   VerificationCodeStatus = 0
	CodeStatusPending   VerificationCodeStatus = 1
	CodeStatusVerified  VerificationCodeStatus = 2
	CodeStatusExpired   VerificationCodeStatus = 3
	CodeStatusLocked    VerificationCodeStatus = 4
	CodeStatusCancelled VerificationCodeStatus = 5
)

var statusValueToString = map[VerificationCodeStatus]string{
	CodeStatusPending:   "pending",
	CodeStatusVerified:  "verified",
	CodeStatusExpired:   "expired",
	CodeStatusLocked:    "locked",
	CodeStatusCancelled: "cancelled",
}

func (s VerificationCodeStatus) String() string {
	if v, ok := statusValueToString[s]; ok {
		return v
	}
	return "unknown"
}

func (s VerificationCodeStatus) IsValid() bool {
	return s >= CodeStatusPending && s <= CodeStatusCancelled
}

func ParseVerificationCodeStatus(value string) (VerificationCodeStatus, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pending":
		return CodeStatusPending, nil
	case "verified":
		return CodeStatusVerified, nil
	case "expired":
		return CodeStatusExpired, nil
	case "locked":
		return CodeStatusLocked, nil
	case "cancelled":
		return CodeStatusCancelled, nil
	default:
		return CodeStatusUnknown, fmt.Errorf("unsupported verification code status: %s", value)
	}
}

// VerificationCode verification code model
type VerificationCode struct {
	ID                int64                  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CodeID            string                 `gorm:"column:code_id;not null;uniqueIndex" json:"codeId"`
	Target            string                 `gorm:"column:target;not null;index" json:"target"`
	TargetType        VerificationTargetType `gorm:"column:target_type;type:smallint;not null" json:"targetType"`
	CodeHash          string                 `gorm:"column:code_hash;not null" json:"-"`
	Purpose           VerificationPurpose    `gorm:"column:purpose;type:smallint;not null" json:"purpose"`
	ExpiresAt         time.Time              `gorm:"column:expires_at;not null" json:"expiresAt"`
	VerifiedAt        *time.Time             `gorm:"column:verified_at" json:"verifiedAt"`
	Status            VerificationCodeStatus `gorm:"column:status;type:smallint;not null;default:1" json:"status"`
	SendIP            string                 `gorm:"column:send_ip" json:"sendIp"`
	SendDeviceID      string                 `gorm:"column:send_device_id" json:"sendDeviceId"`
	AttemptCount      int                    `gorm:"column:attempt_count;not null;default:0" json:"attemptCount"`
	Provider          string                 `gorm:"column:provider" json:"provider"`
	ProviderMessageID string                 `gorm:"column:provider_message_id" json:"providerMessageId"`
	CreatedAt         time.Time              `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt         time.Time              `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}

func (v *VerificationCode) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

func (v *VerificationCode) IsPending() bool {
	return v.Status == CodeStatusPending
}

func (v *VerificationCode) IsVerified() bool {
	return v.Status == CodeStatusVerified
}

func (v *VerificationCode) IsLocked() bool {
	return v.Status == CodeStatusLocked
}
