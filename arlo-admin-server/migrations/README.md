# 数据库迁移说明

## 基线 v1（当前）

全新安装**只执行一份基线**即可：

```bash
mysql --default-character-set=utf8mb4 -u root -p < 001_baseline_v1.sql
```

- 文件：`001_baseline_v1.sql`
- 内容：当前最终表结构 + 标准种子（admin / 菜单 / 字典 / 配置 / 定时任务等）
- 默认库名：`arlo_admin`
- 默认账号：`admin` / `admin123`（以种子 bcrypt 为准）

## 增量补丁（v1 之后）

后续变更使用：

```
002_简短英文描述.sql
003_...
```

- 三位序号全局递增，**禁止复用**
- 已有库只跑尚未执行的增量
- 补丁尽量幂等

当前最新增量序号：无（基线即为 v1 起点；下一份从 **002** 开始）。

## 历史迭代

`001`～`026` 时代的零散补丁已归档到：

```
archive/pre_v1/
```

仅供对照，**新环境不要再按旧顺序执行**。

## 约定

1. 含中文 COMMENT 时务必：`mysql --default-character-set=utf8mb4 …`
2. 菜单 `component` 对应前端 `src/views/{component}.vue`（不要带 `.vue`；可带或不带前导 `/`）
3. 按钮权限为 `type=3`，需写入 `sys_role_menu`，否则前端 `v-permission` 无权限码
4. 新增 type=2 菜单后重启后端或触发 Casbin `ReloadPolicies`
