# Arlo Admin 生产部署

| 路径 | 说明 |
|------|------|
| [Docker Compose 一键](./docker/) | Nginx + 前端 + API + MySQL + Redis |
| [单机 Linux](./linux/README.md) | systemd + Nginx，无容器 |
| [宝塔面板](./baota/README.md) | 面板建站 / 反代 / SSL |

## 上线前检查清单（各路径共用）

- [ ] 更换 MySQL / Redis / JWT 密钥（勿用仓库示例值）
- [ ] 登录后立即修改默认管理员密码（`admin` / `admin123`）
- [ ] `CORS_ORIGIN` / Nginx `server_name` / 实际访问域名一致
- [ ] 配置 HTTPS（证书或面板一键 SSL）
- [ ] 确认生产关闭 Swagger（Compose / `config.prod.yaml` 已关）
- [ ] 备份策略：MySQL dump + `uploads`（多副本请改 OSS）
- [ ] 若用短信验证码：将 `sms.provider` 改为 `aliyun`/`tencent` 并填密钥
- [ ] **WebSocket 反代**：`location /api/` 已含 `Upgrade` / `Connection`（见下方说明）；否则站内信 WS 握手 400

## WebSocket（站内信未读推送）

前端连接：`wss://你的域名/api/v1/ws?token=<accessToken>`（与页面同源，经 Nginx 反代到 API `:8090`）。

REST 与 WebSocket 共用 `location /api/`，`proxy_pass` 指向 `http://127.0.0.1:8090/api/`（后端路由含 `/api` 前缀）：

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:8090/api/;   # Docker 内为 http://api:8090/api/
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_buffering off;
    proxy_read_timeout 3600s;
}
```

**线上握手失败 `Unexpected response code: 400`** 几乎都是 **`/api/` 反代缺少 `Upgrade` / `Connection` / `proxy_http_version 1.1`**。

**可选：分开配置**（WebSocket 与 API 需不同超时时）：

```nginx
location = /api/v1/ws {
    proxy_pass http://127.0.0.1:8090/api/v1/ws;   # Docker 内为 http://api:8090/api/v1/ws
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_buffering off;
    proxy_read_timeout 3600s;
}

location /api/ {
    proxy_pass http://127.0.0.1:8090/api/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 60s;
}
```

WS 段须写在 `location /api/` 之前。

改完后 `nginx -t && nginx -s reload`（或宝塔保存并重载）。

### 本地开发 vs 生产（配置在哪）

| 环境 | 谁做转发 | 配置文件 |
|------|----------|----------|
| **本地 `npm run dev`** | **Vite 开发服务器**（不是 Nginx） | [`arlo-admin-web/vite.config.ts`](../arlo-admin-web/vite.config.ts) 里 `server.proxy['/api']`，其中 **`ws: true`** 负责 WebSocket |
| **生产 / 宝塔 / Docker** | **Nginx** | 本文下方 Nginx 片段，或 [`docker/nginx.conf`](./docker/nginx.conf)、[`linux/nginx-site.conf`](./linux/nginx-site.conf)、[`baota/README.md`](./baota/README.md) §7 |

本地流程：`浏览器 ws://localhost:5173/api/v1/ws` → Vite 代理 → `http://localhost:8090/api/v1/ws`（后端）。  
生产流程：`浏览器 wss://域名/api/v1/ws` → Nginx 反代 → `127.0.0.1:8090`。  

前端 [`stores/message.ts`](../arlo-admin-web/src/stores/message.ts) 连接当前站点同源的 `/api/v1/ws`，无单独 WS 环境变量。

**验证**：浏览器 DevTools → Network → WS，状态应为 **101 Switching Protocols**；Messages 里可见服务端 ping/pong 或 `unread_changed` 事件。

**其它注意**：
- 前置 **CDN / 云 WAF** 需开启 WebSocket（Cloudflare 默认支持；部分套餐需勾选）
- API 需能直连 `:8090`（或容器内 `api:8090`）；仅改 Nginx 不够时检查防火墙
- Compose / Linux 模板已含上述片段：[docker/nginx.conf](./docker/nginx.conf)、[linux/nginx-site.conf](./linux/nginx-site.conf)

## Docker Compose 一键部署

详见 **[docker/README.md](./docker/README.md)**。快捷方式：

```bash
cp deployments/docker/.env.example deployments/docker/.env
# 编辑 MYSQL_PASSWORD / JWT_SECRET / CORS_ORIGIN
make docker-up
```
