package storage

import (
	"context"
	"fmt"
	"io"

	"arlo-admin/internal/config"
)

// Driver 存储驱动接口 — 所有存储后端（本地/OSS/...）统一实现此接口
type Driver interface {
	// Save 保存文件（key: 存储路径/对象键），返回 nil 表示成功
	Save(ctx context.Context, key string, reader io.Reader) error

	// Open 打开文件，返回可读流；调用方负责 Close
	Open(ctx context.Context, key string) (io.ReadCloser, error)

	// Remove 删除文件；文件不存在应返回 nil（幂等）
	Remove(ctx context.Context, key string) error
}

// NewDriver 根据配置创建存储驱动实例
//
// driver=local  → LocalDriver（本地磁盘）
// driver=oss    → OSSDriver（阿里云OSS）
// 其他值       → 返回 error
func NewDriver(cfg *config.StorageConfig) (Driver, error) {
	switch cfg.Driver {
	case "local":
		return NewLocalDriver(cfg.Local.Path), nil
	case "oss":
		return NewOSDriver(&cfg.OSS)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}
}
