# 宝塔面板部署

在已安装宝塔的 **Linux** 主机上发布 Arlo Admin。运行机一般是 `linux/amd64`（常见云主机）；少数 ARM 机器为 `linux/arm64`。

构建可在本机完成后再上传，不必在宝塔服务器上装 Go / Node（当然也可以在服务器上直接编）。

## 1. 面板环境

- 软件商店安装：**Nginx**、**MySQL 8**、**Redis**
- （可选）Supervisor，或使用 systemd 托管 API 进程

先确认服务器架构（影响后端交叉编译）：

```bash
uname -m
# x86_64 → amd64
# aarch64 / arm64 → arm64
```

---

## 2. 构建产物（本机）

需要：

- Go 1.21+（编后端）
- Node.js 20+（编前端）
- 仓库源码

### 2.1 后端：编译给 Linux 用的二进制

宝塔跑的是 Linux，**在 macOS / Windows 上不能直接 `make build`**（编出来是本机系统的程序，服务器跑不了）。要用交叉编译，指定 `GOOS=linux` 和对应的 `GOARCH`。

在仓库的 `arlo-admin-server/` 目录执行：

**目标：Linux x86_64（最常见）**

| 你在哪台机器上编 | 命令 |
|------------------|------|
| macOS / Linux / Windows（Git Bash 或 PowerShell） | 见下方同一条 |

```bash
cd arlo-admin-server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/arlo-admin ./cmd/server/main.go
```

**目标：Linux ARM64**（`uname -m` 为 `aarch64` 时）

```bash
cd arlo-admin-server
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/arlo-admin ./cmd/server/main.go
```

**若构建机本身就是目标 Linux 服务器**（同架构），可以直接：

```bash
cd arlo-admin-server
make build
# 产物：build/arlo-admin
```

说明：

| 项 | 含义 |
|----|------|
| `CGO_ENABLED=0` | 纯静态/无 C 依赖，方便拷到不同 Linux 发行版 |
| `GOOS=linux` | 目标操作系统 |
| `GOARCH=amd64` / `arm64` | 目标 CPU；必须和宝塔机器一致 |
| Windows | 在 **PowerShell** 里写 `$env:CGO_ENABLED=0; $env:GOOS="linux"; $env:GOARCH="amd64"; go build -o build/arlo-admin ./cmd/server/main.go`；或在 **Git Bash / WSL** 里用上面的 Unix 写法 |

产物路径：`arlo-admin-server/build/arlo-admin`（无扩展名，不是 `.exe`）。

### 2.2 前端：与操作系统无关

前端构建结果是静态文件，Mac / Windows / Linux 命令相同：

```bash
cd arlo-admin-web
npm ci
npm run build
# 产物目录：dist/
```

### 2.3 需要上传的内容

| 本地路径 | 上传到服务器 |
|----------|----------------|
| `arlo-admin-server/build/arlo-admin` | `bin/arlo-admin` |
| `arlo-admin-server/configs/`（至少含 `config.yaml`、`config.prod.yaml`、`rbac_model.conf`） | `configs/` |
| `arlo-admin-web/dist/` 内的全部文件 | `web/` |
| `arlo-admin-server/migrations/001_baseline_v1.sql` | 导入数据库用（不必长期放站点目录） |

---

## 3. 数据库

推荐安装 **MySQL 8**（与项目一致）。若是 MySQL 5.7 / MariaDB，基线已统一为 `utf8mb4_unicode_ci`，一般可导入。

1. 宝塔 → 数据库 → 添加库 `arlo_admin`（字符集 utf8mb4），记下账号密码  
2. 在服务器上执行基线（或用 phpMyAdmin 导入，字符集选 **utf8mb4**）：

```bash
mysql --default-character-set=utf8mb4 -u用户 -p < /path/to/001_baseline_v1.sql
```

基线 SQL 内含 `CREATE DATABASE` / `USE` 时，也可对 root 直接导入整文件。  
若导入报 `Unknown collation: utf8mb4_0900_ai_ci`，请拉取最新仓库的 `001_baseline_v1.sql` 再导，或把 SQL 里该排序规则全局替换为 `utf8mb4_unicode_ci`。

---

## 4. 上传与目录

建议目录：

```text
/www/wwwroot/arlo-admin/
  bin/arlo-admin
  configs/
  web/          # 前端 dist 内容
  logs/
  uploads/
```

```bash
# 示例：在服务器上整理权限
mkdir -p /www/wwwroot/arlo-admin/{bin,configs,web,logs,uploads}
chmod +x /www/wwwroot/arlo-admin/bin/arlo-admin
# 运行用户（如 www）需对 logs、uploads 可写
chown -R www:www /www/wwwroot/arlo-admin/logs /www/wwwroot/arlo-admin/uploads
```

`logs/`、`uploads/` **不是** Docker 那种挂载，就是项目目录下的普通文件夹：

| 目录 | 用途 |
|------|------|
| `logs/` | 应用日志（`logs/app.log`） |
| `uploads/` | 本地文件存储（未开 OSS 时） |

