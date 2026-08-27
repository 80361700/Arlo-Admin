package auth

import (
	"arlo-admin/internal/domain/repository"
	"arlo-admin/internal/modules/auth/handler"
	"arlo-admin/internal/modules/auth/service"
	logrepo "arlo-admin/internal/modules/log/repository"
	configrepo "arlo-admin/internal/modules/sysconfig/repository"
	configsvc "arlo-admin/internal/modules/sysconfig/service"
	"arlo-admin/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册认证模块路由
// 依赖注入链: Repository → Service → Handler
func RegisterRoutes(r *gin.RouterGroup) {
	userRepo := repository.NewUserRepository()
	roleRepo := repository.NewRoleRepository()
	menuRepo := repository.NewMenuRepository()
	loginLogRepo := logrepo.NewLoginLogRepository()
	configSvc := configsvc.NewConfigService(configrepo.NewConfigRepository())
	authSvc := service.NewAuthService(userRepo, roleRepo, menuRepo, loginLogRepo, configSvc)
	authHandler := handler.NewAuthHandler(authSvc)

	auth := r.Group("/auth")

	// 公开接口（无需 JWT）
	// refresh 必须公开：access token 过期后请求拦截器仍会带 Authorization，
	// 若 refresh 也走 JWTAuth，会因过期 token 直接 401，永远无法续期
	auth.GET("/captcha", authHandler.Captcha)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	// 需要认证的接口
	protected := auth.Group("")
	protected.Use(middleware.JWTAuth())
	{
		protected.POST("/logout", authHandler.Logout)
		protected.GET("/info", authHandler.UserInfo)
		protected.GET("/menus", authHandler.Menus)
		protected.PUT("/profile", authHandler.UpdateProfile)
		protected.PUT("/password", authHandler.ChangePassword)
	}
}
