package grpc

import (
	"context"

	commonpb "github.com/anychat/server/api/proto/common"
	userpb "github.com/anychat/server/api/proto/user"
	"github.com/anychat/server/internal/user/dto"
	"github.com/anychat/server/internal/user/model"
	"github.com/anychat/server/internal/user/service"
	"github.com/anychat/server/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserServer user gRPC server
type UserServer struct {
	userpb.UnimplementedUserServiceServer
	svc service.UserService
}

// NewUserServer creates user gRPC server
func NewUserServer(service service.UserService) *UserServer {
	return &UserServer{
		svc: service,
	}
}

// GetProfile retrieves personal profile
func (s *UserServer) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.UserProfileResponse, error) {
	resp, err := s.svc.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, convertError(err)
	}

	pbResp := &userpb.UserProfileResponse{
		UserId:    resp.UserID,
		Nickname:  resp.Nickname,
		Avatar:    resp.Avatar,
		Signature: resp.Signature,
		Gender:    int32(resp.Gender),
		Region:    resp.Region,
		QrcodeUrl: resp.QRCodeURL,
		CreatedAt: timestamppb.New(resp.CreatedAt),
	}

	if resp.Birthday != nil {
		pbResp.Birthday = timestamppb.New(*resp.Birthday)
	}
	if resp.Phone != nil {
		pbResp.Phone = resp.Phone
	}
	if resp.Email != nil {
		pbResp.Email = resp.Email
	}

	return pbResp, nil
}

