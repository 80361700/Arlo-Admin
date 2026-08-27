package member

import (
	configrepo "arlo-admin/internal/modules/sysconfig/repository"
	configsvc "arlo-admin/internal/modules/sysconfig/service"
	"arlo-admin/internal/modules/member/handler"
	"arlo-admin/internal/modules/member/repository"
	"arlo-admin/internal/modules/member/service"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, enforcer *casbinpkg.Enforcer) {
	memberRepo := repository.NewMemberRepository()
	cfgSvc := configsvc.NewConfigService(configrepo.NewConfigRepository())
	memberService := service.NewMemberService(memberRepo, cfgSvc)
	memberHandler := handler.NewMemberHandler(memberService)

	// 游客接口：发送验证码、登录、刷新Token
	memberGroup := r.Group("/member")
	{
		memberGroup.POST("/send-code", memberHandler.SendCode)
		memberGroup.POST("/login", memberHandler.Login)
		memberGroup.POST("/refresh", memberHandler.Refresh)
	}

	// 会员认证接口：个人信息
	memberAuth := r.Group("/member")
	memberAuth.Use(middleware.MemberAuth())
	{
		memberAuth.GET("/info", memberHandler.GetInfo)
		memberAuth.PUT("/profile", memberHandler.UpdateProfile)
	}

	// 管理员接口：会员列表 / 录入 / 详情 / 编辑 / 重置密码 / 启停 / 删除
	systemGroup := r.Group("/system/member")
	systemGroup.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))
	{
		systemGroup.GET("/list", memberHandler.List)
		systemGroup.POST("", memberHandler.AdminCreate)
		systemGroup.PUT("", memberHandler.AdminUpdate)
		systemGroup.PUT("/password", memberHandler.UpdatePassword)
		systemGroup.GET("/:id", memberHandler.GetDetail)
		systemGroup.PUT("/:id/status", memberHandler.UpdateStatus)
		systemGroup.DELETE("/:id", memberHandler.Delete)
	}
}
