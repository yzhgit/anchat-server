package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	friendv1 "flamingo/api/friend/v1"
	userv1 "flamingo/api/user/v1"

	"flamingo/app/user/internal/handler"
	"flamingo/app/user/internal/model"
	"flamingo/pkg/auth"
	"flamingo/pkg/broker"
	confpkg "flamingo/pkg/config"
	"flamingo/pkg/crypto"
	"flamingo/pkg/errors"
	"flamingo/pkg/validator"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

var _ handler.UserService = (*userServiceImpl)(nil)

// userServiceImpl user service implementation
type userServiceImpl struct {
	db            *gorm.DB
	authConf      confpkg.Auth
	profileRepo   UserProfileRepository
	settingsRepo  UserSettingsRepository
	qrcodeRepo    UserQRCodeRepository
	pushTokenRepo UserPushTokenRepository
	userRepo      UserRepository
	deviceRepo    UserDeviceRepository
	sessionRepo   UserSessionRepository
	friendClient  friendv1.FriendServiceClient
	verifyManager VerificationManager
	jwtManager    *auth.JWTManager
	broker        broker.Broker
	log           *log.Helper
}

// NewUserService creates user service
func NewUserService(
	db *gorm.DB,
	authConf confpkg.Auth,
	profileRepo UserProfileRepository,
	settingsRepo UserSettingsRepository,
	qrcodeRepo UserQRCodeRepository,
	pushTokenRepo UserPushTokenRepository,
	userRepo UserRepository,
	deviceRepo UserDeviceRepository,
	sessionRepo UserSessionRepository,
	friendClient friendv1.FriendServiceClient,
	verifyManager VerificationManager,
	jwtManager *auth.JWTManager,
	broker broker.Broker,
	logger log.Logger,
) handler.UserService {
	return &userServiceImpl{
		db:            db,
		authConf:      authConf,
		profileRepo:   profileRepo,
		settingsRepo:  settingsRepo,
		qrcodeRepo:    qrcodeRepo,
		pushTokenRepo: pushTokenRepo,
		userRepo:      userRepo,
		deviceRepo:    deviceRepo,
		sessionRepo:   sessionRepo,
		friendClient:  friendClient,
		verifyManager: verifyManager,
		jwtManager:    jwtManager,
		broker:        broker,
		log:           log.NewHelper(logger),
	}
}

// ==================== Init ====================

// initUserDataInTx initializes user data with explicit repos (usable inside a tx)
func (s *userServiceImpl) initUserDataInTx(
	ctx context.Context, userID, nickname string,
	profileRepo UserProfileRepository, settingsRepo UserSettingsRepository, qrcodeRepo UserQRCodeRepository,
) error {
	if nickname == "" {
		nickname = fmt.Sprintf("User_%s", userID[:8])
	}

	profile := &model.UserProfile{
		UserID:   userID,
		Nickname: nickname,
		Gender:   model.GenderUnknown,
	}
	if err := profileRepo.Create(ctx, profile); err != nil {
		return err
	}

	settings := &model.UserSettings{
		UserID:                userID,
		NotificationEnabled:   true,
		SoundEnabled:          true,
		VibrationEnabled:      true,
		MessagePreviewEnabled: true,
		FriendVerifyRequired:  true,
		SearchByPhone:         true,
		SearchByID:            true,
		Language:              "zh_CN",
	}
	if err := settingsRepo.Create(ctx, settings); err != nil {
		return err
	}

	_, err := s.refreshQRCodeInTx(ctx, userID, profileRepo, qrcodeRepo)
	return err
}

