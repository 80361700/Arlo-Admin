package datascope

import (
	"context"

	"gorm.io/gorm"
)

// DBSource 基于 GORM 直查系统表的数据权限数据源
type DBSource struct {
	DB *gorm.DB
}

func (s *DBSource) FindUserRoles(ctx context.Context, userID uint64) ([]RoleScope, error) {
	var rows []RoleScope
	err := s.DB.WithContext(ctx).
		Table("sys_role r").
		Select("r.id AS id, r.data_scope AS data_scope").
		Joins("JOIN sys_user_role ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.status = 1 AND r.deleted_at IS NULL", userID).
		Scan(&rows).Error
	return rows, err
}

func (s *DBSource) FindRoleDeptIDs(ctx context.Context, roleID uint64) ([]uint64, error) {
	var ids []uint64
	err := s.DB.WithContext(ctx).
		Table("sys_role_dept").
		Where("role_id = ?", roleID).
		Pluck("dept_id", &ids).Error
	return ids, err
}

func (s *DBSource) FindUserDeptID(ctx context.Context, userID uint64) (uint64, error) {
	var deptID uint64
	err := s.DB.WithContext(ctx).
		Table("sys_user").
		Select("dept_id").
		Where("id = ? AND deleted_at IS NULL", userID).
		Scan(&deptID).Error
	return deptID, err
}

func (s *DBSource) FindDescendantDeptIDs(ctx context.Context, deptID uint64) ([]uint64, error) {
	type row struct {
		ID       uint64
		ParentID uint64
	}
	var depts []row
	if err := s.DB.WithContext(ctx).
		Table("sys_dept").
		Select("id, parent_id").
		Where("deleted_at IS NULL").
		Scan(&depts).Error; err != nil {
		return nil, err
	}
	children := make(map[uint64][]uint64)
	for _, d := range depts {
		children[d.ParentID] = append(children[d.ParentID], d.ID)
	}
	var result []uint64
	var walk func(uint64)
	walk = func(id uint64) {
		result = append(result, id)
		for _, c := range children[id] {
			walk(c)
		}
	}
	walk(deptID)
	return result, nil
}

// BuildFromDB 便捷方法：用数据库构建当前用户的数据权限
func BuildFromDB(ctx context.Context, db *gorm.DB, userID uint64) (*Provider, error) {
	return Build(ctx, &DBSource{DB: db}, userID)
}
