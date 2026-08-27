package handler

import (
	"arlo-admin/internal/modules/monitor/service"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

type ServerHandler struct {
	svc *service.ServerService
}

func NewServerHandler(svc *service.ServerService) *ServerHandler {
	return &ServerHandler{svc: svc}
}

// GetServer 服务监控指标
// @Summary 服务监控
// @Tags Monitor
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /monitor/server [get]
func (h *ServerHandler) GetServer(c *gin.Context) {
	resp := h.svc.GetServerInfo(c.Request.Context())
	response.Success(c, resp)
}
