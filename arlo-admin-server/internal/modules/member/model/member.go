package model

import (
	"time"

	"gorm.io/gorm"
)

// Member C端会员用户模型，对应 member 表
type Member struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Phone     string         `gorm:"uniqueIndex;size:20;not null;default:''" json:"phone"`
	Password  string         `gorm:"size:128;not null;default:''" json:"-"`
	Nickname  string         `gorm:"size:64;not null;default:''" json:"nickname"`
	Avatar    string         `gorm:"size:255;not null;default:''" json:"avatar"`
	Gender    int8           `gorm:"not null;default:0" json:"gender"`
	Openid    string         `gorm:"uniqueIndex;size:64;default:null" json:"openid"`
	Unionid   string         `gorm:"size:64;not null;default:''" json:"unionid"`
	MpOpenid  string         `gorm:"size:64;not null;default:''" json:"mpOpenid"`
	Source    string         `gorm:"size:16;not null;default:'h5'" json:"source"`
	Status    int8           `gorm:"not null;default:1" json:"status"`
	LastLogin *time.Time     `gorm:"" json:"lastLogin"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Member) TableName() string {
	return "sys_member"
}

func (m *Member) IsEnabled() bool {
	return m.Status == 1
}
