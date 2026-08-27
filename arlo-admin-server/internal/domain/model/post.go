package model

import (
	"time"

	"gorm.io/gorm"
)

// Post 岗位模型，对应 sys_post 表
type Post struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Code      string         `gorm:"uniqueIndex;size:32;not null" json:"code"`
	Name      string         `gorm:"size:64;not null" json:"name"`
	Sort      int            `gorm:"not null;default:0" json:"sort"`
	Status    int8           `gorm:"not null;default:1" json:"status"`
	Remark    string         `gorm:"size:255;not null" json:"remark"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Post) TableName() string { return "sys_post" }
