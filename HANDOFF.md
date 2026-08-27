# arlo-admin 项目交接文档

> 给后续开发（含 Cursor）用的框架说明。以**当前代码**为准；历史演进痕迹只保留约定，不鼓励大重构。

**定位**：通用后台管理**底座**（权限 / 组织 / 日志 / 文件 / 配置 / 消息 / 监控），业务功能按模块增量扩展。

---

## 1. 项目概览

| 维度 | 技术选型 |
|------|---------|
| 后端 | Go 1.21+ / Gin / GORM v2 |
| 权限 | JWT (HMAC-SHA256) + Casbin RBAC + 数据权限 `datascope` |
| 数据 | MySQL 8.0 + Redis 7 |
| 日志 / 配置 | Zap / Viper |
| 前端 | Vue 3 + TypeScript + Element Plus + Vite + Pinia |

仓库两个子项目：

| 目录 | 说明 |
|------|------|
| `arlo-admin-server/` | 后端 API |
| `arlo-admin-web/` | 管理端前端 |
| `deployments/` | 生产部署（Compose / Linux / 宝塔） |

---

## 2. 目录结构（当前）

```
arlo-admin/
├── HANDOFF.md
├── deployments/                 # 生产部署：Compose / Linux / 宝塔（见 README）
├── arlo-admin-server/
│   ├── cmd/server/main.go
│   ├── configs/                 # config.yaml / config.prod.yaml / rbac_model.conf
│   ├── migrations/              # 见 migrations/README.md（基线 001_baseline_v1）
│   ├── internal/
│   │   ├── app/                 # 生命周期：配置 → DB → Redis → Casbin → 调度 → 路由
│   │   ├── config/
│   │   ├── database/            # MySQL + Redis
│   │   ├── router/              # Gin 引擎 + 全局中间件 + 挂载各模块路由
│   │   ├── domain/              # ★ 跨模块共享的组织主数据（见 §3.2）
│   │   │   ├── model/           # User / Role / Menu / Dept / Post …
│   │   │   └── repository/
│   │   ├── job/                 # ★ 进程内调度引擎（非 HTTP 模块）
│   │   └── modules/             # 业务 HTTP 模块
│   │       ├── auth/            # 登录 / 刷新 / 当前用户（无本地 model，用 domain）
│   │       ├── system/          # 用户角色菜单部门岗位 + 字典（字典在本模块 model）
│   │       ├── log/             # 登录日志 / 操作日志
│   │       ├── message/         # 站内信 + 通知公告
│   │       ├── sysconfig/
│   │       ├── file/
│   │       ├── member/          # 会员（部分能力仍为预留）
│   │       ├── monitor/         # 在线用户 / 服务监控
│   │       └── job/             # 定时任务 CRUD API（调度执行在 internal/job）
│   ├── pkg/                     # 框架级能力（无业务）
│   │   ├── jwt / casbin / middleware / response / errors / logger
│   │   ├── datascope / storage / excel / security / captcha
│   │   ├── onlinesession / tokenblacklist / …
│   │   └── utils/
│   └── uploads/                 # 本地上传目录（运行时）
│
└── arlo-admin-web/
    └── src/
        ├── api/modules/         # 与后端模块大致对应
        ├── components/          # ProTable / ProFormDialog / RichEditor / FilePicker …
        ├── layout/              # Sidebar / Navbar
        ├── views/
        │   ├── login / dashboard / error
        │   └── system/          # 现阶段管理页多挂于此；新业务建议 views/{业务域}/
        ├── stores/              # auth / app / message
        ├── themes/ + styles/themes/  # 浅色 / 深色等主题
        └── router/
```

---

## 3. 核心架构约定

### 3.1 业务模块分层（六文件）

仅被**一个**模块使用的表，放在该模块下：

```
internal/modules/{module}/
├── model/        # 可选：仅本模块表
├── dto/
├── repository/   # 可选：仅本模块仓储
├── service/
├── handler/
└── routes.go     # 依赖组装 + 注册路由
```

依赖方向（单向）：

```
routes.go → handler → service → repository → MySQL
```

硬性规则：

- handler **禁止**写 SQL
- service **禁止**碰 `gin.Context`
- DTO 时间字段用 `string`，格式 `2006-01-02 15:04:05`（`pkg/utils` 的 `FormatTime` / `FormatPtrTime`）

### 3.2 `internal/domain`——共享内核（有意设计，不是半成品）

**目的**：避免 `auth` 与 `system`（以及 Casbin 加载）各自复制 User/Role/Menu 等模型。

| 放 `domain` | 放 `modules/{x}/model` |
|-------------|-------------------------|
| 跨模块共享的表：用户、角色、菜单、部门、岗位及关联 | 只被一个模块使用的表：字典、文件、站内信、日志、配置、会员、任务… |

规则：

1. **禁止**在 `auth` / `system` 再定义一份 `User` / `Role` / `Menu` 等
2. 新表：若会被 ≥2 个模块引用 → 进 `domain`；否则进对应 `modules/{x}`
3. `domain` **不是**完整 DDD 领域层，只是共享主数据（shared kernel）；勿往里堆业务 Service

