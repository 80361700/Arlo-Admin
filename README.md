# Arlo Admin

通用后台管理**底座**：账号权限、组织、日志、文件、配置、消息、监控等开箱能力齐全，业务功能按模块增量扩展。

## 仓库结构

| 目录 | 说明 |
|------|------|
| [arlo-admin-server](./arlo-admin-server/) | 后端 API（Go + Gin + GORM + Casbin） |
| [arlo-admin-web](./arlo-admin-web/) | 管理端前端（Vue 3 + Element Plus） |
| [deployments](./deployments/) | 生产部署（Compose / Linux / 宝塔） |
| [HANDOFF.md](./HANDOFF.md) | 架构约定与开发交接（给后续开发 / AI 用） |

早期规划草稿已废弃，**以代码与 HANDOFF 为准**。

## 技术栈

- **后端**：Go、Gin、GORM、MySQL 8、Redis、JWT、Casbin、Zap、Viper
- **前端**：Vue 3、TypeScript、Vite、Pinia、Element Plus

## 快速开始

### 1. 数据库

```bash
mysql --default-character-set=utf8mb4 -u root -p < arlo-admin-server/migrations/001_baseline_v1.sql
# 详见 arlo-admin-server/migrations/README.md
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
| [deployments/README.md](./deployments/README.md) | 生产部署（Compose / Linux / 宝塔） |
| [arlo-admin-server/README.md](./arlo-admin-server/README.md) | 后端目录、运行、迁移 |
| [arlo-admin-web/README.md](./arlo-admin-web/README.md) | 前端目录、开发代理、主题与约定 |
| [HANDOFF.md](./HANDOFF.md) | 分层、`domain`、权限、坑点、扩展方式 |

## 定位说明

适合作为内部或商业项目的管理后台骨架；会员微信等部分能力仍为预留。生产上线前请完成密钥轮换、HTTPS、备份与监控（见部署文档检查清单）。
