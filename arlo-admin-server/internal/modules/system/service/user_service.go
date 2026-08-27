package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"arlo-admin/internal/domain/model"
	"arlo-admin/internal/domain/repository"
	configsvc "arlo-admin/internal/modules/sysconfig/service"
	"arlo-admin/internal/modules/system/dto"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/datascope"
	perrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/security"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户管理服务
type UserService struct {
	userRepo  *repository.UserRepository
	roleRepo  *repository.RoleRepository
	deptRepo  *repository.DeptRepository
	postRepo  *repository.PostRepository
	enforcer  *casbinpkg.Enforcer
	configSvc *configsvc.ConfigService
}

func NewUserService(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	deptRepo *repository.DeptRepository,
	postRepo *repository.PostRepository,
	enforcer *casbinpkg.Enforcer,
	configSvc *configsvc.ConfigService,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		deptRepo:  deptRepo,
		postRepo:  postRepo,
		enforcer:  enforcer,
		configSvc: configSvc,
	}
}

// List 分页查询用户列表（自动应用数据权限）
func (s *UserService) List(ctx context.Context, req *dto.UserListRequest, currentUserID uint64) (*dto.PageResponse, error) {
	// 构建数据权限过滤器
	scope, _ := s.buildDataScope(ctx, currentUserID)

	users, total, err := s.userRepo.List(ctx, req.Username, req.Nickname, req.Phone, req.Status, req.DeptID, scope, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]dto.UserResponse, 0, len(users))
	if len(users) > 0 {
		userIDs := make([]uint64, len(users))
		deptIDs := make([]uint64, 0, len(users))
		deptIDSet := make(map[uint64]bool)
		for i, u := range users {
			userIDs[i] = u.ID
			if u.DeptID > 0 && !deptIDSet[u.DeptID] {
				deptIDSet[u.DeptID] = true
				deptIDs = append(deptIDs, u.DeptID)
			}
		}
		roleMap, _ := s.userRepo.FindRolesByUserIDs(ctx, userIDs)
		postMap, _ := s.userRepo.FindPostsByUserIDs(ctx, userIDs)
		deptNameMap, _ := s.deptRepo.FindByIDs(ctx, deptIDs)
		for _, u := range users {
			roles := roleMap[u.ID]
			posts := postMap[u.ID]
			roleNames := make([]string, len(roles))
			roleIDs := make([]uint64, len(roles))
			for i, r := range roles {
				roleNames[i] = r.Name
				roleIDs[i] = r.ID
			}
			postNames := make([]string, len(posts))
			postIDs := make([]uint64, len(posts))
			for i, p := range posts {
				postNames[i] = p.Name
				postIDs[i] = p.ID
			}
			list = append(list, dto.UserResponse{
				ID:        u.ID,
				Username:  u.Username,
				Nickname:  u.Nickname,
				Avatar:    u.Avatar,
				Email:     u.Email,
				Phone:     u.Phone,
				Gender:    u.Gender,
				DeptID:    u.DeptID,
				DeptName:  deptNameMap[u.DeptID],
				Status:    u.Status,
				Remark:    u.Remark,
				RoleIDs:   roleIDs,
				RoleNames: roleNames,
				PostIDs:   postIDs,
				PostNames: postNames,
				CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}
	return &dto.PageResponse{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// GetDetail 获取用户详情
func (s *UserService) GetDetail(ctx context.Context, id uint64) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, perrors.New(perrors.ErrUserNotFound, "用户不存在")
		}
		return nil, err
	}
	roleIDs, _ := s.userRepo.FindUserRoleIDs(ctx, id)
	postIDs, _ := s.userRepo.FindUserPostIDs(ctx, id)
	roleMap, _ := s.userRepo.FindRolesByUserIDs(ctx, []uint64{id})
	postMap, _ := s.userRepo.FindPostsByUserIDs(ctx, []uint64{id})

	roles := roleMap[id]
	posts := postMap[id]
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}
	postNames := make([]string, len(posts))
	for i, p := range posts {
		postNames[i] = p.Name
	}
	deptName := ""
	if user.DeptID > 0 {
		deptNameMap, _ := s.deptRepo.FindByIDs(ctx, []uint64{user.DeptID})
		deptName = deptNameMap[user.DeptID]
	}
	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Email:     user.Email,
		Phone:     user.Phone,
		Gender:    user.Gender,
		DeptID:    user.DeptID,
		DeptName:  deptName,
		Status:    user.Status,
		Remark:    user.Remark,
		RoleIDs:   roleIDs,
		RoleNames: roleNames,
		PostIDs:   postIDs,
		PostNames: postNames,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// Create 创建用户
func (s *UserService) Create(ctx context.Context, req *dto.CreateUserRequest) error {
	req.Username = strings.TrimSpace(req.Username)
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Username == "" {
		return perrors.New(perrors.BadRequest, "用户名不能为空")
	}
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username, 0)
	if err != nil {
		return err
	}
	if exists {
		return perrors.New(perrors.ErrUserExists, fmt.Sprintf("用户名 %s 已存在", req.Username))
	}
	if err := s.validatePassword(ctx, req.Password); err != nil {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now()
	user := &model.User{
		Username:      req.Username,
		Password:      string(hashed),
		Nickname:      req.Nickname,
		Avatar:        req.Avatar,
		Email:         req.Email,
		Phone:         req.Phone,
		Gender:        req.Gender,
		DeptID:        req.DeptID,
		Status:        req.Status,
		Remark:        req.Remark,
		PwdUpdatedAt:  &now,
		MustChangePwd: 1,
	}
	if user.Status == 0 {
		user.Status = 1
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}
	if len(req.RoleIDs) > 0 {
		_ = s.userRepo.AssignRoles(ctx, user.ID, req.RoleIDs)
	}
	if len(req.PostIDs) > 0 {
		_ = s.userRepo.AssignPosts(ctx, user.ID, req.PostIDs)
	}
	s.reloadPolicies(ctx)
	return nil
}

// Update 更新用户
func (s *UserService) Update(ctx context.Context, req *dto.UpdateUserRequest) error {
	user, err := s.userRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrUserNotFound, "用户不存在")
		}
		return err
	}
	exists, err := s.userRepo.ExistsByUsername(ctx, user.Username, req.ID)
	if err != nil {
		return err
	}
	if exists {
		return perrors.New(perrors.ErrUserExists, "用户名重复")
	}
	user.Nickname = req.Nickname
	user.Avatar = req.Avatar
	user.Email = req.Email
	user.Phone = req.Phone
	user.Gender = req.Gender
	user.DeptID = req.DeptID
	user.Status = req.Status
	user.Remark = req.Remark
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}
	if err := s.userRepo.AssignRoles(ctx, req.ID, req.RoleIDs); err != nil {
		return err
	}
	if err := s.userRepo.AssignPosts(ctx, req.ID, req.PostIDs); err != nil {
		return err
	}
	s.reloadPolicies(ctx)
	return nil
}