// refreshQRCodeInTx creates a QR code and updates the profile's qrcode_url
// using explicit repos (usable inside a tx).
func (s *userServiceImpl) refreshQRCodeInTx(
	ctx context.Context, userID string,
	profileRepo UserProfileRepository, qrcodeRepo UserQRCodeRepository,
) (*userv1.QRCodeResponse, error) {
	qrcodeToken, err := crypto.GenerateQRCodeToken(userID)
	if err != nil {
		return nil, err
	}

	qrcodeURL := fmt.Sprintf("anychat://qrcode?token=%s", qrcodeToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	qrcode := &model.UserQRCode{
		UserID:      userID,
		QRCodeToken: qrcodeToken,
		QRCodeURL:   qrcodeURL,
		ExpiresAt:   expiresAt,
	}
	if err := qrcodeRepo.Create(ctx, qrcode); err != nil {
		return nil, err
	}

	if err := profileRepo.UpdateQRCode(ctx, userID, qrcodeURL); err != nil {
		return nil, err
	}

	return &userv1.QRCodeResponse{
		QrcodeUrl: qrcodeURL,
		ExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

// ==================== Auth Methods ====================

// SendVerificationCode sends verification code
func (s *userServiceImpl) SendVerificationCode(ctx context.Context, req *userv1.SendVerificationCodeRequest) (*userv1.SendVerificationCodeResponse, error) {
	resp, err := s.verifyManager.SendCode(ctx, &SendCodeRequest{
		Target:     req.Target,
		TargetType: model.VerificationTargetType(req.TargetType),
		Purpose:    model.VerificationPurpose(req.Purpose),
		DeviceID:   req.DeviceId,
	}, req.IpAddress)
	if err != nil {
		return nil, err
	}

	return &userv1.SendVerificationCodeResponse{
		CodeId:    resp.CodeID,
		ExpiresIn: resp.ExpiresIn,
	}, nil
}

// Register user registration
func (s *userServiceImpl) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	phoneNumber := req.GetPhoneNumber()
	email := req.GetEmail()
	nickname := req.GetNickname()

	deviceType := model.DeviceType(req.DeviceType)

	// Validate and verify code
	var targetType model.VerificationTargetType
	var verifyTarget string
	if phoneNumber != "" {
		// Normalise to canonical E164 so the DB uniqueness constraint and
		// the verification-code lookup both operate on the same value.
		e164, err := validator.NormalizePhone(phoneNumber)
		if err != nil {
			return nil, errors.NewBusiness(errors.CodePhoneFormatInvalid, "invalid phone number format")
		}
		phoneNumber = e164
		targetType = model.TargetTypeSMS
		verifyTarget = phoneNumber
	} else {
		if !validator.ValidateEmail(email) {
			return nil, errors.NewBusiness(errors.CodeEmailFormatInvalid, "invalid email format")
		}
		targetType = model.TargetTypeEmail
		verifyTarget = email
	}

	if err := s.verifyCode(ctx, verifyTarget, targetType, model.PurposeRegister, req.VerifyCode); err != nil {
		return nil, err
	}

	// Check if user already exists
	if phoneNumber != "" {
		if _, err := s.userRepo.GetByPhone(ctx, phoneNumber); err == nil {
			return nil, errors.NewBusiness(errors.CodeUserExists, "user already exists")
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	if email != "" {
		if _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
			return nil, errors.NewBusiness(errors.CodeUserExists, "user already exists")
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	userID := uuid.New().String()
	deviceID := req.DeviceId
	if deviceID == "" {
		deviceID = uuid.New().String()
	}
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var accessToken, refreshToken string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		userRepoTx := s.userRepo.WithTx(tx)
		profileRepoTx := s.profileRepo.WithTx(tx)
		settingsRepoTx := s.settingsRepo.WithTx(tx)
		qrcodeRepoTx := s.qrcodeRepo.WithTx(tx)
		deviceRepoTx := s.deviceRepo.WithTx(tx)
		sessionRepoTx := s.sessionRepo.WithTx(tx)

		user := &model.User{
			ID:           userID,
			Phone:        toPtr(phoneNumber),
			Email:        toPtr(email),
			PasswordHash: passwordHash,
			Status:       model.UserStatusNormal,
		}
		if txErr := userRepoTx.Create(ctx, user); txErr != nil {
			return fmt.Errorf("create user: %w", txErr)
		}

		if txErr := s.initUserDataInTx(ctx, userID, nickname, profileRepoTx, settingsRepoTx, qrcodeRepoTx); txErr != nil {
			return fmt.Errorf("init user data: %w", txErr)
		}

		device := &model.UserDevice{
			UserID:        userID,
			DeviceID:      deviceID,
			DeviceType:    deviceType,
			ClientVersion: req.ClientVersion,
		}
		if txErr := deviceRepoTx.Create(ctx, device); txErr != nil {
			return fmt.Errorf("create device: %w", txErr)
		}

		var txErr error
		accessToken, txErr = s.jwtManager.GenerateAccessToken(userID, deviceID, int16(deviceType))
		if txErr != nil {
			return fmt.Errorf("generate access token: %w", txErr)
		}
		refreshToken, txErr = s.jwtManager.GenerateRefreshToken(userID, deviceID, int16(deviceType))
		if txErr != nil {
			return fmt.Errorf("generate refresh token: %w", txErr)
		}

		accessTokenExpiresAt := time.Now().Add(s.authConf.AccessTokenExpire.AsDuration())
		refreshTokenExpiresAt := time.Now().Add(s.authConf.RefreshTokenExpire.AsDuration())

		session := &model.UserSession{
			UserID:                userID,
			DeviceID:              deviceID,
			AccessToken:           accessToken,
			RefreshToken:          refreshToken,
			AccessTokenExpiresAt:  accessTokenExpiresAt,
			RefreshTokenExpiresAt: refreshTokenExpiresAt,
		}
		if txErr := sessionRepoTx.Create(ctx, session); txErr != nil {
			return fmt.Errorf("create session: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &userv1.RegisterResponse{
		UserId:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.authConf.AccessTokenExpire.AsDuration().Seconds()),
	}, nil
}

// Login user login
func (s *userServiceImpl) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	l := s.log.WithContext(ctx)
	deviceType := model.DeviceType(req.DeviceType)

	lookupAccount, err := s.normalizeAccountForLookup(req.Account)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByAccount(ctx, lookupAccount)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewBusiness(errors.CodeUserNotFound, "user not found")
	}
	if !user.IsActive() {
		return nil, errors.NewBusiness(errors.CodeAccountDisabled, "account has been disabled")
	}
	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.NewBusiness(errors.CodePasswordError, "incorrect password")
	}

	if err := s.handleSameTypeDeviceKick(ctx, user.ID, req.DeviceId, deviceType); err != nil {
		l.Warnf("failed to kick old device: %v", err)
	}

	existingDevice, err := s.deviceRepo.GetByUserIDAndDeviceID(ctx, user.ID, req.DeviceId)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if existingDevice == nil {
		device := &model.UserDevice{
			UserID:        user.ID,
			DeviceID:      req.DeviceId,
			DeviceType:    deviceType,
			ClientVersion: req.ClientVersion,
		}
		if err := s.deviceRepo.Create(ctx, device); err != nil {
			return nil, err
		}
	} else {
		if err := s.deviceRepo.UpdateLastLogin(ctx, user.ID, req.DeviceId, req.IpAddress); err != nil {
			return nil, err
		}
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, req.DeviceId, int16(deviceType))
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, req.DeviceId, int16(deviceType))
	if err != nil {
		return nil, err
	}

	accessTokenExpiresAt := time.Now().Add(s.authConf.AccessTokenExpire.AsDuration())
	refreshTokenExpiresAt := time.Now().Add(s.authConf.RefreshTokenExpire.AsDuration())

	// update or create session
	session, err := s.sessionRepo.GetByUserIDAndDeviceID(ctx, user.ID, req.DeviceId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			session = &model.UserSession{
				UserID:                user.ID,
				DeviceID:              req.DeviceId,
				AccessToken:           accessToken,
				RefreshToken:          refreshToken,
				AccessTokenExpiresAt:  accessTokenExpiresAt,
				RefreshTokenExpiresAt: refreshTokenExpiresAt,
			}
			if err := s.sessionRepo.Create(ctx, session); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		session.AccessToken = accessToken
		session.RefreshToken = refreshToken
		session.AccessTokenExpiresAt = accessTokenExpiresAt
		session.RefreshTokenExpiresAt = refreshTokenExpiresAt
		if err := s.sessionRepo.Update(ctx, session); err != nil {
			return nil, err
		}
	}

	return &userv1.LoginResponse{
		UserId:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.authConf.AccessTokenExpire.AsDuration().Seconds()),
		User: &userv1.UserInfo{
			UserId: user.ID,
			Phone:  user.Phone,
			Email:  user.Email,
		},
	}, nil
}

