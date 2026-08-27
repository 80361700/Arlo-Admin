package handler

import (
	"arlo-admin/internal/modules/log/dto"
	"arlo-admin/internal/modules/log/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/excel"
	"arlo-admin/pkg/middleware"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// LogHandler 日志管理处理器
type LogHandler struct {
	svc *service.LogService
}

// NewLogHandler 创建 LogHandler
func NewLogHandler(svc *service.LogService) *LogHandler {
	return &LogHandler{svc: svc}
}

// LoginLogList 登录日志列表
// @Summary      登录日志列表
// @Tags         Log
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "页码"        example(1)
// @Param        pageSize  query     int     false  "每页条数"     example(10)
// @Param        username  query     string  false  "用户名"
// @Param        status    query     int     false  "状态（0=失败 1=成功）"
// @Success      200       {object}  response.Response{data=dto.LoginLogListResponse}
// @Failure      400       {object}  response.Response
// @Router       /log/login/list [get]
func (h *LogHandler) LoginLogList(c *gin.Context) {
	var req dto.LoginLogQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}

	uid, _ := middleware.GetCurrentUser(c)
	resp, err := h.svc.GetLoginLogs(c.Request.Context(), &req, uid)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询登录日志失败")
		return
	}
	response.Success(c, resp)
}

// OperationLogList 操作日志列表
// @Summary      操作日志列表
// @Tags         Log
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "页码"        example(1)
// @Param        pageSize  query     int     false  "每页条数"     example(10)
// @Param        username   query     string  false  "操作人"
// @Param        module     query     string  false  "操作模块"
// @Param        url        query     string  false  "请求地址"
// @Param        startTime  query     string  false  "开始时间"  example(2026-07-01 00:00:00)
// @Param        endTime    query     string  false  "结束时间"  example(2026-07-20 23:59:59)
// @Param        status     query     int     false  "状态（0=失败 1=成功）"
// @Success      200       {object}  response.Response{data=dto.OperationLogListResponse}
// @Failure      400       {object}  response.Response
// @Router       /log/operation/list [get]
func (h *LogHandler) OperationLogList(c *gin.Context) {
	var req dto.OperationLogQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误: "+err.Error())
		return
	}

	uid, _ := middleware.GetCurrentUser(c)
	resp, err := h.svc.GetOperationLogs(c.Request.Context(), &req, uid)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询操作日志失败")
		return
	}
	response.Success(c, resp)
}

// ExportLoginLogs 导出登录日志
// @Summary      导出登录日志
// @Tags         Log
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Router       /log/login/export [get]
func (h *LogHandler) ExportLoginLogs(c *gin.Context) {
	var req dto.LoginLogQuery
	_ = c.ShouldBindQuery(&req)
	uid, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.ExportLoginLogs(c.Request.Context(), &req, uid)
	if err != nil {
		response.Error(c, apperrors.Internal, "导出失败: "+err.Error())
		return
	}
	excel.WriteDownload(c, "login_logs.xlsx", data)
}

// ExportOperationLogs 导出操作日志
// @Summary      导出操作日志
// @Tags         Log
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Router       /log/operation/export [get]
func (h *LogHandler) ExportOperationLogs(c *gin.Context) {
	var req dto.OperationLogQuery
	_ = c.ShouldBindQuery(&req)
	uid, _ := middleware.GetCurrentUser(c)
	data, err := h.svc.ExportOperationLogs(c.Request.Context(), &req, uid)
	if err != nil {
		response.Error(c, apperrors.Internal, "导出失败: "+err.Error())
		return
	}
	excel.WriteDownload(c, "operation_logs.xlsx", data)
}
