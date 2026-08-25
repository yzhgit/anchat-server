package model

import "time"

// GroupSetting represents a group settings model
type GroupSetting struct {
	GroupID            string    `gorm:"column:group_id;primaryKey" json:"groupId"`
	JoinVerify         bool      `gorm:"column:join_verify;default:true" json:"joinVerify"`
	AllowMemberInvite  bool      `gorm:"column:allow_member_invite;default:true" json:"allowMemberInvite"`
	AllowViewHistory   bool      `gorm:"column:allow_view_history;default:true" json:"allowViewHistory"`
	AllowAddFriend     bool      `gorm:"column:allow_add_friend;default:true" json:"allowAddFriend"`
	AllowMemberModify  bool      `gorm:"column:allow_member_modify;default:false" json:"allowMemberModify"`
	ShowMemberNickname bool      `gorm:"column:show_member_nickname;default:true" json:"showMemberNickname"`
	CreatedAt          time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt          time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName returns the table name
func (GroupSetting) TableName() string {
	return "group_settings"
}

// DefaultGroupSetting creates default group settings
func DefaultGroupSetting(groupID string) *GroupSetting {
	return &GroupSetting{
		GroupID:            groupID,
		JoinVerify:         true,
		AllowMemberInvite:  true,
		AllowViewHistory:   true,
		AllowAddFriend:     true,
		AllowMemberModify:  false,
		ShowMemberNickname: true,
	}
}