// Logout user logout
func (s *userServiceImpl) Logout(ctx context.Context, userID string, deviceID string) error {
	return s.sessionRepo.DeleteByUserIDAndDeviceID(ctx, userID, deviceID)
}

// RefreshToken refresh access token
func (s *userServiceImpl) RefreshToken(ctx context.Context, refreshTokenIn string) (*userv1.RefreshTokenResponse, error) {
	// validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(refreshTokenIn)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeRefreshTokenInvalid, "invalid refresh token")
	}

	// find session
	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshTokenIn)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeRefreshTokenInvalid, "invalid refresh token")
	}

	// check expiration
	if session.IsRefreshTokenExpired() {
		return nil, errors.NewBusiness(errors.CodeRefreshTokenExpired, "refresh token expired")
	}

	// generate new tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(claims.UserID, claims.DeviceID, claims.DeviceType)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(claims.UserID, claims.DeviceID, claims.DeviceType)
	if err != nil {
		return nil, err
	}

	RefreshTokenExpiresAt := time.Now().Add(s.authConf.RefreshTokenExpire.AsDuration())
	accessTokenExpiresAt := time.Now().Add(s.authConf.AccessTokenExpire.AsDuration())

	session.RefreshToken = refreshToken
	session.RefreshTokenExpiresAt = RefreshTokenExpiresAt
	session.AccessToken = accessToken
	session.AccessTokenExpiresAt = accessTokenExpiresAt
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}

	return &userv1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.authConf.AccessTokenExpire.AsDuration().Seconds()),
	}, nil
}

