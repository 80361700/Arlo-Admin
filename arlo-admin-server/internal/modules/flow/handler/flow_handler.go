package handler

import (
	"errors"
	"strconv"

	"arlo-admin/internal/modules/flow/dto"
	"arlo-admin/internal/modules/flow/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/middleware"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FlowHandler struct {
	svc *service.FlowService
}

func NewFlowHandler(svc *service.FlowService) *FlowHandler {
	return &FlowHandler{svc: svc}
}

func (h *FlowHandler) ListCategoryTree(c *gin.Context) {
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.ListCategoryTree(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) ListCategoryOptions(c *gin.Context) {
	data, err := h.svc.ListCategoryOptions(c.Request.Context())
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	data, err := h.svc.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apperrors.Internal, "创建失败")
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) UpdateCategory(c *gin.Context) {
	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateCategory(c.Request.Context(), &req); err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) DeleteCategory(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) ListProcessOptions(c *gin.Context) {
	var excludeID uint64
	if v := c.Query("excludeId"); v != "" {
		excludeID, _ = strconv.ParseUint(v, 10, 64)
	}
	data, err := h.svc.ListProcessOptions(c.Request.Context(), excludeID)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) GetProcess(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.GetProcess(c.Request.Context(), id, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) SaveProcess(c *gin.Context) {
	var req dto.SaveProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.SaveProcess(c.Request.Context(), &req, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) UpdateProcessState(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	var req dto.UpdateProcessStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.UpdateState(c.Request.Context(), id, req.State, userID); err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) CheckProcess(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.CheckProcess(c.Request.Context(), id, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) DeleteProcess(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.DeleteProcess(c.Request.Context(), id, userID); err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) CloneProcess(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.CloneProcess(c.Request.Context(), id, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) ListProcessHistories(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	userID, _ := middleware.GetCurrentUser(c)
	list, total, err := h.svc.ListProcessHistories(c.Request.Context(), id, page, pageSize, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, response.PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *FlowHandler) GetProcessHistory(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	historyID, err := strconv.ParseUint(c.Param("historyId"), 10, 64)
	if err != nil || historyID == 0 {
		response.Error(c, apperrors.BadRequest, "无效历史ID")
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.GetProcessHistory(c.Request.Context(), id, historyID, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) CheckoutProcessHistory(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	historyID, err := strconv.ParseUint(c.Param("historyId"), 10, 64)
	if err != nil || historyID == 0 {
		response.Error(c, apperrors.BadRequest, "无效历史ID")
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.CheckoutProcessHistory(c.Request.Context(), id, historyID, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, data)
}

// ---------- 表单模板 ----------

func (h *FlowHandler) ListFormCategoryTree(c *gin.Context) {
	data, err := h.svc.ListFormCategoryTree(c.Request.Context())
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) ListFormCategoryOptions(c *gin.Context) {
	data, err := h.svc.ListFormCategoryOptions(c.Request.Context())
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) CreateFormCategory(c *gin.Context) {
	var req dto.CreateFormCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	data, err := h.svc.CreateFormCategory(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apperrors.Internal, "创建失败")
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) UpdateFormCategory(c *gin.Context) {
	var req dto.UpdateFormCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateFormCategory(c.Request.Context(), &req); err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) DeleteFormCategory(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	if err := h.svc.DeleteFormCategory(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) ListFormOptions(c *gin.Context) {
	data, err := h.svc.ListFormOptions(c.Request.Context())
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) GetForm(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	data, err := h.svc.GetForm(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) SaveForm(c *gin.Context) {
	var req dto.SaveFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.SaveForm(c.Request.Context(), &req, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) UpdateFormState(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	var req dto.UpdateFormStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateFormState(c.Request.Context(), id, req.Status); err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) SaveFormSchema(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	var req dto.SaveFormSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.SaveFormSchema(c.Request.Context(), id, req.FormSchema, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) DeleteForm(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	if err := h.svc.DeleteForm(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	response.Success(c, nil)
}

func parseID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(c, apperrors.NotFound, "记录不存在")
	case errors.Is(err, service.ErrCategoryHasProcess):
		response.Error(c, apperrors.BadRequest, err.Error())
	case errors.Is(err, service.ErrProcessKeyExists):
		response.Error(c, apperrors.BadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidModel):
		msg := err.Error()
		if msg == "" || msg == service.ErrInvalidModel.Error() {
			msg = "流程模型无效，请检查流程设计"
		}
		response.Error(c, apperrors.BadRequest, msg)
	case errors.Is(err, service.ErrCategoryHasForm):
		response.Error(c, apperrors.BadRequest, err.Error())
	case errors.Is(err, service.ErrFormCodeExists):
		response.Error(c, apperrors.BadRequest, err.Error())
	case errors.Is(err, service.ErrFormBound):
		response.Error(c, apperrors.BadRequest, err.Error())
	case errors.Is(err, service.ErrHistoryMismatch):
		response.Error(c, apperrors.BadRequest, err.Error())
	case errors.Is(err, service.ErrProcessForbidden):
		response.Error(c, apperrors.Forbidden, err.Error())
	default:
		response.Error(c, apperrors.Internal, "操作失败")
	}
}
