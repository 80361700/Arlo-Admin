package sysconfig

import (
	"arlo-admin/internal/modules/sysconfig/handler"
	"arlo-admin/internal/modules/sysconfig/repository"
	"arlo-admin/internal/modules/sysconfig/service"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册系统配置模块路由
func RegisterRoutes(r *gin.RouterGroup, enforcer *casbinpkg.Enforcer) {
	repo := repository.NewConfigRepository()
	svc := service.NewConfigService(repo)
	h := handler.NewConfigHandler(svc)

	group := r.Group("/sysconfig")
	// 公开接口：登录页读取系统名称 / 验证码开关等
	group.GET("/public", h.GetPublic)

	authGroup := group.Group("")
	authGroup.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))
	{
		authGroup.GET("/list", h.List)
		authGroup.GET("/:key", h.GetByKey) // :key 必须放在 /list、/public 之后
		authGroup.POST("", h.Create)
		authGroup.PUT("", h.Update)
		authGroup.DELETE("/:id", h.Delete)
	}
}
