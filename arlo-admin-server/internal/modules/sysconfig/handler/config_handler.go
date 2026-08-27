package handler

import (
	"strconv"

	"arlo-admin/internal/modules/sysconfig/dto"
	"arlo-admin/internal/modules/sysconfig/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 配置管理处理器
type ConfigHandler struct {
	svc *service.ConfigService
}

// NewConfigHandler 创建 ConfigHandler
func NewConfigHandler(svc *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

// GetPublic 公开配置（登录页等免鉴权）
// @Summary      公开系统配置
// @Tags         Config
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response{data=dto.PublicConfigResponse}
// @Router       /sysconfig/public [get]
func (h *ConfigHandler) GetPublic(c *gin.Context) {
	response.Success(c, h.svc.GetPublic(c.Request.Context()))
}

// List 配置列表
// @Summary      配置列表
// @Tags         Config
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name  query     string  false  "配置名称"
// @Param        key   query     string  false  "配置键"
// @Param        type  query     int     false  "配置类型"
// @Success      200   {object}  response.Response{data=[]dto.ConfigResponse}
// @Router       /sysconfig/list [get]
func (h *ConfigHandler) List(c *gin.Context) {
	var req dto.ConfigListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}

	list, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询配置列表失败")
		return
	}
	response.Success(c, list)
}

// GetByKey 按 key 获取配置
// @Summary      按 key 获取配置
// @Tags         Config
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        key   path      string  true  "配置键"
// @Success      200   {object}  response.Response{data=dto.ConfigResponse}
// @Failure      400   {object}  response.Response
// @Router       /sysconfig/{key} [get]
func (h *ConfigHandler) GetByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}

	config, err := h.svc.GetByKey(c.Request.Context(), key)
	if err != nil {
		response.Error(c, apperrors.Internal, err.Error())
		return
	}
	response.Success(c, config)
}

// Create 创建配置
// @Summary      创建配置
// @Tags         Config
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateConfigRequest  true  "配置信息"
// @Success      200   {object}  response.Response{data=dto.ConfigResponse}
// @Failure      400   {object}  response.Response
// @Router       /sysconfig [post]
func (h *ConfigHandler) Create(c *gin.Context) {
	var req dto.CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}

	config, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.Success(c, config)
}

// Update 更新配置
// @Summary      更新配置
// @Tags         Config
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.UpdateConfigRequest  true  "配置信息"
// @Success      200   {object}  response.Response{data=dto.ConfigResponse}
// @Failure      400   {object}  response.Response
// @Router       /sysconfig [put]
func (h *ConfigHandler) Update(c *gin.Context) {
	var req dto.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}

	config, err := h.svc.Update(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.Success(c, config)
}

// Delete 删除配置
// @Summary      删除配置
// @Tags         Config
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "配置ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /sysconfig/{id} [delete]
func (h *ConfigHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}
