package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"arlo-admin/internal/modules/flow/dto"
	"arlo-admin/internal/modules/flow/repository"
	"arlo-admin/internal/modules/flow/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/middleware"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *FlowHandler) ListLaunchProcesses(c *gin.Context) {
	name := c.Query("name")
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.ListLaunchProcesses(c.Request.Context(), name, userID)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) GetLaunchForm(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.GetLaunchForm(c.Request.Context(), id, userID)
	if err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) LaunchProcess(c *gin.Context) {
	var req dto.LaunchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, userName := middleware.GetCurrentUser(c)
	if n, _, ok := repository.NewOrgStore().GetUser(userID); ok && n != "" {
		userName = n
	}
	id, err := h.svc.Launch(c.Request.Context(), &req, userID, userName)
	if err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, gin.H{"instanceId": id})
}

func (h *FlowHandler) ActivateDraft(c *gin.Context) {
	var req dto.ActivateDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.ActivateDraft(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) DeleteDraft(c *gin.Context) {
	var req dto.DeleteDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.DeleteDraft(c.Request.Context(), req.InstanceID, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) UpdateDraft(c *gin.Context) {
	var req dto.UpdateDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.UpdateDraft(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) ListPendingTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	f := parseApproveListFilter(c)
	userID, _ := middleware.GetCurrentUser(c)
	list, total, err := h.svc.ListPending(c.Request.Context(), userID, f, page, pageSize)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func (h *FlowHandler) ListMyApplications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	f := parseApproveListFilter(c)
	userID, _ := middleware.GetCurrentUser(c)
	list, total, err := h.svc.ListMyApplications(c.Request.Context(), userID, f, page, pageSize)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func (h *FlowHandler) ListMonitor(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	f := parseApproveListFilter(c)
	onlyActive := c.DefaultQuery("onlyActive", "1") != "0"
	userID, _ := middleware.GetCurrentUser(c)
	list, total, err := h.svc.ListMonitor(c.Request.Context(), userID, f, onlyActive, page, pageSize)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func (h *FlowHandler) ListReceived(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	f := parseApproveListFilter(c)
	userID, _ := middleware.GetCurrentUser(c)
	list, total, err := h.svc.ListReceived(c.Request.Context(), userID, f, page, pageSize)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func (h *FlowHandler) MarkReceivedRead(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.MarkReceivedRead(c.Request.Context(), req.ID, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) ListApproved(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	f := parseApproveListFilter(c)
	userID, _ := middleware.GetCurrentUser(c)
	list, total, err := h.svc.ListApproved(c.Request.Context(), userID, f, page, pageSize)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func (h *FlowHandler) ListClaimable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	f := parseApproveListFilter(c)
	userID, _ := middleware.GetCurrentUser(c)
	list, total, err := h.svc.ListClaimable(c.Request.Context(), userID, f, page, pageSize)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func (h *FlowHandler) ClaimTask(c *gin.Context) {
	var req dto.ClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.Claim(c.Request.Context(), req.TaskID, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) ConsentTask(c *gin.Context) {
	var req dto.ConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.Consent(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) ResubmitTask(c *gin.Context) {
	var req dto.ResubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.Resubmit(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) RejectTask(c *gin.Context) {
	var req dto.RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.Reject(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) RevokeInstance(c *gin.Context) {
	var req dto.RevokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.Revoke(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) TransferTask(c *gin.Context) {
	var req dto.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.Transfer(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) TerminateInstance(c *gin.Context) {
	var req dto.TerminateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.Terminate(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) AdminTransferTask(c *gin.Context) {
	var req dto.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.AdminTransfer(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) RollbackTask(c *gin.Context) {
	var req dto.RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.Rollback(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) UrgeInstance(c *gin.Context) {
	var req dto.UrgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, userName := middleware.GetCurrentUser(c)
	if n, _, ok := repository.NewOrgStore().GetUser(userID); ok && n != "" {
		userName = n
	}
	if err := h.svc.Urge(c.Request.Context(), &req, userID, userName); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) AppendActor(c *gin.Context) {
	var req dto.AppendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.AppendActor(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) RemoveActor(c *gin.Context) {
	var req dto.RemoveActorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.RemoveActor(c.Request.Context(), &req, userID); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) AddRuntimeCC(c *gin.Context) {
	var req dto.RuntimeCCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, userName := middleware.GetCurrentUser(c)
	if n, _, ok := repository.NewOrgStore().GetUser(userID); ok && n != "" {
		userName = n
	}
	if err := h.svc.AddRuntimeCC(c.Request.Context(), &req, userID, userName); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) AddComment(c *gin.Context) {
	var req dto.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, userName := middleware.GetCurrentUser(c)
	if n, _, ok := repository.NewOrgStore().GetUser(userID); ok && n != "" {
		userName = n
	}
	if err := h.svc.AddComment(c.Request.Context(), &req, userID, userName); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) BatchConsent(c *gin.Context) {
	var req dto.BatchConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.BatchConsent(c.Request.Context(), &req, userID)
	if err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) BatchReject(c *gin.Context) {
	var req dto.BatchRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.BatchReject(c.Request.Context(), &req, userID)
	if err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) GetMyDelegate(c *gin.Context) {
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.GetMyDelegate(c.Request.Context(), userID)
	if err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) SetMyDelegate(c *gin.Context) {
	var req dto.SetDelegateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	userID, _ := middleware.GetCurrentUser(c)
	if err := h.svc.SetMyDelegate(c.Request.Context(), userID, &req); err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *FlowHandler) GetInstanceDetail(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	taskID, _ := strconv.ParseUint(c.Query("taskId"), 10, 64)
	userID, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.GetInstanceDetail(c.Request.Context(), id, userID, taskID)
	if err != nil {
		writeApproveErr(c, err)
		return
	}
	response.Success(c, data)
}

func (h *FlowHandler) TickDelayTasks(c *gin.Context) {
	delayN, timeoutN, remindN, err := h.svc.TickTimers(c.Request.Context())
	if err != nil {
		response.Error(c, apperrors.Internal, "处理失败")
		return
	}
	response.Success(c, gin.H{
		"delay":    delayN,
		"timeout":  timeoutN,
		"remind":   remindN,
		"processed": delayN + timeoutN + remindN,
	})
}

func writeApproveErr(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, service.ErrProcessDisabled) ||
		errors.Is(err, service.ErrTaskNotFound) ||
		errors.Is(err, service.ErrCannotHandle) {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.Error(c, apperrors.BadRequest, err.Error())
}

// parseApproveListFilter 审批列表筛选：keyword/createBy/instanceState/beginTime/endTime
func parseApproveListFilter(c *gin.Context) repository.ApproveListFilter {
	begin, end := parseListTimeRange(c)
	f := repository.ApproveListFilter{
		Keyword:  c.Query("keyword"),
		CreateBy: c.Query("createBy"),
		Begin:    begin,
		End:      end,
	}
	if s := strings.TrimSpace(c.Query("instanceState")); s != "" {
		if v, err := strconv.ParseInt(s, 10, 8); err == nil {
			st := int8(v)
			f.InstanceState = &st
		}
	}
	return f
}

// parseListTimeRange 列表时间筛选：beginTime/endTime 支持 YYYY-MM-DD（结束日含当天）
func parseListTimeRange(c *gin.Context) (begin, end *time.Time) {
	begin = parseDayBound(c.Query("beginTime"), false)
	end = parseDayBound(c.Query("endTime"), true)
	return begin, end
}

func parseDayBound(raw string, endOfDay bool) *time.Time {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	loc := time.Local
	t, err := time.ParseInLocation("2006-01-02", s, loc)
	if err != nil {
		t, err = time.ParseInLocation("2006-01-02 15:04:05", s, loc)
		if err != nil {
			return nil
		}
		return &t
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	return &t
}
