package handler

import (
	"strconv"
	"strings"

	"arlo-admin/internal/modules/demo/dto"
	"arlo-admin/internal/modules/demo/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/middleware"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.List(c.Request.Context(), &q, userID)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, userName := middleware.GetCurrentUser(c)
	id, err := h.svc.Create(c.Request.Context(), &req, userID, userName)
	if err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *OrderHandler) Delete(c *gin.Context) {
	var body struct {
		IDs []uint64 `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Delete(c.Request.Context(), body.IDs); err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.SuccessWithMsg(c, "删除成功", nil)
}

func (h *OrderHandler) ListProcessOptions(c *gin.Context) {
	data, err := h.svc.ListProcessOptions(c.Request.Context())
	if err != nil {
		response.Error(c, apperrors.Internal, "查询流程失败")
		return
	}
	response.Success(c, data)
}

func (h *OrderHandler) GetProcessPreview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	data, err := h.svc.GetProcessPreview(c.Request.Context(), id)
	if err != nil {
		msg := err.Error()
		code := apperrors.BadRequest
		if strings.Contains(msg, "不存在") {
			code = apperrors.NotFound
		}
		response.Error(c, code, msg)
		return
	}
	response.Success(c, data)
}

func (h *OrderHandler) GetLaunchForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.GetLaunchForm(c.Request.Context(), id, userID)
	if err != nil {
		msg := err.Error()
		code := apperrors.BadRequest
		if strings.Contains(msg, "不存在") {
			code = apperrors.NotFound
		}
		if strings.Contains(msg, "无权") {
			code = apperrors.Forbidden
		}
		response.Error(c, code, msg)
		return
	}
	response.Success(c, data)
}

func (h *OrderHandler) Launch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	var req dto.LaunchRequest
	_ = c.ShouldBindJSON(&req)
	userID, userName := middleware.GetCurrentUser(c)
	instanceID, err := h.svc.Launch(c.Request.Context(), id, &req, userID, userName)
	if err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"instanceId": instanceID})
}
