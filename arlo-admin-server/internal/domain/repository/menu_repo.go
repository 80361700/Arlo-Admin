package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/domain/model"

	"gorm.io/gorm"
)

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository() *MenuRepository {
	return &MenuRepository{db: database.DB}
}

func (r *MenuRepository) Create(ctx context.Context, menu *model.Menu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

func (r *MenuRepository) Update(ctx context.Context, menu *model.Menu) error {
	return r.db.WithContext(ctx).Save(menu).Error
}

func (r *MenuRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("menu_id = ?", id).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Menu{}, id).Error
	})
}

func (r *MenuRepository) FindByID(ctx context.Context, id uint64) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.WithContext(ctx).First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *MenuRepository) FindAll(ctx context.Context) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

// FindPermissionsByRoleIDs 根据角色 ID 列表查询权限标识（菜单 type=2 + 按钮 type=3）
func (r *MenuRepository) FindPermissionsByRoleIDs(ctx context.Context, roleIDs []uint64) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}

	var permissions []string
	err := r.db.WithContext(ctx).
		Table("sys_role_menu").
		Select("DISTINCT sys_menu.permission").
		Joins("JOIN sys_menu ON sys_menu.id = sys_role_menu.menu_id AND sys_menu.status = 1").
		Where("sys_role_menu.role_id IN ? AND sys_menu.type IN (2, 3) AND sys_menu.permission != ''", roleIDs).
		Pluck("permission", &permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// FindMenusByRoleIDs 根据角色 ID 列表查询完整菜单树（前端使用）
func (r *MenuRepository) FindMenusByRoleIDs(ctx context.Context, roleIDs []uint64) ([]*model.Menu, error) {
	if len(roleIDs) == 0 {
		return []*model.Menu{}, nil
	}

	var menus []*model.Menu
	err := r.db.WithContext(ctx).
		Table("sys_menu").
		Select("DISTINCT sys_menu.*").
		Joins("JOIN sys_role_menu ON sys_role_menu.menu_id = sys_menu.id").
		Where("sys_role_menu.role_id IN ? AND sys_menu.status = 1 AND sys_menu.visible = 1", roleIDs).
		Order("sys_menu.sort ASC").
		Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

// FindAllVisible 查询全部启用且可见的菜单（用于补全父级目录）
func (r *MenuRepository) FindAllVisible(ctx context.Context) ([]*model.Menu, error) {
	var menus []*model.Menu
	err := r.db.WithContext(ctx).
		Where("status = 1 AND visible = 1").
		Order("sort ASC").
		Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) HasChildren(ctx context.Context, parentID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Menu{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count > 0, err
}

// FindDescendantIDs 递归查找某菜单的所有子孙 ID（含自身）
func (r *MenuRepository) FindDescendantIDs(ctx context.Context, menuID uint64) ([]uint64, error) {
	menus, err := r.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	childrenMap := make(map[uint64][]uint64)
	for _, m := range menus {
		childrenMap[m.ParentID] = append(childrenMap[m.ParentID], m.ID)
	}
	var result []uint64
	var walk func(uint64)
	walk = func(id uint64) {
		result = append(result, id)
		for _, child := range childrenMap[id] {
			walk(child)
		}
	}
	walk(menuID)
	return result, nil
}
