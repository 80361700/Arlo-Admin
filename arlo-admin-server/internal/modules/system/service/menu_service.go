package service

import (
	"context"

	"arlo-admin/internal/domain/model"
	"arlo-admin/internal/domain/repository"
	"arlo-admin/internal/modules/system/dto"

	casbinpkg "arlo-admin/pkg/casbin"
	perrors "arlo-admin/pkg/errors"

	"gorm.io/gorm"
)

// MenuService 菜单管理服务
type MenuService struct {
	menuRepo *repository.MenuRepository
	enforcer *casbinpkg.Enforcer
}

func NewMenuService(menuRepo *repository.MenuRepository, enforcer *casbinpkg.Enforcer) *MenuService {
	return &MenuService{menuRepo: menuRepo, enforcer: enforcer}
}

func (s *MenuService) reloadPolicies(ctx context.Context) {
	if s.enforcer != nil {
		_ = s.enforcer.ReloadPolicies(ctx)
	}
}

// GetTree 获取菜单树
func (s *MenuService) GetTree(ctx context.Context) ([]*dto.MenuTreeResponse, error) {
	menus, err := s.menuRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildTree(menus, 0), nil
}

func (s *MenuService) buildTree(menus []model.Menu, parentID uint64) []*dto.MenuTreeResponse {
	var tree []*dto.MenuTreeResponse
	for _, m := range menus {
		if m.ParentID == parentID {
			node := &dto.MenuTreeResponse{
				ID:         m.ID,
				ParentID:   m.ParentID,
				Name:       m.Name,
				Type:       m.Type,
				Path:       m.Path,
				Component:  m.Component,
				Icon:       m.Icon,
				Sort:       m.Sort,
				Permission: m.Permission,
				Status:     m.Status,
				Visible:    m.Visible,
				KeepAlive:  m.KeepAlive,
			}
			children := s.buildTree(menus, m.ID)
			if len(children) > 0 {
				node.Children = children
			}
			tree = append(tree, node)
		}
	}
	return tree
}

// Create 创建菜单
func (s *MenuService) Create(ctx context.Context, req *dto.CreateMenuRequest) error {
	menu := &model.Menu{
		ParentID:   req.ParentID,
		Name:       req.Name,
		Type:       req.Type,
		Path:       req.Path,
		Component:  req.Component,
		Icon:       req.Icon,
		Sort:       req.Sort,
		Permission: req.Permission,
		Status:     req.Status,
		Visible:    req.Visible,
		KeepAlive:  req.KeepAlive,
	}
	if menu.Status == 0 {
		menu.Status = 1
	}
	if menu.Visible == 0 {
		menu.Visible = 1
	}
	if menu.KeepAlive == 0 {
		menu.KeepAlive = 1
	}
	if err := s.menuRepo.Create(ctx, menu); err != nil {
		return err
	}
	s.reloadPolicies(ctx)
	return nil
}

// Update 更新菜单
func (s *MenuService) Update(ctx context.Context, req *dto.UpdateMenuRequest) error {
	menu, err := s.menuRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrMenuExists, "菜单不存在")
		}
		return err
	}
	if req.ParentID != 0 && req.ParentID == req.ID {
		return perrors.New(perrors.BadRequest, "上级菜单不能是自己")
	}
	// 禁止把上级设为自己的子孙，避免成环
	if req.ParentID != 0 {
		descendantIDs, err := s.menuRepo.FindDescendantIDs(ctx, req.ID)
		if err != nil {
			return err
		}
		for _, id := range descendantIDs {
			if id == req.ParentID {
				return perrors.New(perrors.BadRequest, "上级菜单不能是自己的子菜单")
			}
		}
	}
	menu.ParentID = req.ParentID
	menu.Name = req.Name
	menu.Type = req.Type
	menu.Path = req.Path
	menu.Component = req.Component
	menu.Icon = req.Icon
	menu.Sort = req.Sort
	menu.Permission = req.Permission
	menu.Status = req.Status
	menu.Visible = req.Visible
	menu.KeepAlive = req.KeepAlive
	if err := s.menuRepo.Update(ctx, menu); err != nil {
		return err
	}
	s.reloadPolicies(ctx)
	return nil
}

// Delete 删除菜单
func (s *MenuService) Delete(ctx context.Context, id uint64) error {
	_, err := s.menuRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrMenuExists, "菜单不存在")
		}
		return err
	}
	has, err := s.menuRepo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if has {
		return perrors.New(perrors.ErrHasChildren, "存在子菜单，无法删除")
	}
	if err := s.menuRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.reloadPolicies(ctx)
	return nil
}
