package handler

import (
	"io"
	"strconv"
	"time"

	"arlo-admin/internal/modules/system/dto"
	"arlo-admin/internal/modules/system/service"
	"arlo-admin/pkg/excel"
	perrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/middleware"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理 HTTP 处理器
type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// List 用户分页列表
// @Summary      用户分页列表
// @Tags         System/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "页码"     default(1)
// @Param        pageSize  query     int     false  "每页条数"  default(10)
// @Param        username  query     string  false  "用户名"
// @Param        nickname  query     string  false  "昵称"
// @Param        phone     query     string  false  "手机号"
// @Param        status    query     int     false  "状态(0禁用 1启用)"
// @Param        deptId    query     int     false  "部门ID"
// @Success      200       {object}  response.Response{data=response.PageData}
// @Failure      401       {object}  response.Response
// @Failure      403       {object}  response.Response
// @Router       /system/user/list [get]
func (h *UserHandler) List(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.List(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// GetDetail 用户详情
// @Summary      获取用户详情
// @Tags         System/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  response.Response{data=dto.UserResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /system/user/{id} [get]
func (h *UserHandler) GetDetail(c *gin.Context) {
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

// Create 创建用户
// @Summary      创建用户
// @Tags         System/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.CreateUserRequest  true  "创建用户请求"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      403      {object}  response.Response
// @Router       /system/user [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
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

// Update 更新用户
// @Summary      更新用户信息
// @Tags         System/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.UpdateUserRequest  true  "更新用户请求"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      403      {object}  response.Response
// @Router       /system/user [put]
func (h *UserHandler) Update(c *gin.Context) {
	var req dto.UpdateUserRequest
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

// Delete 删除用户
// @Summary      删除用户
// @Tags         System/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /system/user/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
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

// UpdatePassword 修改密码
// @Summary      修改用户密码
// @Tags         System/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.UpdateUserPasswordRequest  true  "修改密码请求"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      403      {object}  response.Response
// @Router       /system/user/password [put]
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	var req dto.UpdateUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdatePassword(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// Export 导出用户 Excel
// @Summary      导出用户
// @Tags         System/User
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Router       /system/user/export [get]
func (h *UserHandler) Export(c *gin.Context) {
	var req dto.UserListRequest
	_ = c.ShouldBindQuery(&req)
	uid, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.ExportUsers(c.Request.Context(), &req, uid)
	if err != nil {
		response.Error(c, perrors.Internal, "导出失败: "+err.Error())
		return
	}
	excel.WriteDownload(c, "users_"+time.Now().Format("20060102150405")+".xlsx", data)
}

// ImportTemplate 下载导入模板
// @Summary      用户导入模板
// @Tags         System/User
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Router       /system/user/import/template [get]
func (h *UserHandler) ImportTemplate(c *gin.Context) {
	data, err := h.svc.ImportTemplate()
	if err != nil {
		response.Error(c, perrors.Internal, "生成模板失败")
		return
	}
	excel.WriteDownload(c, "user_import_template.xlsx", data)
}

// Import 导入用户
// @Summary      导入用户
// @Tags         System/User
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Excel 文件"
// @Success      200   {object}  response.Response
// @Router       /system/user/import [post]
func (h *UserHandler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, perrors.BadRequest, "请上传 Excel 文件")
		return
	}
	f, err := file.Open()
	if err != nil {
		response.Error(c, perrors.BadRequest, "读取文件失败")
		return
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		response.Error(c, perrors.BadRequest, "读取文件失败")
		return
	}
	ok, errs, err := h.svc.ImportUsers(c.Request.Context(), buf)
	if err != nil {
		response.Error(c, perrors.BadRequest, "解析失败: "+err.Error())
		return
	}
	response.SuccessWithMsg(c, "导入完成", gin.H{"success": ok, "errors": errs})
}

// Unlock 解除账号登录锁定
// @Summary      解锁用户
// @Tags         System/User
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  response.Response
// @Router       /system/user/{id}/unlock [put]
func (h *UserHandler) Unlock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, perrors.BadRequest, "无效的用户ID")
		return
	}
	if err := h.svc.Unlock(c.Request.Context(), id); err != nil {
		if ae, ok := err.(*perrors.AppError); ok {
			response.Error(c, ae.Code, ae.Msg)
			return
		}
		response.Error(c, perrors.Internal, "解锁失败")
		return
	}
	response.Success(c, nil)
}
