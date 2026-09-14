package service

import (
	"context"
	"strings"

	"arlo-admin/internal/domain/model"
	"arlo-admin/internal/domain/repository"
	"arlo-admin/internal/modules/system/dto"

	perrors "arlo-admin/pkg/errors"

	"gorm.io/gorm"
)

// DeptService 部门管理服务
type DeptService struct {
	deptRepo *repository.DeptRepository
	userRepo *repository.UserRepository
}

func NewDeptService(deptRepo *repository.DeptRepository, userRepo *repository.UserRepository) *DeptService {
	return &DeptService{deptRepo: deptRepo, userRepo: userRepo}
}

// GetTree 获取部门树
func (s *DeptService) GetTree(ctx context.Context) ([]*dto.DeptTreeResponse, error) {
	depts, err := s.deptRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	leaderIDs := make([]uint64, 0)
	seen := make(map[uint64]bool)
	for _, d := range depts {
		if d.LeaderID > 0 && !seen[d.LeaderID] {
			seen[d.LeaderID] = true
			leaderIDs = append(leaderIDs, d.LeaderID)
		}
	}
	contactMap, err := s.userRepo.FindContactsByIDs(ctx, leaderIDs)
	if err != nil {
		return nil, err
	}
	return s.buildTree(depts, 0, contactMap), nil
}

// buildTree 构建部门树结构
func (s *DeptService) buildTree(depts []model.Dept, parentID uint64, contactMap map[uint64]repository.UserContact) []*dto.DeptTreeResponse {
	var tree []*dto.DeptTreeResponse
	for _, d := range depts {
		if d.ParentID == parentID {
			contact := contactMap[d.LeaderID]
			node := &dto.DeptTreeResponse{
				ID:       d.ID,
				ParentID: d.ParentID,
				Name:     d.Name,
				Code:     d.Code,
				Sort:     d.Sort,
				LeaderID: d.LeaderID,
				Leader:   contact.Name,
				Phone:    contact.Phone,
				Email:    contact.Email,
				Status:   d.Status,
				Remark:   d.Remark,
			}
			children := s.buildTree(depts, d.ID, contactMap)
			if len(children) > 0 {
				node.Children = children
			}
			tree = append(tree, node)
		}
	}
	return tree
}

func (s *DeptService) validateLeaderID(ctx context.Context, leaderID uint64) error {
	if leaderID == 0 {
		return nil
	}
	_, err := s.userRepo.FindByID(ctx, leaderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.BadRequest, "负责人用户不存在")
		}
		return err
	}
	return nil
}

func (s *DeptService) validateCode(ctx context.Context, code string, excludeID uint64) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}
	exists, err := s.deptRepo.ExistsByCode(ctx, code, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return perrors.New(perrors.BadRequest, "部门编码已存在")
	}
	return nil
}

// Create 创建部门
func (s *DeptService) Create(ctx context.Context, req *dto.CreateDeptRequest) error {
	if err := s.validateLeaderID(ctx, req.LeaderID); err != nil {
		return err
	}
	code := strings.TrimSpace(req.Code)
	if err := s.validateCode(ctx, code, 0); err != nil {
		return err
	}
	dept := &model.Dept{
		ParentID: req.ParentID,
		Name:     req.Name,
		Code:     code,
		Sort:     req.Sort,
		LeaderID: req.LeaderID,
		Status:   req.Status,
		Remark:   strings.TrimSpace(req.Remark),
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
	if err := s.validateLeaderID(ctx, req.LeaderID); err != nil {
		return err
	}
	code := strings.TrimSpace(req.Code)
	if err := s.validateCode(ctx, code, req.ID); err != nil {
		return err
	}
	dept.ParentID = req.ParentID
	dept.Name = req.Name
	dept.Code = code
	dept.Sort = req.Sort
	dept.LeaderID = req.LeaderID
	dept.Status = req.Status
	dept.Remark = strings.TrimSpace(req.Remark)
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
