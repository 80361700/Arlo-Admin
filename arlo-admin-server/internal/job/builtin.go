package job

import (
	"context"
	"fmt"
	"sync"

	flowrepo "arlo-admin/internal/modules/flow/repository"
	flowsvc "arlo-admin/internal/modules/flow/service"
	logrepo "arlo-admin/internal/modules/log/repository"
	"arlo-admin/internal/modules/log/service"
)

var builtinOnce sync.Once

// RegisterBuiltin 注册内置处理器
func RegisterBuiltin() {
	builtinOnce.Do(func() {
		Register("log_cleanup", "清理登录/操作日志", "按 retainDays 清理过期登录日志与操作日志", handleLogCleanup)
		Register("flow_tick", "流程定时推进", "延时节点到期推进、审批超时自动通过/拒绝、审批提醒", handleFlowTick)
	})
}

func handleLogCleanup(ctx context.Context, params string) (string, error) {
	days := ParseRetainDays(params, 90)
	svc := service.NewLogService(
		logrepo.NewLoginLogRepository(),
		logrepo.NewOperationLogRepository(),
	)
	loginN, opN, err := svc.CleanupLogs(ctx, days)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("retainDays=%d loginDeleted=%d operationDeleted=%d", days, loginN, opN), nil
}

func handleFlowTick(ctx context.Context, _ string) (string, error) {
	svc := flowsvc.NewFlowService(flowrepo.NewFlowRepository())
	delayN, timeoutN, remindN, err := svc.TickTimers(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("delay=%d timeout=%d remind=%d", delayN, timeoutN, remindN), nil
}
