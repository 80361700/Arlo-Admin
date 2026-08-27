package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"arlo-admin/internal/config"
	"arlo-admin/internal/database"
	"arlo-admin/internal/domain/repository"
	"arlo-admin/internal/job"
	jobrepo "arlo-admin/internal/modules/job/repository"
	"arlo-admin/internal/router"
	"arlo-admin/pkg/appruntime"
	casbinpkg "arlo-admin/pkg/casbin"
	"arlo-admin/pkg/logger"
	"arlo-admin/pkg/sms"
)

type App struct {
	cfg        *config.Config
	configPath string
	scheduler  *job.Scheduler
}

func NewApp(configPath string, prodConfigPath string) *App {
	cfg, err := config.Load(configPath, prodConfigPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	a := &App{cfg: cfg, configPath: configPath}
	a.normalizeLocalPaths()
	return a
}

// resolveFromProjectRoot 将相对路径解析为基于配置文件所在项目根的绝对路径。
// 配置通常在 <root>/configs/config.yaml，因此 root = dirname(dirname(configPath))。
func (a *App) resolveFromProjectRoot(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	absConfig, err := filepath.Abs(a.configPath)
	if err != nil {
		return p
	}
	root := filepath.Dir(filepath.Dir(absConfig))
	return filepath.Join(root, p)
}

// normalizeLocalPaths 把日志、上传、Casbin 模型等相对路径钉到项目根，避免宝塔等工作目录不是项目根时写散落。
func (a *App) normalizeLocalPaths() {
	a.cfg.Casbin.ModelPath = a.resolveFromProjectRoot(a.cfg.Casbin.ModelPath)
	a.cfg.Log.FilePath = a.resolveFromProjectRoot(a.cfg.Log.FilePath)
	if a.cfg.Storage.Driver == "" || a.cfg.Storage.Driver == "local" {
		a.cfg.Storage.Local.Path = a.resolveFromProjectRoot(a.cfg.Storage.Local.Path)
	}
	_ = os.MkdirAll(filepath.Dir(a.cfg.Log.FilePath), 0o755)
	if a.cfg.Storage.Driver == "" || a.cfg.Storage.Driver == "local" {
		_ = os.MkdirAll(a.cfg.Storage.Local.Path, 0o755)
	}
}

func (a *App) Run() {
	appruntime.StartedAt = time.Now()

	// Initialize logger
	logger.Init(&a.cfg.Log)
	defer logger.Sync()

	// 短信渠道（mock / aliyun / tencent）
	if err := sms.Init(&a.cfg.SMS); err != nil {
		if a.cfg.Server.Mode == "debug" {
			logger.Logger.Warn("sms init failed, fallback to mock: " + err.Error())
			_ = sms.Init(&config.SMSConfig{Provider: "mock"})
		} else {
			logger.Logger.Fatal("sms init failed: " + err.Error())
		}
	}

	// Initialize MySQL
	if a.cfg.Database.Host != "" {
		if _, err := database.InitMySQL(&a.cfg.Database); err != nil {
			logger.Logger.Warn("mysql connection failed, continuing without database: " + err.Error())
		}
		defer database.CloseMySQL()
	} else {
		logger.Logger.Info("mysql not configured, skipping database initialization")
	}

	// Initialize Redis（验证码 / 会话 / 黑名单等依赖；生产必须可用）
	if a.cfg.Redis.Host != "" {
		if _, err := database.InitRedis(&a.cfg.Redis); err != nil {
			msg := "redis connection failed: " + err.Error()
			if a.cfg.Server.Mode == "debug" {
				logger.Logger.Warn(msg + " (debug: continuing without redis)")
			} else {
				logger.Logger.Fatal(msg)
			}
		}
		defer database.CloseRedis()
	} else {
		logger.Logger.Info("redis not configured, skipping redis initialization")
		if a.cfg.Server.Mode != "debug" {
			logger.Logger.Fatal("redis is required in non-debug mode")
		}
	}

	// Initialize Casbin（DB 已配置时必须成功，否则受保护接口全部不可用）
	var casbinEnforcer *casbinpkg.Enforcer
	if database.DB != nil {
		roleRepo := repository.NewRoleRepository()
		enforcer, err := casbinpkg.NewEnforcer(a.cfg.Casbin.ModelPath, roleRepo)
		if err != nil {
			logger.Logger.Error("casbin init failed: " + err.Error())
			if a.cfg.Server.Mode != "debug" {
				logger.Logger.Fatal("casbin is required when database is available")
			}
		} else if err := enforcer.LoadPolicies(context.Background()); err != nil {
			logger.Logger.Error("casbin load policies failed: " + err.Error())
			if a.cfg.Server.Mode != "debug" {
				logger.Logger.Fatal("casbin policy load failed")
			}
		} else {
			casbinEnforcer = enforcer
			logger.Logger.Info("casbin initialized with policies loaded")
		}
		if casbinEnforcer == nil {
			logger.Logger.Warn("casbin enforcer is nil: protected APIs will return 权限组件未就绪")
		}
	} else if a.cfg.Server.Mode != "debug" {
		logger.Logger.Fatal("database is required in non-debug mode")
	}

	// 定时任务（DB 驱动；无 DB 则空调度器）
	var jobRepo *jobrepo.JobRepository
	if database.DB != nil {
		jobRepo = jobrepo.NewJobRepository()
	}
	a.scheduler = job.Start(jobRepo)

	// Setup router
	r := router.Setup(a.cfg.Server.Mode, casbinEnforcer, a.scheduler)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  a.cfg.Server.ReadTimeout,
		WriteTimeout: a.cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Logger.Info(fmt.Sprintf("server starting on :%d", a.cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("server failed to start: " + err.Error())
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info("shutting down server...")
	a.scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("server forced to shutdown: " + err.Error())
	}

	logger.Logger.Info("server exited gracefully")
}
