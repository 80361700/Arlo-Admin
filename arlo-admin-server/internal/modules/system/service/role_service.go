package service

import (
	"context"
	"fmt"

	"arlo-admin/internal/domain/model"
	"arlo-admin/internal/domain/repository"
	"arlo-admin/internal/modules/system/dto"

	casbinpkg "arlo-admin/pkg/casbin"
	perrors "arlo-admin/pkg/errors"

	"gorm.io/gorm"
)

// RoleService 角色管理服务
type RoleService struct {
	roleRepo *repository.RoleRepository
	deptRepo *repository.DeptRepository
	enforcer *casbinpkg.Enforcer
}

func NewRoleService(roleRepo *repository.RoleRepository, deptRepo *repository.DeptRepository, enforcer *casbinpkg.Enforcer) *RoleService {
	return &RoleService{roleRepo: roleRepo, deptRepo: deptRepo, enforcer: enforcer}
}

// reloadPolicies 安全地重新加载 Casbin 策略
func (s *RoleService) reloadPolicies(ctx context.Context) {
	if s.enforcer != nil {
		_ = s.enforcer.ReloadPolicies(ctx)
	}
}

// helper: model → response
func (s *RoleService) toResponse(r *model.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID:        r.ID,
		Name:      r.Name,
		Code:      r.Code,
		Sort:      r.Sort,
		Status:    r.Status,
		Remark:    r.Remark,
		DataScope: r.DataScope,
		CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// List 分页查询角色列表
func (s *RoleService) List(ctx context.Context, req *dto.RoleListRequest) (*dto.PageResponse, error) {
	roles, total, err := s.roleRepo.List(ctx, req.Name, req.Code, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]dto.RoleResponse, 0, len(roles))
	for i := range roles {
		list = append(list, s.toResponse(&roles[i]))
	}
	return &dto.PageResponse{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// GetAll 获取全部角色（下拉选择等场景）
func (s *RoleService) GetAll(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := s.roleRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]dto.RoleResponse, 0, len(roles))
	for i := range roles {
		list = append(list, dto.RoleResponse{
			ID:   roles[i].ID,
			Name: roles[i].Name,
			Code: roles[i].Code,
		})
	}
	return list, nil
}

// GetDetail 获取角色详情（含数据权限部门）
func (s *RoleService) GetDetail(ctx context.Context, id uint64) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, perrors.New(perrors.ErrRoleExists, "角色不存在")
		}
		return nil, err
	}

	resp := s.toResponse(role)

	// 附加自定义数据权限的部门ID
	if role.DataScope == 2 {
		deptIDs, _ := s.roleRepo.FindRoleDeptIDs(ctx, id)
		resp.DeptIDs = deptIDs
	}

	return &resp, nil
}

// GetRoleMenuIDs 获取角色拥有的菜单ID列表
func (s *RoleService) GetRoleMenuIDs(ctx context.Context, roleID uint64) ([]uint64, error) {
	return s.roleRepo.FindRoleMenuIDs(ctx, roleID)
}

// Create 创建角色（含数据权限部门分配）
func (s *RoleService) Create(ctx context.Context, req *dto.CreateRoleRequest) error {
	exists, err := s.roleRepo.ExistsByCode(ctx, req.Code, 0)
	if err != nil {
		return err
	}
	if exists {
		return perrors.New(perrors.ErrRoleExists, fmt.Sprintf("角色编码 %s 已存在", req.Code))
	}

	ds := req.DataScope
	if ds < 1 || ds > 5 {
		ds = 1 // 默认全部数据
	}

	role := &model.Role{
		Name:      req.Name,
		Code:      req.Code,
		Sort:      req.Sort,
		Status:    req.Status,
		Remark:    req.Remark,
		DataScope: ds,
	}
	if role.Status == 0 {
		role.Status = 1
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return err
	}

	// 自定义部门时分配部门
	if ds == 2 && len(req.DeptIDs) > 0 {
		_ = s.roleRepo.AssignDepts(ctx, role.ID, req.DeptIDs)
	}

	s.reloadPolicies(ctx)
	return nil
}

// Update 更新角色（含数据权限部门重分配）
func (s *RoleService) Update(ctx context.Context, req *dto.UpdateRoleRequest) error {
	role, err := s.roleRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrRoleExists, "角色不存在")
		}
		return err
	}
	exists, err := s.roleRepo.ExistsByCode(ctx, req.Code, req.ID)
	if err != nil {
		return err
	}
	if exists {
		return perrors.New(perrors.ErrRoleExists, fmt.Sprintf("角色编码 %s 已存在", req.Code))
	}

	ds := req.DataScope
	if ds < 1 || ds > 5 {
		ds = role.DataScope // 未传则保持原值
	}

	role.Name = req.Name
	role.Code = req.Code
	role.Sort = req.Sort
	role.Status = req.Status
	role.Remark = req.Remark
	role.DataScope = ds

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return err
	}

	// 重分配数据权限部门（非自定义时清空）
	if ds == 2 {
		_ = s.roleRepo.AssignDepts(ctx, req.ID, req.DeptIDs)
	} else {
		_ = s.roleRepo.AssignDepts(ctx, req.ID, nil)
	}

	s.reloadPolicies(ctx)
	return nil
}

// Delete 删除角色
func (s *RoleService) Delete(ctx context.Context, id uint64) error {
	_, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrRoleExists, "角色不存在")
		}
		return err
	}
	hasUser, err := s.roleRepo.HasUserAssigned(ctx, id)
	if err != nil {
		return err
	}
	if hasUser {
		return perrors.New(perrors.ErrRoleAssigned, "角色已被分配，无法删除")
	}
	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.reloadPolicies(ctx)
	return nil
}

// AssignMenus 分配角色菜单
func (s *RoleService) AssignMenus(ctx context.Context, req *dto.AssignRoleMenusRequest) error {
	_, err := s.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrRoleExists, "角色不存在")
		}
		return err
	}
	if err := s.roleRepo.AssignMenus(ctx, req.RoleID, req.MenuIDs); err != nil {
		return err
	}
	s.reloadPolicies(ctx)
	return nil
}