当前引用方：`modules/auth`、`modules/system`、`pkg/casbin`、`internal/app` 等。

### 3.3 `internal/job` vs `modules/job`

| 路径 | 职责 |
|------|------|
| `internal/job` | 进程内调度引擎（cron、执行、防并发）；**无 HTTP** |
| `modules/job` | 定时任务的管理 API（列表/启停/手动触发等） |

新增内置任务：在 `internal/job` 注册 handler；元数据进 `sys_job`（见迁移 `024`）。

### 3.4 中间件与路由挂载

全局（`router.Setup`）：

```
CORS → RequestID → RequestLogger → Recovery
```

`/api/v1` 组额外：`OperationLog`（异步写操作日志）。

模块在 `routes.go` 内自行套 `JWTAuth` / `CasbinAuth`。常见情况：

| 范围 | 认证 |
|------|------|
| `/auth/login` 等公开接口 | 无 |
| `/auth/*` 需登录 | JWT |
| `/system/*`、`/sysconfig/*`、文件管理写操作等 | JWT + Casbin |
| 文件下载 / 部分公开读 | 见 §6（可能仅 JWT 或公开策略） |
| `/log/*`、`/message/*` 等 | 以各模块 `routes.go` 为准 |

模块注册入口：`internal/router/router.go`。

### 3.5 JWT

- Access ~2h / Refresh ~7d（`config.yaml` 可配）
- `Subject` 区分 `access` / `refresh`
- 登出：客户端丢 token + Redis 黑名单 / 踢下线（`tokenblacklist`、`onlinesession`）
- 下载场景额外支持 `?token=`（见 §6.1）

### 3.6 Casbin RBAC

- 模型：`configs/rbac_model.conf`（`keyMatch2` + `regexMatch`）
- 启动从 DB 加载到内存；改菜单权限后需 `ReloadPolicies`（或重启）
- 按钮权限：菜单 `type=3`，且写入 `sys_role_menu`，否则 `/auth/info` 的 `permissions` 无该码，前端 `v-permission` 会隐藏

### 3.7 数据权限（`pkg/datascope`）

- 角色字段 `data_scope` + `sys_role_dept`；多角色取最宽松
- **已接入列表**：用户、文件、操作/登录日志、公告、站内信发送端等
- **刻意不接**：字典/岗位/部门/角色/菜单/系统配置（全局主数据）；会员（缺归属字段）

### 3.8 配置

```
开发: go run cmd/server/main.go
生产: APP_ENV=prod …  → config.yaml + merge config.prod.yaml
```

---

## 4. 数据库与迁移

### 4.1 执行方式

详见 **`arlo-admin-server/migrations/README.md`**。

- **全新安装**：只执行 `001_baseline_v1.sql`（表结构 + 标准种子）
- **后续变更**：追加 `002_*.sql` 起的增量补丁，序号递增、禁止复用
- 旧迭代 `001`～`026` 已归档至 `migrations/archive/pre_v1/`，勿再对新库连跑
- 含中文 COMMENT 的 SQL：**必须**用 `utf8mb4` 客户端导入

### 4.2 核心表（基线 v1，共 22 张）

| 表 | 说明 |
|----|------|
| sys_user / sys_role / sys_user_role / sys_menu / sys_role_menu | 账号与功能权限 |
| sys_dept / sys_post / sys_user_post / sys_role_dept | 组织与数据范围 |
| sys_dict_type / sys_dict_data | 字典 |
| sys_config | 系统配置（常配 Redis 缓存） |
| sys_notice | 通知公告 |
| sys_message / sys_message_read / sys_message_hide | 站内信；广播个人隐藏见 §6.2 |
| sys_file | 文件（MD5 去重、access_key、公开策略等） |
| login_log → **sys_login_log** / **sys_operation_log** | 日志 |
| sys_member | 会员（独立于后台用户） |
| sys_job / sys_job_log | 定时任务 |

种子：`admin` / `admin123`，角色 `super_admin`，菜单与监控/任务等以基线文件为准。

---

## 5. 后端模块能力一览

以 `internal/modules/*/routes.go` 与前端 `api/modules/*` 为准；此处不逐条抄接口。

| 模块 | 能力摘要 |
|------|----------|
| auth | 登录/登出/刷新/当前用户信息与权限码；验证码等按配置 |
| system | 用户（含改密/解锁/Excel 导入导出）、角色（含菜单与 data_scope）、部门树、菜单树、岗位、字典 |
| log | 登录/操作日志列表与导出 |
| message | 站内信（收发/已读/删除/未读数）、通知公告（发布/撤回） |
| sysconfig | 配置 CRUD；部分公开读取策略见模块实现 |
| file | 上传（MD5 去重）、列表、删除、下载；公开访问与 access_key |
| member | 管理端会员列表等；**客户端密码/微信登录等仍为预留，勿当已闭环** |
| monitor | 在线用户、强制下线、服务监控 |
| job | 任务管理 API；执行引擎在 `internal/job`（如 `log_cleanup`） |

前端对应页面目前多在 `views/system/`（含 monitor、member、message…）；**新业务域**建议新建 `views/{domain}/`，避免无限塞进 system。

