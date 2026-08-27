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
	Sort      int            `gorm:"not null;default:0" json:"sort"`
	Leader    string         `gorm:"size:32;not null" json:"leader"`
	Phone     string         `gorm:"size:20;not null" json:"phone"`
	Email     string         `gorm:"size:64;not null" json:"email"`
	Status    int8           `gorm:"not null;default:1" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Dept) TableName() string { return "sys_dept" }
