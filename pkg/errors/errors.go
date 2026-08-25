package errors

import "errors"

// Common error codes
const (
	CodeSuccess       = 0
	CodeParamError    = 1   // Parameter error
	CodeInternalError = 2   // Internal error
	CodeUnauthorized  = 401 // Unauthorized
	CodeForbidden     = 403 // Forbidden
	CodeNotFound      = 404 // Resource not found
)

// User Service error codes (10xxx)
const (
	CodeUserExists          = 10101 // User already exists
	CodePasswordWeak        = 10103 // Password too weak
	CodeUserNotFound        = 10104 // User not found
	CodePasswordError       = 10105 // Incorrect password
	CodeAccountDisabled     = 10106 // Account disabled
	CodeRefreshTokenInvalid = 10107 // Invalid RefreshToken
	CodeRefreshTokenExpired = 10108 // RefreshToken expired
	CodeTokenInvalid        = 10109 // Invalid Token
	CodeTokenExpired        = 10110 // Token expired

	// Verification code sub-domain error codes (102xx)
	CodeSendRateLimited        = 10201 // Sending too frequently
	CodeSendLimitReached       = 10202 // Verification code send limit reached
	CodeTargetFormatInvalid    = 10203 // Invalid target format
	CodeSMSServiceError        = 10204 // SMS service error
	CodeEmailServiceError      = 10205 // Email service error
	CodeVerifyCodeError        = 10206 // Incorrect verification code
	CodeVerifyCodeExpired      = 10207 // Verification code expired
	CodeVerifyCodeAlreadyUsed  = 10208 // Verification code already used, please get a new one
	CodeVerifyCodeNotFound     = 10209 // Verification code not found
	CodeVerifyAttemptsExceeded = 10210 // Too many verification attempts

	CodeNicknameUsed        = 10301 // Nickname already used
	CodeNicknameSensitive   = 10302 // Nickname contains sensitive words
	CodeUserProfileNotFound = 10303 // User not found
	CodeQRCodeExpired       = 10304 // QR code expired
	CodeQRCodeInvalid       = 10305 // Invalid QR code
	CodePhoneFormatInvalid  = 10306 // Invalid phone number format
	CodePhoneAlreadyBound   = 10307 // Phone number already bound
	CodeEmailFormatInvalid  = 10308 // Invalid email format
	CodeEmailAlreadyBound   = 10309 // Email already bound
	CodeOldPhoneNotMatch    = 10310 // Old phone number does not match
	CodeOldEmailNotMatch    = 10311 // Old email does not match
)

// Friend Service error codes (20xxx)
const (
	CodeAlreadyFriend         = 20101 // Already friends
	CodeBlockedByUser         = 20102 // Blocked by user
	CodeDuplicateRequest      = 20103 // Duplicate request
	CodeFriendNotFound        = 20104 // Friend not found
	CodeRequestNotFound       = 20105 // Request not found
	CodeCannotAddSelf         = 20106 // Cannot add yourself as friend
	CodeRequestProcessed      = 20107 // Request already processed
	CodeRequestExpired        = 20108 // Request expired
	CodeFriendLimitReached    = 20109 // Friend limit reached
	CodeTargetFriendLimit     = 20110 // Target friend limit reached
	CodeBlacklistLimitReached = 20111 // Blacklist limit reached
	CodeAlreadyInBlacklist    = 20112 // Already in blacklist
	CodeNotInBlacklist        = 20113 // Not in blacklist
	CodeUserBlocked           = 20114 // User blocked
	CodeRequestExists         = 20115 // Request already exists
	CodePermissionDenied      = 20116 // Permission denied
	CodeNotFriend             = 20117 // Not a friend
)

// Group Service error codes (30xxx)
const (
	CodeGroupNotFound            = 30101 // Group not found
	CodeGroupDissolved           = 30102 // Group dissolved
	CodeGroupMemberTooFew        = 30103 // Insufficient group members
	CodeGroupMemberLimitReached  = 30104 // Group member limit reached
	CodeNotGroupMember           = 30105 // Not a group member
	CodeAlreadyGroupMember       = 30106 // Already a group member
	CodeNoOwnerPermission        = 30107 // No owner permission
	CodeNoAdminPermission        = 30108 // No admin permission
	CodeCannotRemoveOwner        = 30109 // Cannot remove owner
	CodeCannotRemoveAdmin        = 30110 // Cannot remove admin
	CodeGroupNameSensitive       = 30111 // Group name contains sensitive words
	CodeAnnouncementSensitive    = 30112 // Group announcement contains sensitive words
	CodeJoinRequestNotFound      = 30113 // Join request not found
	CodeJoinRequestProcessed     = 30114 // Join request already processed
	CodeMemberMuted              = 30115 // Member muted
	CodeCannotQuitOwnGroup       = 30116 // Cannot quit your own group
	CodeGroupQRExpired           = 30117 // Group QR code expired
	CodeGroupQRInvalid           = 30118 // Invalid group QR code
	CodeGroupPinnedLimitExceeded = 30119 // Group pinned limit exceeded
)

