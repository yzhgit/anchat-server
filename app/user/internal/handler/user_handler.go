package handler

import (
	"context"

	userv1 "flamingo/api/user/v1"

	"flamingo/app/user/internal/model"
	"flamingo/pkg/crypto"
	"flamingo/pkg/errors"
	"flamingo/pkg/md"

	"github.com/go-kratos/kratos/v2/log"
	empty "github.com/golang/protobuf/ptypes/empty"
)

// UserService user service interface
type UserService interface {
	SendVerificationCode(ctx context.Context, req *userv1.SendVerificationCodeRequest) (*userv1.SendVerificationCodeResponse, error)
	Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error)
	Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error)
	Logout(ctx context.Context, userID string, deviceID string) error
	RefreshToken(ctx context.Context, refreshToken string) (*userv1.RefreshTokenResponse, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword, deviceID string) error
	ResetPassword(ctx context.Context, account, verifyCode, newPassword string) error

	// User profile
	GetProfile(ctx context.Context, userID string, req *userv1.GetProfileRequest) (*userv1.UserProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req *userv1.UpdateProfileRequest) (*userv1.UserProfileResponse, error)
	GetUserInfo(ctx context.Context, userID string, req *userv1.GetUserInfoRequest) (*userv1.UserInfoResponse, error)
	SearchUsers(ctx context.Context, req *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error)

	// User settings
	GetSettings(ctx context.Context, userID string, req *userv1.GetSettingsRequest) (*userv1.UserSettingsResponse, error)
	UpdateSettings(ctx context.Context, userID string, req *userv1.UpdateSettingsRequest) (*userv1.UserSettingsResponse, error)

	// QR code
	RefreshQRCode(ctx context.Context, userID string) (*userv1.QRCodeResponse, error)
	GetUserByQRCode(ctx context.Context, code string) (*userv1.UserInfoResponse, error)

	// Push token
	UpdatePushToken(ctx context.Context, userID string, req *userv1.UpdatePushTokenRequest) error

	// Account binding
	BindPhone(ctx context.Context, userID string, req *userv1.BindPhoneRequest) (*userv1.BindPhoneResponse, error)
	ChangePhone(ctx context.Context, userID string, req *userv1.ChangePhoneRequest) (*userv1.ChangePhoneResponse, error)
	BindEmail(ctx context.Context, userID string, req *userv1.BindEmailRequest) (*userv1.BindEmailResponse, error)
	ChangeEmail(ctx context.Context, userID string, req *userv1.ChangeEmailRequest) (*userv1.ChangeEmailResponse, error)
}

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	svc UserService
	log *log.Helper
}

// NewUserServer creates user gRPC handler
func NewUserHandler(svc UserService, logger log.Logger) *UserHandler {
	return &UserHandler{
		svc: svc,
		log: log.NewHelper(logger),
	}
}

