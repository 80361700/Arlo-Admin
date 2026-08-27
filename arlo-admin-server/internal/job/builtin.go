package job

import (
	"context"
	"fmt"
	"sync"

	logrepo "arlo-admin/internal/modules/log/repository"
	"arlo-admin/internal/modules/log/service"
)

var builtinOnce sync.Once

// RegisterBuiltin 注册内置处理器
func RegisterBuiltin() {
	builtinOnce.Do(func() {
		Register("log_cleanup", "清理登录/操作日志", "按 retainDays 清理过期登录日志与操作日志", handleLogCleanup)
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
