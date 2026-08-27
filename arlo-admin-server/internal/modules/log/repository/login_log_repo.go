package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/log/model"
	"arlo-admin/pkg/datascope"

	"gorm.io/gorm"
)

// LoginLogRepository 登录日志数据访问层
type LoginLogRepository struct {
	db *gorm.DB
}

// NewLoginLogRepository 创建 LoginLogRepository
func NewLoginLogRepository() *LoginLogRepository {
	return &LoginLogRepository{db: database.DB}
}

// Create 写入登录日志
func (r *LoginLogRepository) Create(ctx context.Context, log *model.LoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// List 分页查询登录日志（按数据权限过滤）
func (r *LoginLogRepository) List(ctx context.Context, username string, status *int8, startTime, endTime string, scope *datascope.Provider, page, pageSize int) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64

	q := r.db.WithContext(ctx).Model(&model.LoginLog{})
	if scope != nil {
		q = scope.ApplyByUsername(q, "username")
	}
	if username != "" {
		q = q.Where("username LIKE ?", "%"+username+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if startTime != "" {
		q = q.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		q = q.Where("created_at <= ?", endTime)
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
func (r *LoginLogRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < DATE_SUB(NOW(), INTERVAL ? DAY)", days).
		Delete(&model.LoginLog{})
	return result.RowsAffected, result.Error
}