// ChangePassword changes the user's password
func (s *userServiceImpl) ChangePassword(ctx context.Context, userID, oldPassword, newPassword, deviceID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.NewBusiness(errors.CodeUserNotFound, "user not found")
	}

	// verify old password
	if !crypto.CheckPassword(oldPassword, user.PasswordHash) {
		return errors.NewBusiness(errors.CodePasswordError, "incorrect password")
	}

	// generate new password hash
	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// update password
	if err := s.userRepo.UpdatePassword(ctx, userID, passwordHash); err != nil {
		return err
	}

	// force logout other devices (excluding current device)
	return s.forceLogoutOtherDevices(ctx, userID, deviceID, "password_changed")
}

// ResetPassword resets the user's password using a verification code.
// The user-existence check happens before VerifyCode so that, once the code
// is consumed, the update is guaranteed to succeed — otherwise the user would
// be permanently locked out (code spent, password unchanged).
func (s *userServiceImpl) ResetPassword(ctx context.Context, account, verifyCode, newPassword string) error {
	normalized, err := s.normalizeAccountForLookup(account)
	if err != nil {
		return err
	}
	account = normalized
	targetType := model.TargetTypeSMS
	if isEmail(account) {
		targetType = model.TargetTypeEmail
	}

	// check user exists BEFORE consuming the verification code
	user, err := s.userRepo.GetByAccount(ctx, account)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeUserNotFound, "user not found")
		}
		return err
	}

	_, err = s.verifyManager.VerifyCode(ctx, &VerifyCodeRequest{
		Target:     account,
		TargetType: targetType,
		Code:       verifyCode,
		Purpose:    model.PurposeResetPassword,
	})
	if err != nil {
		return err
	}

	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	userRepoTx := s.userRepo.WithTx(s.db)
	sessionRepoTx := s.sessionRepo.WithTx(s.db)
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		userRepoTx = s.userRepo.WithTx(tx)
		sessionRepoTx = s.sessionRepo.WithTx(tx)

		if err := userRepoTx.UpdatePassword(ctx, user.ID, passwordHash); err != nil {
			return fmt.Errorf("update password: %w", err)
		}

		// invalidate all user sessions (force logout)
		if err := sessionRepoTx.DeleteByUserID(ctx, user.ID); err != nil {
			return fmt.Errorf("delete sessions: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return s.forceLogoutAllDevices(ctx, user.ID, "password_reset")
}

// forceLogoutAllDevices forces logout of all devices
func (s *userServiceImpl) forceLogoutAllDevices(ctx context.Context, userID, reason string) error {
	l := s.log.WithContext(ctx)
	devices, err := s.deviceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, device := range devices {
		if err := s.sessionRepo.DeleteByUserIDAndDeviceID(ctx, userID, device.DeviceID); err != nil {
			l.Warnw("msg", "Failed to delete session", "error", err, "deviceID", device.DeviceID)
			continue
		}

		notif := broker.NewNotification(
			broker.TypeAuthForceLogout,
			userID,
			broker.PriorityHigh,
		)
		notif.Payload = map[string]interface{}{
			"device_id":   device.DeviceID,
			"device_type": device.DeviceType.String(),
			"reason":      reason,
		}
		if err := s.broker.PublishToUser(userID, notif); err != nil {
			l.Warnw("msg", "Failed to publish force logout notification", "error", err)
		}
	}

	return nil
}

// ==================== Profile Methods ====================

// GetProfile retrieves personal profile
func (s *userServiceImpl) GetProfile(ctx context.Context, userID string, req *userv1.GetProfileRequest) (*userv1.UserProfileResponse, error) {
	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeUserProfileNotFound, "user profile not found")
		}
		return nil, err
	}

	resp := s.toUserProfileResponse(profile)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeUserNotFound, "user not found")
		}
		return nil, err
	}
	resp.Phone = user.Phone
	resp.Email = user.Email

	return resp, nil
}

// UpdateProfile updates personal profile
func (s *userServiceImpl) UpdateProfile(ctx context.Context, userID string, req *userv1.UpdateProfileRequest) (*userv1.UserProfileResponse, error) {
	// Get current profile
	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update nickname
	if req.Nickname != nil {
		nickname := *req.Nickname
		if !validator.ValidateNickname(nickname) {
			return nil, errors.NewBusiness(errors.CodeParamError, "invalid nickname format")
		}
		if validator.ContainsSensitiveWords(nickname) {
			return nil, errors.NewBusiness(errors.CodeNicknameSensitive, "nickname contains sensitive words")
		}
		// Check if nickname is already taken
		exists, err := s.profileRepo.CheckNicknameExists(ctx, nickname, userID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.NewBusiness(errors.CodeNicknameUsed, "nickname is already taken")
		}
		profile.Nickname = nickname
	}

	// Update avatar
	if req.Avatar != nil {
		profile.Avatar = *req.Avatar
	}

	// Update signature
	if req.Signature != nil {
		profile.Signature = *req.Signature
	}

	// Update gender
	if req.Gender != nil {
		if !validator.ValidateGender(int(*req.Gender)) {
			return nil, errors.NewBusiness(errors.CodeParamError, "invalid gender value")
		}
		profile.Gender = int(*req.Gender)
	}

	// Update birthday
	if req.Birthday != nil {
		birth := req.Birthday.AsTime()
		profile.Birthday = &birth
	}

	// Update region
	if req.Region != nil {
		profile.Region = *req.Region
	}

	// Save update
	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}

	return s.toUserProfileResponse(profile), nil
}

