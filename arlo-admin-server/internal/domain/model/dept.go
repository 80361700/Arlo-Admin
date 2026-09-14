package model

import (
	"time"

	"gorm.io/gorm"
)

// Dept 部门模型，对应 sys_dept 表
type Dept struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	ParentID  uint64         `gorm:"not null;default:0;index" json:"parentId"`
	Name      string         `gorm:"size:64;not null" json:"name"`
	Code      string         `gorm:"size:64;not null" json:"code"`
	Sort      int            `gorm:"not null;default:0" json:"sort"`
	LeaderID  uint64         `gorm:"not null;default:0" json:"leaderId"`
	Status    int8           `gorm:"not null;default:1" json:"status"`
	Remark    string         `gorm:"size:255;not null" json:"remark"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Dept) TableName() string { return "sys_dept" }
