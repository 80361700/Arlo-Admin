package service

import (
	"context"

	"arlo-admin/internal/domain/model"
	"arlo-admin/internal/domain/repository"
	"arlo-admin/internal/modules/system/dto"

	perrors "arlo-admin/pkg/errors"

	"gorm.io/gorm"
)

// DeptService 部门管理服务
type DeptService struct {
	deptRepo *repository.DeptRepository
}

func NewDeptService(deptRepo *repository.DeptRepository) *DeptService {
	return &DeptService{deptRepo: deptRepo}
}

// GetTree 获取部门树
func (s *DeptService) GetTree(ctx context.Context) ([]*dto.DeptTreeResponse, error) {
	depts, err := s.deptRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildTree(depts, 0), nil
}

// buildTree 构建部门树结构
func (s *DeptService) buildTree(depts []model.Dept, parentID uint64) []*dto.DeptTreeResponse {
	var tree []*dto.DeptTreeResponse
	for _, d := range depts {
		if d.ParentID == parentID {
			node := &dto.DeptTreeResponse{
				ID:       d.ID,
				ParentID: d.ParentID,
				Name:     d.Name,
				Sort:     d.Sort,
				Leader:   d.Leader,
				Phone:    d.Phone,
				Email:    d.Email,
				Status:   d.Status,
			}
			children := s.buildTree(depts, d.ID)
			if len(children) > 0 {
				node.Children = children
			}
			tree = append(tree, node)
		}
	}
	return tree
}

// Create 创建部门
func (s *DeptService) Create(ctx context.Context, req *dto.CreateDeptRequest) error {
	dept := &model.Dept{
		ParentID: req.ParentID,
		Name:     req.Name,
		Sort:     req.Sort,
		Leader:   req.Leader,
		Phone:    req.Phone,
		Email:    req.Email,
		Status:   req.Status,
	}
	if dept.Status == 0 {
		dept.Status = 1
	}
	return s.deptRepo.Create(ctx, dept)
}

// Update 更新部门
func (s *DeptService) Update(ctx context.Context, req *dto.UpdateDeptRequest) error {
	dept, err := s.deptRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrDeptExists, "部门不存在")
		}
		return err
	}
	// 不能将子部门设为自己的父部门
	if req.ParentID != 0 && req.ParentID == req.ID {
		return perrors.New(perrors.BadRequest, "上级部门不能是自己")
	}
	// 禁止把上级设为自己的子孙，避免成环
	if req.ParentID != 0 {
		descendantIDs, err := s.deptRepo.FindDescendantIDs(ctx, req.ID)
		if err != nil {
			return err
		}
		for _, id := range descendantIDs {
			if id == req.ParentID {
				return perrors.New(perrors.BadRequest, "上级部门不能是自己的子部门")
			}
		}
	}
	dept.ParentID = req.ParentID
	dept.Name = req.Name
	dept.Sort = req.Sort
	dept.Leader = req.Leader
	dept.Phone = req.Phone
	dept.Email = req.Email
	dept.Status = req.Status
	return s.deptRepo.Update(ctx, dept)
}

// Delete 删除部门
func (s *DeptService) Delete(ctx context.Context, id uint64) error {
	_, err := s.deptRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrDeptExists, "部门不存在")
		}
		return err
	}
	has, err := s.deptRepo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if has {
		return perrors.New(perrors.ErrHasChildren, "存在子部门，无法删除")
	}
	return s.deptRepo.Delete(ctx, id)
}
