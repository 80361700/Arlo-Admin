package main

import (
	"flag"
	"os"

	"arlo-admin/internal/app"
)

// @title           arlo-admin API
// @version         1.0.0
// @description     管理后台 API 接口文档
// @contact.name    arlo-admin
// @host            localhost:8090
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// 命令行参数
	configPath := flag.String("config", "configs/config.yaml", "基础配置文件路径")
	prodConfig := flag.String("prod-config", "", "生产环境覆盖配置文件路径")
	flag.Parse()

	// 如果 APP_ENV=prod 且未通过 flag 指定 prod-config，自动使用 configs/config.prod.yaml
	env := os.Getenv("APP_ENV")
	prodPath := *prodConfig
	if prodPath == "" && (env == "prod" || env == "production") {
		prodPath = "configs/config.prod.yaml"
	}

	application := app.NewApp(*configPath, prodPath)
	application.Run()
}
