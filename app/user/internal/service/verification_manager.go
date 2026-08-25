package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"flamingo/app/user/internal/model"

	"flamingo/pkg/cache"
	confpkg "flamingo/pkg/config"
	"flamingo/pkg/errors"
	"flamingo/pkg/sender"
	"flamingo/pkg/validator"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// VerificationManager defines the methods User Service depends on from
// VerificationManager. Defined here so it can be mocked in tests.
type VerificationManager interface {
	SendCode(ctx context.Context, req *SendCodeRequest, ipAddress string) (*SendCodeResponse, error)
	VerifyCode(ctx context.Context, req *VerifyCodeRequest) (*VerifyCodeResponse, error)
	CheckCodeStatus(ctx context.Context, codeID string) (*CheckCodeStatusResponse, error)
}

// compile-time checks for interface satisfaction
var _ VerificationManager = (*verificationManagerImpl)(nil)

type verificationManagerImpl struct {
	conf         confpkg.Verify
	env          confpkg.Environment
	codeRepo     VerificationCodeRepository
	templateRepo VerificationTemplateRepository
	cache        cache.Cache
	smsSender    sender.SMSSender
	emailSender  sender.EmailSender
	log          *log.Helper
}

func NewVerificationManager(
	logger log.Logger,
	conf confpkg.Verify,
	env confpkg.Environment,
	codeRepo VerificationCodeRepository,
	templateRepo VerificationTemplateRepository,
	cache cache.Cache,
	smsSender sender.SMSSender,
	emailSender sender.EmailSender,
) VerificationManager {
	if conf.Code.Length == 0 {
		conf.Code.Length = 6
	}
	if conf.Code.ExpireSeconds == 0 {
		conf.Code.ExpireSeconds = 300
	}
	if conf.Code.MaxAttempts == 0 {
		conf.Code.MaxAttempts = 5
	}
	if conf.RateLimit.TargetPerMinute == 0 {
		conf.RateLimit.TargetPerMinute = 1
	}
	if conf.RateLimit.TargetPerDay == 0 {
		conf.RateLimit.TargetPerDay = 10
	}
	if conf.RateLimit.IPPerHour == 0 {
		conf.RateLimit.IPPerHour = 200
	}
	if conf.RateLimit.DevicePerDay == 0 {
		conf.RateLimit.DevicePerDay = 100
	}

	// Production safety checks — fail fast rather than silently degrade.
	if env.IsProd() {
		if conf.Code.HashSecret == "" {
			panic("verification: hash_secret must be configured in production")
		}
		if conf.Code.DebugFixedCode != "" || conf.Code.AllowDevBypass {
			panic("verification: DebugFixedCode / AllowDevBypass must not be enabled in production")
		}
		if conf.SMS.Provider == "" {
			panic("verification: SMS provider must be configured in production (noop sender will not deliver codes)")
		}
	}

	return &verificationManagerImpl{
		conf:         conf,
		env:          env,
		codeRepo:     codeRepo,
		templateRepo: templateRepo,
		cache:        cache,
		smsSender:    smsSender,
		emailSender:  emailSender,
		log:          log.NewHelper(logger),
	}
}

