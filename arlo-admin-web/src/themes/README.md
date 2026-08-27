# 全局主题

外观写在 CSS；TS 只登记主题 id / 名称，供切换下拉使用。

## 目录

| 路径 | 作用 |
|------|------|
| `src/themes/index.ts` | 主题列表（id、label） |
| `src/styles/themes/_*.scss` | 各主题完整外观 |
| `src/stores/app.ts` | 当前主题、切换、持久化 |

根节点 class 约定：`theme-{id}`（如 `theme-light`）。

## 新增主题

1. 新建 `src/styles/themes/_xxx.scss`（根选择器 `.theme-xxx`）
2. `styles/themes/index.scss` 里 `@use 'xxx'`
3. `themes/index.ts` 的 `themes` 数组加 `{ id: 'xxx', label: '名称' }`
