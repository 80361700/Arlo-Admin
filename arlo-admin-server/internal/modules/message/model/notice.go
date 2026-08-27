package model

import (
	"time"

	"gorm.io/gorm"
)

// Notice 通知公告模型，对应 sys_notice 表
type Notice struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:128;not null;default:''" json:"title"`
	Content     string         `gorm:"type:text" json:"content"`
	Type        int8           `gorm:"not null;default:1" json:"type"`      // 1=通知, 2=公告
	Level       int8           `gorm:"not null;default:1" json:"level"`     // 1=普通, 2=重要, 3=紧急
	Status      int8           `gorm:"not null;default:0" json:"status"`    // 0=草稿, 1=已发布, 2=已撤回
	PublisherID uint64         `gorm:"not null;default:0" json:"publisherId"`
	Publisher   string         `gorm:"size:32;not null;default:''" json:"publisher"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Notice) TableName() string {
	return "sys_notice"
}