func (s *verificationManagerImpl) SendCode(ctx context.Context, req *SendCodeRequest, ipAddress string) (*SendCodeResponse, error) {
	l := s.log.WithContext(ctx)
	target, err := s.validateAndNormalizeTarget(req.Target, req.TargetType)
	if err != nil {
		return nil, err
	}
	if err := s.validatePurpose(req.Purpose); err != nil {
		return nil, err
	}
	if s.cache == nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "verification cache is not configured")
	}

	rollbackKeys, err := s.applyRateLimits(ctx, target, req.TargetType, req.Purpose, req.DeviceID, ipAddress)
	if err != nil {
		return nil, err
	}

	if err := s.cancelPreviousCode(ctx, target, req.TargetType, req.Purpose); err != nil {
		s.rollbackRateLimits(ctx, rollbackKeys)
		return nil, err
	}

	code := s.generateCode()
	codeID := fmt.Sprintf("vc_%d_%s", time.Now().UnixNano(), randString(8))
	expiresAt := time.Now().Add(time.Duration(s.conf.Code.ExpireSeconds) * time.Second)
	codeHash := s.hashCode(req.Purpose, target, code)

	cacheKey := s.codeCacheKey(req.Purpose, target)
	if err := s.cache.HSet(ctx, cacheKey,
		"code_id", codeID,
		"code_hash", codeHash,
		"target_type", req.TargetType.String(),
		"expires_at", expiresAt.UTC().Format(time.RFC3339),
		"attempts", "0",
		"max_attempts", fmt.Sprintf("%d", s.conf.Code.MaxAttempts),
		"device_id", req.DeviceID,
	); err != nil {
		s.rollbackRateLimits(ctx, rollbackKeys)
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to store verification code")
	}
	if err := s.cache.Expire(ctx, cacheKey, time.Duration(s.conf.Code.ExpireSeconds)*time.Second); err != nil {
		_ = s.cache.Del(ctx, cacheKey)
		s.rollbackRateLimits(ctx, rollbackKeys)
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to set verification code expiry")
	}

	record := &model.VerificationCode{
		CodeID:       codeID,
		Target:       target,
		TargetType:   req.TargetType,
		CodeHash:     codeHash,
		Purpose:      req.Purpose,
		ExpiresAt:    expiresAt,
		Status:       model.CodeStatusPending,
		SendIP:       ipAddress,
		SendDeviceID: req.DeviceID,
	}
	if err := s.codeRepo.Create(ctx, record); err != nil {
		_ = s.cache.Del(ctx, cacheKey)
		s.rollbackRateLimits(ctx, rollbackKeys)
		l.Errorw("msg", "failed to create verification record", "error", err)
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to create verification record")
	}

	if err := s.dispatchCode(ctx, target, req.TargetType, req.Purpose, code); err != nil {
		_ = s.cache.Del(ctx, cacheKey)
		s.rollbackRateLimits(ctx, rollbackKeys)
		_ = s.codeRepo.UpdateStatus(ctx, codeID, model.CodeStatusCancelled)
		return nil, err
	}

	return &SendCodeResponse{
		CodeID:    codeID,
		ExpiresIn: int64(s.conf.Code.ExpireSeconds),
		Sent:      true,
		Message:   "verification code sent",
	}, nil
}

func (s *verificationManagerImpl) VerifyCode(ctx context.Context, req *VerifyCodeRequest) (*VerifyCodeResponse, error) {
	l := s.log.WithContext(ctx)
	target, err := s.validateAndNormalizeTarget(req.Target, req.TargetType)
	if err != nil {
		return nil, err
	}
	if err := s.validatePurpose(req.Purpose); err != nil {
		return nil, err
	}
	if s.cache == nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "verification cache is not configured")
	}

	cacheKey := s.codeCacheKey(req.Purpose, target)
	fields, err := s.cache.HGetAll(ctx, cacheKey)
	if err != nil {
		l.Errorw("msg", "failed to load verification code from cache", "error", err)
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to load verification code")
	}

	if len(fields) == 0 {
		if s.shouldAllowDevBypass(req.Code) {
			l.Warnw("msg", "Dev bypass verification code accepted",
				"target", maskTarget(target, req.TargetType),
				"targetType", req.TargetType.String(),
				"purpose", req.Purpose.String(),
			)
			return &VerifyCodeResponse{
				Valid:   true,
				CodeID:  "dev-bypass",
				Message: "verification successful",
			}, nil
		}
		return nil, s.resolveMissingCodeError(ctx, target, req.TargetType, req.Purpose)
	}

	codeID := fields["code_id"]
	if fields["target_type"] != req.TargetType.String() {
		return nil, errors.NewBusiness(errors.CodeVerifyCodeNotFound, "verification code not found for this target")
	}

	expiresAt, err := time.Parse(time.RFC3339, fields["expires_at"])
	if err != nil {
		l.Errorw("msg", "invalid verification expiry in cache", "codeID", codeID, "error", err)
		return nil, errors.NewBusiness(errors.CodeInternalError, "invalid verification code expiry")
	}
	if time.Now().After(expiresAt) {
		_ = s.cache.Del(ctx, cacheKey)
		_ = s.codeRepo.UpdateStatus(ctx, codeID, model.CodeStatusExpired)
		return nil, errors.NewBusiness(errors.CodeVerifyCodeExpired, "verification code has expired")
	}

	if subtle.ConstantTimeCompare(
		[]byte(s.hashCode(req.Purpose, target, req.Code)),
		[]byte(fields["code_hash"]),
	) != 1 {
		attempts, incrErr := s.cache.HIncrBy(ctx, cacheKey, "attempts", 1)
		if incrErr != nil {
			l.Errorw("msg", "failed to increment verification attempts", "codeID", codeID, "error", incrErr)
			return nil, errors.NewBusiness(errors.CodeInternalError, "failed to track verification attempt")
		}
		_ = s.codeRepo.IncrementAttempts(ctx, codeID)

		maxAttempts := int64(s.conf.Code.MaxAttempts)
		if fields["max_attempts"] != "" {
			if parsed, parseErr := parseInt64(fields["max_attempts"]); parseErr == nil {
				maxAttempts = parsed
			}
		}
		if attempts >= maxAttempts {
			_ = s.cache.Del(ctx, cacheKey)
			_ = s.codeRepo.UpdateStatus(ctx, codeID, model.CodeStatusLocked)
			return nil, errors.NewBusiness(errors.CodeVerifyAttemptsExceeded, "too many verification attempts, code locked")
		}

		return nil, errors.NewBusiness(errors.CodeVerifyCodeError, "incorrect verification code")
	}

	now := time.Now()
	if err := s.codeRepo.UpdateVerifiedAt(ctx, codeID, now); err != nil {
		l.Errorw("msg", "failed to mark verification code as used", "codeID", codeID, "error", err)
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to mark verification code as used")
	}
	_ = s.cache.Del(ctx, cacheKey)

	return &VerifyCodeResponse{
		Valid:   true,
		CodeID:  codeID,
		Message: "verification successful",
	}, nil
}

