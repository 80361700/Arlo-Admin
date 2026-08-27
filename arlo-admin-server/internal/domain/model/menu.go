package model

import (
	"time"

	"gorm.io/gorm"
)

// Menu 菜单权限模型，对应 sys_menu 表
type Menu struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	ParentID   uint64         `gorm:"not null;default:0;index" json:"parentId"`
	Name       string         `gorm:"size:32;not null" json:"name"`
	Type       int8           `gorm:"not null;default:1" json:"type"`
	Path       string         `gorm:"size:128;not null" json:"path"`
	Component  string         `gorm:"size:128;not null" json:"component"`
	Icon       string         `gorm:"size:64;not null" json:"icon"`
	Sort       int            `gorm:"not null;default:0" json:"sort"`
	Permission string         `gorm:"size:128;not null" json:"permission"`
	Status     int8           `gorm:"not null;default:1" json:"status"`
	Visible    int8           `gorm:"not null;default:1" json:"visible"`
	KeepAlive  int8           `gorm:"not null;default:1" json:"keepAlive"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Menu) TableName() string { return "sys_menu" }

// IsEnabled 菜单是否启用
func (m *Menu) IsEnabled() bool {
	return m.Status == 1
}
