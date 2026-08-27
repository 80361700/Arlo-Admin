package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// LocalDriver 本地磁盘存储驱动
type LocalDriver struct {
	basePath string // 存储根目录，如 "uploads/"
}

// NewLocalDriver 创建本地存储驱动
func NewLocalDriver(basePath string) *LocalDriver {
	if basePath == "" {
		basePath = "uploads/"
	}
	return &LocalDriver{basePath: basePath}
}

// Save 将数据保存到本地磁盘；自动创建父目录
func (d *LocalDriver) Save(_ context.Context, key string, reader io.Reader) error {
	fullPath := filepath.Join(d.basePath, key)

	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		os.Remove(fullPath) // 写入失败时清理
		return err
	}

	return nil
}

// Open 打开本地文件
func (d *LocalDriver) Open(_ context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(d.basePath, key)
	return os.Open(fullPath)
}

// Remove 删除本地文件（文件不存在不算错误）
func (d *LocalDriver) Remove(_ context.Context, key string) error {
	fullPath := filepath.Join(d.basePath, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
