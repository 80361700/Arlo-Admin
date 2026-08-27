package handler

import (
	"errors"
	"strconv"
	"strings"

	"arlo-admin/internal/modules/job/dto"
	"arlo-admin/internal/modules/job/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobHandler struct {
	svc *service.JobService
}

func NewJobHandler(svc *service.JobService) *JobHandler {
	return &JobHandler{svc: svc}
}

func (h *JobHandler) List(c *gin.Context) {
	var q dto.JobListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	resp, err := h.svc.List(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, resp)
}

func (h *JobHandler) Detail(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	resp, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, apperrors.NotFound, "任务不存在")
			return
		}
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, resp)
}

func (h *JobHandler) Create(c *gin.Context) {
	var req dto.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	resp, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		writeJobErr(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *JobHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	var req dto.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), id, &req); err != nil {
		writeJobErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *JobHandler) UpdateStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	var req dto.UpdateJobStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		writeJobErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *JobHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		writeJobErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *JobHandler) Run(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	if err := h.svc.RunOnce(c.Request.Context(), id); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "正在执行") {
			response.Error(c, apperrors.BadRequest, msg)
			return
		}
		// 业务执行失败也返回 200 + 提示：日志已写入
		response.Success(c, gin.H{"ok": false, "msg": msg})
		return
	}
	response.Success(c, gin.H{"ok": true, "msg": "执行完成"})
}

func (h *JobHandler) Handlers(c *gin.Context) {
	response.Success(c, h.svc.ListHandlers())
}

func (h *JobHandler) LogList(c *gin.Context) {
	var q dto.JobLogQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}
	resp, err := h.svc.ListLogs(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, resp)
}

func (h *JobHandler) LogDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("logId"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, apperrors.BadRequest, "无效ID")
		return
	}
	resp, err := h.svc.GetLog(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, apperrors.NotFound, "日志不存在")
			return
		}
		response.Error(c, apperrors.Internal, "查询失败")
		return
	}
	response.Success(c, resp)
}

func parseID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

func writeJobErr(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, apperrors.NotFound, "任务不存在")
		return
	}
	if service.IsBadHandler(err) {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	msg := err.Error()
	if strings.Contains(msg, "cron") || strings.HasPrefix(msg, "分:") || strings.HasPrefix(msg, "时:") ||
		strings.Contains(msg, "cron 需为") {
		response.Error(c, apperrors.BadRequest, "Cron 无效: "+msg)
		return
	}
	response.Error(c, apperrors.Internal, "操作失败")
}