// Delete 删除用户
func (s *UserService) Delete(ctx context.Context, id uint64) error {
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrUserNotFound, "用户不存在")
		}
		return err
	}
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.reloadPolicies(ctx)
	return nil
}

func (s *UserService) reloadPolicies(ctx context.Context) {
	if s.enforcer != nil {
		_ = s.enforcer.ReloadPolicies(ctx)
	}
}

// UpdatePassword 修改用户密码（管理员重置，强制下次改密）
func (s *UserService) UpdatePassword(ctx context.Context, req *dto.UpdateUserPasswordRequest) error {
	_, err := s.userRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrUserNotFound, "用户不存在")
		}
		return err
	}
	if err := s.validatePassword(ctx, req.Password); err != nil {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, req.ID, string(hashed), false)
}

// Unlock 解除登录锁定
func (s *UserService) Unlock(ctx context.Context, id uint64) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrUserNotFound, "用户不存在")
		}
		return err
	}
	_ = security.UnlockAccount(ctx, user.Username)
	return nil
}

func (s *UserService) validatePassword(ctx context.Context, password string) error {
	minLen, complex := 6, false
	if s.configSvc != nil {
		minLen = s.configSvc.PwdMinLength(ctx)
		complex = s.configSvc.PwdRequireComplexity(ctx)
	}
	if err := security.ValidatePassword(password, minLen, complex); err != nil {
		return perrors.New(perrors.ErrPasswordWeak, err.Error())
	}
	return nil
}

// buildDataScope 根据用户角色的 data_scope 构建数据权限过滤器
func (s *UserService) buildDataScope(ctx context.Context, userID uint64) (*datascope.Provider, error) {
	return datascope.BuildFromDB(ctx, s.userRepo.DB(), userID)
}
