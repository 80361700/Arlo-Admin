# 单机 Linux 部署（systemd + Nginx）

适用于已有 MySQL 8、Redis、Nginx 的裸机或云主机（无 Docker）。

## 1. 准备依赖

- Go 1.21+（仅构建机需要；运行机只需二进制）
- Node.js 20+（仅构建前端）
- MySQL 8、Redis、Nginx

## 2. 数据库

```bash
mysql --default-character-set=utf8mb4 -u root -p < arlo-admin-server/migrations/001_baseline_v1.sql
```

## 3. 构建

```bash
# 后端
cd arlo-admin-server
make build
# 产物：build/arlo-admin

# 前端
cd ../arlo-admin-web
npm ci
npm run build
# 产物：dist/
```

将 `build/arlo-admin`、`configs/`、`dist/` 拷到服务器，例如：

```text
/opt/arlo-admin/
  bin/arlo-admin
  configs/config.yaml
  configs/config.prod.yaml
  configs/rbac_model.conf
  web/                 # 前端 dist 内容
  logs/
  uploads/
```

```bash
sudo mkdir -p /opt/arlo-admin/{bin,configs,web,logs,uploads}
# 拷贝二进制与配置后：
sudo chown -R arlo:arlo /opt/arlo-admin
```

## 4. 生产配置

编辑 `configs/config.prod.yaml`：

- `database` / `redis` 主机与密码（一般是 `127.0.0.1`）
- `jwt.secret` 换成长随机串
- `server.corsOrigins` 填真实前端域名（同源反代时可填该域名）
- `server.mode: release`，`enableSwagger: false`

## 5. systemd

参考 [`arlo-admin.service`](./arlo-admin.service)：由 unit 设置 `WorkingDirectory`、`APP_ENV=prod` 及配置路径。

```bash
sudo cp deployments/linux/arlo-admin.service /etc/systemd/system/
# 若安装目录不是 /opt/arlo-admin，改 unit 里的路径
sudo systemctl daemon-reload
sudo systemctl enable --now arlo-admin
sudo systemctl status arlo-admin
```

前台排查：

```bash
cd /opt/arlo-admin
sudo -u arlo env APP_ENV=prod ./bin/arlo-admin \
  --config=/opt/arlo-admin/configs/config.yaml \
  --prod-config=/opt/arlo-admin/configs/config.prod.yaml
```

## 6. Nginx

参考 [`nginx-site.conf`](./nginx-site.conf)：静态目录指向 `/opt/arlo-admin/web`，`/api/`（含 WebSocket）与 `/health` 反代到 `127.0.0.1:8090`。

```bash
sudo ln -s /path/to/nginx-site.conf /etc/nginx/sites-enabled/arlo-admin
sudo nginx -t && sudo systemctl reload nginx
```

TLS：用 Let's Encrypt（certbot）或云厂商证书，在 server 块增加 `listen 443 ssl`。

## 7. 发布更新

```bash
# 构建机产出新二进制与 dist 后
sudo systemctl stop arlo-admin
# 覆盖 bin/ 与 web/
sudo systemctl start arlo-admin
```

增量 SQL（`002_*.sql` 起）按 [`migrations/README.md`](../../arlo-admin-server/migrations/README.md) 手工执行。