func (s *verificationManagerImpl) CheckCodeStatus(ctx context.Context, codeID string) (*CheckCodeStatusResponse, error) {
	code, err := s.codeRepo.GetByCodeID(ctx, codeID)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeVerifyCodeNotFound, "verification code not found")
	}

	return &CheckCodeStatusResponse{
		Status:    code.Status,
		ExpiresAt: code.ExpiresAt,
	}, nil
}

func (s *verificationManagerImpl) dispatchCode(ctx context.Context, target string, targetType model.VerificationTargetType, purpose model.VerificationPurpose, code string) error {
	l := s.log.WithContext(ctx)
	if !s.isReleaseMode() {
		l.Infow("msg", "verification code generated for local environment", "target", maskTarget(target, targetType), "targetType", targetType.String(), "purpose", purpose.String(), "code", maskCode(code))
	}

	templateID := ""
	emailSubject := "AnyChat Verification Code"
	emailContent := fmt.Sprintf("Your verification code is: %s, valid for %d minutes.", code, s.conf.Code.ExpireSeconds/60)
	if s.templateRepo != nil {
		template, err := s.templateRepo.GetByPurpose(ctx, purpose)
		if err == nil {
			templateID = template.SMSTemplateID
			if template.EmailSubject != "" {
				emailSubject = template.EmailSubject
			}
			if template.EmailContent != "" {
				emailContent = strings.ReplaceAll(template.EmailContent, "{code}", code)
			}
		}
	}

	switch targetType {
	case model.TargetTypeSMS:
		if s.smsSender == nil {
			if !s.isReleaseMode() {
				return nil
			}
			return errors.NewBusiness(errors.CodeSMSServiceError, "SMS service is not configured")
		}
		if err := s.smsSender.Send(target, templateID, code); err != nil {
			l.Errorw("msg", "failed to send sms verification code", "error", err)
			return errors.NewBusiness(errors.CodeSMSServiceError, "failed to send SMS verification code")
		}
	case model.TargetTypeEmail:
		if s.emailSender == nil {
			if !s.isReleaseMode() {
				return nil
			}
			return errors.NewBusiness(errors.CodeEmailServiceError, "email service is not configured")
		}
		if err := s.emailSender.Send(target, emailSubject, emailContent); err != nil {
			l.Errorw("msg", "failed to send email verification code", "error", err)
			return errors.NewBusiness(errors.CodeEmailServiceError, "failed to send email verification code")
		}
	default:
		return errors.NewBusiness(errors.CodeTargetFormatInvalid, "unsupported verification target type")
	}

	return nil
}

func (s *verificationManagerImpl) cancelPreviousCode(ctx context.Context, target string, targetType model.VerificationTargetType, purpose model.VerificationPurpose) error {
	l := s.log.WithContext(ctx)
	code, err := s.codeRepo.GetLatestByTarget(ctx, target, targetType, purpose)
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		l.Errorw("msg", "failed to query previous verification code", "error", err)
		return errors.NewBusiness(errors.CodeInternalError, "failed to query previous verification code")
	}
	if code.Status == model.CodeStatusPending {
		if err := s.codeRepo.UpdateStatus(ctx, code.CodeID, model.CodeStatusCancelled); err != nil {
			l.Errorw("msg", "failed to cancel previous verification code", "codeID", code.CodeID, "error", err)
			return errors.NewBusiness(errors.CodeInternalError, "failed to cancel previous verification code")
		}
	}
	return nil
}

