package handler

import (
	"strconv"

	"arlo-admin/internal/modules/system/dto"
	"arlo-admin/internal/modules/system/service"
	perrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// DictHandler 字典管理 HTTP 处理器
type DictHandler struct {
	svc *service.DictService
}

func NewDictHandler(svc *service.DictService) *DictHandler {
	return &DictHandler{svc: svc}
}

// --- 字典类型 ---

// ListDictTypes 字典类型分页列表 GET /api/v1/system/dict/type/list
func (h *DictHandler) ListDictTypes(c *gin.Context) {
	var req dto.DictTypeListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	data, err := h.svc.ListDictTypes(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// GetAllDictTypes 全部字典类型 GET /api/v1/system/dict/type/all
func (h *DictHandler) GetAllDictTypes(c *gin.Context) {
	data, err := h.svc.GetAllDictTypes(c.Request.Context())
	if err != nil {
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// GetDictType 获取单个字典类型 GET /api/v1/system/dict/type/:id
func (h *DictHandler) GetDictType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, perrors.BadRequest, "ID格式错误")
		return
	}
	data, err := h.svc.GetDictType(c.Request.Context(), id)
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

// CreateDictType 创建字典类型 POST /api/v1/system/dict/type
func (h *DictHandler) CreateDictType(c *gin.Context) {
	var req dto.CreateDictTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.CreateDictType(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// UpdateDictType 更新字典类型 PUT /api/v1/system/dict/type
func (h *DictHandler) UpdateDictType(c *gin.Context) {
	var req dto.UpdateDictTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateDictType(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteDictType 删除字典类型 DELETE /api/v1/system/dict/type/:id
func (h *DictHandler) DeleteDictType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, perrors.BadRequest, "ID格式错误")
		return
	}
	if err := h.svc.DeleteDictType(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// --- 字典数据 ---

// ListDictDatas 字典数据分页列表 GET /api/v1/system/dict/data/list
func (h *DictHandler) ListDictDatas(c *gin.Context) {
	var req dto.DictDataListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	data, err := h.svc.ListDictDatas(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// GetDictDatasByTypeID 按类型 ID 获取字典数据 GET /api/v1/system/dict/data/type/:id
func (h *DictHandler) GetDictDatasByTypeID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, perrors.BadRequest, "ID格式错误")
		return
	}
	data, err := h.svc.GetDictDatasByTypeID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, data)
}

// GetDictDatasByCode 按字典编码获取启用项 GET /api/v1/system/dict/data/code/:code
func (h *DictHandler) GetDictDatasByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.Error(c, perrors.BadRequest, "字典编码不能为空")
		return
	}
	data, err := h.svc.GetDictDatasByCode(c.Request.Context(), code)
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

// GetDictData 获取单个字典数据 GET /api/v1/system/dict/data/:id
func (h *DictHandler) GetDictData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, perrors.BadRequest, "ID格式错误")
		return
	}
	data, err := h.svc.GetDictData(c.Request.Context(), id)
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

// CreateDictData 创建字典数据 POST /api/v1/system/dict/data
func (h *DictHandler) CreateDictData(c *gin.Context) {
	var req dto.CreateDictDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.CreateDictData(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// UpdateDictData 更新字典数据 PUT /api/v1/system/dict/data
func (h *DictHandler) UpdateDictData(c *gin.Context) {
	var req dto.UpdateDictDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, perrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateDictData(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteDictData 删除字典数据 DELETE /api/v1/system/dict/data/:id
func (h *DictHandler) DeleteDictData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, perrors.BadRequest, "ID格式错误")
		return
	}
	if err := h.svc.DeleteDictData(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*perrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Msg)
			return
		}
		response.Error(c, perrors.Internal, err.Error())
		return
	}
	response.Success(c, nil)
}
