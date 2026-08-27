package handler

import (
	"errors"
	"strings"

	"arlo-admin/internal/modules/auth/dto"
	"arlo-admin/internal/modules/auth/service"
	"arlo-admin/pkg/captcha"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login 用户登录
// @Summary      用户登录
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LoginRequest  true  "登录请求"
// @Success      200      {object}  response.Response{data=dto.LoginResponse}
// @Failure      400      {object}  response.Response
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "请输入用户名和密码")
		return
	}

	// 提前将用户名注入上下文，确保操作日志能记录操作人
	c.Set("username", req.Username)

	// 从请求中提取客户端信息
	req.IP = c.ClientIP()
	parseUserAgent(c.GetHeader("User-Agent"), &req)

	resp, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAccountLocked):
			response.Error(c, apperrors.ErrAccountLocked, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			response.Error(c, apperrors.ErrUserNotFound, apperrors.GetMsg(apperrors.ErrUserNotFound))
		case errors.Is(err, service.ErrPasswordWrong):
			msg := err.Error()
			if msg == service.ErrPasswordWrong.Error() {
				msg = apperrors.GetMsg(apperrors.ErrPasswordWrong)
			}
			response.Error(c, apperrors.ErrPasswordWrong, msg)
		case errors.Is(err, service.ErrUserDisabled):
			response.Error(c, apperrors.ErrUserDisabled, apperrors.GetMsg(apperrors.ErrUserDisabled))
		case errors.Is(err, service.ErrCaptchaInvalid):
			response.Error(c, apperrors.BadRequest, "验证码错误")
		default:
			response.Error(c, apperrors.Internal, "登录失败")
		}
		return
	}

	response.SuccessWithMsg(c, "登录成功", resp)
}

// Logout 用户登出
// @Summary      用户登出
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.LogoutRequest  false  "可选 refreshToken"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	accessToken, _ := c.Get("accessToken")
	at, _ := accessToken.(string)

	var req dto.LogoutRequest
	// body 可选；无 body 时只作废 access
	_ = c.ShouldBindJSON(&req)

	_ = h.svc.Logout(c.Request.Context(), at, req.RefreshToken)
	response.SuccessWithMsg(c, "登出成功", nil)
}

// Refresh 刷新令牌
// @Summary      刷新访问令牌
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RefreshRequest  true  "刷新令牌请求"
// @Success      200      {object}  response.Response{data=dto.RefreshResponse}
// @Failure      401      {object}  response.Response
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "请提供刷新令牌")
		return
	}

	resp, err := h.svc.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRefreshTokenType):
			response.Error(c, apperrors.ErrTokenInvalid, "非法的刷新令牌类型")
		default:
			response.Error(c, apperrors.ErrTokenInvalid, apperrors.GetMsg(apperrors.ErrTokenInvalid))
		}
		return
	}

	response.Success(c, resp)
}

// UserInfo 获取当前用户信息
// @Summary      获取当前用户信息
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=dto.UserInfoResponse}
// @Failure      401  {object}  response.Response
// @Router       /auth/info [get]
func (h *AuthHandler) UserInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, apperrors.Unauthorized, apperrors.GetMsg(apperrors.Unauthorized))
		return
	}

	uid, ok := userID.(uint64)
	if !ok {
		response.Error(c, apperrors.Internal, "用户标识错误")
		return
	}

	info, err := h.svc.GetUserInfo(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, apperrors.ErrUserNotFound, apperrors.GetMsg(apperrors.ErrUserNotFound))
		return
	}

	response.Success(c, info)
}

// UpdateProfile 更新当前用户个人资料
// @Summary      更新个人资料
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.UpdateProfileRequest  true  "个人资料"
// @Success      200  {object}  response.Response
// @Router       /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateProfile(c.Request.Context(), uid, &req); err != nil {
		response.Error(c, apperrors.Internal, "更新失败")
		return
	}
	response.SuccessWithMsg(c, "更新成功", nil)
}

// ChangePassword 修改当前用户密码
// @Summary      修改密码
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.ChangePasswordRequest  true  "修改密码"
// @Success      200  {object}  response.Response
// @Router       /auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), uid, &req); err != nil {
		if errors.Is(err, service.ErrOldPasswordWrong) {
			response.Error(c, apperrors.BadRequest, "原密码错误")
			return
		}
		if errors.Is(err, service.ErrPasswordWeak) {
			response.Error(c, apperrors.ErrPasswordWeak, err.Error())
			return
		}
		response.Error(c, apperrors.Internal, "修改密码失败")
		return
	}
	response.SuccessWithMsg(c, "密码修改成功", nil)
}

func currentUserID(c *gin.Context) (uint64, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, apperrors.Unauthorized, apperrors.GetMsg(apperrors.Unauthorized))
		return 0, false
	}
	uid, ok := userID.(uint64)
	if !ok {
		response.Error(c, apperrors.Internal, "用户标识错误")
		return 0, false
	}
	return uid, true
}

// Menus 获取当前用户菜单树（按角色过滤）
// @Summary      获取当前用户菜单
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]dto.MenuTreeNode}
// @Failure      401  {object}  response.Response
// @Router       /auth/menus [get]
func (h *AuthHandler) Menus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, apperrors.Unauthorized, apperrors.GetMsg(apperrors.Unauthorized))
		return
	}
	uid, ok := userID.(uint64)
	if !ok {
		response.Error(c, apperrors.Internal, "用户标识错误")
		return
	}

	menus, err := h.svc.GetUserMenus(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, apperrors.Internal, "获取菜单失败")
		return
	}
	response.Success(c, menus)
}

// Captcha 获取图形验证码
// @Summary      获取图形验证码
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response{data=dto.CaptchaResponse}
// @Router       /auth/captcha [get]
func (h *AuthHandler) Captcha(c *gin.Context) {
	resp, err := h.svc.GenerateCaptcha()
	if err != nil {
		msg := "验证码生成失败"
		if errors.Is(err, captcha.ErrStoreUnavailable) {
			msg = captcha.ErrStoreUnavailable.Error()
		}
		response.Error(c, apperrors.Internal, msg)
		return
	}
	response.Success(c, resp)
}

// parseUserAgent 从 User-Agent 字符串提取浏览器和操作系统信息
func parseUserAgent(ua string, req *dto.LoginRequest) {
	if ua == "" {
		return
	}
	// 操作系统检测
	switch {
	case strings.Contains(ua, "Mac OS") || strings.Contains(ua, "Macintosh"):
		req.OS = "macOS"
	case strings.Contains(ua, "Windows NT 10"):
		req.OS = "Windows 10"
	case strings.Contains(ua, "Windows NT"):
		req.OS = "Windows"
	case strings.Contains(ua, "Linux") && strings.Contains(ua, "Android"):
		req.OS = "Android"
	case strings.Contains(ua, "Linux"):
		req.OS = "Linux"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		req.OS = "iOS"
	}
	// 浏览器检测
	switch {
	case strings.Contains(ua, "Edg/"):
		req.Browser = "Edge"
	case strings.Contains(ua, "Chrome/"):
		req.Browser = "Chrome"
	case strings.Contains(ua, "Safari/") && !strings.Contains(ua, "Chrome"):
		req.Browser = "Safari"
	case strings.Contains(ua, "Firefox/"):
		req.Browser = "Firefox"
	}
}
