package grpc

import (
	"context"

	authpb "github.com/anychat/server/api/proto/auth"
	commonpb "github.com/anychat/server/api/proto/common"
	"github.com/anychat/server/internal/auth/dto"
	"github.com/anychat/server/internal/auth/model"
	"github.com/anychat/server/internal/auth/service"
	"github.com/anychat/server/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer auth gRPC server
type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	svc service.AuthService
}

// NewAuthServer creates auth gRPC server
func NewAuthServer(service service.AuthService) *AuthServer {
	return &AuthServer{
		svc: service,
	}
}

// SendVerificationCode sends verification code
func (s *AuthServer) SendVerificationCode(ctx context.Context, req *authpb.SendVerificationCodeRequest) (*authpb.SendVerificationCodeResponse, error) {
	targetType := model.VerificationTargetType(req.TargetType)
	if !targetType.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid target_type")
	}
	purpose := model.VerificationPurpose(req.Purpose)
	if !purpose.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid purpose")
	}

	dtoReq := &dto.SendVerificationCodeRequest{
		Target:     req.Target,
		TargetType: targetType,
		Purpose:    purpose,
		DeviceID:   req.DeviceId,
		IPAddress:  req.IpAddress,
	}

	resp, err := s.svc.SendVerificationCode(ctx, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &authpb.SendVerificationCodeResponse{
		CodeId:    resp.CodeID,
		ExpiresIn: resp.ExpiresIn,
	}, nil
}

// Register user registration
func (s *AuthServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	deviceType := model.DeviceType(req.DeviceType)
	if !deviceType.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid device_type")
	}

	// Proto -> DTO conversion
	dtoReq := &dto.RegisterRequest{
		Password:      req.Password,
		VerifyCode:    req.VerifyCode,
		DeviceType:    deviceType,
		DeviceID:      req.DeviceId,
		ClientVersion: req.ClientVersion,
	}
	if req.PhoneNumber != nil {
		dtoReq.PhoneNumber = *req.PhoneNumber
	}
	if req.Email != nil {
		dtoReq.Email = *req.Email
	}
	if req.Nickname != nil {
		dtoReq.Nickname = *req.Nickname
	}

	// call service layer
	resp, err := s.svc.Register(ctx, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	// DTO -> Proto conversion
	return &authpb.RegisterResponse{
		UserId:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}

// Login user login
func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	deviceType := model.DeviceType(req.DeviceType)
	if !deviceType.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid device_type")
	}

	// Proto -> DTO conversion
	dtoReq := &dto.LoginRequest{
		Account:       req.Account,
		Password:      req.Password,
		DeviceType:    deviceType,
		DeviceID:      req.DeviceId,
		ClientVersion: req.ClientVersion,
		IpAddress:     req.IpAddress,
	}

	// call service layer
	resp, err := s.svc.Login(ctx, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	// DTO -> Proto conversion
	pbResp := &authpb.LoginResponse{
		UserId:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}

	if resp.User != nil {
		pbResp.User = &commonpb.UserInfo{
			UserId:   resp.User.UserID,
			Nickname: resp.User.Nickname,
			Avatar:   resp.User.Avatar,
		}
		if resp.User.Phone != nil {
			pbResp.User.Phone = resp.User.Phone
		}
		if resp.User.Email != nil {
			pbResp.User.Email = resp.User.Email
		}
	}

	return pbResp, nil
}

// Logout user logout
func (s *AuthServer) Logout(ctx context.Context, req *authpb.LogoutRequest) (*commonpb.Empty, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.LogoutRequest{
		DeviceID: req.DeviceId,
	}

	// call service layer
	err := s.svc.Logout(ctx, req.UserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// RefreshToken refresh access token
func (s *AuthServer) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	// call service layer
	resp, err := s.svc.RefreshToken(ctx, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	// DTO -> Proto conversion
	return &authpb.RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}

// ChangePassword change password
func (s *AuthServer) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*commonpb.Empty, error) {
	// Proto -> DTO conversion
	dtoReq := &dto.ChangePasswordRequest{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
		DeviceID:    req.DeviceId,
	}

	// call service layer
	err := s.svc.ChangePassword(ctx, req.UserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// ResetPassword reset password (forgot password)
func (s *AuthServer) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*commonpb.Empty, error) {
	dtoReq := &dto.ResetPasswordRequest{
		Account:     req.Account,
		VerifyCode:  req.VerifyCode,
		NewPassword: req.NewPassword,
	}

	err := s.svc.ResetPassword(ctx, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// ValidateToken validates token (called by gateway)
func (s *AuthServer) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	// call service layer
	claims, err := s.svc.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	// return validation result
	return &authpb.ValidateTokenResponse{
		Valid:      true,
		UserId:     claims.UserID,
		DeviceId:   claims.DeviceID,
		DeviceType: authpb.DeviceType(claims.DeviceType),
	}, nil
}

// convertError converts business errors to gRPC errors
func convertError(err error) error {
	if bizErr, ok := err.(*errors.Business); ok {
		switch bizErr.Code {
		case errors.CodeParamError:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeUserNotFound:
			return status.Error(codes.NotFound, bizErr.Message)
		case errors.CodeUnauthorized:
			return status.Error(codes.Unauthenticated, bizErr.Message)
		case errors.CodeUserExists:
			return status.Error(codes.AlreadyExists, bizErr.Message)
		case errors.CodePasswordError:
			return status.Error(codes.Unauthenticated, bizErr.Message)
		case errors.CodePasswordWeak:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeAccountDisabled:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeTargetFormatInvalid,
			errors.CodeVerifyCodeError,
			errors.CodeVerifyCodeExpired,
			errors.CodeVerifyCodeAlreadyUsed,
			errors.CodeVerifyCodeNotFound:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeSendRateLimited,
			errors.CodeSendLimitReached,
			errors.CodeVerifyAttemptsExceeded:
			return status.Error(codes.ResourceExhausted, bizErr.Message)
		case errors.CodeRefreshTokenInvalid, errors.CodeRefreshTokenExpired:
			return status.Error(codes.Unauthenticated, bizErr.Message)
		default:
			return status.Error(codes.Internal, bizErr.Message)
		}
	}
	return status.Error(codes.Internal, err.Error())
}
