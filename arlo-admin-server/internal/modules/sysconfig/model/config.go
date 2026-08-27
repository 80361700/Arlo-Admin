package model

import (
	"time"

	"gorm.io/gorm"
)

// SysConfig 系统配置模型，对应 sys_config 表
type SysConfig struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:64;not null" json:"name"`             // 配置名称
	Key       string         `gorm:"uniqueIndex;size:64;not null" json:"key"`  // 配置键
	Value     string         `gorm:"type:text" json:"value"`                   // 配置值
	Type      int8           `gorm:"not null;default:1" json:"type"`           // 类型: 1=文本, 2=JSON, 3=开关, 4=图片
	Remark    string         `gorm:"size:255;not null;default:''" json:"remark"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SysConfig) TableName() string { return "sys_config" }
