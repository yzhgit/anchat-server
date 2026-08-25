package model

import (
	"fmt"
	"strings"
	"time"
)

// FriendRequestSource is the friend request source enum.
type FriendRequestSource int16

const (
	FriendRequestSourceUnknown  FriendRequestSource = 0
	FriendRequestSourceSearch   FriendRequestSource = 1
	FriendRequestSourceQRCode   FriendRequestSource = 2
	FriendRequestSourceGroup    FriendRequestSource = 3
	FriendRequestSourceContacts FriendRequestSource = 4
)

var friendRequestSourceValueToString = map[FriendRequestSource]string{
	FriendRequestSourceSearch:   "search",
	FriendRequestSourceQRCode:   "qrcode",
	FriendRequestSourceGroup:    "group",
	FriendRequestSourceContacts: "contacts",
}

func (s FriendRequestSource) String() string {
	if value, ok := friendRequestSourceValueToString[s]; ok {
		return value
	}
	return "unknown"
}

func (s FriendRequestSource) IsValid() bool {
	return s >= FriendRequestSourceSearch && s <= FriendRequestSourceContacts
}

func ParseFriendRequestSource(value string) (FriendRequestSource, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "search":
		return FriendRequestSourceSearch, nil
	case "qrcode":
		return FriendRequestSourceQRCode, nil
	case "group":
		return FriendRequestSourceGroup, nil
	case "contacts":
		return FriendRequestSourceContacts, nil
	default:
		return FriendRequestSourceUnknown, fmt.Errorf("unsupported friend request source: %s", value)
	}
}

// FriendRequestStatus is the friend request status enum.
type FriendRequestStatus int16

const (
	FriendRequestStatusUnknown  FriendRequestStatus = 0
	FriendRequestStatusPending  FriendRequestStatus = 1
	FriendRequestStatusAccepted FriendRequestStatus = 2
	FriendRequestStatusRejected FriendRequestStatus = 3
	FriendRequestStatusExpired  FriendRequestStatus = 4
)

var friendRequestStatusValueToString = map[FriendRequestStatus]string{
	FriendRequestStatusPending:  "pending",
	FriendRequestStatusAccepted: "accepted",
	FriendRequestStatusRejected: "rejected",
	FriendRequestStatusExpired:  "expired",
}

func (s FriendRequestStatus) String() string {
	if value, ok := friendRequestStatusValueToString[s]; ok {
		return value
	}
	return "unknown"
}

func (s FriendRequestStatus) IsValid() bool {
	return s >= FriendRequestStatusPending && s <= FriendRequestStatusExpired
}

func ParseFriendRequestStatus(value string) (FriendRequestStatus, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pending":
		return FriendRequestStatusPending, nil
	case "accepted":
		return FriendRequestStatusAccepted, nil
	case "rejected":
		return FriendRequestStatusRejected, nil
	case "expired":
		return FriendRequestStatusExpired, nil
	default:
		return FriendRequestStatusUnknown, fmt.Errorf("unsupported friend request status: %s", value)
	}
}

// FriendRequestAction is the friend request handling action enum.
type FriendRequestAction int16

const (
	FriendRequestActionUnknown FriendRequestAction = 0
	FriendRequestActionAccept  FriendRequestAction = 1
	FriendRequestActionReject  FriendRequestAction = 2
)

func (a FriendRequestAction) IsValid() bool {
	return a == FriendRequestActionAccept || a == FriendRequestActionReject
}

// FriendRequestQueryType is the query type for listing friend requests.
type FriendRequestQueryType int16

const (
	FriendRequestQueryTypeUnknown  FriendRequestQueryType = 0
	FriendRequestQueryTypeReceived FriendRequestQueryType = 1
	FriendRequestQueryTypeSent     FriendRequestQueryType = 2
)

func (t FriendRequestQueryType) IsValid() bool {
	return t == FriendRequestQueryTypeReceived || t == FriendRequestQueryTypeSent
}

// FriendRequest is the friend request model
type FriendRequest struct {
	ID         int64               `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FromUserID string              `gorm:"column:from_user_id;not null;index" json:"fromUserId"`
	ToUserID   string              `gorm:"column:to_user_id;not null;index:idx_friend_requests_to_user" json:"toUserId"`
	Message    string              `gorm:"column:message;size:200" json:"message"`
	Source     FriendRequestSource `gorm:"column:source;type:smallint;default:1" json:"source"`
	Status     FriendRequestStatus `gorm:"column:status;type:smallint;default:1;index:idx_friend_requests_to_user" json:"status"`
	CreatedAt  time.Time           `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;index" json:"createdAt"`
	UpdatedAt  time.Time           `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName is the table name
func (FriendRequest) TableName() string {
	return "friend_requests"
}

// IsPending returns true if the request is pending
func (fr *FriendRequest) IsPending() bool {
	return fr.Status == FriendRequestStatusPending
}

// IsAccepted returns true if the request is accepted
func (fr *FriendRequest) IsAccepted() bool {
	return fr.Status == FriendRequestStatusAccepted
}

// IsRejected returns true if the request is rejected
func (fr *FriendRequest) IsRejected() bool {
	return fr.Status == FriendRequestStatusRejected
}
