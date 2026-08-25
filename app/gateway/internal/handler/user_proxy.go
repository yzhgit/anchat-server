package handler

import (
	"context"

	userv1 "flamingo/api/user/v1"

	empty "github.com/golang/protobuf/ptypes/empty"
)

// userProxy implements UserServiceHTTPServer by forwarding to the gRPC client.
type userProxy struct {
	client userv1.UserServiceClient
}

func (p *userProxy) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	return p.client.Login(ctx, req)
}

func (p *userProxy) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	return p.client.Register(ctx, req)
}

func (p *userProxy) Logout(ctx context.Context, req *userv1.LogoutRequest) (*empty.Empty, error) {
	return p.client.Logout(ctx, req)
}

func (p *userProxy) RefreshToken(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
	return p.client.RefreshToken(ctx, req)
}

func (p *userProxy) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*empty.Empty, error) {
	return p.client.ChangePassword(ctx, req)
}

func (p *userProxy) ResetPassword(ctx context.Context, req *userv1.ResetPasswordRequest) (*empty.Empty, error) {
	return p.client.ResetPassword(ctx, req)
}

func (p *userProxy) SendVerificationCode(ctx context.Context, req *userv1.SendVerificationCodeRequest) (*userv1.SendVerificationCodeResponse, error) {
	return p.client.SendVerificationCode(ctx, req)
}

func (p *userProxy) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.UserProfileResponse, error) {
	return p.client.GetProfile(ctx, req)
}

func (p *userProxy) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UserProfileResponse, error) {
	return p.client.UpdateProfile(ctx, req)
}

func (p *userProxy) BindPhone(ctx context.Context, req *userv1.BindPhoneRequest) (*userv1.BindPhoneResponse, error) {
	return p.client.BindPhone(ctx, req)
}

func (p *userProxy) ChangePhone(ctx context.Context, req *userv1.ChangePhoneRequest) (*userv1.ChangePhoneResponse, error) {
	return p.client.ChangePhone(ctx, req)
}

func (p *userProxy) BindEmail(ctx context.Context, req *userv1.BindEmailRequest) (*userv1.BindEmailResponse, error) {
	return p.client.BindEmail(ctx, req)
}

func (p *userProxy) ChangeEmail(ctx context.Context, req *userv1.ChangeEmailRequest) (*userv1.ChangeEmailResponse, error) {
	return p.client.ChangeEmail(ctx, req)
}

func (p *userProxy) GetUserInfo(ctx context.Context, req *userv1.GetUserInfoRequest) (*userv1.UserInfoResponse, error) {
	return p.client.GetUserInfo(ctx, req)
}

func (p *userProxy) SearchUsers(ctx context.Context, req *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	return p.client.SearchUsers(ctx, req)
}

func (p *userProxy) GetSettings(ctx context.Context, req *userv1.GetSettingsRequest) (*userv1.UserSettingsResponse, error) {
	return p.client.GetSettings(ctx, req)
}

func (p *userProxy) UpdateSettings(ctx context.Context, req *userv1.UpdateSettingsRequest) (*userv1.UserSettingsResponse, error) {
	return p.client.UpdateSettings(ctx, req)
}

func (p *userProxy) RefreshQRCode(ctx context.Context, req *userv1.RefreshQRCodeRequest) (*userv1.QRCodeResponse, error) {
	return p.client.RefreshQRCode(ctx, req)
}

func (p *userProxy) GetUserByQRCode(ctx context.Context, req *userv1.GetUserByQRCodeRequest) (*userv1.UserInfoResponse, error) {
	return p.client.GetUserByQRCode(ctx, req)
}

func (p *userProxy) UpdatePushToken(ctx context.Context, req *userv1.UpdatePushTokenRequest) (*empty.Empty, error) {
	return p.client.UpdatePushToken(ctx, req)
}
