package service

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/message/dto"
	"arlo-admin/internal/modules/message/model"
	"arlo-admin/internal/modules/message/repository"
	"arlo-admin/pkg/datascope"
	"arlo-admin/pkg/utils"
	"arlo-admin/pkg/ws"
)

// MessageService 站内信服务
type MessageService struct {
	repo *repository.MessageRepository
}

// NewMessageService 创建 MessageService
func NewMessageService(repo *repository.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

// Send 发送站内信（支持批量接收者）
func (s *MessageService) Send(ctx context.Context, req *dto.SendMessageRequest, senderID uint64, sender string) error {
	if req.Type == 0 {
		req.Type = 1
	}

	receiverIDs := req.ReceiverIDs
	if len(receiverIDs) == 0 {
		// 全部用户：只创建一条 receiver_id=0 的记录
		msg := &model.Message{
			Title:      req.Title,
			Content:    req.Content,
			Type:       req.Type,
			SenderID:   senderID,
			Sender:     sender,
			ReceiverID: 0,
		}
		if err := s.repo.Create(ctx, msg); err != nil {
			return err
		}
		ws.Default().NotifyAll()
		return nil
	}

	// 批量发送给指定用户（事务）
	if err := s.repo.CreateBatch(ctx, receiverIDs, req.Title, req.Content, req.Type, senderID, sender); err != nil {
		return err
	}
	ws.Default().NotifyUsers(receiverIDs...)
	return nil
}

// MarkAsRead 标记已读
func (s *MessageService) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	if err := s.repo.MarkAsRead(ctx, id, userID); err != nil {
		return err
	}
	ws.Default().NotifyUsers(userID)
	return nil
}

// MarkAllAsRead 全部标记已读
func (s *MessageService) MarkAllAsRead(ctx context.Context, userID uint64) error {
	if err := s.repo.MarkAllAsRead(ctx, userID); err != nil {
		return err
	}
	ws.Default().NotifyUsers(userID)
	return nil
}

// Delete 按场景删除：sent=仅发送方不可见，received=仅当前收件人不可见
func (s *MessageService) Delete(ctx context.Context, id uint64, userID uint64, side string) error {
	if side == "" {
		side = repository.DeleteSideReceived
	}
	scope, _ := datascope.BuildFromDB(ctx, database.DB, userID)
	dataScopeAll := scope != nil && scope.Scope == datascope.ScopeAll
	if err := s.repo.DeleteBySide(ctx, id, userID, side, dataScopeAll); err != nil {
		return err
	}
	if side == repository.DeleteSideReceived {
		ws.Default().NotifyUsers(userID)
	}
	return nil
}

// List 分页查询消息
func (s *MessageService) List(ctx context.Context, userID uint64, req *dto.MessageListQuery) (*dto.MessageListResponse, error) {
	// 发送记录使用聚合查询（按消息分组，显示 receiverCount）
	if req.Direction != nil && *req.Direction == 2 {
		return s.listSentMessages(ctx, userID, req)
	}

	// 我的消息：正常查询
	msgs, total, err := s.repo.List(ctx, userID, req.IsRead, req.Direction, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]dto.MessageResponse, len(msgs))
	for i, m := range msgs {
		list[i] = toMessageResponse(&m)
	}
	return &dto.MessageListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// listSentMessages 发送记录聚合查询（按数据权限过滤）
func (s *MessageService) listSentMessages(ctx context.Context, userID uint64, req *dto.MessageListQuery) (*dto.MessageListResponse, error) {
	scope, _ := datascope.BuildFromDB(ctx, database.DB, userID)
	msgs, total, err := s.repo.ListSent(ctx, userID, scope, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]dto.MessageResponse, len(msgs))
	for i, m := range msgs {
		list[i] = toMessageResponse(&m)
	}
	return &dto.MessageListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UnreadCount 未读消息数
func (s *MessageService) UnreadCount(ctx context.Context, userID uint64) (*dto.UnreadCountResponse, error) {
	count, err := s.repo.UnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.UnreadCountResponse{Count: count}, nil
}

// toMessageResponse model → dto（统一处理时间格式）
func toMessageResponse(m *model.Message) dto.MessageResponse {
	return dto.MessageResponse{
		ID:            m.ID,
		Title:         m.Title,
		Content:       m.Content,
		Type:          m.Type,
		SenderID:      m.SenderID,
		Sender:        m.Sender,
		ReceiverID:    m.ReceiverID,
		ReceiverCount: m.ReceiverCount,
		IsRead:        m.IsRead,
		ReadAt:        utils.FormatPtrTime(m.ReadAt),
		CreatedAt:     utils.FormatTime(m.CreatedAt),
	}
}
