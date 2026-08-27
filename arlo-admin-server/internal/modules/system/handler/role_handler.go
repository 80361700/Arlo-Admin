package handler

import (
	"strconv"

	"arlo-admin/internal/modules/system/dto"
	"arlo-admin/internal/modules/system/service"
	perrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// RoleHandler 角色管理 HTTP 处理器
type RoleHandler struct {
	svc *service.RoleService
}

func NewRoleHandler(svc *service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// List 角色分页列表 GET /api/v1/system/role/list
func (h *RoleHandler) List(c *gin.Context) {
	var req dto.RoleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	data, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// GetAll 全部角色 GET /api/v1/system/role/all
func (h *RoleHandler) GetAll(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// GetDetail 角色详情 GET /api/v1/system/role/:id
func (h *RoleHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, perrors.BadRequest, "ID格式错误")
		return
	}
	data, err := h.svc.GetDetail(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// GetMenus 角色权限菜单 GET /api/v1/system/role/:id/menus
func (h *RoleHandler) GetMenus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, perrors.BadRequest, "ID格式错误")
		return
	}
	data, err := h.svc.GetRoleMenuIDs(c.Request.Context(), id)
	if err != nil {
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// Create 创建角色 POST /api/v1/system/role
func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
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

// Update 更新角色 PUT /api/v1/system/role
func (h *RoleHandler) Update(c *gin.Context) {
	var req dto.UpdateRoleRequest
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

// Delete 删除角色 DELETE /api/v1/system/role/:id
func (h *RoleHandler) Delete(c *gin.Context) {
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

// AssignMenus 分配角色菜单 POST /api/v1/system/role/assignMenus
func (h *RoleHandler) AssignMenus(c *gin.Context) {
	var req dto.AssignRoleMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.AssignMenus(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}
