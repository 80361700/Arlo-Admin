package handler

import (
	"arlo-admin/internal/modules/monitor/dto"
	"arlo-admin/internal/modules/monitor/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

type OnlineHandler struct {
	svc *service.OnlineService
}

func NewOnlineHandler(svc *service.OnlineService) *OnlineHandler {
	return &OnlineHandler{svc: svc}
}

// List 在线用户列表
func (h *OnlineHandler) List(c *gin.Context) {
	var req dto.OnlineQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	resp, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询在线用户失败")
		return
	}
	response.Success(c, resp)
}

// Kick 强制下线
func (h *OnlineHandler) Kick(c *gin.Context) {
	var req dto.KickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Kick(c.Request.Context(), &req); err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.SuccessWithMsg(c, "已强制下线", nil)
}
