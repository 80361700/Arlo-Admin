# Docker Compose 一键部署

本目录包含 Compose 编排与镜像构建所需全部文件。

| 文件 | 说明 |
|------|------|
| `docker-compose.yml` | mysql / redis / api / web |
| `.env.example` | 密钥模板（复制为 `.env`） |
| `Dockerfile.api` | 后端镜像 |
| `Dockerfile.web` | 前端 + Nginx 镜像 |
| `nginx.conf` | web 容器内 Nginx |
| `entrypoint-api.sh` | 启动时用环境变量生成 `config.prod.yaml` |
| `config.prod.docker.yaml.tmpl` | 生产配置模板 |

## 依赖

- Docker Engine 20+ 与 Docker Compose v2
- 开放宿主机 `HTTP_PORT`（默认 80）

## 步骤

```bash
# 在仓库根目录
cp deployments/docker/.env.example deployments/docker/.env
# 编辑 .env：MYSQL_PASSWORD、JWT_SECRET、CORS_ORIGIN、HTTP_PORT

make docker-up
# 或：
# docker compose -f deployments/docker/docker-compose.yml --env-file deployments/docker/.env up -d --build
```

首次启动 MySQL 会自动执行 [`arlo-admin-server/migrations/001_baseline_v1.sql`](../../arlo-admin-server/migrations/001_baseline_v1.sql)。

访问：`http://localhost`（或服务器 IP / 域名）。健康检查：`/health`。

默认账号：`admin` / `admin123`（登录后立即改密）。

## 常用命令

```bash
make docker-ps
make docker-logs
make docker-down          # 停服务，保留数据卷
make docker-down-volumes  # 停服务并删除数据（危险）
```

## 架构

```
浏览器 → web(Nginx:80) → 静态 dist
                      ↘ /api（含 WebSocket）、/health → api:8090 → MySQL / Redis
```

- 仅 `web` 映射到宿主机；MySQL / Redis **不对公网暴露**
- [`nginx.conf`](./nginx.conf) 在 `location /api/` 内已含 WebSocket 所需 `Upgrade` 头；自建 Nginx 见 [部署总览 · WebSocket](../README.md#websocket站内信未读推送)
- API 密钥由 `entrypoint-api.sh` 从环境变量写入临时 `config.prod.yaml`
- 文件默认落在 Docker 卷 `api-uploads`；多实例请改 OSS

## HTTPS

默认仅 HTTP。可在前置负载均衡 / CDN 终结 TLS，或改 [`nginx.conf`](./nginx.conf) 启用 443 并挂载证书。

## 注意点

- `.env` 中密码勿含未转义的 `$` / 复杂引号
- 改密后重建 API 容器即可；**不要**随便 `down -v`
- 基线 SQL 仅在空数据卷首次初始化时执行
