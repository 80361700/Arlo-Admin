package handler

import (
	"strconv"

	"arlo-admin/internal/modules/system/dto"
	"arlo-admin/internal/modules/system/service"
	perrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// DeptHandler 部门管理 HTTP 处理器
type DeptHandler struct {
	svc *service.DeptService
}

func NewDeptHandler(svc *service.DeptService) *DeptHandler {
	return &DeptHandler{svc: svc}
}

// GetTree 获取部门树 GET /api/v1/system/dept/tree
func (h *DeptHandler) GetTree(c *gin.Context) {
	data, err := h.svc.GetTree(c.Request.Context())
	if err != nil {
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// Create 创建部门 POST /api/v1/system/dept
func (h *DeptHandler) Create(c *gin.Context) {
	var req dto.CreateDeptRequest
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

// Update 更新部门 PUT /api/v1/system/dept
func (h *DeptHandler) Update(c *gin.Context) {
	var req dto.UpdateDeptRequest
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

// Delete 删除部门 DELETE /api/v1/system/dept/:id
func (h *DeptHandler) Delete(c *gin.Context) {
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
