package storage

import (
	"context"
	"fmt"
	"io"

	"arlo-admin/internal/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSDriver 阿里云OSS存储驱动
type OSDriver struct {
	client *oss.Client
	bucket *oss.Bucket
	cfg    *config.OSSStorageConfig
}

// NewOSDriver 创建阿里云 OSS 存储驱动
func NewOSDriver(cfg *config.OSSStorageConfig) (*OSDriver, error) {
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.BucketName == "" {
		return nil, fmt.Errorf("oss config incomplete: endpoint/accessKeyId/bucketName required")
	}

	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("oss client init failed: %w", err)
	}

	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("oss bucket init failed: %w", err)
	}

	return &OSDriver{client: client, bucket: bucket, cfg: cfg}, nil
}

// Save 上传文件到OSS
func (d *OSDriver) Save(_ context.Context, key string, reader io.Reader) error {
	return d.bucket.PutObject(key, reader)
}

// Open 从OSS下载文件流
func (d *OSDriver) Open(_ context.Context, key string) (io.ReadCloser, error) {
	body, err := d.bucket.GetObject(key)
	if err != nil {
		// 区分"文件不存在"错误
		if ossErr, ok := err.(oss.ServiceError); ok && ossErr.StatusCode == 404 {
			return nil, fmt.Errorf("oss object not found: %s", key)
		}
		return nil, fmt.Errorf("oss download failed: %w", err)
	}
	return body, nil
}

// Remove 从OSS删除文件
func (d *OSDriver) Remove(_ context.Context, key string) error {
	err := d.bucket.DeleteObject(key)
	if err != nil {
		return fmt.Errorf("oss delete failed: %w", err)
	}
	return nil
}
