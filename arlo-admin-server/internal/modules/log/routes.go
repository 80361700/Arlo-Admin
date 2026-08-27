package log

import (
	"arlo-admin/internal/modules/log/handler"
	"arlo-admin/internal/modules/log/repository"
	"arlo-admin/internal/modules/log/service"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册日志模块路由
func RegisterRoutes(r *gin.RouterGroup, enforcer *casbinpkg.Enforcer) {
	loginLogRepo := repository.NewLoginLogRepository()
	opLogRepo := repository.NewOperationLogRepository()
	logSvc := service.NewLogService(loginLogRepo, opLogRepo)
	logHandler := handler.NewLogHandler(logSvc)

	lg := r.Group("/log")
	lg.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))
	{
		lg.GET("/login/list", logHandler.LoginLogList)
		lg.GET("/login/export", logHandler.ExportLoginLogs)
		lg.GET("/operation/list", logHandler.OperationLogList)
		lg.GET("/operation/export", logHandler.ExportOperationLogs)
	}
}
