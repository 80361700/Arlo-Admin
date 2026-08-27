package repository

import (
	"context"
	"time"

	"arlo-admin/internal/database"
	"arlo-admin/internal/domain/model"
	"arlo-admin/pkg/datascope"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{db: database.DB}
}

// DB 暴露底层连接（供数据权限等跨模块构建使用）
func (r *UserRepository) DB() *gorm.DB {
	return r.db
}

// --- 用户基础 CRUD ---

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Select("*").Omit("password", "created_at").Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string, excludeID uint64) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}

// UpdatePassword 更新密码；selfChange=true 时清除强制改密并刷新 pwd_updated_at
func (r *UserRepository) UpdatePassword(ctx context.Context, id uint64, password string, selfChange bool) error {
	now := time.Now()
	updates := map[string]interface{}{
		"password":       password,
		"pwd_updated_at": now,
	}
	if selfChange {
		updates["must_change_pwd"] = 0
	} else {
		updates["must_change_pwd"] = 1
	}
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("last_login", &now).Error
}

// UpdateProfile 更新个人资料（不含角色/状态等管理字段；部门不可自改）
func (r *UserRepository) UpdateProfile(ctx context.Context, userID uint64, nickname string, gender int8, phone, email, remark, avatar string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"nickname": nickname,
		"gender":   gender,
		"phone":    phone,
		"email":    email,
		"remark":   remark,
		"avatar":   avatar,
	}).Error
}

// FindDeptName 查询部门名称
func (r *UserRepository) FindDeptName(ctx context.Context, deptID uint64) string {
	if deptID == 0 {
		return ""
	}
	var name string
	_ = r.db.WithContext(ctx).Table("sys_dept").Select("name").Where("id = ? AND deleted_at IS NULL", deptID).Scan(&name).Error
	return name
}

// FindPostNamesByUserID 查询用户岗位名称列表
func (r *UserRepository) FindPostNamesByUserID(ctx context.Context, userID uint64) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).
		Table("sys_user_post").
		Select("sys_post.name").
		Joins("JOIN sys_post ON sys_post.id = sys_user_post.post_id AND sys_post.deleted_at IS NULL").
		Where("sys_user_post.user_id = ?", userID).
		Pluck("sys_post.name", &names).Error
	if err != nil {
		return nil, err
	}
	return names, nil
}

// List 用户分页列表（支持多条件筛选 + 数据权限过滤）
func (r *UserRepository) List(ctx context.Context, username, nickname, phone string, status *int8, deptID *uint64, scope *datascope.Provider, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	q := r.db.WithContext(ctx).Model(&model.User{})

	// 数据权限：部门类范围按 dept_id；仅本人只能看自己的账号
	if scope != nil {
		switch scope.Scope {
		case datascope.ScopeSelf:
			q = q.Where("id = ?", scope.UserID)
		default:
			q = scope.Apply(q) // 默认 DeptColumn=dept_id
		}
	}

	if username != "" {
		q = q.Where("username LIKE ?", "%"+username+"%")
	}
	if nickname != "" {
		q = q.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	if phone != "" {
		q = q.Where("phone LIKE ?", "%"+phone+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if deptID != nil {
		q = q.Where("dept_id = ?", *deptID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// --- 用户角色关联 ---

func (r *UserRepository) AssignRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		userRoles := make([]model.UserRole, 0, len(roleIDs))
		for _, rid := range roleIDs {
			userRoles = append(userRoles, model.UserRole{UserID: userID, RoleID: rid})
		}
		return tx.Create(&userRoles).Error
	})
}

func (r *UserRepository) FindUserRoleIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	var urs []model.UserRole
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&urs).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(urs))
	for _, ur := range urs {
		ids = append(ids, ur.RoleID)
	}
	return ids, nil
}

func (r *UserRepository) FindRolesByUserIDs(ctx context.Context, userIDs []uint64) (map[uint64][]model.Role, error) {
	if len(userIDs) == 0 {
		return map[uint64][]model.Role{}, nil
	}
	var urs []model.UserRole
	if err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&urs).Error; err != nil {
		return nil, err
	}
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
	if len(roleIDs) == 0 {
		result := make(map[uint64][]model.Role, len(userIDs))
		for _, uid := range userIDs {
			result[uid] = []model.Role{}
		}
		return result, nil
	}
	var roles []model.Role
	if err := r.db.WithContext(ctx).Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}
	roleMap := make(map[uint64]model.Role, len(roles))
	for _, r := range roles {
		roleMap[r.ID] = r
	}
	result := make(map[uint64][]model.Role, len(userIDs))
	for _, uid := range userIDs {
		roleIDs := userRoleMap[uid]
		userRoles := make([]model.Role, 0, len(roleIDs))
		for _, rid := range roleIDs {
			if r, ok := roleMap[rid]; ok {
				userRoles = append(userRoles, r)
			}
		}
		result[uid] = userRoles
	}
	return result, nil
}

// --- 用户岗位关联 ---

func (r *UserRepository) AssignPosts(ctx context.Context, userID uint64, postIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserPost{}).Error; err != nil {
			return err
		}
		if len(postIDs) == 0 {
			return nil
		}
		userPosts := make([]model.UserPost, 0, len(postIDs))
		for _, pid := range postIDs {
			userPosts = append(userPosts, model.UserPost{UserID: userID, PostID: pid})
		}
		return tx.Create(&userPosts).Error
	})
}

func (r *UserRepository) FindUserPostIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	var ups []model.UserPost
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&ups).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(ups))
	for _, up := range ups {
		ids = append(ids, up.PostID)
	}
	return ids, nil
}

func (r *UserRepository) FindPostsByUserIDs(ctx context.Context, userIDs []uint64) (map[uint64][]model.Post, error) {
	if len(userIDs) == 0 {
		return map[uint64][]model.Post{}, nil
	}
	var ups []model.UserPost
	if err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&ups).Error; err != nil {
		return nil, err
	}
	if len(ups) == 0 {
		result := make(map[uint64][]model.Post, len(userIDs))
		for _, uid := range userIDs {
			result[uid] = []model.Post{}
		}
		return result, nil
	}
	postIDs := make([]uint64, 0, len(ups))
	postIDSet := make(map[uint64]bool)
	userPostMap := make(map[uint64][]uint64)
	for _, up := range ups {
		userPostMap[up.UserID] = append(userPostMap[up.UserID], up.PostID)
		if !postIDSet[up.PostID] {
			postIDSet[up.PostID] = true
			postIDs = append(postIDs, up.PostID)
		}
	}
	var posts []model.Post
	if err := r.db.WithContext(ctx).Where("id IN ?", postIDs).Find(&posts).Error; err != nil {
		return nil, err
	}
	postMap := make(map[uint64]model.Post, len(posts))
	for _, p := range posts {
		postMap[p.ID] = p
	}
	result := make(map[uint64][]model.Post, len(userIDs))
	for _, uid := range userIDs {
		uids := userPostMap[uid]
		userPosts := make([]model.Post, 0, len(uids))
		for _, pid := range uids {
			if p, ok := postMap[pid]; ok {
				userPosts = append(userPosts, p)
			}
		}
		result[uid] = userPosts
	}
	return result, nil
}
