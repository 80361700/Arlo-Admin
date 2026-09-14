package demo

import (
	"arlo-admin/internal/modules/demo/handler"
	"arlo-admin/internal/modules/demo/repository"
	"arlo-admin/internal/modules/demo/service"
	flowrepo "arlo-admin/internal/modules/flow/repository"
	flowsvc "arlo-admin/internal/modules/flow/service"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 演示业务（采购单 + 按 KEY 发起）
func RegisterRoutes(r *gin.RouterGroup, enforcer *casbinpkg.Enforcer) {
	orderRepo := repository.NewOrderRepository()
	flowRepository := flowrepo.NewFlowRepository()
	flowService := flowsvc.NewFlowService(flowRepository)
	orderSvc := service.NewOrderService(orderRepo, flowService, flowRepository)
	h := handler.NewOrderHandler(orderSvc)

	g := r.Group("/business/purchase-order")
	g.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))
	{
		g.GET("/list", h.List)
		g.POST("", h.Create)
		g.POST("/delete", h.Delete)
		g.GET("/process-options", h.ListProcessOptions)
		g.GET("/:id/process-preview", h.GetProcessPreview)
		g.GET("/:id/launch-form", h.GetLaunchForm)
		g.POST("/:id/launch", h.Launch)
	}
}
