package model

import (
	"time"

	"gorm.io/gorm"
)

// SysFile 文件模型，对应 sys_file 表
type SysFile struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	AccessKey  string         `gorm:"size:32;not null;uniqueIndex" json:"accessKey"`     // 访问钥（URL 用，不可猜）
	Name       string         `gorm:"size:128;not null" json:"name"`                     // 原始文件名
	Path       string         `gorm:"size:255;not null" json:"-"`                        // 存储路径（不返回给前端）
	Size       int64          `gorm:"not null;default:0" json:"size"`                    // 文件大小(字节)
	MimeType   string         `gorm:"size:128;not null" json:"mimeType"`                 // MIME类型
	Category   string         `gorm:"size:32;not null;default:other" json:"category"`    // 文件分类: image/video/audio/document/other
	IsPublic   int8           `gorm:"not null;default:1;index" json:"isPublic"`          // 是否公开（默认公开，可改为私有）
	MD5        string         `gorm:"size:32;not null;index" json:"md5"`                 // MD5校验值
	UploaderID uint64         `gorm:"not null;default:0;index" json:"uploaderId"`
	Uploader   string         `gorm:"size:32;not null" json:"uploader"`
	CreatedAt  time.Time      `json:"createdAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SysFile) TableName() string { return "sys_file" }
