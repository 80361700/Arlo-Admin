package handler

import (
	"strconv"

	"arlo-admin/internal/modules/message/dto"
	"arlo-admin/internal/modules/message/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/middleware"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// NoticeHandler 通知公告处理器
type NoticeHandler struct {
	svc *service.NoticeService
}

// NewNoticeHandler 创建 NoticeHandler
func NewNoticeHandler(svc *service.NoticeService) *NoticeHandler {
	return &NoticeHandler{svc: svc}
}

func getUserInfo(c *gin.Context) (uint64, string) {
	return middleware.GetCurrentUser(c)
}

// List 通知公告列表 GET /api/v1/message/notice/list
func (h *NoticeHandler) List(c *gin.Context) {
	var req dto.NoticeListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}

	uid, _ := getUserInfo(c)
	resp, err := h.svc.List(c.Request.Context(), &req, uid)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询公告列表失败")
		return
	}
	response.Success(c, resp)
}

// Get 通知公告详情 GET /api/v1/message/notice/:id
func (h *NoticeHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效的ID")
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apperrors.Internal, "公告不存在")
		return
	}
	response.Success(c, resp)
}

// Create 创建通知公告 POST /api/v1/message/notice
func (h *NoticeHandler) Create(c *gin.Context) {
	var req dto.CreateNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}

	uid, uname := getUserInfo(c)
	resp, err := h.svc.Create(c.Request.Context(), &req, uid, uname)
	if err != nil {
		response.Error(c, apperrors.Internal, "创建公告失败")
		return
	}
	response.SuccessWithMsg(c, "创建成功", resp)
}

// Update 更新通知公告 PUT /api/v1/message/notice/:id
func (h *NoticeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效的ID")
		return
	}

	var req dto.UpdateNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.SuccessWithMsg(c, "更新成功", resp)
}

// Delete 删除通知公告 DELETE /api/v1/message/notice/:id
func (h *NoticeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效的ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, apperrors.Internal, "删除公告失败")
		return
	}
	response.SuccessWithMsg(c, "删除成功", nil)
}

// Publish 发布公告 PUT /api/v1/message/notice/:id/publish
func (h *NoticeHandler) Publish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效的ID")
		return
	}

	if err := h.svc.Publish(c.Request.Context(), id); err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.SuccessWithMsg(c, "发布成功", nil)
}

// Revoke 撤回公告 PUT /api/v1/message/notice/:id/revoke
func (h *NoticeHandler) Revoke(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效的ID")
		return
	}

	if err := h.svc.Revoke(c.Request.Context(), id); err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.SuccessWithMsg(c, "撤回成功", nil)
}
