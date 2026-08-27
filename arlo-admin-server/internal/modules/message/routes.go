package message

import (
	"arlo-admin/internal/modules/message/handler"
	"arlo-admin/internal/modules/message/repository"
	"arlo-admin/internal/modules/message/service"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册消息模块路由
func RegisterRoutes(r *gin.RouterGroup, enforcer *casbinpkg.Enforcer) {
	noticeRepo := repository.NewNoticeRepository()
	noticeSvc := service.NewNoticeService(noticeRepo)
	noticeHandler := handler.NewNoticeHandler(noticeSvc)

	msgRepo := repository.NewMessageRepository()
	msgSvc := service.NewMessageService(msgRepo)
	msgHandler := handler.NewMessageHandler(msgSvc)

	notice := r.Group("/message/notice")
	notice.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))
	{
		notice.GET("/list", noticeHandler.List)
		notice.GET("/:id", noticeHandler.Get)
		notice.POST("", noticeHandler.Create)
		notice.PUT("/:id", noticeHandler.Update)
		notice.DELETE("/:id", noticeHandler.Delete)
		notice.PUT("/:id/publish", noticeHandler.Publish)
		notice.PUT("/:id/revoke", noticeHandler.Revoke)
	}

	msg := r.Group("/message")
	msg.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))
	{
		msg.GET("/list", msgHandler.List)
		msg.POST("/send", msgHandler.Send)
		msg.PUT("/:id/read", msgHandler.MarkRead)
		msg.PUT("/read-all", msgHandler.MarkAllRead)
		msg.DELETE("/:id", msgHandler.Delete)
		msg.GET("/unread-count", msgHandler.UnreadCount)
	}
}
