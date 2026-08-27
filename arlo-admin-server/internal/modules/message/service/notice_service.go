package service

import (
	"context"
	"errors"
	"fmt"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/message/dto"
	"arlo-admin/internal/modules/message/model"
	"arlo-admin/internal/modules/message/repository"
	"arlo-admin/pkg/datascope"
	"arlo-admin/pkg/utils"

	"gorm.io/gorm"
)

// NoticeService 通知公告服务
type NoticeService struct {
	repo *repository.NoticeRepository
}

// NewNoticeService 创建 NoticeService
func NewNoticeService(repo *repository.NoticeRepository) *NoticeService {
	return &NoticeService{repo: repo}
}

// Create 创建通知公告
func (s *NoticeService) Create(ctx context.Context, req *dto.CreateNoticeRequest, userID uint64, username string) (*dto.NoticeResponse, error) {
	notice := &model.Notice{
		Title:       req.Title,
		Content:     req.Content,
		Type:        req.Type,
		Level:       req.Level,
		Status:      0, // 草稿
		PublisherID: userID,
		Publisher:   username,
	}
	if notice.Type == 0 {
		notice.Type = 1
	}
	if notice.Level == 0 {
		notice.Level = 1
	}
	if err := s.repo.Create(ctx, notice); err != nil {
		return nil, err
	}
	return toNoticeResponse(notice), nil
}

// Update 更新通知公告（仅草稿/已撤回可编辑）
func (s *NoticeService) Update(ctx context.Context, id uint64, req *dto.UpdateNoticeRequest) (*dto.NoticeResponse, error) {
	notice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("公告不存在")
		}
		return nil, err
	}
	if notice.Status == 1 {
		return nil, errors.New("已发布的公告请先撤回再编辑")
	}
	notice.Title = req.Title
	notice.Content = req.Content
	if req.Type > 0 {
		notice.Type = req.Type
	}
	if req.Level > 0 {
		notice.Level = req.Level
	}
	if err := s.repo.Update(ctx, notice); err != nil {
		return nil, err
	}
	return toNoticeResponse(notice), nil
}

// Delete 删除通知公告（软删除）
func (s *NoticeService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

// GetByID 查询详情
func (s *NoticeService) GetByID(ctx context.Context, id uint64) (*dto.NoticeResponse, error) {
	notice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toNoticeResponse(notice), nil
}

// List 分页列表（按发布人数据权限过滤）
func (s *NoticeService) List(ctx context.Context, req *dto.NoticeListQuery, currentUserID uint64) (*dto.NoticeListResponse, error) {
	scope, _ := datascope.BuildFromDB(ctx, database.DB, currentUserID)
	notices, total, err := s.repo.List(ctx, req.Title, req.Status, scope, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]dto.NoticeResponse, len(notices))
	for i, n := range notices {
		list[i] = *toNoticeResponse(&n)
	}
	return &dto.NoticeListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// Publish 发布公告（草稿/已撤回可发布）
func (s *NoticeService) Publish(ctx context.Context, id uint64) error {
	notice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("公告不存在")
		}
		return err
	}
	if notice.Status == 1 {
		return errors.New("公告已发布")
	}
	if notice.Status != 0 && notice.Status != 2 {
		return fmt.Errorf("当前状态不可发布")
	}
	return s.repo.UpdateStatus(ctx, id, 1)
}

// Revoke 撤回公告（仅已发布可撤回）
func (s *NoticeService) Revoke(ctx context.Context, id uint64) error {
	notice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("公告不存在")
		}
		return err
	}
	if notice.Status != 1 {
		return errors.New("仅已发布的公告可撤回")
	}
	return s.repo.UpdateStatus(ctx, id, 2)
}

func toNoticeResponse(n *model.Notice) *dto.NoticeResponse {
	return &dto.NoticeResponse{
		ID:          n.ID,
		Title:       n.Title,
		Content:     n.Content,
		Type:        n.Type,
		Level:       n.Level,
		Status:      n.Status,
		PublisherID: n.PublisherID,
		Publisher:   n.Publisher,
		CreatedAt:   utils.FormatTime(n.CreatedAt),
		UpdatedAt:   utils.FormatTime(n.UpdatedAt),
	}
}
