#!/bin/bash
# 宝塔 / Supervisor 启动脚本：先进入项目根，再启动（避免相对路径找不到配置）
set -euo pipefail
ROOT="${ARLO_ADMIN_ROOT:-/www/wwwroot/arlo-admin}"
cd "$ROOT"
mkdir -p "$ROOT/logs" "$ROOT/uploads"
exec /usr/bin/env APP_ENV=prod "$ROOT/bin/arlo-admin" \
  --config="$ROOT/configs/config.yaml" \
  --prod-config="$ROOT/configs/config.prod.yaml"
