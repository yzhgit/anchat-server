package model

import "time"

// JoinRequestStatus represents join request status.
type JoinRequestStatus int16

// JoinRequestStatus values.
const (
	JoinRequestStatusUnknown  JoinRequestStatus = 0
	JoinRequestStatusPending  JoinRequestStatus = 1
	JoinRequestStatusAccepted JoinRequestStatus = 2
	JoinRequestStatusRejected JoinRequestStatus = 3
)

// GroupJoinRequest represents a group join request model
type GroupJoinRequest struct {
	ID        int64             `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID   string            `gorm:"column:group_id;not null" json:"groupId"`
	UserID    string            `gorm:"column:user_id;not null" json:"userId"`
	InviterID string            `gorm:"column:inviter_id" json:"inviterId"` // NULL means user-initiated request
	Message   string            `gorm:"column:message;size:200" json:"message"`
	Status    JoinRequestStatus `gorm:"column:status;type:smallint;not null;default:1" json:"status"`
	CreatedAt time.Time         `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time         `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName returns the table name
func (GroupJoinRequest) TableName() string {
	return "group_join_requests"
}

// IsPending checks if pending
func (r *GroupJoinRequest) IsPending() bool {
	return r.Status == JoinRequestStatusPending
}

// IsProcessed checks if processed
func (r *GroupJoinRequest) IsProcessed() bool {
	return r.Status == JoinRequestStatusAccepted || r.Status == JoinRequestStatusRejected
}

// IsInvitation checks if this is an invitation (vs user-initiated request)
func (r *GroupJoinRequest) IsInvitation() bool {
	return r.InviterID != ""
}
