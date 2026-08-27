package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/system/model"

	"gorm.io/gorm"
)

type DictRepository struct {
	db *gorm.DB
}

func NewDictRepository() *DictRepository {
	return &DictRepository{db: database.DB}
}

// --- 字典类型 ---

func (r *DictRepository) CreateDictType(ctx context.Context, dt *model.DictType) error {
	return r.db.WithContext(ctx).Create(dt).Error
}

func (r *DictRepository) UpdateDictType(ctx context.Context, dt *model.DictType) error {
	return r.db.WithContext(ctx).Save(dt).Error
}

func (r *DictRepository) DeleteDictType(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dict_type_id = ?", id).Delete(&model.DictData{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.DictType{}, id).Error
	})
}

func (r *DictRepository) FindDictTypeByID(ctx context.Context, id uint64) (*model.DictType, error) {
	var dt model.DictType
	err := r.db.WithContext(ctx).First(&dt, id).Error
	if err != nil {
		return nil, err
	}
	return &dt, nil
}

func (r *DictRepository) FindDictTypeByCode(ctx context.Context, code string) (*model.DictType, error) {
	var dt model.DictType
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&dt).Error
	if err != nil {
		return nil, err
	}
	return &dt, nil
}

func (r *DictRepository) FindAllDictTypes(ctx context.Context) ([]model.DictType, error) {
	var dts []model.DictType
	err := r.db.WithContext(ctx).Order("id ASC").Find(&dts).Error
	return dts, err
}

func (r *DictRepository) ExistsDictTypeByCode(ctx context.Context, code string, excludeID uint64) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.DictType{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *DictRepository) ListDictTypes(ctx context.Context, name, code string, status *int8, page, pageSize int) ([]model.DictType, int64, error) {
	var dts []model.DictType
	var total int64
	q := r.db.WithContext(ctx).Model(&model.DictType{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		q = q.Where("code LIKE ?", "%"+code+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&dts).Error; err != nil {
		return nil, 0, err
	}
	return dts, total, nil
}

// --- 字典数据 ---

func (r *DictRepository) CreateDictData(ctx context.Context, dd *model.DictData) error {
	return r.db.WithContext(ctx).Create(dd).Error
}

func (r *DictRepository) UpdateDictData(ctx context.Context, dd *model.DictData) error {
	return r.db.WithContext(ctx).Save(dd).Error
}

func (r *DictRepository) DeleteDictData(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.DictData{}, id).Error
}

func (r *DictRepository) FindDictDataByID(ctx context.Context, id uint64) (*model.DictData, error) {
	var dd model.DictData
	err := r.db.WithContext(ctx).First(&dd, id).Error
	if err != nil {
		return nil, err
	}
	return &dd, nil
}

func (r *DictRepository) FindDictDataByTypeID(ctx context.Context, dictTypeID uint64) ([]model.DictData, error) {
	var dds []model.DictData
	err := r.db.WithContext(ctx).Where("dict_type_id = ?", dictTypeID).Order("sort ASC, id ASC").Find(&dds).Error
	return dds, err
}

// FindEnabledDictDataByTypeID 仅启用项（供业务下拉）
func (r *DictRepository) FindEnabledDictDataByTypeID(ctx context.Context, dictTypeID uint64) ([]model.DictData, error) {
	var dds []model.DictData
	err := r.db.WithContext(ctx).
		Where("dict_type_id = ? AND status = 1", dictTypeID).
		Order("sort ASC, id ASC").
		Find(&dds).Error
	return dds, err
}

func (r *DictRepository) ListDictDatas(ctx context.Context, dictTypeID *uint64, label string, status *int8, page, pageSize int) ([]model.DictData, int64, error) {
	var dds []model.DictData
	var total int64
	q := r.db.WithContext(ctx).Model(&model.DictData{})
	if dictTypeID != nil {
		q = q.Where("dict_type_id = ?", *dictTypeID)
	}
	if label != "" {
		q = q.Where("label LIKE ?", "%"+label+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Order("sort ASC, id ASC").Offset(offset).Limit(pageSize).Find(&dds).Error; err != nil {
		return nil, 0, err
	}
	return dds, total, nil
}