// GetUserInfo retrieves user info (query other users)
func (s *userServiceImpl) GetUserInfo(ctx context.Context, userID string, req *userv1.GetUserInfoRequest) (*userv1.UserInfoResponse, error) {
	profile, err := s.profileRepo.GetByUserID(ctx, req.TargetUserId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeUserProfileNotFound, "user profile not found")
		}
		return nil, err
	}

	isFriend := false
	isBlocked := false

	// Only query friend relationship and block status when requester exists and dependency is available

	blockedResp, err := s.friendClient.IsBlocked(ctx, &friendv1.IsBlockedRequest{
		UserId:       userID,
		TargetUserId: req.TargetUserId,
	})
	if err != nil {
		return nil, err
	}
	isBlocked = blockedResp.IsBlocked
	if isBlocked {
		return nil, errors.NewBusiness(errors.CodePermissionDenied, "you have been blocked by this user")
	}

	friendResp, err := s.friendClient.IsFriend(ctx, &friendv1.IsFriendRequest{
		UserId:       userID,
		TargetUserId: req.TargetUserId,
	})
	if err != nil {
		return nil, err
	}
	isFriend = friendResp.IsFriend

	return &userv1.UserInfoResponse{
		UserId:    profile.UserID,
		Nickname:  profile.Nickname,
		Avatar:    profile.Avatar,
		Signature: profile.Signature,
		Gender:    int32(profile.Gender),
		Region:    profile.Region,
		IsFriend:  isFriend,
		IsBlocked: isBlocked,
	}, nil
}

// SearchUsers searches for users
func (s *userServiceImpl) SearchUsers(ctx context.Context, req *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	profiles, total, err := s.profileRepo.SearchByKeyword(ctx, req.Keyword, pageSize, offset)
	if err != nil {
		return nil, err
	}

	users := make([]*userv1.UserBriefInfo, 0, len(profiles))
	for _, p := range profiles {
		users = append(users, &userv1.UserBriefInfo{
			UserId:    p.UserID,
			Nickname:  p.Nickname,
			Avatar:    p.Avatar,
			Signature: p.Signature,
		})
	}

	return &userv1.SearchUsersResponse{
		Total: total,
		Users: users,
	}, nil
}

// ==================== Settings Methods ====================

// GetSettings retrieves personal settings
func (s *userServiceImpl) GetSettings(ctx context.Context, userID string, req *userv1.GetSettingsRequest) (*userv1.UserSettingsResponse, error) {
	settings, err := s.settingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeUserProfileNotFound, "user profile not found")
		}
		return nil, err
	}

	return &userv1.UserSettingsResponse{
		UserId:                settings.UserID,
		NotificationEnabled:   settings.NotificationEnabled,
		SoundEnabled:          settings.SoundEnabled,
		VibrationEnabled:      settings.VibrationEnabled,
		MessagePreviewEnabled: settings.MessagePreviewEnabled,
		FriendVerifyRequired:  settings.FriendVerifyRequired,
		SearchByPhone:         settings.SearchByPhone,
		SearchById:            settings.SearchByID,
		Language:              settings.Language,
	}, nil
}

// UpdateSettings updates personal settings
func (s *userServiceImpl) UpdateSettings(ctx context.Context, userID string, req *userv1.UpdateSettingsRequest) (*userv1.UserSettingsResponse, error) {
	settings, err := s.settingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.NotificationEnabled != nil {
		settings.NotificationEnabled = *req.NotificationEnabled
	}
	if req.SoundEnabled != nil {
		settings.SoundEnabled = *req.SoundEnabled
	}
	if req.VibrationEnabled != nil {
		settings.VibrationEnabled = *req.VibrationEnabled
	}
	if req.MessagePreviewEnabled != nil {
		settings.MessagePreviewEnabled = *req.MessagePreviewEnabled
	}
	if req.FriendVerifyRequired != nil {
		settings.FriendVerifyRequired = *req.FriendVerifyRequired
	}
	if req.SearchByPhone != nil {
		settings.SearchByPhone = *req.SearchByPhone
	}
	if req.SearchById != nil {
		settings.SearchByID = *req.SearchById
	}
	if req.Language != nil {
		settings.Language = *req.Language
	}

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}

	return &userv1.UserSettingsResponse{
		UserId:                settings.UserID,
		NotificationEnabled:   settings.NotificationEnabled,
		SoundEnabled:          settings.SoundEnabled,
		VibrationEnabled:      settings.VibrationEnabled,
		MessagePreviewEnabled: settings.MessagePreviewEnabled,
		FriendVerifyRequired:  settings.FriendVerifyRequired,
		SearchByPhone:         settings.SearchByPhone,
		SearchById:            settings.SearchByID,
		Language:              settings.Language,
	}, nil
}

