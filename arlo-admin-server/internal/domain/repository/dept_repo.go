package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/domain/model"

	"gorm.io/gorm"
)

type DeptRepository struct {
	db *gorm.DB
}

func NewDeptRepository() *DeptRepository {
	return &DeptRepository{db: database.DB}
}

func (r *DeptRepository) Create(ctx context.Context, dept *model.Dept) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *DeptRepository) Update(ctx context.Context, dept *model.Dept) error {
	return r.db.WithContext(ctx).Save(dept).Error
}

func (r *DeptRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Dept{}, id).Error
}

func (r *DeptRepository) FindByID(ctx context.Context, id uint64) (*model.Dept, error) {
	var dept model.Dept
	err := r.db.WithContext(ctx).First(&dept, id).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// FindByIDs 批量查询部门，返回 deptID → name 映射
func (r *DeptRepository) FindByIDs(ctx context.Context, ids []uint64) (map[uint64]string, error) {
	result := make(map[uint64]string)
	if len(ids) == 0 {
		return result, nil
	}
	var depts []model.Dept
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&depts).Error
	if err != nil {
		return result, err
	}
	for _, d := range depts {
		result[d.ID] = d.Name
	}
	return result, nil
}

func (r *DeptRepository) FindAll(ctx context.Context) ([]model.Dept, error) {
	var depts []model.Dept
	err := r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&depts).Error
	return depts, err
}

func (r *DeptRepository) HasChildren(ctx context.Context, parentID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Dept{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count > 0, err
}

// FindDescendantIDs 递归查找某个部门的所有子孙部门ID（含自身）
// 用于 data_scope=3（本部门及以下）的权限过滤
func (r *DeptRepository) FindDescendantIDs(ctx context.Context, deptID uint64) ([]uint64, error) {
	depts, err := r.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// 构建 parentID → children 索引
	childrenMap := make(map[uint64][]uint64)
	for _, d := range depts {
		childrenMap[d.ParentID] = append(childrenMap[d.ParentID], d.ID)
	}

	// 递归收集
	var result []uint64
	var collect func(pid uint64)
	collect = func(pid uint64) {
		result = append(result, pid)
		for _, child := range childrenMap[pid] {
			collect(child)
		}
	}
	collect(deptID)
	return result, nil
}
