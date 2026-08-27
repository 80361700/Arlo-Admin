package router

import (
	_ "arlo-admin/docs"
	"arlo-admin/internal/config"
	corejob "arlo-admin/internal/job"
	"arlo-admin/internal/modules/auth"
	"arlo-admin/internal/modules/file"
	jobmod "arlo-admin/internal/modules/job"
	"arlo-admin/internal/modules/log"
	"arlo-admin/internal/modules/member"
	"arlo-admin/internal/modules/message"
	"arlo-admin/internal/modules/monitor"
	"arlo-admin/internal/modules/sysconfig"
	"arlo-admin/internal/modules/system"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/middleware"
	"arlo-admin/pkg/response"
	"arlo-admin/pkg/ws"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func Setup(mode string, enforcer *casbinpkg.Enforcer, sch *corejob.Scheduler) *gin.Engine {
	gin.SetMode(mode)

	r := gin.New()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())

	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status": "ok",
			"name":   "arlo-admin",
		})
	})

	if shouldEnableSwagger(mode) {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := r.Group("/api/v1")
	api.Use(middleware.OperationLog())

	// WebSocket：query token 鉴权，不走 OperationLog 体解析
	r.GET("/api/v1/ws", ws.HandleWS)

	auth.RegisterRoutes(api)
	member.RegisterRoutes(api, enforcer)
	system.RegisterRoutes(api, enforcer)
	log.RegisterRoutes(api, enforcer)
	monitor.RegisterRoutes(api, enforcer)
	jobmod.RegisterRoutes(api, enforcer, sch)
	message.RegisterRoutes(api, enforcer)
	sysconfig.RegisterRoutes(api, enforcer)
	file.RegisterRoutes(api, enforcer)

	return r
}

func shouldEnableSwagger(mode string) bool {
	if config.GlobalConfig != nil {
		return config.GlobalConfig.Server.EnableSwagger
	}
	return mode == gin.DebugMode
}