// UpdateProfile updates personal profile
func (s *UserServer) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UserProfileResponse, error) {
	dtoReq := &dto.UpdateProfileRequest{}

	if req.Nickname != nil {
		dtoReq.Nickname = req.Nickname
	}
	if req.Avatar != nil {
		dtoReq.Avatar = req.Avatar
	}
	if req.Signature != nil {
		dtoReq.Signature = req.Signature
	}
	if req.Gender != nil {
		gender := int(*req.Gender)
		dtoReq.Gender = &gender
	}
	if req.Birthday != nil {
		birthday := req.Birthday.AsTime()
		dtoReq.Birthday = &birthday
	}
	if req.Region != nil {
		dtoReq.Region = req.Region
	}

	resp, err := s.svc.UpdateProfile(ctx, req.UserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	pbResp := &userpb.UserProfileResponse{
		UserId:    resp.UserID,
		Nickname:  resp.Nickname,
		Avatar:    resp.Avatar,
		Signature: resp.Signature,
		Gender:    int32(resp.Gender),
		Region:    resp.Region,
		QrcodeUrl: resp.QRCodeURL,
		CreatedAt: timestamppb.New(resp.CreatedAt),
	}

	if resp.Birthday != nil {
		pbResp.Birthday = timestamppb.New(*resp.Birthday)
	}
	if resp.Phone != nil {
		pbResp.Phone = resp.Phone
	}
	if resp.Email != nil {
		pbResp.Email = resp.Email
	}

	return pbResp, nil
}

// GetUserInfo retrieves user info
func (s *UserServer) GetUserInfo(ctx context.Context, req *userpb.GetUserInfoRequest) (*userpb.UserInfoResponse, error) {
	resp, err := s.svc.GetUserInfo(ctx, req.UserId, req.TargetUserId)
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.UserInfoResponse{
		UserId:    resp.UserID,
		Nickname:  resp.Nickname,
		Avatar:    resp.Avatar,
		Signature: resp.Signature,
		Gender:    int32(resp.Gender),
		Region:    resp.Region,
		IsFriend:  resp.IsFriend,
		IsBlocked: resp.IsBlocked,
	}, nil
}

// SearchUsers searches for users
func (s *UserServer) SearchUsers(ctx context.Context, req *userpb.SearchUsersRequest) (*userpb.SearchUsersResponse, error) {
	dtoReq := &dto.SearchUsersRequest{
		Keyword:  req.Keyword,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	resp, err := s.svc.SearchUsers(ctx, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	users := make([]*userpb.UserBriefInfo, 0, len(resp.Users))
	for _, u := range resp.Users {
		users = append(users, &userpb.UserBriefInfo{
			UserId:    u.UserID,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Signature: u.Signature,
		})
	}

	return &userpb.SearchUsersResponse{
		Total: resp.Total,
		Users: users,
	}, nil
}

// GetSettings retrieves user settings
func (s *UserServer) GetSettings(ctx context.Context, req *userpb.GetSettingsRequest) (*userpb.UserSettingsResponse, error) {
	resp, err := s.svc.GetSettings(ctx, req.UserId)
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.UserSettingsResponse{
		UserId:                resp.UserID,
		NotificationEnabled:   resp.NotificationEnabled,
		SoundEnabled:          resp.SoundEnabled,
		VibrationEnabled:      resp.VibrationEnabled,
		MessagePreviewEnabled: resp.MessagePreviewEnabled,
		FriendVerifyRequired:  resp.FriendVerifyRequired,
		SearchByPhone:         resp.SearchByPhone,
		SearchById:            resp.SearchByID,
		Language:              resp.Language,
	}, nil
}

// UpdateSettings updates user settings
func (s *UserServer) UpdateSettings(ctx context.Context, req *userpb.UpdateSettingsRequest) (*userpb.UserSettingsResponse, error) {
	dtoReq := &dto.UpdateSettingsRequest{}

	if req.NotificationEnabled != nil {
		dtoReq.NotificationEnabled = req.NotificationEnabled
	}
	if req.SoundEnabled != nil {
		dtoReq.SoundEnabled = req.SoundEnabled
	}
	if req.VibrationEnabled != nil {
		dtoReq.VibrationEnabled = req.VibrationEnabled
	}
	if req.MessagePreviewEnabled != nil {
		dtoReq.MessagePreviewEnabled = req.MessagePreviewEnabled
	}
	if req.FriendVerifyRequired != nil {
		dtoReq.FriendVerifyRequired = req.FriendVerifyRequired
	}
	if req.SearchByPhone != nil {
		dtoReq.SearchByPhone = req.SearchByPhone
	}
	if req.SearchById != nil {
		dtoReq.SearchByID = req.SearchById
	}
	if req.Language != nil {
		dtoReq.Language = req.Language
	}

	resp, err := s.svc.UpdateSettings(ctx, req.UserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.UserSettingsResponse{
		UserId:                resp.UserID,
		NotificationEnabled:   resp.NotificationEnabled,
		SoundEnabled:          resp.SoundEnabled,
		VibrationEnabled:      resp.VibrationEnabled,
		MessagePreviewEnabled: resp.MessagePreviewEnabled,
		FriendVerifyRequired:  resp.FriendVerifyRequired,
		SearchByPhone:         resp.SearchByPhone,
		SearchById:            resp.SearchByID,
		Language:              resp.Language,
	}, nil
}

// RefreshQRCode refreshes QR code
func (s *UserServer) RefreshQRCode(ctx context.Context, req *userpb.RefreshQRCodeRequest) (*userpb.QRCodeResponse, error) {
	resp, err := s.svc.RefreshQRCode(ctx, req.UserId)
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.QRCodeResponse{
		QrcodeUrl: resp.QRCodeURL,
		ExpiresAt: timestamppb.New(resp.ExpiresAt),
	}, nil
}

// GetUserByQRCode retrieves user by QR code
func (s *UserServer) GetUserByQRCode(ctx context.Context, req *userpb.GetUserByQRCodeRequest) (*userpb.UserInfoResponse, error) {
	resp, err := s.svc.GetUserByQRCode(ctx, req.Qrcode)
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.UserInfoResponse{
		UserId:    resp.UserID,
		Nickname:  resp.Nickname,
		Avatar:    resp.Avatar,
		Signature: resp.Signature,
	}, nil
}

// UpdatePushToken updates push token
func (s *UserServer) UpdatePushToken(ctx context.Context, req *userpb.UpdatePushTokenRequest) (*commonpb.Empty, error) {
	dtoReq := &dto.UpdatePushTokenRequest{
		DeviceID:  req.DeviceId,
		PushToken: req.PushToken,
		Platform:  model.PushPlatform(req.Platform),
	}

	err := s.svc.UpdatePushToken(ctx, req.UserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// BindPhone binds phone number
func (s *UserServer) BindPhone(ctx context.Context, req *userpb.BindPhoneRequest) (*userpb.BindPhoneResponse, error) {
	resp, err := s.svc.BindPhone(ctx, req.UserId, &dto.BindPhoneRequest{
		PhoneNumber: req.PhoneNumber,
		VerifyCode:  req.VerifyCode,
	})
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.BindPhoneResponse{
		PhoneNumber: resp.PhoneNumber,
		IsPrimary:   resp.IsPrimary,
	}, nil
}

// ChangePhone changes phone number
func (s *UserServer) ChangePhone(ctx context.Context, req *userpb.ChangePhoneRequest) (*userpb.ChangePhoneResponse, error) {
	dtoReq := &dto.ChangePhoneRequest{
		OldPhoneNumber: req.OldPhoneNumber,
		NewPhoneNumber: req.NewPhoneNumber,
		NewVerifyCode:  req.NewVerifyCode,
		DeviceID:       req.DeviceId,
	}
	if req.OldVerifyCode != nil {
		dtoReq.OldVerifyCode = req.OldVerifyCode
	}

	resp, err := s.svc.ChangePhone(ctx, req.UserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.ChangePhoneResponse{
		OldPhoneNumber: resp.OldPhoneNumber,
		NewPhoneNumber: resp.NewPhoneNumber,
	}, nil
}

// BindEmail binds email
func (s *UserServer) BindEmail(ctx context.Context, req *userpb.BindEmailRequest) (*userpb.BindEmailResponse, error) {
	resp, err := s.svc.BindEmail(ctx, req.UserId, &dto.BindEmailRequest{
		Email:      req.Email,
		VerifyCode: req.VerifyCode,
	})
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.BindEmailResponse{
		Email:     resp.Email,
		IsPrimary: resp.IsPrimary,
	}, nil
}

// ChangeEmail changes email
func (s *UserServer) ChangeEmail(ctx context.Context, req *userpb.ChangeEmailRequest) (*userpb.ChangeEmailResponse, error) {
	dtoReq := &dto.ChangeEmailRequest{
		OldEmail:      req.OldEmail,
		NewEmail:      req.NewEmail,
		NewVerifyCode: req.NewVerifyCode,
		DeviceID:      req.DeviceId,
	}
	if req.OldVerifyCode != nil {
		dtoReq.OldVerifyCode = req.OldVerifyCode
	}

	resp, err := s.svc.ChangeEmail(ctx, req.UserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.ChangeEmailResponse{
		OldEmail: resp.OldEmail,
		NewEmail: resp.NewEmail,
	}, nil
}

// InitUserData initializes user data (called by auth-service)
func (s *UserServer) InitUserData(ctx context.Context, req *userpb.InitUserDataRequest) (*commonpb.Empty, error) {
	err := s.svc.InitUserData(ctx, req.UserId, req.Nickname)
	if err != nil {
		return nil, convertError(err)
	}

	return &commonpb.Empty{}, nil
}

// convertError converts business error to gRPC error
func convertError(err error) error {
	if bizErr, ok := err.(*errors.Business); ok {
		switch bizErr.Code {
		case errors.CodeParamError:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeUserProfileNotFound:
			return status.Error(codes.NotFound, bizErr.Message)
		case errors.CodeNicknameUsed:
			return status.Error(codes.AlreadyExists, bizErr.Message)
		case errors.CodeNicknameSensitive:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodePhoneFormatInvalid,
			errors.CodeEmailFormatInvalid,
			errors.CodeOldPhoneNotMatch,
			errors.CodeOldEmailNotMatch:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodePhoneAlreadyBound, errors.CodeEmailAlreadyBound:
			return status.Error(codes.AlreadyExists, bizErr.Message)
		case errors.CodeUserNotFound:
			return status.Error(codes.NotFound, bizErr.Message)
		case errors.CodePermissionDenied:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeVerifyCodeError,
			errors.CodeVerifyCodeExpired,
			errors.CodeVerifyCodeAlreadyUsed,
			errors.CodeVerifyCodeNotFound:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeVerifyAttemptsExceeded:
			return status.Error(codes.ResourceExhausted, bizErr.Message)
		case errors.CodeQRCodeExpired, errors.CodeQRCodeInvalid:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		default:
			return status.Error(codes.Internal, bizErr.Message)
		}
	}
	return status.Error(codes.Internal, err.Error())
}
