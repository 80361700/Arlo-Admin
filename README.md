# Arlo Admin

**定位**：**审批流引擎** + **通用后台管理底座**。

自研可视化审批流（设计 → 发起 → 办理 → 监控），业务单据可直接挂流程；同时具备账号权限、组织、日志、文件、配置、消息、监控等后台能力，按模块继续扩展即可。

架构与目录约定见 [HANDOFF.md](./HANDOFF.md)。

## 在线体验

- 地址：[http://101.200.43.49/](http://101.200.43.49/)
- 账号：`admin` / `admin123`


## 工作流审批

| 能力 | 说明 |
|------|------|
| 流程 / 表单设计 | 分类；模型与 Schema 落库；定义历史版本与检出 |
| 节点 | 审批（依次 / 会签 / 或签）、抄送、条件分支、延时、触发、子流程 |
| 办理人 | 指定人 / 角色 / 部门负责人 / 发起人自选；加签、减签、转办、委托 |
| 办理动作 | 同意 / 拒绝 / 退回、批量同意与拒绝、催办、撤销、草稿、改单重提、评论、运行时抄送 |
| 列表 | 发起、待办、已办、我的申请、我收到的、认领、流程监控（含终止 / 管理员转办） |
| 运维与扩展 | `flow_tick` 延时推进与超时/提醒；打印；业务单绑 `process_id`（演示：采购单） |

| 端 | 路径 |
|----|------|
| 后端 | `arlo-admin-server/internal/modules/flow/`（`engine/` 推进逻辑） |
| 前端 | `arlo-admin-web/src/views/flow/`、`components/flowProcess/`、`components/designer-extensions/` |
| 迁移 | 全新 `001_baseline`；已有库补跑 `002_flow_approve_runtime.sql` |

## 仓库结构

| 目录 | 说明 |
|------|------|
| [arlo-admin-server](./arlo-admin-server/) | 后端 API（启动 / 迁移 / 扩展模块） |
| [arlo-admin-web](./arlo-admin-web/) | 管理端前端（启动 / 代理 / 约定） |
| [deployments](./deployments/) | 生产部署（Compose / Linux / 宝塔） |
| [HANDOFF.md](./HANDOFF.md) | 架构约定、工作流模块说明、坑点与扩展 |

早期规划草稿已废弃，**以代码与 HANDOFF 为准**。

## 技术栈

- **后端**：Go、Gin、GORM、MySQL 8、Redis、JWT、Casbin、Zap、Viper
- **前端**：Vue 3、TypeScript、Vite、Pinia、Element Plus

## 快速开始

### 1. 数据库

```bash
mysql --default-character-set=utf8mb4 -u root -p < arlo-admin-server/migrations/001_baseline_v1.sql
# 已有旧基线库再执行：002_flow_approve_runtime.sql（见 migrations/README.md）
```

### 2. 后端

```bash
cd arlo-admin-server
# 修改 configs/config.yaml 中的 MySQL / Redis
make dev
# 默认 http://localhost:8090  · 健康检查 /health
```

### 3. 前端

```bash
cd arlo-admin-web
npm install
npm run dev
# 默认 http://localhost:5173
```

默认账号：`admin` / `admin123`（以种子数据为准）。

## 生产部署

详见 **[deployments/README.md](./deployments/README.md)**（Compose 一键、单机 Linux、宝塔）。

```bash
cp deployments/docker/.env.example deployments/docker/.env
# 编辑 MYSQL_PASSWORD / JWT_SECRET / CORS_ORIGIN
make docker-up
# 浏览器打开 http://localhost  · 默认 admin / admin123（登录后立即改密）
```

## 文档导航

| 文档 | 内容 |
|------|------|
| [HANDOFF.md](./HANDOFF.md) | 分层、权限、**工作流模块**、坑点、扩展方式 |
| [deployments/README.md](./deployments/README.md) | 生产部署 |
| [arlo-admin-server/README.md](./arlo-admin-server/README.md) | 后端启动与迁移 |
| [arlo-admin-web/README.md](./arlo-admin-web/README.md) | 前端启动与约定 |

生产上线前请完成密钥轮换、HTTPS、备份与监控（见部署文档检查清单）。
