package model

import (
	"fmt"
	"strings"
	"time"
)

// DeviceType is the user device type enum.
type DeviceType int16

const (
	DeviceTypeUnknown DeviceType = 0
	DeviceTypeIOS     DeviceType = 1
	DeviceTypeAndroid DeviceType = 2
	DeviceTypeWeb     DeviceType = 3
	DeviceTypePC      DeviceType = 4
	DeviceTypeH5      DeviceType = 5
)

var deviceTypeValueToString = map[DeviceType]string{
	DeviceTypeIOS:     "ios",
	DeviceTypeAndroid: "android",
	DeviceTypeWeb:     "web",
	DeviceTypePC:      "pc",
	DeviceTypeH5:      "h5",
}

func (t DeviceType) String() string {
	if s, ok := deviceTypeValueToString[t]; ok {
		return s
	}
	return "unknown"
}

func (t DeviceType) IsValid() bool {
	return t >= DeviceTypeIOS && t <= DeviceTypeH5
}

func ParseDeviceType(value string) (DeviceType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "ios":
		return DeviceTypeIOS, nil
	case "android":
		return DeviceTypeAndroid, nil
	case "web":
		return DeviceTypeWeb, nil
	case "pc":
		return DeviceTypePC, nil
	case "h5":
		return DeviceTypeH5, nil
	default:
		return DeviceTypeUnknown, fmt.Errorf("unsupported device type: %s", value)
	}
}

// UserDevice user device model
type UserDevice struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID        string     `gorm:"column:user_id;not null" json:"userId"`
	DeviceID      string     `gorm:"column:device_id;not null" json:"deviceId"`
	DeviceType    DeviceType `gorm:"column:device_type;type:smallint;not null" json:"deviceType"`
	ClientVersion string     `gorm:"column:client_version" json:"clientVersion"`
	LastLoginAt   *time.Time `gorm:"column:last_login_at" json:"lastLoginAt"`
	LastLoginIP   string     `gorm:"column:last_login_ip" json:"lastLoginIp"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName returns table name
func (UserDevice) TableName() string {
	return "user_devices"
}
