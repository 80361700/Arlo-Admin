package model

import (
	"time"

	"gorm.io/gorm"
)

// FlowCategory 流程分类
type FlowCategory struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"size:64;not null;default:''" json:"name"`
	Sort      int            `gorm:"not null;default:0" json:"sort"`
	Remark    string         `gorm:"size:255;not null;default:''" json:"remark"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FlowCategory) TableName() string { return "flow_category" }

// FlowProcess 流程定义
type FlowProcess struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryID        uint64         `gorm:"not null;default:0;index" json:"categoryId"`
	ProcessKey        string         `gorm:"size:64;not null;uniqueIndex" json:"processKey"`
	ProcessName       string         `gorm:"size:128;not null;default:''" json:"processName"`
	ProcessIcon       string         `gorm:"size:512;not null;default:''" json:"processIcon"`
	ProcessType       string         `gorm:"size:32;not null;default:main" json:"processType"` // main审批 business业务审批 child子流程
	ProcessVersion    int            `gorm:"not null;default:1" json:"processVersion"`
	ProcessState      int8           `gorm:"not null;default:0;index" json:"processState"` // 0禁用 1启用
	Remark            string         `gorm:"size:255;not null;default:''" json:"remark"`
	ModelContent      string         `gorm:"type:longtext" json:"modelContent"`
	ProcessForm       string         `gorm:"type:longtext" json:"processForm"`
	ProcessSetting    string         `gorm:"type:text" json:"processSetting"`
	ProcessPermission string         `gorm:"type:text" json:"processPermission"`
	CreatedBy         uint64         `gorm:"not null;default:0" json:"createdBy"`
	UpdatedBy         uint64         `gorm:"not null;default:0" json:"updatedBy"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FlowProcess) TableName() string { return "flow_process" }

// FlowFormCategory 表单分类
type FlowFormCategory struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"size:64;not null;default:''" json:"name"`
	Sort      int            `gorm:"not null;default:0" json:"sort"`
	Remark    string         `gorm:"size:255;not null;default:''" json:"remark"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FlowFormCategory) TableName() string { return "flow_form_category" }

// FlowForm 表单模板（设计表单；form_type=2 系统表单预留）
type FlowForm struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryID uint64         `gorm:"not null;default:0;index" json:"categoryId"`
	Name       string         `gorm:"size:128;not null;default:''" json:"name"`
	Code       string         `gorm:"size:64;not null;uniqueIndex" json:"code"`
	FormType   int8           `gorm:"not null;default:1" json:"formType"` // 1设计 2系统
	Status     int8           `gorm:"not null;default:1;index" json:"status"` // 0禁用 1正常
	Sort       int            `gorm:"not null;default:0" json:"sort"`
	Remark     string         `gorm:"size:255;not null;default:''" json:"remark"`
	FormSchema string         `gorm:"type:longtext" json:"formSchema"`
	PcURL      string         `gorm:"column:pc_url;size:255;not null;default:''" json:"pcUrl"`
	AppURL     string         `gorm:"column:app_url;size:255;not null;default:''" json:"appUrl"`
	CreatedBy  uint64         `gorm:"not null;default:0" json:"createdBy"`
	UpdatedBy  uint64         `gorm:"not null;default:0" json:"updatedBy"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FlowForm) TableName() string { return "flow_form" }

// FlowProcessHistory 流程定义历史版本快照
type FlowProcessHistory struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProcessID         uint64    `gorm:"not null;default:0;index;uniqueIndex:uk_flow_process_history_ver" json:"processId"`
	ProcessKey        string    `gorm:"size:64;not null;default:''" json:"processKey"`
	ProcessName       string    `gorm:"size:128;not null;default:''" json:"processName"`
	ProcessIcon       string    `gorm:"size:512;not null;default:''" json:"processIcon"`
	ProcessType       string    `gorm:"size:32;not null;default:main" json:"processType"`
	ProcessVersion    int       `gorm:"not null;default:1;uniqueIndex:uk_flow_process_history_ver" json:"processVersion"`
	Remark            string    `gorm:"size:255;not null;default:''" json:"remark"`
	ModelContent      string    `gorm:"type:longtext" json:"modelContent"`
	ProcessForm       string    `gorm:"type:longtext" json:"processForm"`
	ProcessSetting    string    `gorm:"type:text" json:"processSetting"`
	ProcessPermission string    `gorm:"type:text" json:"processPermission"`
	CreatedBy         uint64    `gorm:"not null;default:0" json:"createdBy"`
	CreatedAt         time.Time `json:"createdAt"`
}

func (FlowProcessHistory) TableName() string { return "flow_process_history" }
