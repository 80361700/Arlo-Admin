package model

import (
	"time"

	"gorm.io/gorm"
)

// SysJob 定时任务
type SysJob struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string         `gorm:"size:64;not null" json:"name"`
	Handler    string         `gorm:"size:64;not null;index" json:"handler"`
	Cron       string         `gorm:"size:64;not null" json:"cron"`
	Params     string         `gorm:"size:512;not null;default:''" json:"params"`
	Status     int8           `gorm:"not null;default:1;index" json:"status"` // 0暂停 1启用
	Remark     string         `gorm:"size:255;not null;default:''" json:"remark"`
	LastRunAt  *time.Time     `json:"lastRunAt"`
	LastStatus *int8          `json:"lastStatus"` // 0失败 1成功
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SysJob) TableName() string { return "sys_job" }

// SysJobLog 执行日志
type SysJobLog struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	JobID       uint64    `gorm:"not null;index" json:"jobId"`
	JobName     string    `gorm:"size:64;not null;default:''" json:"jobName"`
	Handler     string    `gorm:"size:64;not null;default:''" json:"handler"`
	TriggerType int8      `gorm:"not null;default:0" json:"triggerType"` // 0调度 1手动
	Status      int8      `gorm:"not null;default:1" json:"status"`      // 0失败 1成功
	Result      string    `gorm:"type:text" json:"result"`
	ErrorMsg    string    `gorm:"size:1000;not null;default:''" json:"errorMsg"`
	DurationMs  int       `gorm:"not null;default:0" json:"durationMs"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (SysJobLog) TableName() string { return "sys_job_log" }
