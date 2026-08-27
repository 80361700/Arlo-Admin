package monitor

import (
	"arlo-admin/internal/modules/monitor/handler"
	"arlo-admin/internal/modules/monitor/service"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册监控模块路由
func RegisterRoutes(r *gin.RouterGroup, enforcer *casbinpkg.Enforcer) {
	onlineSvc := service.NewOnlineService()
	onlineH := handler.NewOnlineHandler(onlineSvc)
	serverH := handler.NewServerHandler(service.NewServerService())

	g := r.Group("/monitor")
	g.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))
	{
		g.GET("/online/list", onlineH.List)
		g.POST("/online/kick", onlineH.Kick)
		g.GET("/server", serverH.GetServer)
	}
}