// SendVerificationCode sends verification code
func (s *UserHandler) SendVerificationCode(ctx context.Context, req *userv1.SendVerificationCodeRequest) (*userv1.SendVerificationCodeResponse, error) {
	targetType := model.VerificationTargetType(req.TargetType)
	if !targetType.IsValid() {
		return nil, errors.BadRequest(ctx, "invalid target_type")
	}
	purpose := model.VerificationPurpose(req.Purpose)
	if !purpose.IsValid() {
		return nil, errors.BadRequest(ctx, "invalid purpose")
	}

	// validate target format based on target type
	if req.Target == "" {
		return nil, errors.BadRequest(ctx, "target is required")
	}

	resp, err := s.svc.SendVerificationCode(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// Register user registration
func (s *UserHandler) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	if req.GetPhoneNumber() == "" && req.GetEmail() == "" {
		return nil, errors.BadRequest(ctx, "phone or email required")
	}
	if req.GetPassword() == "" {
		return nil, errors.BadRequest(ctx, "password is required")
	}
	if !crypto.ValidatePasswordStrength(req.GetPassword()) {
		return nil, errors.BadRequest(ctx, "password does not meet strength requirements")
	}
	deviceType := model.DeviceType(req.DeviceType)
	if !deviceType.IsValid() {
		return nil, errors.BadRequest(ctx, "invalid device_type")
	}

	resp, err := s.svc.Register(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// Login user login
func (s *UserHandler) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	deviceType := model.DeviceType(req.DeviceType)
	if !deviceType.IsValid() {
		return nil, errors.BadRequest(ctx, "invalid device_type")
	}

	if req.Account == "" {
		return nil, errors.BadRequest(ctx, "account is required")
	}
	if req.Password == "" {
		return nil, errors.BadRequest(ctx, "password required")
	}

	resp, err := s.svc.Login(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// Logout user logout
func (s *UserHandler) Logout(ctx context.Context, req *userv1.LogoutRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	err := s.svc.Logout(ctx, userID, req.DeviceId)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return &empty.Empty{}, nil
}

// RefreshToken refresh access token
func (s *UserHandler) RefreshToken(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
	resp, err := s.svc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// ChangePassword change password
func (s *UserHandler) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	if req.OldPassword == "" {
		return nil, errors.BadRequest(ctx, "old_password is required")
	}
	if req.NewPassword == "" {
		return nil, errors.BadRequest(ctx, "new_password is required")
	}
	if !crypto.ValidatePasswordStrength(req.NewPassword) {
		return nil, errors.BadRequest(ctx, "password does not meet strength requirements")
	}
	err := s.svc.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword, req.DeviceId)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return &empty.Empty{}, nil
}

// ResetPassword reset password (forgot password)
func (s *UserHandler) ResetPassword(ctx context.Context, req *userv1.ResetPasswordRequest) (*empty.Empty, error) {
	if req.Account == "" {
		return nil, errors.BadRequest(ctx, "account is required")
	}
	if req.VerifyCode == "" {
		return nil, errors.BadRequest(ctx, "verify_code is required")
	}
	if req.NewPassword == "" {
		return nil, errors.BadRequest(ctx, "new_password is required")
	}
	if !crypto.ValidatePasswordStrength(req.NewPassword) {
		return nil, errors.BadRequest(ctx, "password does not meet strength requirements")
	}
	err := s.svc.ResetPassword(ctx, req.Account, req.VerifyCode, req.NewPassword)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return &empty.Empty{}, nil
}

// GetProfile retrieves personal profile
func (s *UserHandler) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.UserProfileResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.GetProfile(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// UpdateProfile updates personal profile
func (s *UserHandler) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UserProfileResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.UpdateProfile(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// GetUserInfo retrieves user info
func (s *UserHandler) GetUserInfo(ctx context.Context, req *userv1.GetUserInfoRequest) (*userv1.UserInfoResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.GetUserInfo(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// SearchUsers searches for users
func (s *UserHandler) SearchUsers(ctx context.Context, req *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	resp, err := s.svc.SearchUsers(ctx, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// GetSettings retrieves user settings. When req.UserId is set the caller
// requests settings of that user (service-to-service), otherwise it returns
// the current authenticated user's own settings.
func (s *UserHandler) GetSettings(ctx context.Context, req *userv1.GetSettingsRequest) (*userv1.UserSettingsResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		userID = md.MustGetUserID(ctx)
	}
	resp, err := s.svc.GetSettings(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// UpdateSettings updates user settings
func (s *UserHandler) UpdateSettings(ctx context.Context, req *userv1.UpdateSettingsRequest) (*userv1.UserSettingsResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.UpdateSettings(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// RefreshQRCode refreshes QR code
func (s *UserHandler) RefreshQRCode(ctx context.Context, req *userv1.RefreshQRCodeRequest) (*userv1.QRCodeResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.RefreshQRCode(ctx, userID)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// GetUserByQRCode retrieves user by QR code
func (s *UserHandler) GetUserByQRCode(ctx context.Context, req *userv1.GetUserByQRCodeRequest) (*userv1.UserInfoResponse, error) {
	resp, err := s.svc.GetUserByQRCode(ctx, req.Qrcode)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// UpdatePushToken updates push token
func (s *UserHandler) UpdatePushToken(ctx context.Context, req *userv1.UpdatePushTokenRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	err := s.svc.UpdatePushToken(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return &empty.Empty{}, nil
}

// BindPhone binds phone number
func (s *UserHandler) BindPhone(ctx context.Context, req *userv1.BindPhoneRequest) (*userv1.BindPhoneResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.BindPhone(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// ChangePhone changes phone number
func (s *UserHandler) ChangePhone(ctx context.Context, req *userv1.ChangePhoneRequest) (*userv1.ChangePhoneResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.ChangePhone(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// BindEmail binds email
func (s *UserHandler) BindEmail(ctx context.Context, req *userv1.BindEmailRequest) (*userv1.BindEmailResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.BindEmail(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}

// ChangeEmail changes email
func (s *UserHandler) ChangeEmail(ctx context.Context, req *userv1.ChangeEmailRequest) (*userv1.ChangeEmailResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.ChangeEmail(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}
	return resp, nil
}