func (s *verificationManagerImpl) applyRateLimits(ctx context.Context, target string, targetType model.VerificationTargetType, purpose model.VerificationPurpose, deviceID, ipAddress string) ([]string, error) {
	targetHash := s.targetHash(target)
	keys := make([]string, 0, 4)

	targetMinuteKey := fmt.Sprintf("auth:vc:rl:target:%s:%s:1m", purpose.String(), targetHash)
	if err := s.incrementAndCheck(ctx, targetMinuteKey, time.Minute, s.conf.RateLimit.TargetPerMinute, errors.CodeSendRateLimited); err != nil {
		return nil, err
	}
	keys = append(keys, targetMinuteKey)

	targetDayKey := fmt.Sprintf("auth:vc:rl:target:%s:%s:24h", purpose.String(), targetHash)
	if err := s.incrementAndCheck(ctx, targetDayKey, 24*time.Hour, s.conf.RateLimit.TargetPerDay, errors.CodeSendLimitReached); err != nil {
		s.rollbackRateLimits(ctx, keys)
		return nil, err
	}
	keys = append(keys, targetDayKey)

	if ipAddress != "" {
		ipKey := fmt.Sprintf("auth:vc:rl:ip:%s:1h", ipAddress)
		if err := s.incrementAndCheck(ctx, ipKey, time.Hour, s.conf.RateLimit.IPPerHour, errors.CodeSendRateLimited); err != nil {
			s.rollbackRateLimits(ctx, keys)
			return nil, err
		}
		keys = append(keys, ipKey)
	}

	if deviceID != "" {
		deviceKey := fmt.Sprintf("auth:vc:rl:device:%s:24h", deviceID)
		if err := s.incrementAndCheck(ctx, deviceKey, 24*time.Hour, s.conf.RateLimit.DevicePerDay, errors.CodeSendLimitReached); err != nil {
			s.rollbackRateLimits(ctx, keys)
			return nil, err
		}
		keys = append(keys, deviceKey)
	}

	return keys, nil
}

func (s *verificationManagerImpl) rollbackRateLimits(ctx context.Context, keys []string) {
	l := s.log.WithContext(ctx)
	for _, key := range keys {
		if err := s.cache.Del(ctx, key); err != nil {
			l.Warnw("msg", "failed to rollback verification rate limit counter", "key", key, "error", err)
		}
	}
}

// incrementAndCheck atomically increments a rate-limit counter with a TTL.
// Uses SetNX to initialize (sets count=1 with TTL on first creation) and
// INCR for subsequent increments. On limit breach the key is fully removed
// (Del) rather than Decr'd, which would leave a zero-count key lingering.
func (s *verificationManagerImpl) incrementAndCheck(ctx context.Context, key string, ttl time.Duration, limit int, errorCode int) error {
	l := s.log.WithContext(ctx)

	// Initialize the counter if it does not exist.
	ok, err := s.cache.SetNX(ctx, key, 1, ttl)
	if err != nil {
		l.Errorw("msg", "failed to initialize verification rate limit counter", "key", key, "error", err)
		return errors.NewBusiness(errors.CodeInternalError, "failed to initialize rate limit counter")
	}
	if ok {
		if limit > 0 && 1 > limit {
			_ = s.cache.Del(ctx, key)
			return errors.NewBusiness(errorCode, "verification rate limit exceeded")
		}
		return nil
	}

	// Counter already exists — increment.
	current, err := s.cache.Incr(ctx, key)
	if err != nil {
		l.Errorw("msg", "failed to increment verification rate limit counter", "key", key, "error", err)
		return errors.NewBusiness(errors.CodeInternalError, "failed to increment rate limit counter")
	}
	if limit > 0 && current > int64(limit) {
		_ = s.cache.Del(ctx, key)
		return errors.NewBusiness(errorCode, "verification rate limit exceeded")
	}
	return nil
}

func (s *verificationManagerImpl) resolveMissingCodeError(ctx context.Context, target string, targetType model.VerificationTargetType, purpose model.VerificationPurpose) error {
	l := s.log.WithContext(ctx)
	code, err := s.codeRepo.GetLatestByTarget(ctx, target, targetType, purpose)
	if err == gorm.ErrRecordNotFound {
		return errors.NewBusiness(errors.CodeVerifyCodeNotFound, "verification code not found")
	}
	if err != nil {
		l.Errorw("msg", "failed to load verification audit record", "error", err)
		return errors.NewBusiness(errors.CodeInternalError, "failed to load verification audit record")
	}

	switch code.Status {
	case model.CodeStatusVerified:
		return errors.NewBusiness(errors.CodeVerifyCodeAlreadyUsed, "verification code has already been used")
	case model.CodeStatusExpired:
		return errors.NewBusiness(errors.CodeVerifyCodeExpired, "verification code has expired")
	case model.CodeStatusLocked:
		return errors.NewBusiness(errors.CodeVerifyAttemptsExceeded, "verification attempts exceeded, code locked")
	default:
		return errors.NewBusiness(errors.CodeVerifyCodeNotFound, "verification code not found")
	}
}

