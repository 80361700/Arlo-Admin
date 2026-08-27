package handler

import (
	"strconv"

	"arlo-admin/internal/modules/message/dto"
	"arlo-admin/internal/modules/message/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// MessageHandler 站内信处理器
type MessageHandler struct {
	svc *service.MessageService
}

// NewMessageHandler 创建 MessageHandler
func NewMessageHandler(svc *service.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

// List 站内信列表 GET /api/v1/message/list
func (h *MessageHandler) List(c *gin.Context) {
	uid, _ := getUserInfo(c)
	var req dto.MessageListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}

	resp, err := h.svc.List(c.Request.Context(), uid, &req)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询消息列表失败")
		return
	}
	response.Success(c, resp)
}

// Send 发送站内信 POST /api/v1/message/send
func (h *MessageHandler) Send(c *gin.Context) {
	uid, uname := getUserInfo(c)
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.Send(c.Request.Context(), &req, uid, uname); err != nil {
		response.Error(c, apperrors.Internal, "发送消息失败")
		return
	}
	response.SuccessWithMsg(c, "发送成功", nil)
}

// MarkRead 标记已读 PUT /api/v1/message/:id/read
func (h *MessageHandler) MarkRead(c *gin.Context) {
	uid, _ := getUserInfo(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效的ID")
		return
	}

	if err := h.svc.MarkAsRead(c.Request.Context(), id, uid); err != nil {
		response.Error(c, apperrors.Internal, "标记已读失败")
		return
	}
	response.SuccessWithMsg(c, "标记已读成功", nil)
}

// MarkAllRead 全部标记已读 PUT /api/v1/message/read-all
func (h *MessageHandler) MarkAllRead(c *gin.Context) {
	uid, _ := getUserInfo(c)
	if err := h.svc.MarkAllAsRead(c.Request.Context(), uid); err != nil {
		response.Error(c, apperrors.Internal, "标记已读失败")
		return
	}
	response.SuccessWithMsg(c, "全部标记已读成功", nil)
}

// Delete 删除消息 DELETE /api/v1/message/:id?side=sent|received
func (h *MessageHandler) Delete(c *gin.Context) {
	uid, _ := getUserInfo(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效的ID")
		return
	}

	side := c.DefaultQuery("side", "received")
	if err := h.svc.Delete(c.Request.Context(), id, uid, side); err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.SuccessWithMsg(c, "删除成功", nil)
}

// UnreadCount 未读消息数 GET /api/v1/message/unread-count
func (h *MessageHandler) UnreadCount(c *gin.Context) {
	uid, _ := getUserInfo(c)
	resp, err := h.svc.UnreadCount(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询未读数失败")
		return
	}
	response.Success(c, resp)
}