// Message Service error codes (40xxx)
const (
	CodeMessageNotFound         = 40101 // Message not found
	CodeMessageSendFailed       = 40102 // Message send failed
	CodeMessageRecallFailed     = 40103 // Message recall failed
	CodeMessageRecallTimeLimit  = 40104 // Message recall time limit exceeded
	CodeMessageDeleteFailed     = 40105 // Message delete failed
	CodeMessagePermissionDenied = 40106 // Message permission denied
	CodeConversationNotFound    = 40107 // Conversation not found
	CodeSequenceGenerateFailed  = 40108 // Sequence number generation failed
	CodeMarkReadFailed          = 40109 // Mark read failed
	CodeGetUnreadCountFailed    = 40110 // Get unread count failed
	CodeSearchMessageFailed     = 40111 // Search message failed
	CodeInvalidOperation        = 40112 // Invalid operation
	CodeMessageNotInGroup       = 40113 // Message not in this group
)

// File Service error codes (60xxx)
const (
	CodeFileNotFound         = 60101 // File not found
	CodeFileAccessDenied     = 60102 // File access denied
	CodeFileSizeExceeded     = 60103 // File size exceeded
	CodeFileTypeNotAllowed   = 60104 // File type not allowed
	CodeFileUploadFailed     = 60105 // File upload failed
	CodeFileAlreadyExists    = 60106 // File already exists
	CodeInvalidFileID        = 60107 // Invalid file ID
	CodeFileExpired          = 60108 // File expired
	CodeStorageQuotaExceeded = 60109 // Storage quota exceeded
	CodeThumbnailGenFailed   = 60110 // Thumbnail generation failed
)

// RTC Service error codes (70xxx)
const (
	CodeCallNotFound         = 70101 // Call not found
	CodeCallAlreadyActive    = 70102 // Call already active
	CodeCallPermissionDenied = 70103 // Call permission denied
	CodeCallInvalidStatus    = 70104 // Invalid call status
	CodeMeetingNotFound      = 70105 // Meeting room not found
	CodeMeetingPasswordWrong = 70106 // Meeting room password incorrect
	CodeMeetingAlreadyEnded  = 70107 // Meeting room already ended
	CodeMeetingPermission    = 70108 // Meeting room permission denied
	CodeLiveKitTokenFailed   = 70109 // LiveKit token generation failed
	CodeLiveKitRoomFailed    = 70110 // LiveKit room operation failed
)

// Push Service error codes (80xxx)
const (
	CodePushFailed        = 80101 // Push failed
	CodePushTokenNotFound = 80102 // Push token not found
	CodePushConfigInvalid = 80103 // Push config invalid
)

// Admin Service error codes (90xxx)
const (
	CodeAdminNotFound        = 90101 // Admin not found
	CodeAdminDisabled        = 90102 // Admin account disabled
	CodeAdminUsernameExists  = 90103 // Admin username already exists
	CodeAdminInvalidPassword = 90104 // Admin password incorrect
	CodeAdminTokenInvalid    = 90105 // Admin token invalid
	CodeAdminPermission      = 90106 // Admin permission denied
	CodeConfigKeyNotFound    = 90107 // Config key not found
)

