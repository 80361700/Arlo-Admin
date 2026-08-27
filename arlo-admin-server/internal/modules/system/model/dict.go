package model

import (
	"time"

	"gorm.io/gorm"
)

// DictType 字典类型模型，对应 sys_dict_type 表
type DictType struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:64;not null" json:"name"`
	Code      string         `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Status    int8           `gorm:"not null;default:1" json:"status"`
	Remark    string         `gorm:"size:255;not null" json:"remark"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DictType) TableName() string { return "sys_dict_type" }

// DictData 字典数据模型，对应 sys_dict_data 表
type DictData struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	DictTypeID  uint64         `gorm:"not null;default:0;index" json:"dictTypeId"`
	Label       string         `gorm:"size:64;not null" json:"label"`
	Value       string         `gorm:"size:64;not null" json:"value"`
	Sort        int            `gorm:"not null;default:0" json:"sort"`
	IsDefault   int8           `gorm:"not null;default:0" json:"isDefault"`
	Status      int8           `gorm:"not null;default:1" json:"status"`
	Remark      string         `gorm:"size:255;not null" json:"remark"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DictData) TableName() string { return "sys_dict_data" }
