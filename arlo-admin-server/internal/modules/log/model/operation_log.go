package model

import "time"

// OperationLog 操作日志模型，对应 sys_operation_log 表
type OperationLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"not null;default:0" json:"userId"`
	Username  string    `gorm:"size:32;not null;default:''" json:"username"`
	Module    string    `gorm:"size:32;not null;default:''" json:"module"`
	Action    string    `gorm:"size:64;not null;default:''" json:"action"`
	Method    string    `gorm:"size:10;not null;default:''" json:"method"`
	URL       string    `gorm:"size:255;not null;default:''" json:"url"`
	IP        string    `gorm:"size:45;not null;default:''" json:"ip"`
	UserAgent string    `gorm:"size:500;not null;default:''" json:"userAgent"`
	Params    string    `gorm:"type:text" json:"params"`
	Result    string    `gorm:"type:text" json:"result"`
	CostTime  int       `gorm:"not null;default:0" json:"costTime"`
	Status    int8      `gorm:"not null;default:1" json:"status"` // 0=失败, 1=成功
	ErrorMsg  string    `gorm:"type:text" json:"errorMsg"`
	CreatedAt time.Time `json:"createdAt"`
}

func (OperationLog) TableName() string {
	return "sys_operation_log"
}
