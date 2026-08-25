package errors

import (
	"context"
	stderrors "errors"
	"strconv"

	"flamingo/pkg/consts"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// grpcCodeMapping maps business error codes to gRPC status codes.
// The mapping follows industry best practices:
//   - Parameter/validation errors → InvalidArgument
//   - Resource not found → NotFound
//   - Resource already exists → AlreadyExists
//   - Permission denied → PermissionDenied
//   - Rate limit / quota exceeded → ResourceExhausted
//   - Precondition failed (state conflict) → FailedPrecondition
//   - Unknown → Internal
var grpcCodeMapping = map[int]codes.Code{
	// Common
	CodeParamError:    codes.InvalidArgument,
	CodeInternalError: codes.Internal,
	CodeUnauthorized:  codes.Unauthenticated,
	CodeForbidden:     codes.PermissionDenied,
	CodeNotFound:      codes.NotFound,

	// User (10xxx)
	CodeUserExists:          codes.AlreadyExists,
	CodePasswordWeak:        codes.InvalidArgument,
	CodeUserNotFound:        codes.NotFound,
	CodePasswordError:       codes.InvalidArgument,
	CodeAccountDisabled:     codes.PermissionDenied,
	CodeRefreshTokenInvalid: codes.Unauthenticated,
	CodeRefreshTokenExpired: codes.Unauthenticated,
	CodeTokenInvalid:        codes.Unauthenticated,
	CodeTokenExpired:        codes.Unauthenticated,

	CodeSendRateLimited:        codes.ResourceExhausted,
	CodeSendLimitReached:       codes.ResourceExhausted,
	CodeTargetFormatInvalid:    codes.InvalidArgument,
	CodeSMSServiceError:        codes.Internal,
	CodeEmailServiceError:      codes.Internal,
	CodeVerifyCodeError:        codes.InvalidArgument,
	CodeVerifyCodeExpired:      codes.InvalidArgument,
	CodeVerifyCodeAlreadyUsed:  codes.FailedPrecondition,
	CodeVerifyCodeNotFound:     codes.NotFound,
	CodeVerifyAttemptsExceeded: codes.ResourceExhausted,

	CodeNicknameUsed:        codes.AlreadyExists,
	CodeNicknameSensitive:   codes.InvalidArgument,
	CodeUserProfileNotFound: codes.NotFound,
	CodeQRCodeExpired:       codes.FailedPrecondition,
	CodeQRCodeInvalid:       codes.InvalidArgument,
	CodePhoneFormatInvalid:  codes.InvalidArgument,
	CodePhoneAlreadyBound:   codes.AlreadyExists,
	CodeEmailFormatInvalid:  codes.InvalidArgument,
	CodeEmailAlreadyBound:   codes.AlreadyExists,
	CodeOldPhoneNotMatch:    codes.InvalidArgument,
	CodeOldEmailNotMatch:    codes.InvalidArgument,

	// Friend (20xxx)
	CodeAlreadyFriend:         codes.AlreadyExists,
	CodeBlockedByUser:         codes.PermissionDenied,
	CodeDuplicateRequest:      codes.AlreadyExists,
	CodeFriendNotFound:        codes.NotFound,
	CodeRequestNotFound:       codes.NotFound,
	CodeCannotAddSelf:         codes.InvalidArgument,
	CodeRequestProcessed:      codes.FailedPrecondition,
	CodeRequestExpired:        codes.FailedPrecondition,
	CodeFriendLimitReached:    codes.ResourceExhausted,
	CodeTargetFriendLimit:     codes.ResourceExhausted,
	CodeBlacklistLimitReached: codes.ResourceExhausted,
	CodeAlreadyInBlacklist:    codes.AlreadyExists,
	CodeNotInBlacklist:        codes.FailedPrecondition,
	CodeUserBlocked:           codes.PermissionDenied,
	CodeRequestExists:         codes.AlreadyExists,
	CodePermissionDenied:      codes.PermissionDenied,
	CodeNotFriend:             codes.FailedPrecondition,

	// Group (30xxx)
	CodeGroupNotFound:            codes.NotFound,
	CodeGroupDissolved:           codes.FailedPrecondition,
	CodeGroupMemberTooFew:        codes.FailedPrecondition,
	CodeGroupMemberLimitReached:  codes.ResourceExhausted,
	CodeNotGroupMember:           codes.PermissionDenied,
	CodeAlreadyGroupMember:       codes.AlreadyExists,
	CodeNoOwnerPermission:        codes.PermissionDenied,
	CodeNoAdminPermission:        codes.PermissionDenied,
	CodeCannotRemoveOwner:        codes.FailedPrecondition,
	CodeCannotRemoveAdmin:        codes.FailedPrecondition,
	CodeGroupNameSensitive:       codes.InvalidArgument,
	CodeAnnouncementSensitive:    codes.InvalidArgument,
	CodeJoinRequestNotFound:      codes.NotFound,
	CodeJoinRequestProcessed:     codes.FailedPrecondition,
	CodeMemberMuted:              codes.FailedPrecondition,
	CodeCannotQuitOwnGroup:       codes.FailedPrecondition,
	CodeGroupQRExpired:           codes.FailedPrecondition,
	CodeGroupQRInvalid:           codes.InvalidArgument,
	CodeGroupPinnedLimitExceeded: codes.ResourceExhausted,

	// Message (40xxx)
	CodeMessageNotFound:         codes.NotFound,
	CodeMessageSendFailed:       codes.Internal,
	CodeMessageRecallFailed:     codes.Internal,
	CodeMessageRecallTimeLimit:  codes.FailedPrecondition,
	CodeMessageDeleteFailed:     codes.Internal,
	CodeMessagePermissionDenied: codes.PermissionDenied,
	CodeConversationNotFound:    codes.NotFound,
	CodeSequenceGenerateFailed:  codes.Internal,
	CodeMarkReadFailed:          codes.Internal,
	CodeGetUnreadCountFailed:    codes.Internal,
	CodeSearchMessageFailed:     codes.Internal,
	CodeInvalidOperation:        codes.FailedPrecondition,
	CodeMessageNotInGroup:       codes.FailedPrecondition,

	// File (60xxx)
	CodeFileNotFound:         codes.NotFound,
	CodeFileAccessDenied:     codes.PermissionDenied,
	CodeFileSizeExceeded:     codes.InvalidArgument,
	CodeFileTypeNotAllowed:   codes.InvalidArgument,
	CodeFileUploadFailed:     codes.Internal,
	CodeFileAlreadyExists:    codes.AlreadyExists,
	CodeInvalidFileID:        codes.InvalidArgument,
	CodeFileExpired:          codes.FailedPrecondition,
	CodeStorageQuotaExceeded: codes.ResourceExhausted,
	CodeThumbnailGenFailed:   codes.Internal,

	// RTC (70xxx)
	CodeCallNotFound:         codes.NotFound,
	CodeCallAlreadyActive:    codes.AlreadyExists,
	CodeCallPermissionDenied: codes.PermissionDenied,
	CodeCallInvalidStatus:    codes.FailedPrecondition,
	CodeMeetingNotFound:      codes.NotFound,
	CodeMeetingPasswordWrong: codes.InvalidArgument,
	CodeMeetingAlreadyEnded:  codes.FailedPrecondition,
	CodeMeetingPermission:    codes.PermissionDenied,
	CodeLiveKitTokenFailed:   codes.Internal,
	CodeLiveKitRoomFailed:    codes.Internal,

	// Push (80xxx)
	CodePushFailed:        codes.Internal,
	CodePushTokenNotFound: codes.NotFound,
	CodePushConfigInvalid: codes.InvalidArgument,

	// Admin (90xxx)
	CodeAdminNotFound:        codes.NotFound,
	CodeAdminDisabled:        codes.PermissionDenied,
	CodeAdminUsernameExists:  codes.AlreadyExists,
	CodeAdminInvalidPassword: codes.InvalidArgument,
	CodeAdminTokenInvalid:    codes.Unauthenticated,
	CodeAdminPermission:      codes.PermissionDenied,
	CodeConfigKeyNotFound:    codes.NotFound,
}

// ConvertError converts an error to a gRPC status error and sets the business
// error code in trailing metadata so that callers (including the gateway) can
// extract the precise business error code.
//
// For *Business errors: maps to the appropriate gRPC code and attaches
// x-error-code metadata.
//
// For non-business errors: wraps them as codes.Internal.
//
// For nil: returns nil.
//
// BadRequest and Internal are convenience helpers that construct *Business
// errors and immediately route them through ConvertError, ensuring the
// x-error-code metadata is always attached. Use these in handlers for
// parameter validation failures instead of rtc status.Error directly.
func ConvertError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	var bizErr *Business
	if !stderrors.As(err, &bizErr) {
		// Not a business error — wrap as Internal.
		return status.Error(codes.Internal, err.Error())
	}

	code := grpcCodeMapping[bizErr.Code]
	if code == 0 {
		code = codes.Internal
	}

	// Set business error code in trailing metadata for downstream consumers
	// (e.g., gateway to forward as HTTP header).
	metadata.AppendToOutgoingContext(ctx, consts.ErrorBusinessCodeMetadataKey, strconv.Itoa(bizErr.Code))

	return status.Error(code, bizErr.Message)
}

// BadRequest constructs a parameter-validation business error and routes it
// through ConvertError, ensuring the x-error-code metadata is attached.
//
// Use this in handlers for parameter validation failures instead of
// rtc status.Error(codes.InvalidArgument, message) directly.
//
// Example:
//
//	if req.UserId == "" {
//	    return nil, errors.BadRequest(ctx, "user_id is required")
//	}
func BadRequest(ctx context.Context, message string) error {
	return ConvertError(ctx, NewBusiness(CodeParamError, message))
}

// Internal wraps an internal/unexpected error through ConvertError.
// Since the wrapped error is not a *Business, ConvertError maps it to
// codes.Internal and attaches the CodeInternalError metadata.
//
// Use this in handlers/services for unexpected failures instead of
// rtc status.Error(codes.Internal, err.Error()) directly.
//
// Example:
//
//	if err := someOperation(ctx); err != nil {
//	    return nil, errors.Internal(ctx, err)
//	}
func Internal(ctx context.Context, err error) error {
	return ConvertError(ctx, NewBusiness(CodeInternalError, err.Error()))
}