// ==================== QR Code Methods ====================

// RefreshQRCode refreshes QR code
func (s *userServiceImpl) RefreshQRCode(ctx context.Context, userID string) (*userv1.QRCodeResponse, error) {
	// Generate QR code token
	qrcodeToken, err := crypto.GenerateQRCodeToken(userID)
	if err != nil {
		return nil, err
	}

	// Generate QR code URL
	qrcodeURL := fmt.Sprintf("anychat://qrcode?token=%s", qrcodeToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Save QR code record
	qrcode := &model.UserQRCode{
		UserID:      userID,
		QRCodeToken: qrcodeToken,
		QRCodeURL:   qrcodeURL,
		ExpiresAt:   expiresAt,
	}
	if err := s.qrcodeRepo.Create(ctx, qrcode); err != nil {
		return nil, err
	}

	// Update QR code URL in user profile
	if err := s.profileRepo.UpdateQRCode(ctx, userID, qrcodeURL); err != nil {
		return nil, err
	}

	return &userv1.QRCodeResponse{
		QrcodeUrl: qrcodeURL,
		ExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

// GetUserByQRCode retrieves user info by QR code
func (s *userServiceImpl) GetUserByQRCode(ctx context.Context, code string) (*userv1.UserInfoResponse, error) {
	// Find QR code
	qrcode, err := s.qrcodeRepo.GetByToken(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeQRCodeInvalid, "invalid QR code")
		}
		return nil, err
	}

	// Check if expired
	if qrcode.IsExpired() {
		return nil, errors.NewBusiness(errors.CodeQRCodeExpired, "QR code has expired")
	}

	// Get user profile
	profile, err := s.profileRepo.GetByUserID(ctx, qrcode.UserID)
	if err != nil {
		return nil, err
	}

	return &userv1.UserInfoResponse{
		UserId:    profile.UserID,
		Nickname:  profile.Nickname,
		Avatar:    profile.Avatar,
		Signature: profile.Signature,
	}, nil
}

// ==================== Push Token ====================

// UpdatePushToken updates push token
func (s *userServiceImpl) UpdatePushToken(ctx context.Context, userID string, req *userv1.UpdatePushTokenRequest) error {
	token := &model.UserPushToken{
		UserID:    userID,
		DeviceID:  req.DeviceId,
		PushToken: req.PushToken,
		Platform:  model.PushPlatform(req.Platform),
	}
	return s.pushTokenRepo.CreateOrUpdate(ctx, token)
}

// ==================== Account Binding ====================

// BindPhone binds phone number
func (s *userServiceImpl) BindPhone(ctx context.Context, userID string, req *userv1.BindPhoneRequest) (*userv1.BindPhoneResponse, error) {
	e164, err := validator.NormalizePhone(req.PhoneNumber)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodePhoneFormatInvalid, "invalid phone number format")
	}

	user, err := s.requireAuthUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Phone != nil {
		if *user.Phone == e164 {
			return &userv1.BindPhoneResponse{
				PhoneNumber: maskPhone(e164),
				IsPrimary:   true,
			}, nil
		}
		return nil, errors.NewBusiness(errors.CodeParamError, "phone already bound, please use change phone API")
	}

	if err := s.ensurePhoneAvailable(ctx, e164, userID); err != nil {
		return nil, err
	}
	if err := s.verifyCode(ctx, e164, model.TargetTypeSMS, model.PurposeBindPhone, req.VerifyCode); err != nil {
		return nil, err
	}
	if err := s.userRepo.UpdatePhone(ctx, userID, &e164); err != nil {
		return nil, err
	}

	return &userv1.BindPhoneResponse{
		PhoneNumber: maskPhone(e164),
		IsPrimary:   true,
	}, nil
}