请保证目录存在且 `www` 可写。若不用 `start.sh`、工作目录又不在项目根，旧二进制可能把文件写到别的 cwd 下；请用 `start.sh`，或重新编译包含「按项目根解析路径」的版本后再覆盖 `bin/arlo-admin`。

可用宝塔文件管理上传，或 `scp`：

```bash
# 在构建机上（按实际 IP/路径改）
scp arlo-admin-server/build/arlo-admin root@服务器:/www/wwwroot/arlo-admin/bin/
scp -r arlo-admin-server/configs/* root@服务器:/www/wwwroot/arlo-admin/configs/
scp -r arlo-admin-web/dist/* root@服务器:/www/wwwroot/arlo-admin/web/
```

---

## 5. 配置

编辑服务器上的 `configs/config.prod.yaml`：

- `database.host` / `user` / `password` / `dbname` 填宝塔库信息（多为 `127.0.0.1`）
- `redis.host` 一般为 `127.0.0.1`；面板若给 Redis 设了密码则同步填写
- `jwt.secret` 换成长随机串
- `corsOrigins` 填站点域名（如 `https://admin.example.com`）

启动时需 `APP_ENV=prod`（见下一节），程序会合并 `config.yaml` + `config.prod.yaml`。

---

## 6. 守护进程

**方式 A：systemd**（推荐）  
使用 [`../linux/arlo-admin.service`](../linux/arlo-admin.service)，按实际路径改 `WorkingDirectory` / `ExecStart` / `User`。

**方式 B：Supervisor / 宝塔「Go项目」**

推荐用启动脚本（自动 `cd` 到项目根，避免找不到 `rbac_model.conf`）：

1. 将 [`start.sh`](./start.sh) 上传到 `/www/wwwroot/arlo-admin/start.sh`
2. `chmod +x /www/wwwroot/arlo-admin/start.sh`
3. 面板「执行命令」填：

```bash
/www/wwwroot/arlo-admin/start.sh
```

或不用脚本时，执行命令填：

```bash
/usr/bin/env APP_ENV=prod /www/wwwroot/arlo-admin/bin/arlo-admin --config=/www/wwwroot/arlo-admin/configs/config.yaml --prod-config=/www/wwwroot/arlo-admin/configs/config.prod.yaml
```

并确认服务器上存在：`/www/wwwroot/arlo-admin/configs/rbac_model.conf`（与 `config.yaml` 同目录）。  
若仍报找不到该文件，请改用上面的 `start.sh`（宝塔工作目录往往不是项目根）。

- 运行用户：`www`（需能读 `configs/`，能写 `logs/`、`uploads/`）
- 开机启动：是

不要写成 `APP_ENV=prod /www/.../arlo-admin ...`（无 `env`）：宝塔 `nohup` 会把 `APP_ENV=prod` 当成程序名。

文件名注意是 **`config.prod.yaml`**（点分隔），不是 `config-prod.yaml`。

启动失败时先在 SSH 前台排查：

```bash
cd /www/wwwroot/arlo-admin
./start.sh
# 或
ls configs/rbac_model.conf configs/config.yaml configs/config.prod.yaml
```

若提示 `cannot execute binary file`，多半是 `GOARCH` 与服务器 CPU 不一致，按第 2.1 节重编。

---

## 7. 网站与反代

1. 宝塔 → 网站 → 添加站点（域名指向本机）  
2. 根目录设为 `/www/wwwroot/arlo-admin/web`  
3. 设置 → 配置文件，配置 SPA 回退与 API 反代：

```nginx
location / {
    try_files $uri $uri/ /index.html;
}

# API + WebSocket（/api/v1/ws 与 REST 共用）
location /api/ {
    proxy_pass http://127.0.0.1:8090/api/;
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

location /health {
    proxy_pass http://127.0.0.1:8090/health;
}
```

**可选：分开配置**（WebSocket 与 API 需不同超时时，将下面两段替换上面的 `location /api/`，WS 段写在 `/api/` 之前）：

```nginx
location = /api/v1/ws {
    proxy_pass http://127.0.0.1:8090/api/v1/ws;
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

4. SSL → Let's Encrypt 一键申请，强制 HTTPS

**WebSocket 仍报 400 时**：确认 `location /api/` 含 `proxy_http_version 1.1`、`Upgrade`、`Connection "upgrade"`；保存后重载 Nginx；确认 API 进程在 `8090` 运行（`curl http://127.0.0.1:8090/health`）；登录后在 DevTools → Network → WS 查看状态是否为 **101**。

---

## 8. 验证

- 浏览器打开站点，使用 `admin` / `admin123` 登录后立刻改密  
- `https://你的域名/health` 应返回 JSON `status: ok`
- 登录后 DevTools → Network → **WS**，`/api/v1/ws` 状态应为 **101**（不是 400）

---

## 9. 更新

1. 本机按第 2 节重新构建  
2. 覆盖服务器上的 `bin/arlo-admin` 与 `web/`  
3. 有增量 SQL（`002_*.sql` 起）时先执行再重启守护进程  
4. 重启 Supervisor / systemd  
