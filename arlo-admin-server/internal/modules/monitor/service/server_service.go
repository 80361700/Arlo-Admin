package service

import (
	"context"

	"arlo-admin/internal/config"
	"arlo-admin/internal/modules/monitor/collector"
	"arlo-admin/internal/modules/monitor/dto"
)

type ServerService struct{}

func NewServerService() *ServerService {
	return &ServerService{}
}

func (s *ServerService) GetServerInfo(ctx context.Context) *dto.ServerMonitorResponse {
	mode := ""
	if config.GlobalConfig != nil {
		mode = config.GlobalConfig.Server.Mode
	}
	return &dto.ServerMonitorResponse{
		CPU:   collector.CollectCPU(),
		Mem:   collector.CollectMem(),
		Disk:  collector.CollectDisk(),
		Sys:   collector.CollectSys(),
		Go:    collector.CollectGo(),
		App:   collector.CollectApp(mode),
		DB:    collector.CollectDB(ctx),
		Redis: collector.CollectRedis(ctx),
	}
}