// ChangePhone changes phone number
func (s *userServiceImpl) ChangePhone(ctx context.Context, userID string, req *userv1.ChangePhoneRequest) (*userv1.ChangePhoneResponse, error) {
	newE164, err := validator.NormalizePhone(req.NewPhoneNumber)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodePhoneFormatInvalid, "invalid phone number format")
	}
	oldE164, err := validator.NormalizePhone(req.OldPhoneNumber)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodePhoneFormatInvalid, "invalid phone number format")
	}

	user, err := s.requireAuthUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Phone == nil || *user.Phone != oldE164 {
		return nil, errors.NewBusiness(errors.CodeOldPhoneNotMatch, "old phone number does not match")
	}
	if newE164 == oldE164 {
		return nil, errors.NewBusiness(errors.CodeParamError, "new phone number cannot be the same as old phone number")
	}

	if req.OldVerifyCode != nil && *req.OldVerifyCode != "" {
		if err := s.verifyCode(ctx, oldE164, model.TargetTypeSMS, model.PurposeChangePhone, *req.OldVerifyCode); err != nil {
			return nil, err
		}
	}
	if err := s.ensurePhoneAvailable(ctx, newE164, userID); err != nil {
		return nil, err
	}
	if err := s.verifyCode(ctx, newE164, model.TargetTypeSMS, model.PurposeChangePhone, req.NewVerifyCode); err != nil {
		return nil, err
	}
	if err := s.userRepo.UpdatePhone(ctx, userID, &newE164); err != nil {
		return nil, err
	}
	if err := s.invalidateSessionsAfterContactChange(ctx, userID, req.DeviceId); err != nil {
		return nil, err
	}

	return &userv1.ChangePhoneResponse{
		OldPhoneNumber: maskPhone(oldE164),
		NewPhoneNumber: maskPhone(newE164),
	}, nil
}

// BindEmail binds email
func (s *userServiceImpl) BindEmail(ctx context.Context, userID string, req *userv1.BindEmailRequest) (*userv1.BindEmailResponse, error) {
	if !validator.ValidateEmail(req.Email) {
		return nil, errors.NewBusiness(errors.CodeEmailFormatInvalid, "invalid email format")
	}

	user, err := s.requireAuthUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Email != nil {
		if *user.Email == req.Email {
			return &userv1.BindEmailResponse{
				Email:     maskEmail(req.Email),
				IsPrimary: true,
			}, nil
		}
		return nil, errors.NewBusiness(errors.CodeParamError, "email already bound, please use change email API")
	}

	if err := s.ensureEmailAvailable(ctx, req.Email, userID); err != nil {
		return nil, err
	}
	if err := s.verifyCode(ctx, req.Email, model.TargetTypeEmail, model.PurposeBindEmail, req.VerifyCode); err != nil {
		return nil, err
	}
	if err := s.userRepo.UpdateEmail(ctx, userID, &req.Email); err != nil {
		return nil, err
	}

	return &userv1.BindEmailResponse{
		Email:     maskEmail(req.Email),
		IsPrimary: true,
	}, nil
}

// ChangeEmail changes email
func (s *userServiceImpl) ChangeEmail(ctx context.Context, userID string, req *userv1.ChangeEmailRequest) (*userv1.ChangeEmailResponse, error) {
	if !validator.ValidateEmail(req.NewEmail) {
		return nil, errors.NewBusiness(errors.CodeEmailFormatInvalid, "invalid email format")
	}

	user, err := s.requireAuthUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Email == nil || *user.Email != req.OldEmail {
		return nil, errors.NewBusiness(errors.CodeOldEmailNotMatch, "old email does not match")
	}
	if req.NewEmail == req.OldEmail {
		return nil, errors.NewBusiness(errors.CodeParamError, "new email cannot be the same as old email")
	}

	if req.OldVerifyCode != nil && *req.OldVerifyCode != "" {
		if err := s.verifyCode(ctx, req.OldEmail, model.TargetTypeEmail, model.PurposeChangeEmail, *req.OldVerifyCode); err != nil {
			return nil, err
		}
	}
	if err := s.ensureEmailAvailable(ctx, req.NewEmail, userID); err != nil {
		return nil, err
	}
	if err := s.verifyCode(ctx, req.NewEmail, model.TargetTypeEmail, model.PurposeChangeEmail, req.NewVerifyCode); err != nil {
		return nil, err
	}
	if err := s.userRepo.UpdateEmail(ctx, userID, &req.NewEmail); err != nil {
		return nil, err
	}
	if err := s.invalidateSessionsAfterContactChange(ctx, userID, req.DeviceId); err != nil {
		return nil, err
	}

	return &userv1.ChangeEmailResponse{
		OldEmail: maskEmail(req.OldEmail),
		NewEmail: req.NewEmail,
	}, nil
}

// ==================== Helper Methods ====================

func (s *userServiceImpl) toUserProfileResponse(profile *model.UserProfile) *userv1.UserProfileResponse {
	resp := &userv1.UserProfileResponse{
		UserId:    profile.UserID,
		Nickname:  profile.Nickname,
		Avatar:    profile.Avatar,
		Signature: profile.Signature,
		Gender:    int32(profile.Gender),
		Region:    profile.Region,
		QrcodeUrl: profile.QRCodeURL,
		CreatedAt: timestamppb.New(profile.CreatedAt),
	}
	if profile.Birthday != nil {
		resp.Birthday = timestamppb.New(*profile.Birthday)
	}
	return resp
}

