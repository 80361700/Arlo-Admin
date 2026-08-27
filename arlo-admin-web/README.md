# arlo-admin-web

Arlo Admin 管理端前端：动态菜单、权限指令、通用表格/表单、主题切换等。

在线体验：[http://101.200.43.49/](http://101.200.43.49/)（`admin` / `admin123`）

仓库总览见根目录 [README.md](../README.md)；架构细节见 [HANDOFF.md](../HANDOFF.md)。

## 技术栈

Vue 3 · TypeScript · Vite · Pinia · Vue Router · Element Plus · Sass

## 目录一览

```
arlo-admin-web/
├── public/              # 静态资源（如默认 logo vite.svg）
├── src/
│   ├── api/             # Axios 封装 + 各模块接口
│   ├── components/      # ProTable / ProFormDialog / FilePicker …
│   ├── composables/
│   ├── directives/      # v-permission
│   ├── layout/          # Sidebar / Navbar
│   ├── router/
│   ├── stores/          # auth / app / message
│   ├── styles/          # 全局样式；主题外观在 styles/themes/
│   ├── themes/          # 主题清单（id / label）
│   ├── utils/
│   └── views/           # 页面（现阶段多在 system/；新业务建议按域分目录）
├── .env.development
├── .env.production
└── vite.config.ts
```

## 运行

```bash
npm install
npm run dev      # http://localhost:5173
npm run build    # 生产构建
npm run preview  # 预览构建结果
```

开发代理：`.env.development` 中 `VITE_API_BASE_URL=/api`，后端地址在 `vite.config.ts` 的 `server.proxy`（默认指向本机 `8090`）。  
**WebSocket（站内信未读）** 也走同一代理：`proxy['/api']` 需设 **`ws: true`**，浏览器连 `ws://localhost:5173/api/v1/ws`，由 Vite 转发到后端；**生产环境**在 Nginx 的 `location /api/` 内加 `Upgrade` 头即可，见 [`deployments/README.md`](../deployments/README.md#websocket站内信未读推送)。

默认登录：`admin` / `admin123`（以后端种子为准）。

## 常用约定

- 菜单 `component` 对应 `src/views/{component}.vue`（不要带 `.vue` 后缀）
- 按钮权限：后端菜单 `type=3` + 角色已分配，前端用 `v-permission`
- 主题：改外观写 `themes/styles/_*.scss`，在 `themes/index.ts` 登记；当前主题在 `stores/app`
- 通用列表：优先复用 `ProTable`（筛选清空后勿残留旧 query，见 HANDOFF）
- 文件字段存 `accessKey`，预览用 `AuthAvatar` / `useAuthFileSrc` 等，勿再拼旧数字 ID 下载路径

## 环境变量

| 变量 | 说明 |
|------|------|
| `VITE_API_BASE_URL` | 请求前缀，本地一般为 `/api` |
| `VITE_APP_TITLE` | 浏览器标题等展示名 |