// Error message mapping
var errorMessages = map[int]string{
	CodeSuccess:       "Success",
	CodeParamError:    "Parameter error",
	CodeInternalError: "Internal error",
	CodeUnauthorized:  "Unauthorized",
	CodeForbidden:     "Forbidden",
	CodeNotFound:      "Resource not found",

	CodeUserExists:          "User already exists",
	CodePasswordWeak:        "Password too weak",
	CodeUserNotFound:        "User not found",
	CodePasswordError:       "Incorrect password",
	CodeAccountDisabled:     "Account disabled",
	CodeRefreshTokenInvalid: "Invalid RefreshToken",
	CodeRefreshTokenExpired: "RefreshToken expired",
	CodeTokenInvalid:        "Invalid Token",
	CodeTokenExpired:        "Token expired",

	CodeNicknameUsed:        "Nickname already used",
	CodeNicknameSensitive:   "Nickname contains sensitive words",
	CodeUserProfileNotFound: "User not found",
	CodeQRCodeExpired:       "QR code expired",
	CodeQRCodeInvalid:       "Invalid QR code",
	CodePhoneFormatInvalid:  "Invalid phone number format",
	CodePhoneAlreadyBound:   "Phone number already bound",
	CodeEmailFormatInvalid:  "Invalid email format",
	CodeEmailAlreadyBound:   "Email already bound",
	CodeOldPhoneNotMatch:    "Old phone number does not match",
	CodeOldEmailNotMatch:    "Old email does not match",

	CodeAlreadyFriend:         "Already friends",
	CodeBlockedByUser:         "Blocked by user",
	CodeDuplicateRequest:      "Duplicate request",
	CodeFriendNotFound:        "Friend not found",
	CodeRequestNotFound:       "Request not found",
	CodeCannotAddSelf:         "Cannot add yourself as friend",
	CodeRequestProcessed:      "Request already processed",
	CodeRequestExpired:        "Request expired",
	CodeFriendLimitReached:    "Friend limit reached",
	CodeTargetFriendLimit:     "Target friend limit reached",
	CodeBlacklistLimitReached: "Blacklist limit reached",
	CodeAlreadyInBlacklist:    "Already in blacklist",
	CodeNotInBlacklist:        "Not in blacklist",
	CodeUserBlocked:           "User blocked",
	CodeRequestExists:         "Request already exists",
	CodePermissionDenied:      "Permission denied",
	CodeNotFriend:             "Not a friend",

	CodeGroupNotFound:            "Group not found",
	CodeGroupDissolved:           "Group dissolved",
	CodeGroupMemberTooFew:        "Insufficient group members",
	CodeGroupMemberLimitReached:  "Group member limit reached",
	CodeNotGroupMember:           "Not a group member",
	CodeAlreadyGroupMember:       "Already a group member",
	CodeNoOwnerPermission:        "No owner permission",
	CodeNoAdminPermission:        "No admin permission",
	CodeCannotRemoveOwner:        "Cannot remove owner",
	CodeCannotRemoveAdmin:        "Cannot remove admin",
	CodeGroupNameSensitive:       "Group name contains sensitive words",
	CodeAnnouncementSensitive:    "Group announcement contains sensitive words",
	CodeJoinRequestNotFound:      "Join request not found",
	CodeJoinRequestProcessed:     "Join request already processed",
	CodeMemberMuted:              "Member muted",
	CodeCannotQuitOwnGroup:       "Cannot quit your own group",
	CodeGroupQRExpired:           "Group QR code expired",
	CodeGroupQRInvalid:           "Invalid group QR code",
	CodeGroupPinnedLimitExceeded: "Group pinned limit exceeded, please unpin some first",

	CodeMessageNotFound:         "Message not found",
	CodeMessageSendFailed:       "Message send failed",
	CodeMessageRecallFailed:     "Message recall failed",
	CodeMessageRecallTimeLimit:  "Message recall time limit exceeded",
	CodeMessageDeleteFailed:     "Message delete failed",
	CodeMessagePermissionDenied: "Message permission denied",
	CodeConversationNotFound:    "Conversation not found",
	CodeSequenceGenerateFailed:  "Sequence number generation failed",
	CodeMarkReadFailed:          "Mark read failed",
	CodeGetUnreadCountFailed:    "Get unread count failed",
	CodeSearchMessageFailed:     "Search message failed",
	CodeInvalidOperation:        "Invalid operation",
	CodeMessageNotInGroup:       "Message not in this group",

	CodeFileNotFound:         "File not found",
	CodeFileAccessDenied:     "File access denied",
	CodeFileSizeExceeded:     "File size exceeded",
	CodeFileTypeNotAllowed:   "File type not allowed",
	CodeFileUploadFailed:     "File upload failed",
	CodeFileAlreadyExists:    "File already exists",
	CodeInvalidFileID:        "Invalid file ID",
	CodeFileExpired:          "File expired",
	CodeStorageQuotaExceeded: "Storage quota exceeded",
	CodeThumbnailGenFailed:   "Thumbnail generation failed",

	CodeSendRateLimited:        "Sending too frequently, please try again later",
	CodeSendLimitReached:       "Verification code send limit reached",
	CodeTargetFormatInvalid:    "Invalid target format",
	CodeSMSServiceError:        "SMS service error, please try again later",
	CodeEmailServiceError:      "Email service error, please try again later",
	CodeVerifyCodeError:        "Incorrect verification code",
	CodeVerifyCodeExpired:      "Verification code expired",
	CodeVerifyCodeAlreadyUsed:  "Verification code already used, please get a new one",
	CodeVerifyCodeNotFound:     "Verification code not found",
	CodeVerifyAttemptsExceeded: "Too many verification attempts, please get a new verification code",
}

// GetMessage returns the error message for a given code
func GetMessage(code int) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "Unknown error"
}

// Business represents a business error
type Business struct {
	Code    int
	Message string
}

func (e *Business) Error() string {
	return e.Message
}

// NewBusiness creates a new business error
func NewBusiness(code int, message string) error {
	if message == "" {
		message = GetMessage(code)
	}
	return &Business{
		Code:    code,
		Message: message,
	}
}

// IsBusiness checks if the error is a business error
func IsBusiness(err error) bool {
	var bizErr *Business
	return errors.As(err, &bizErr)
}

// GetBusinessCode retrieves the business error code
func GetBusinessCode(err error) int {
	var bizErr *Business
	if errors.As(err, &bizErr) {
		return bizErr.Code
	}
	return CodeInternalError
}
