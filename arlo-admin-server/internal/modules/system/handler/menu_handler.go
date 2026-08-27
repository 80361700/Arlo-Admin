package handler

import (
	"strconv"

	"arlo-admin/internal/modules/system/dto"
	"arlo-admin/internal/modules/system/service"
	perrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// MenuHandler 菜单管理 HTTP 处理器
type MenuHandler struct {
	svc *service.MenuService
}

func NewMenuHandler(svc *service.MenuService) *MenuHandler {
	return &MenuHandler{svc: svc}
}

// GetTree 获取菜单树 GET /api/v1/system/menu/tree
func (h *MenuHandler) GetTree(c *gin.Context) {
	data, err := h.svc.GetTree(c.Request.Context())
	if err != nil {
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// Create 创建菜单 POST /api/v1/system/menu
func (h *MenuHandler) Create(c *gin.Context) {
	var req dto.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Create(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// Update 更新菜单 PUT /api/v1/system/menu
func (h *MenuHandler) Update(c *gin.Context) {
	var req dto.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// Delete 删除菜单 DELETE /api/v1/system/menu/:id
func (h *MenuHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, perrors.BadRequest, "ID格式错误")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}
