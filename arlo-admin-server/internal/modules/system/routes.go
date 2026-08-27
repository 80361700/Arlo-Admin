package system

import (
	"arlo-admin/internal/domain/repository"
	configrepo "arlo-admin/internal/modules/sysconfig/repository"
	configsvc "arlo-admin/internal/modules/sysconfig/service"
	"arlo-admin/internal/modules/system/handler"
	sysrepo "arlo-admin/internal/modules/system/repository"
	"arlo-admin/internal/modules/system/service"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册系统模块路由
func RegisterRoutes(r *gin.RouterGroup, enforcer *casbinpkg.Enforcer) {
	// --- 依赖注入链 ---
	userRepo := repository.NewUserRepository()
	roleRepo := repository.NewRoleRepository()
	deptRepo := repository.NewDeptRepository()
	menuRepo := repository.NewMenuRepository()
	postRepo := repository.NewPostRepository()
	dictRepo := sysrepo.NewDictRepository()
	cfgSvc := configsvc.NewConfigService(configrepo.NewConfigRepository())

	userSvc := service.NewUserService(userRepo, roleRepo, deptRepo, postRepo, enforcer, cfgSvc)
	roleSvc := service.NewRoleService(roleRepo, deptRepo, enforcer)
	deptSvc := service.NewDeptService(deptRepo)
	menuSvc := service.NewMenuService(menuRepo, enforcer)
	postSvc := service.NewPostService(postRepo)
	dictSvc := service.NewDictService(dictRepo)

	userHandler := handler.NewUserHandler(userSvc)
	roleHandler := handler.NewRoleHandler(roleSvc)
	deptHandler := handler.NewDeptHandler(deptSvc)
	menuHandler := handler.NewMenuHandler(menuSvc)
	postHandler := handler.NewPostHandler(postSvc)
	dictHandler := handler.NewDictHandler(dictSvc)

	// --- 下拉选项接口：仅需登录 ---
	// 用户管理等表单会复用角色/岗位/部门列表，不能要求同时拥有「岗位管理」等菜单
	options := r.Group("/system")
	options.Use(middleware.JWTAuth())
	{
		options.GET("/role/all", roleHandler.GetAll)
		options.GET("/post/all", postHandler.GetAll)
		options.GET("/dept/tree", deptHandler.GetTree)
		options.GET("/dict/type/all", dictHandler.GetAllDictTypes)
		options.GET("/dict/data/code/:code", dictHandler.GetDictDatasByCode)
	}

	// --- 业务 CRUD：JWT + Casbin ---
	system := r.Group("/system")
	system.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))

	// 用户管理
	{
		user := system.Group("/user")
		user.GET("/list", userHandler.List)
		user.GET("/export", userHandler.Export)
		user.GET("/import/template", userHandler.ImportTemplate)
		user.POST("/import", userHandler.Import)
		user.PUT("/:id/unlock", userHandler.Unlock)
		user.GET("/:id", userHandler.GetDetail)
		user.POST("", userHandler.Create)
		user.PUT("", userHandler.Update)
		user.DELETE("/:id", userHandler.Delete)
		user.PUT("/password", userHandler.UpdatePassword)
	}

	// 角色管理
	{
		role := system.Group("/role")
		role.GET("/list", roleHandler.List)
		role.GET("/:id", roleHandler.GetDetail)
		role.GET("/:id/menus", roleHandler.GetMenus)
		role.POST("", roleHandler.Create)
		role.PUT("", roleHandler.Update)
		role.DELETE("/:id", roleHandler.Delete)
		role.POST("/assignMenus", roleHandler.AssignMenus)
	}

	// 部门管理（tree 已放到 options 组）
	{
		dept := system.Group("/dept")
		dept.POST("", deptHandler.Create)
		dept.PUT("", deptHandler.Update)
		dept.DELETE("/:id", deptHandler.Delete)
	}

	// 菜单管理
	{
		menu := system.Group("/menu")
		menu.GET("/tree", menuHandler.GetTree)
		menu.POST("", menuHandler.Create)
		menu.PUT("", menuHandler.Update)
		menu.DELETE("/:id", menuHandler.Delete)
	}

	// 岗位管理
	{
		post := system.Group("/post")
		post.GET("/list", postHandler.List)
		post.GET("/:id", postHandler.GetDetail)
		post.POST("", postHandler.Create)
		post.PUT("", postHandler.Update)
		post.DELETE("/:id", postHandler.Delete)
	}

	// 字典管理
	{
		dictType := system.Group("/dict/type")
		dictType.GET("/list", dictHandler.ListDictTypes)
		dictType.GET("/:id", dictHandler.GetDictType)
		dictType.POST("", dictHandler.CreateDictType)
		dictType.PUT("", dictHandler.UpdateDictType)
		dictType.DELETE("/:id", dictHandler.DeleteDictType)

		dictData := system.Group("/dict/data")
		dictData.GET("/list", dictHandler.ListDictDatas)
		dictData.GET("/type/:id", dictHandler.GetDictDatasByTypeID)
		dictData.GET("/:id", dictHandler.GetDictData)
		dictData.POST("", dictHandler.CreateDictData)
		dictData.PUT("", dictHandler.UpdateDictData)
		dictData.DELETE("/:id", dictHandler.DeleteDictData)
	}
}
