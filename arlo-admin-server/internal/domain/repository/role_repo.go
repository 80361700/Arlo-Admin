package repository

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/domain/model"

	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{db: database.DB}
}

func (r *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *RoleRepository) Update(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *RoleRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&model.RoleDept{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Role{}, id).Error
	})
}

func (r *RoleRepository) FindByID(ctx context.Context, id uint64) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) FindAll(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&roles).Error
	return roles, err
}

// FindRolesByUserID 查询用户拥有的所有角色（启用状态）
func (r *RoleRepository) FindRolesByUserID(ctx context.Context, userID uint64) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN sys_user_role ON sys_user_role.role_id = sys_role.id").
		Where("sys_user_role.user_id = ? AND sys_role.status = 1", userID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// FindAllRoles 查询所有启用状态的角色（Casbin 策略加载用）
func (r *RoleRepository) FindAllRoles(ctx context.Context) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.WithContext(ctx).Where("status = 1").Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// RoleMenuResult 角色菜单关联（Casbin 策略加载用）
type RoleMenuResult struct {
	RoleCode   string
	MenuPath   string
	Permission string
}

// FindAllRoleMenus 查询所有角色菜单关联，用于构建 Casbin 策略
// 仅加载 type=2（菜单页），目录 path（如 /system）过宽会覆盖整个模块 API
func (r *RoleRepository) FindAllRoleMenus(ctx context.Context) ([]RoleMenuResult, error) {
	var results []RoleMenuResult
	err := r.db.WithContext(ctx).
		Table("sys_role_menu").
		Select("sys_role.code AS role_code, sys_menu.path AS menu_path, sys_menu.permission").
		Joins("JOIN sys_role ON sys_role.id = sys_role_menu.role_id AND sys_role.status = 1").
		Joins("JOIN sys_menu ON sys_menu.id = sys_role_menu.menu_id AND sys_menu.status = 1").
		Where("sys_menu.type = 2 AND sys_menu.path != ''").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// UserRoleResult 用户角色关联（Casbin 分组加载用）
type UserRoleResult struct {
	UserID   uint64
	RoleCode string
}

// FindAllUserRoles 查询所有用户角色关系
func (r *RoleRepository) FindAllUserRoles(ctx context.Context) ([]UserRoleResult, error) {
	var results []UserRoleResult
	err := r.db.WithContext(ctx).
		Table("sys_user_role").
		Select("sys_user_role.user_id, sys_role.code AS role_code").
		Joins("JOIN sys_role ON sys_role.id = sys_user_role.role_id AND sys_role.status = 1").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *RoleRepository) ExistsByCode(ctx context.Context, code string, excludeID uint64) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.Role{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *RoleRepository) List(ctx context.Context, name, code string, status *int8, page, pageSize int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Role{})
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
	if err := q.Order("sort ASC, id ASC").Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

// --- 角色菜单关联 ---

func (r *RoleRepository) FindRoleMenuIDs(ctx context.Context, roleID uint64) ([]uint64, error) {
	var rms []model.RoleMenu
	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&rms).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(rms))
	for _, rm := range rms {
		ids = append(ids, rm.MenuID)
	}
	return ids, nil
}

func (r *RoleRepository) AssignMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		if len(menuIDs) == 0 {
			return nil
		}
		rms := make([]model.RoleMenu, 0, len(menuIDs))
		for _, mid := range menuIDs {
			rms = append(rms, model.RoleMenu{RoleID: roleID, MenuID: mid})
		}
		return tx.Create(&rms).Error
	})
}

func (r *RoleRepository) HasUserAssigned(ctx context.Context, roleID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserRole{}).Where("role_id = ?", roleID).Count(&count).Error
	return count > 0, err
}

// --- 角色部门关联（数据权限） ---

// FindRoleDeptIDs 查询角色关联的部门ID列表
func (r *RoleRepository) FindRoleDeptIDs(ctx context.Context, roleID uint64) ([]uint64, error) {
	var rds []model.RoleDept
	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&rds).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(rds))
	for _, rd := range rds {
		ids = append(ids, rd.DeptID)
	}
	return ids, nil
}

// AssignDepts 分配角色数据权限部门（先删后插）
func (r *RoleRepository) AssignDepts(ctx context.Context, roleID uint64, deptIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleDept{}).Error; err != nil {
			return err
		}
		if len(deptIDs) == 0 {
			return nil
		}
		rds := make([]model.RoleDept, 0, len(deptIDs))
		for _, did := range deptIDs {
			rds = append(rds, model.RoleDept{RoleID: roleID, DeptID: did})
		}
		return tx.Create(&rds).Error
	})
}

// FindRolesByUserIDs 批量查询用户角色（含 data_scope 字段）
func (r *RoleRepository) FindRolesByUserIDs(ctx context.Context, userIDs []uint64) (map[uint64][]model.Role, error) {
	if len(userIDs) == 0 {
		return map[uint64][]model.Role{}, nil
	}

	// 1. 查 user_role 关联
	var urs []model.UserRole
	if err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&urs).Error; err != nil {
		return nil, err
	}
	if len(urs) == 0 {
		result := make(map[uint64][]model.Role, len(userIDs))
		for _, uid := range userIDs {
			result[uid] = []model.Role{}
		}
		return result, nil
	}

	// 2. 收集 roleID 并去重
	roleIDs := make([]uint64, 0, len(urs))
	roleIDSet := make(map[uint64]bool)
	userRoleMap := make(map[uint64][]uint64)
	for _, ur := range urs {
		userRoleMap[ur.UserID] = append(userRoleMap[ur.UserID], ur.RoleID)
		if !roleIDSet[ur.RoleID] {
			roleIDSet[ur.RoleID] = true
			roleIDs = append(roleIDs, ur.RoleID)
		}
	}

	// 3. 查角色（含 data_scope）
	var roles []model.Role
	if err := r.db.WithContext(ctx).Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}
	roleMap := make(map[uint64]model.Role, len(roles))
	for _, r := range roles {
		roleMap[r.ID] = r
	}

	// 4. 组装结果
	result := make(map[uint64][]model.Role, len(userIDs))
	for _, uid := range userIDs {
		rids := userRoleMap[uid]
		userRoles := make([]model.Role, 0, len(rids))
		for _, rid := range rids {
			if r, ok := roleMap[rid]; ok {
				userRoles = append(userRoles, r)
			}
		}
		result[uid] = userRoles
	}
	return result, nil
}
