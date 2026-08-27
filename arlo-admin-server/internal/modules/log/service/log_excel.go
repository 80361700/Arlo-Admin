package service

import (
	"context"
	"fmt"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/log/dto"
	"arlo-admin/pkg/datascope"
	"arlo-admin/pkg/excel"
	"arlo-admin/pkg/utils"
)

// ExportLoginLogs 导出登录日志（最多 10000 条）
func (s *LogService) ExportLoginLogs(ctx context.Context, req *dto.LoginLogQuery, currentUserID uint64) ([]byte, error) {
	scope, _ := datascope.BuildFromDB(ctx, database.DB, currentUserID)
	logs, _, err := s.loginLogRepo.List(ctx, req.Username, req.Status, req.StartTime, req.EndTime, scope, 1, 10000)
	if err != nil {
		return nil, err
	}
	rows := make([][]interface{}, 0, len(logs))
	for _, l := range logs {
		status := "失败"
		if l.Status == 1 {
			status = "成功"
		}
		rows = append(rows, []interface{}{
			l.Username, l.IP, l.Location, l.Browser, l.OS, status, l.Msg, utils.FormatTime(l.CreatedAt),
		})
	}
	return excel.Write(excel.Sheet{
		Name:    "登录日志",
		Headers: []string{"用户名", "IP", "地点", "浏览器", "操作系统", "状态", "消息", "时间"},
		Rows:    rows,
	})
}

// ExportOperationLogs 导出操作日志（最多 10000 条）
func (s *LogService) ExportOperationLogs(ctx context.Context, req *dto.OperationLogQuery, currentUserID uint64) ([]byte, error) {
	scope, _ := datascope.BuildFromDB(ctx, database.DB, currentUserID)
	logs, _, err := s.operationLogRepo.List(ctx, req.Username, req.Module, req.URL, req.StartTime, req.EndTime, req.Status, scope, 1, 10000)
	if err != nil {
		return nil, err
	}
	rows := make([][]interface{}, 0, len(logs))
	for _, l := range logs {
		status := "失败"
		if l.Status == 1 {
			status = "成功"
		}
		rows = append(rows, []interface{}{
			l.Username, l.Module, l.Action, l.Method, l.URL, l.IP,
			fmt.Sprintf("%d", l.CostTime), status, l.ErrorMsg, utils.FormatTime(l.CreatedAt),
		})
	}
	return excel.Write(excel.Sheet{
		Name:    "操作日志",
		Headers: []string{"操作人", "模块", "操作", "方法", "URL", "IP", "耗时(ms)", "状态", "错误信息", "时间"},
		Rows:    rows,
	})
}
