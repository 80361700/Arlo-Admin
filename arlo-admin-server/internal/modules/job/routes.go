package job

import (
	corejob "arlo-admin/internal/job"
	"arlo-admin/internal/modules/job/handler"
	"arlo-admin/internal/modules/job/repository"
	"arlo-admin/internal/modules/job/service"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册定时任务管理路由（依赖已启动的 Scheduler）
func RegisterRoutes(r *gin.RouterGroup, enforcer *casbinpkg.Enforcer, sch *corejob.Scheduler) {
	repo := repository.NewJobRepository()
	svc := service.NewJobService(repo, sch)
	h := handler.NewJobHandler(svc)

	g := r.Group("/monitor/job")
	g.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))
	{
		g.GET("/list", h.List)
		g.GET("/handlers", h.Handlers)
		g.GET("/log/list", h.LogList)
		g.GET("/log/:logId", h.LogDetail)
		g.GET("/:id", h.Detail)
		g.POST("", h.Create)
		g.PUT("/:id", h.Update)
		g.PUT("/:id/status", h.UpdateStatus)
		g.POST("/:id/run", h.Run)
		g.DELETE("/:id", h.Delete)
	}
}
