package model

import "time"

// LoginLog 登录日志模型，对应 sys_login_log 表
type LoginLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:32;not null;default:''" json:"username"`
	IP        string    `gorm:"size:45;not null;default:''" json:"ip"`
	Location  string    `gorm:"size:128;not null;default:''" json:"location"`
	Browser   string    `gorm:"size:64;not null;default:''" json:"browser"`
	OS        string    `gorm:"size:32;not null;default:''" json:"os"`
	Status    int8      `gorm:"not null;default:1" json:"status"` // 0=失败, 1=成功
	Msg       string    `gorm:"size:255;not null;default:''" json:"msg"`
	CreatedAt time.Time `json:"createdAt"`
}

func (LoginLog) TableName() string {
	return "sys_login_log"
}
