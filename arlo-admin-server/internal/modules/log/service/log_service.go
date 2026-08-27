package service

import (
	"context"

	"arlo-admin/internal/database"
	"arlo-admin/internal/modules/log/dto"
	"arlo-admin/internal/modules/log/model"
	"arlo-admin/internal/modules/log/repository"
	"arlo-admin/pkg/datascope"
	"arlo-admin/pkg/utils"
)

// LogService 日志服务
type LogService struct {
	loginLogRepo     *repository.LoginLogRepository
	operationLogRepo *repository.OperationLogRepository
}

// NewLogService 创建 LogService
func NewLogService(loginLogRepo *repository.LoginLogRepository, operationLogRepo *repository.OperationLogRepository) *LogService {
	return &LogService{
		loginLogRepo:     loginLogRepo,
		operationLogRepo: operationLogRepo,
	}
}

// GetLoginLogs 分页查询登录日志
func (s *LogService) GetLoginLogs(ctx context.Context, req *dto.LoginLogQuery, currentUserID uint64) (*dto.LoginLogListResponse, error) {
	scope, _ := datascope.BuildFromDB(ctx, database.DB, currentUserID)
	logs, total, err := s.loginLogRepo.List(ctx, req.Username, req.Status, req.StartTime, req.EndTime, scope, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	list := make([]dto.LoginLogResponse, len(logs))
	for i, l := range logs {
		list[i] = dto.LoginLogResponse{
			ID:        l.ID,
			Username:  l.Username,
			IP:        l.IP,
			Location:  l.Location,
			Browser:   l.Browser,
			OS:        l.OS,
			Status:    l.Status,
			Msg:       l.Msg,
			CreatedAt: utils.FormatTime(l.CreatedAt),
		}
	}

	return &dto.LoginLogListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetOperationLogs 分页查询操作日志
func (s *LogService) GetOperationLogs(ctx context.Context, req *dto.OperationLogQuery, currentUserID uint64) (*dto.OperationLogListResponse, error) {
	scope, _ := datascope.BuildFromDB(ctx, database.DB, currentUserID)
	logs, total, err := s.operationLogRepo.List(ctx, req.Username, req.Module, req.URL, req.StartTime, req.EndTime, req.Status, scope, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	list := make([]dto.OperationLogResponse, len(logs))
	for i, l := range logs {
		list[i] = dto.OperationLogResponse{
			ID:        l.ID,
			UserID:    l.UserID,
			Username:  l.Username,
			Module:    l.Module,
			Action:    l.Action,
			Method:    l.Method,
			URL:       l.URL,
			IP:        l.IP,
			UserAgent: l.UserAgent,
			Params:    l.Params,
			Result:    l.Result,
			CostTime:  l.CostTime,
			Status:    l.Status,
			ErrorMsg:  l.ErrorMsg,
			CreatedAt: utils.FormatTime(l.CreatedAt),
		}
	}

	return &dto.OperationLogListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// RecordLoginLog 记录登录日志（供外部调用）
func (s *LogService) RecordLoginLog(ctx context.Context, log *model.LoginLog) error {
	return s.loginLogRepo.Create(ctx, log)
}

// CleanupLogs 清理过期日志
func (s *LogService) CleanupLogs(ctx context.Context, days int) (loginDeleted, operationDeleted int64, err error) {
	loginDeleted, err = s.loginLogRepo.DeleteOlderThan(ctx, days)
	if err != nil {
		return 0, 0, err
	}
	operationDeleted, err = s.operationLogRepo.DeleteOlderThan(ctx, days)
	if err != nil {
		return loginDeleted, 0, err
	}
	return loginDeleted, operationDeleted, nil
}