func (s *userServiceImpl) requireAuthUser(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeUserNotFound, "user not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *userServiceImpl) ensurePhoneAvailable(ctx context.Context, phone, excludeUserID string) error {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err == nil && user.ID != excludeUserID {
		return errors.NewBusiness(errors.CodePhoneAlreadyBound, "phone number is already bound to another account")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func (s *userServiceImpl) ensureEmailAvailable(ctx context.Context, email, excludeUserID string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && user.ID != excludeUserID {
		return errors.NewBusiness(errors.CodeEmailAlreadyBound, "email is already bound to another account")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func (s *userServiceImpl) verifyCode(ctx context.Context, target string, targetType model.VerificationTargetType, purpose model.VerificationPurpose, code string) error {
	resp, err := s.verifyManager.VerifyCode(ctx, &VerifyCodeRequest{
		Target:     target,
		TargetType: targetType,
		Purpose:    purpose,
		Code:       code,
	})
	if err != nil {
		return err
	}
	if !resp.Valid {
		return errors.NewBusiness(errors.CodeVerifyCodeError, "invalid verification code")
	}
	return nil
}

func (s *userServiceImpl) invalidateSessionsAfterContactChange(ctx context.Context, userID, deviceID string) error {
	return s.sessionRepo.DeleteByUserIDExceptDeviceID(ctx, userID, deviceID)
}

func (s *userServiceImpl) handleSameTypeDeviceKick(ctx context.Context, userID, deviceID string, deviceType model.DeviceType) error {
	l := s.log.WithContext(ctx)
	devices, err := s.deviceRepo.GetByUserIDAndDeviceType(ctx, userID, deviceType)
	if err != nil {
		return err
	}
	for _, device := range devices {
		if device.DeviceID == deviceID {
			continue
		}
		if err := s.sessionRepo.DeleteByUserIDAndDeviceID(ctx, userID, device.DeviceID); err != nil {
			l.Warnw("msg", "Failed to delete old session", "error", err, "deviceID", device.DeviceID)
		}

		notif := broker.NewNotification(
			broker.TypeAuthForceLogout,
			userID,
			broker.PriorityHigh,
		)
		notif.Payload = map[string]interface{}{
			"device_id":   device.DeviceID,
			"device_type": device.DeviceType.String(),
			"reason":      "new_device_login",
		}
		if err := s.broker.PublishToUser(userID, notif); err != nil {
			l.Warnw("msg", "Failed to publish force logout notification", "error", err)
		}
	}
	return nil
}

// forceLogoutOtherDevices forces logout of other devices
func (s *userServiceImpl) forceLogoutOtherDevices(ctx context.Context, userID, excludeDeviceID, reason string) error {
	l := s.log.WithContext(ctx)
	devices, err := s.deviceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, device := range devices {
		if device.DeviceID == excludeDeviceID {
			continue
		}

		if err := s.sessionRepo.DeleteByUserIDAndDeviceID(ctx, userID, device.DeviceID); err != nil {
			l.Warnw("msg", "Failed to delete session", "error", err, "deviceID", device.DeviceID)
			continue
		}

		notif := broker.NewNotification(
			broker.TypeAuthForceLogout,
			userID,
			broker.PriorityHigh,
		)
		notif.Payload = map[string]interface{}{
			"device_id":   device.DeviceID,
			"device_type": device.DeviceType.String(),
			"reason":      reason,
		}
		if err := s.broker.PublishToUser(userID, notif); err != nil {
			l.Warnw("msg", "Failed to publish force logout notification", "error", err)
		}
	}

	return nil
}

func isEmail(account string) bool {
	return strings.Contains(account, "@")
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return "***"
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func maskEmail(email string) string {
	at := 0
	for i := 0; i < len(email); i++ {
		if email[i] == '@' {
			at = i
			break
		}
	}
	if at <= 0 || at == len(email)-1 {
		return "***"
	}
	name := email[:at]
	domain := email[at+1:]
	if len(name) <= 2 {
		return "***@" + domain
	}
	return name[:2] + "***@" + domain
}

// toPtr returns a pointer to the given string value
func toPtr(s string) *string {
	return &s
}

// normalizeAccountForLookup normalises an "account" value for use as a lookup
// key. Email is lowercased; phone is normalised to canonical E164 form so that
// a bare 11-digit China number ("13800138000") matches the stored E164 value
// ("+8613800138000"). This keeps the read path consistent with the write path.
func (s *userServiceImpl) normalizeAccountForLookup(account string) (string, error) {
	account = strings.TrimSpace(account)
	if isEmail(account) {
		return strings.ToLower(account), nil
	}
	e164, err := validator.NormalizePhone(account)
	if err != nil {
		return "", errors.NewBusiness(errors.CodeParamError, "invalid account format")
	}
	return e164, nil
}