---

## 6. 已知约定与坑

### 6.1 文件下载与 Token

- 管理端预览/下载常需 JWT；`<img>` / 编辑器内嵌可能用 `?token=` + `inline`
- 部分下载路径**仅 JWT、不过 Casbin**，避免无「文件管理菜单权限」的角色在富文本里 403
- 公开文件另有 access_key / 公开策略（迁移 `020`～`023`），以 `file` 模块为准
- 前端拼 URL 时 token 需 `encodeURIComponent`

### 6.2 站内信删除与发送

| 场景 | 行为 |
|------|------|
| 发送方删除 | `sender_deleted`（不影响收件人） |
| 指定收件人删除 | `receiver_deleted` |
| **广播**收件人删除 | 写入 `sys_message_hide`（个人隐藏，不能改公共行） |

发送请求只认 `receiverIds`：省略或空数组 = 广播；非空 = 指定用户批量发送。  
库字段 `sys_message.receiver_id`（0=广播）仍是行数据，不是请求兼容字段。  
发送记录列表有聚合逻辑（同批发送），见 message repository。

### 6.3 角色菜单树（前端）

`check-strictly=false` 父子联动；回显只设叶子；提交 `checked + halfChecked`。

### 6.4 操作日志

中间件读 GET Query / 写操作 Body（ReadAll 后写回）；**默认不落完整响应体**（体积与敏感信息）。

### 6.5 上传

- 先算 MD5，命中则复用存储记录
- 前端 `uploadFile` 须去掉强制 `application/json`，让浏览器带 multipart boundary

### 6.6 ProTable 搜索

查询参数会带上全部筛选项；清空的字段以 `undefined` 发出，避免父页 `Object.assign(query, p)` 残留旧值。

### 6.7 主题与布局

`stores/app`：`themeId`（`styles`/`themes` 外观）+ `layoutMode`（`side` | `mix` | `topbar`）。  
顶栏画笔悬停打开外观面板（主题/导航预览图标），本地键 `arlo-theme` / `arlo-layout-mode`（旧值 `top` 自动映射为 `mix`）。  

- 侧栏：左侧完整菜单 + 顶栏面包屑  
- 混合：顶栏 Logo+一级，左侧二级；首页默认第一模块；无面包屑  
- 顶栏：顶栏 Logo+完整菜单树（下拉），无侧栏；内容区顶部无底色面包屑  

无首页顶栏项（首页在 Logo/用户菜单）。布局改造前备份：`.backup/layout-before-topnav-20260806/`。

---

## 7. 进度与后续（底座视角）

### 已具备（可当底座）

- [x] 骨架、JWT、Casbin、数据权限
- [x] 组织与系统管理、日志、消息、配置、文件
- [x] 在线用户 / 踢下线 / 服务监控 / 进程内定时任务
- [x] 用户与登录日志 Excel 导入导出
- [x] 前端动态路由、权限指令、通用 Pro 组件、主题

### 建议后续（非阻塞业务开发）

1. 单元测试（优先 service）与 API 文档保持（已有 swag 草稿能力时按需生成）
2. 会员客户端鉴权闭环（若真要做 C 端）
3. 消息 WebSocket（当前未读角标为轮询）
4. 生产加固：备份、监控告警、HTTPS；编排见仓库根 `deployments/README.md`
5. 可选：新环境 schema 快照，减少从 001 起跑历史补丁

**不建议**：为「目录好看」强行大搬家 `domain`。

---

## 8. 运行指南

### 后端

```bash
cd arlo-admin-server
# 建库：执行基线 SQL（见 migrations/README.md）
mysql --default-character-set=utf8mb4 -u root -p < migrations/001_baseline_v1.sql
# 改 configs/config.yaml
go run cmd/server/main.go
# http://localhost:8090/health
```

### 前端

```bash
cd arlo-admin-web
npm install
npm run dev
# http://localhost:5173  默认 admin / admin123
```

---

## 9. 速查：想改什么去哪

| 想做的事 | 位置 |
|----------|------|
| 新增业务模块 | 仿 `modules/message/`；共享组织实体用 `domain`，勿复制 User |
| 改组织用户/角色/菜单模型或仓储 | `internal/domain/` |
| 挂路由 / 全局中间件 | `internal/router/router.go` + 模块 `routes.go` |
| 加内置定时任务 | `internal/job` 注册 + `sys_job` 数据 |
| 改任务管理 API | `modules/job` |
| 改表结构 | **新**迁移 `NNN_*.sql`（序号递增、勿复用）+ 对应 model |
| 改配置项 | `configs/config.yaml` + `internal/config` |
| 改错误码 / 响应格式 | `pkg/errors` / `pkg/response` |
| 改 JWT / Casbin / 数据权限 | `pkg/jwt` / `pkg/casbin` / `pkg/datascope` + `rbac_model.conf` |
| 加管理页 | `views/{域}/` + `api/modules/`；菜单 `component` 勿带 `.vue` |
| 改通用表格 / 筛选 | `components/ProTable.vue` |
| 改主题 | `styles/themes/` + `stores/app` |
