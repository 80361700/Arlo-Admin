package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/log/model"
	"arlo-admin/pkg/datascope"

	"gorm.io/gorm"
)

// OperationLogRepository 操作日志数据访问层
type OperationLogRepository struct {
	db *gorm.DB
}

// NewOperationLogRepository 创建 OperationLogRepository
func NewOperationLogRepository() *OperationLogRepository {
	return &OperationLogRepository{db: database.DB}
}

// Create 写入操作日志
func (r *OperationLogRepository) Create(ctx context.Context, log *model.OperationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// List 分页查询操作日志（按数据权限过滤）
func (r *OperationLogRepository) List(ctx context.Context, username, module, url, startTime, endTime string, status *int8, scope *datascope.Provider, page, pageSize int) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	q := r.db.WithContext(ctx).Model(&model.OperationLog{})
	if scope != nil {
		q = scope.ApplyByOwner(q, "user_id")
	}
	if username != "" {
		q = q.Where("username LIKE ?", "%"+username+"%")
	}
	if module != "" {
		q = q.Where("module LIKE ?", "%"+module+"%")
	}
	if url != "" {
		q = q.Where("url LIKE ?", "%"+url+"%")
	}
	if startTime != "" {
		q = q.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		q = q.Where("created_at <= ?", endTime)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteOlderThan 删除N天前的日志
func (r *OperationLogRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < DATE_SUB(NOW(), INTERVAL ? DAY)", days).
		Delete(&model.OperationLog{})
	return result.RowsAffected, result.Error
}
