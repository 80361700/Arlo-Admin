package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/demo/model"

	"gorm.io/gorm"
)

type OrderRepository struct{}

func NewOrderRepository() *OrderRepository { return &OrderRepository{} }

func (r *OrderRepository) db() *gorm.DB { return database.DB }

func (r *OrderRepository) Create(ctx context.Context, o *model.DemoPurchaseOrder) error {
	return r.db().WithContext(ctx).Create(o).Error
}

func (r *OrderRepository) Update(ctx context.Context, o *model.DemoPurchaseOrder) error {
	return r.db().WithContext(ctx).Save(o).Error
}

func (r *OrderRepository) Get(ctx context.Context, id uint64) (*model.DemoPurchaseOrder, error) {
	var o model.DemoPurchaseOrder
	if err := r.db().WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) Delete(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db().WithContext(ctx).Where("id IN ?", ids).Delete(&model.DemoPurchaseOrder{}).Error
}

func (r *OrderRepository) List(ctx context.Context, title string, status *int8, page, pageSize int) ([]model.DemoPurchaseOrder, int64, error) {
	q := r.db().WithContext(ctx).Model(&model.DemoPurchaseOrder{})
	if title != "" {
		q = q.Where("title LIKE ?", "%"+title+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.DemoPurchaseOrder
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
