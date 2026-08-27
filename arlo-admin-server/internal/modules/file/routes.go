package file

import (
	"arlo-admin/internal/config"
	"arlo-admin/internal/modules/file/handler"
	"arlo-admin/internal/modules/file/repository"
	"arlo-admin/internal/modules/file/service"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/logger"
	"arlo-admin/pkg/middleware"
	"arlo-admin/pkg/storage"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册文件管理模块路由
func RegisterRoutes(r *gin.RouterGroup, enforcer *casbinpkg.Enforcer) {
	// 创建存储驱动（根据配置选择 local 或 oss）
	driver, err := createDriver()
	if err != nil {
		logger.Logger.Warn("file storage driver init failed, falling back to local: " + err.Error())
		driver = storage.NewLocalDriver("uploads/")
	}

	repo := repository.NewFileRepository()
	svc := service.NewFileService(repo, driver, getMaxSize(), getAllowedExts())
	h := handler.NewFileHandler(svc)

	// 管理操作：需要 JWT + Casbin 权限校验
	adminGroup := r.Group("/file")
	adminGroup.Use(middleware.JWTAuth(), middleware.CasbinAuth(enforcer))
	{
		adminGroup.POST("/upload", h.Upload)
		adminGroup.GET("/list", h.List)
		adminGroup.PUT("/:id/public", h.SetPublic)
		adminGroup.DELETE("/:id", h.Delete)
	}

	// 统一访问：/file/{accessKey}；公开免登，私有需 Authorization: Bearer
	r.GET("/file/:accessKey", h.Serve)
}

// createDriver 根据全局配置创建存储驱动
func createDriver() (storage.Driver, error) {
	if config.GlobalConfig == nil {
		return storage.NewLocalDriver("uploads/"), nil
	}
	return storage.NewDriver(&config.GlobalConfig.Storage)
}

// getMaxSize 获取最大上传大小配置
func getMaxSize() int64 {
	if config.GlobalConfig == nil || config.GlobalConfig.Storage.MaxSize <= 0 {
		return 50 << 20 // 默认 50MB
	}
	return config.GlobalConfig.Storage.MaxSize
}

func getAllowedExts() []string {
	if config.GlobalConfig == nil {
		return nil
	}
	return config.GlobalConfig.Storage.AllowedExts
}
