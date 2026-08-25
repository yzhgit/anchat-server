package model

import (
	"time"
)

// VerificationTemplate verification template model
type VerificationTemplate struct {
	ID            int64               `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Purpose       VerificationPurpose `gorm:"column:purpose;type:smallint;not null;uniqueIndex" json:"purpose"`
	Name          string              `gorm:"column:name;not null" json:"name"`
	SMSTemplateID string              `gorm:"column:sms_template_id" json:"smsTemplateId"`
	SMSContent    string              `gorm:"column:sms_content" json:"smsContent"`
	EmailSubject  string              `gorm:"column:email_subject" json:"emailSubject"`
	EmailContent  string              `gorm:"column:email_content" json:"emailContent"`
	IsActive      bool                `gorm:"column:is_active;not null;default:true" json:"isActive"`
	CreatedAt     time.Time           `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt     time.Time           `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (VerificationTemplate) TableName() string {
	return "verification_templates"
}
