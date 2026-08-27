# arlo-admin-server

Arlo Admin 后端服务：REST API、JWT 认证、Casbin 功能权限、数据权限、文件存储、站内信、定时任务与监控等。

在线体验：[http://101.200.43.49/](http://101.200.43.49/)（`admin` / `admin123`）

仓库总览见根目录 [README.md](../README.md)；架构细节见 [HANDOFF.md](../HANDOFF.md)。

## 目录一览

```
arlo-admin-server/
├── cmd/server/          # 入口
├── configs/             # config.yaml / config.prod.yaml / rbac_model.conf
├── internal/
│   ├── app/             # 生命周期
│   ├── domain/          # 跨模块共享：用户/角色/菜单/部门/岗位
│   ├── job/             # 进程内调度引擎
│   ├── modules/         # HTTP 业务模块（auth/system/file/…）
│   └── router/          # Gin 路由挂载
├── pkg/                 # jwt / casbin / middleware / datascope / storage …
├── migrations/          # SQL 迁移（见 migrations/README.md）
├── docs/                # Swagger 生成物
├── scripts/             # 辅助脚本（如 test_api.sh）
└── Makefile
```

## 运行

```bash
# 依赖：Go 1.21+、MySQL 8、Redis
# 先按 migrations/README.md 建库并执行补丁

# 开发
make dev
# 或
go run ./cmd/server/main.go

# 编译
make build          # 输出 build/arlo-admin

# 生产（需先 build，并设置 APP_ENV=prod）
make prod
```

默认监听：**8090**（`configs/config.yaml` → `server.port`）。  
健康检查：`GET /health`。Swagger 在开启时访问 `/swagger/index.html`。

生产配置：`APP_ENV=prod` 时先加载 `config.yaml`，再合并 `config.prod.yaml`。

## 迁移

- 全新安装：只执行 [`migrations/001_baseline_v1.sql`](./migrations/001_baseline_v1.sql)
- 说明与增量约定：[`migrations/README.md`](./migrations/README.md)
- 含中文 COMMENT 时请用：`mysql --default-character-set=utf8mb4 …`
- 基线之后的变更从 **002_*.sql** 起追加

## 扩展业务模块

1. 在 `internal/modules/{name}/` 按 handler → service → repository → model/dto 增加代码  
2. 组织主数据（User/Role/Menu 等）复用 `internal/domain`，不要复制一份  
3. 在 `internal/router/router.go` 挂载路由  
4. 需要菜单/权限码时写迁移并分配角色

进程内定时任务：执行逻辑在 `internal/job`，管理 API 在 `modules/job`。

## 部署

生产编排在仓库根目录 **[../deployments/README.md](../deployments/README.md)**（Docker Compose 一键、Linux、宝塔）。  
`logs/`、`uploads/`、`build/` 为本地/运行产物，已在 `.gitignore`。
