package repository

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewUserProfileRepository,
	NewUserPushTokenRepository,
	NewUserQRCodeRepository,
	NewUserSettingsRepository,
	NewUserDeviceRepository,
	NewUserSessionRepository,
	NewUserRepository,
	NewVerificationCodeRepository,
	NewVerificationTemplateRepository,
)
