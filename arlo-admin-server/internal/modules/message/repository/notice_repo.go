package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/message/model"
	"arlo-admin/pkg/datascope"

	"gorm.io/gorm"
)

// NoticeRepository 通知公告数据访问层
type NoticeRepository struct {
	db *gorm.DB
}

// NewNoticeRepository 创建 NoticeRepository
func NewNoticeRepository() *NoticeRepository {
	return &NoticeRepository{db: database.DB}
}

func (r *NoticeRepository) Create(ctx context.Context, notice *model.Notice) error {
	return r.db.WithContext(ctx).Create(notice).Error
}

func (r *NoticeRepository) Update(ctx context.Context, notice *model.Notice) error {
	return r.db.WithContext(ctx).Save(notice).Error
}

func (r *NoticeRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Notice{}, id).Error
}

func (r *NoticeRepository) FindByID(ctx context.Context, id uint64) (*model.Notice, error) {
	var notice model.Notice
	err := r.db.WithContext(ctx).First(&notice, id).Error
	if err != nil {
		return nil, err
	}
	return &notice, nil
}

func (r *NoticeRepository) List(ctx context.Context, title string, status *int8, scope *datascope.Provider, page, pageSize int) ([]model.Notice, int64, error) {
	var notices []model.Notice
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Notice{})
	if scope != nil {
		q = scope.ApplyByOwner(q, "publisher_id")
	}
	if title != "" {
		q = q.Where("title LIKE ?", "%"+title+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&notices).Error; err != nil {
		return nil, 0, err
	}

	return notices, total, nil
}

func (r *NoticeRepository) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	return r.db.WithContext(ctx).Model(&model.Notice{}).Where("id = ?", id).Update("status", status).Error
}
