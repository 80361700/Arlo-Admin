package handler

import (
	"strconv"

	"arlo-admin/internal/modules/member/dto"
	"arlo-admin/internal/modules/member/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/middleware"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	memberService *service.MemberService
}

func NewMemberHandler(memberService *service.MemberService) *MemberHandler {
	return &MemberHandler{memberService: memberService}
}

// SendCode POST /api/v1/member/send-code
func (h *MemberHandler) SendCode(c *gin.Context) {
	var req dto.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "手机号格式错误")
		return
	}

	if err := h.memberService.SendCode(c.Request.Context(), req.Phone); err != nil {
		response.Error(c, apperrors.ErrCodeSendFailed, err.Error())
		return
	}

	response.Success(c, nil)
}

// Login POST /api/v1/member/login
func (h *MemberHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}

	result, err := h.memberService.Login(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		response.Error(c, apperrors.ErrLoginFailed, err.Error())
		return
	}

	response.Success(c, result)
}

// Refresh POST /api/v1/member/refresh
func (h *MemberHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}

	result, err := h.memberService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, apperrors.Unauthorized, err.Error())
		return
	}

	response.Success(c, result)
}

// GetInfo GET /api/v1/member/info
func (h *MemberHandler) GetInfo(c *gin.Context) {
	memberID, _ := middleware.GetCurrentMember(c)
	if memberID == 0 {
		response.Error(c, apperrors.Unauthorized, "未登录")
		return
	}

	info, err := h.memberService.GetInfo(c.Request.Context(), memberID)
	if err != nil {
		response.Error(c, apperrors.NotFound, err.Error())
		return
	}

	response.Success(c, info)
}

// UpdateProfile PUT /api/v1/member/profile
func (h *MemberHandler) UpdateProfile(c *gin.Context) {
	memberID, _ := middleware.GetCurrentMember(c)
	if memberID == 0 {
		response.Error(c, apperrors.Unauthorized, "未登录")
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}

	if err := h.memberService.UpdateProfile(c.Request.Context(), memberID, &req); err != nil {
		response.Error(c, apperrors.Internal, err.Error())
		return
	}

	response.Success(c, nil)
}

// List GET /api/v1/system/member/list — 管理员查看会员列表
func (h *MemberHandler) List(c *gin.Context) {
	var req dto.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	list, total, err := h.memberService.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apperrors.Internal, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":     list,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	})
}

// GetDetail GET /api/v1/system/member/:id — 管理员查看会员详情
func (h *MemberHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "ID格式错误")
		return
	}
	detail, err := h.memberService.GetDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apperrors.NotFound, err.Error())
		return
	}
	response.Success(c, detail)
}

// AdminCreate POST /api/v1/system/member — 管理员手动录入会员
func (h *MemberHandler) AdminCreate(c *gin.Context) {
	var req dto.AdminCreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}
	if err := h.memberService.AdminCreate(c.Request.Context(), &req); err != nil {
		response.Error(c, apperrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// AdminUpdate PUT /api/v1/system/member — 管理员更新会员资料
func (h *MemberHandler) AdminUpdate(c *gin.Context) {
	var req dto.AdminUpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}
	if err := h.memberService.AdminUpdate(c.Request.Context(), &req); err != nil {
		response.Error(c, apperrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// UpdatePassword PUT /api/v1/system/member/password — 管理员重置会员密码
func (h *MemberHandler) UpdatePassword(c *gin.Context) {
	var req dto.UpdateMemberPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}
	if err := h.memberService.UpdatePassword(c.Request.Context(), &req); err != nil {
		response.Error(c, apperrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// UpdateStatus PUT /api/v1/system/member/:id/status — 管理员禁用/启用会员
func (h *MemberHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "ID格式错误")
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}

	if err := h.memberService.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		response.Error(c, apperrors.Internal, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete DELETE /api/v1/system/member/:id — 管理员删除会员
func (h *MemberHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "ID格式错误")
		return
	}
	if err := h.memberService.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, apperrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}