func (s *verificationManagerImpl) validateAndNormalizeTarget(target string, targetType model.VerificationTargetType) (string, error) {
	normalized := strings.TrimSpace(target)
	switch targetType {
	case model.TargetTypeSMS:
		// Normalize to canonical E164 form so every representation of the
		// same number (+86-1381234, 1381234, 861381234) collapses into the
		// same rate-limit / code-lookup bucket.
		e164, err := validator.NormalizePhone(normalized)
		if err != nil {
			return "", errors.NewBusiness(errors.CodeTargetFormatInvalid, "invalid phone number format")
		}
		return e164, nil
	case model.TargetTypeEmail:
		normalized = strings.ToLower(normalized)
		if !validator.ValidateEmail(normalized) {
			return "", errors.NewBusiness(errors.CodeTargetFormatInvalid, "invalid email format")
		}
	default:
		return "", errors.NewBusiness(errors.CodeTargetFormatInvalid, "invalid target type")
	}
	return normalized, nil
}

func (s *verificationManagerImpl) validatePurpose(purpose model.VerificationPurpose) error {
	if purpose.IsValid() {
		return nil
	}
	return errors.NewBusiness(errors.CodeParamError, "verification purpose not supported")
}

func (s *verificationManagerImpl) codeCacheKey(purpose model.VerificationPurpose, target string) string {
	return fmt.Sprintf("auth:vc:%s:%s", purpose.String(), s.targetHash(target))
}

func (s *verificationManagerImpl) targetHash(target string) string {
	sum := sha256.Sum256([]byte(target))
	return hex.EncodeToString(sum[:])
}

func (s *verificationManagerImpl) hashCode(purpose model.VerificationPurpose, target, code string) string {
	secret := s.conf.Code.HashSecret
	if secret == "" {
		secret = "anychat-verification-secret"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strings.Join([]string{purpose.String(), target, code}, ":")))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *verificationManagerImpl) generateCode() string {
	if s.shouldUseDebugCode() {
		return s.conf.Code.DebugFixedCode
	}

	const codeChars = "0123456789"
	return randomString(codeChars, s.conf.Code.Length, s.log)
}

func (s *verificationManagerImpl) shouldUseDebugCode() bool {
	return !s.isReleaseMode() && s.conf.Code.DebugFixedCode != ""
}

func (s *verificationManagerImpl) shouldAllowDevBypass(code string) bool {
	return !s.isReleaseMode() && s.conf.Code.AllowDevBypass && s.conf.Code.DebugFixedCode != "" && code == s.conf.Code.DebugFixedCode
}

func (s *verificationManagerImpl) isReleaseMode() bool {
	return s.env.IsProd()
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return randomString(letters, n, nil)
}

func parseInt64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func maskTarget(target string, targetType model.VerificationTargetType) string {
	if target == "" {
		return ""
	}
	if targetType == model.TargetTypeEmail {
		parts := strings.SplitN(target, "@", 2)
		if len(parts) != 2 {
			return "***"
		}
		name := parts[0]
		if len(name) <= 2 {
			return "***@" + parts[1]
		}
		return name[:2] + "***@" + parts[1]
	}
	if len(target) <= 4 {
		return "***"
	}
	return target[:1] + "***" + target[len(target)-1:]
}

// maskCode redacts a verification code for safe logging (e.g. "****45" for "123456").
func maskCode(code string) string {
	if len(code) <= 2 {
		return "***"
	}
	return strings.Repeat("*", len(code)-2) + code[len(code)-2:]
}

// randomString picks `length` chars from `charset` using crypto/rand.
// Entropy exhaustion is fatal — it means the OS CSPRNG is broken.
func randomString(charset string, length int, log *log.Helper) string {
	n := big.NewInt(int64(len(charset)))
	result := make([]byte, length)
	for i := range result {
		randVal, err := rand.Int(rand.Reader, n)
		if err != nil {
			log.Errorw("msg", "crypto/rand exhausted, cannot generate random string", "error", err)
			panic("crypto/rand exhausted")
		}
		result[i] = charset[randVal.Int64()]
	}
	return string(result)
}
