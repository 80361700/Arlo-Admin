package flow

import (
	"arlo-admin/internal/modules/flow/handler"
	"arlo-admin/internal/modules/flow/repository"
	"arlo-admin/internal/modules/flow/service"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 流程定义（设计器落库）
func RegisterRoutes(r *gin.RouterGroup, enforcer *casbinpkg.Enforcer) {
	repo := repository.NewFlowRepository()
	svc := service.NewFlowService(repo)
	h := handler.NewFlowHandler(svc)

	g := r.Group("/flow")
	g.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))
	{
		// 分类
		g.GET("/category/tree", h.ListCategoryTree)
		g.GET("/category/options", h.ListCategoryOptions)
		g.POST("/category", h.CreateCategory)
		g.PUT("/category", h.UpdateCategory)
		g.DELETE("/category/:id", h.DeleteCategory)

		// 流程定义
		g.GET("/process/options", h.ListProcessOptions)
		g.GET("/process/:id/histories", h.ListProcessHistories)
		g.GET("/process/:id/histories/:historyId", h.GetProcessHistory)
		g.POST("/process/:id/histories/:historyId/checkout", h.CheckoutProcessHistory)
		g.GET("/process/:id", h.GetProcess)
		g.POST("/process", h.SaveProcess)
		g.PUT("/process/:id/state", h.UpdateProcessState)
		g.GET("/process/:id/check", h.CheckProcess)
		g.POST("/process/:id/clone", h.CloneProcess)
		g.DELETE("/process/:id", h.DeleteProcess)

		// 审批运行时
		g.GET("/approve/launch/list", h.ListLaunchProcesses)
		g.GET("/approve/launch/:id", h.GetLaunchForm)
		g.POST("/approve/launch", h.LaunchProcess)
		g.POST("/approve/draft/activate", h.ActivateDraft)
		g.POST("/approve/draft/update", h.UpdateDraft)
		g.POST("/approve/draft/delete", h.DeleteDraft)
		g.GET("/approve/pending", h.ListPendingTasks)
		g.GET("/approve/mine", h.ListMyApplications)
		g.GET("/approve/monitor", h.ListMonitor)
		g.GET("/approve/received", h.ListReceived)
		g.POST("/approve/received/read", h.MarkReceivedRead)
		g.GET("/approve/approved", h.ListApproved)
		g.GET("/approve/claim", h.ListClaimable)
		g.POST("/approve/claim", h.ClaimTask)
		g.GET("/approve/instance/:id", h.GetInstanceDetail)
		g.POST("/approve/consent", h.ConsentTask)
		g.POST("/approve/batch-consent", h.BatchConsent)
		g.POST("/approve/resubmit", h.ResubmitTask)
		g.POST("/approve/reject", h.RejectTask)
		g.POST("/approve/batch-reject", h.BatchReject)
		g.POST("/approve/revoke", h.RevokeInstance)
		g.POST("/approve/transfer", h.TransferTask)
		g.POST("/approve/terminate", h.TerminateInstance)
		g.POST("/approve/admin-transfer", h.AdminTransferTask)
		g.POST("/approve/rollback", h.RollbackTask)
		g.POST("/approve/urge", h.UrgeInstance)
		g.POST("/approve/append", h.AppendActor)
		g.POST("/approve/remove-actor", h.RemoveActor)
		g.POST("/approve/cc", h.AddRuntimeCC)
		g.POST("/approve/comment", h.AddComment)
		g.GET("/approve/delegate", h.GetMyDelegate)
		g.PUT("/approve/delegate", h.SetMyDelegate)
		g.POST("/approve/tick-delay", h.TickDelayTasks) // 兼容旧路径
		g.POST("/approve/tick-timers", h.TickDelayTasks)

		// 表单分类
		g.GET("/form-category/tree", h.ListFormCategoryTree)
		g.GET("/form-category/options", h.ListFormCategoryOptions)
		g.POST("/form-category", h.CreateFormCategory)
		g.PUT("/form-category", h.UpdateFormCategory)
		g.DELETE("/form-category/:id", h.DeleteFormCategory)

		// 表单模板
		g.GET("/form/options", h.ListFormOptions)
		g.GET("/form/:id", h.GetForm)
		g.POST("/form", h.SaveForm)
		g.PUT("/form/:id/state", h.UpdateFormState)
		g.PUT("/form/:id/schema", h.SaveFormSchema)
		g.DELETE("/form/:id", h.DeleteForm)
	}
}
