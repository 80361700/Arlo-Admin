package service

import (
	"context"
	"fmt"

	"arlo-admin/internal/domain/model"
	"arlo-admin/internal/domain/repository"
	"arlo-admin/internal/modules/system/dto"

	perrors "arlo-admin/pkg/errors"

	"gorm.io/gorm"
)

// PostService 岗位管理服务
type PostService struct {
	postRepo *repository.PostRepository
}

func NewPostService(postRepo *repository.PostRepository) *PostService {
	return &PostService{postRepo: postRepo}
}

// List 分页查询岗位列表
func (s *PostService) List(ctx context.Context, req *dto.PostListRequest) (*dto.PageResponse, error) {
	posts, total, err := s.postRepo.List(ctx, req.Code, req.Name, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]dto.PostResponse, 0, len(posts))
	for _, p := range posts {
		list = append(list, dto.PostResponse{
			ID:        p.ID,
			Code:      p.Code,
			Name:      p.Name,
			Sort:      p.Sort,
			Status:    p.Status,
			Remark:    p.Remark,
			CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &dto.PageResponse{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// GetAll 获取全部岗位（下拉选择）
func (s *PostService) GetAll(ctx context.Context) ([]dto.PostResponse, error) {
	posts, err := s.postRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]dto.PostResponse, 0, len(posts))
	for _, p := range posts {
		list = append(list, dto.PostResponse{
			ID:   p.ID,
			Name: p.Name,
			Code: p.Code,
		})
	}
	return list, nil
}

// GetDetail 获取岗位详情
func (s *PostService) GetDetail(ctx context.Context, id uint64) (*dto.PostResponse, error) {
	post, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, perrors.New(perrors.NotFound, "岗位不存在")
		}
		return nil, err
	}
	return &dto.PostResponse{
		ID:        post.ID,
		Code:      post.Code,
		Name:      post.Name,
		Sort:      post.Sort,
		Status:    post.Status,
		Remark:    post.Remark,
		CreatedAt: post.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// Create 创建岗位
func (s *PostService) Create(ctx context.Context, req *dto.CreatePostRequest) error {
	exists, err := s.postRepo.ExistsByCode(ctx, req.Code, 0)
	if err != nil {
		return err
	}
	if exists {
		return perrors.New(perrors.ErrDictTypeExists, fmt.Sprintf("岗位编码 %s 已存在", req.Code))
	}
	post := &model.Post{
		Code:   req.Code,
		Name:   req.Name,
		Sort:   req.Sort,
		Status: req.Status,
		Remark: req.Remark,
	}
	if post.Status == 0 {
		post.Status = 1
	}
	return s.postRepo.Create(ctx, post)
}

// Update 更新岗位
func (s *PostService) Update(ctx context.Context, req *dto.UpdatePostRequest) error {
	post, err := s.postRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.NotFound, "岗位不存在")
		}
		return err
	}
	exists, err := s.postRepo.ExistsByCode(ctx, req.Code, req.ID)
	if err != nil {
		return err
	}
	if exists {
		return perrors.New(perrors.ErrDictTypeExists, fmt.Sprintf("岗位编码 %s 已存在", req.Code))
	}
	post.Code = req.Code
	post.Name = req.Name
	post.Sort = req.Sort
	post.Status = req.Status
	post.Remark = req.Remark
	return s.postRepo.Update(ctx, post)
}

// Delete 删除岗位
func (s *PostService) Delete(ctx context.Context, id uint64) error {
	_, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.NotFound, "岗位不存在")
		}
		return err
	}
	return s.postRepo.Delete(ctx, id)
}
