# 数据库迁移说明

## 基线 v1（当前）

全新安装**只执行一份基线**即可：

```bash
mysql --default-character-set=utf8mb4 -u root -p < 001_baseline_v1.sql
```

- 文件：`001_baseline_v1.sql`
- 内容：当前最终表结构 + 标准种子（含工作流定义/运行时、审批菜单、演示单据等）
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

### 升级到当前审批能力（已跑过旧版 `001`）

原先分散的 `002`～`020` 已合并为**一份**：

```bash
mysql --default-character-set=utf8mb4 -u root -p arlo_admin < 002_flow_approve_runtime.sql
```

- 含：运行时表、审批/监控/演示菜单、`flow_tick`、评论、委托、抄送 `read_at` 等
- `001` 里已有的部门字段 / 流程定义与表单**不再重复**
- 幂等，可重复执行；若库已逐条跑过旧 `009`～`020`，再跑本文件也安全

## 约定

1. 含中文 COMMENT 时务必：`mysql --default-character-set=utf8mb4 …`
2. 菜单 `component` 对应前端 `src/views/{component}.vue`（不要带 `.vue`；可带或不带前导 `/`）
3. 按钮权限为 `type=3`，需写入 `sys_role_menu`，否则前端 `v-permission` 无权限码
4. 新增 type=2 菜单后重启后端或触发 Casbin `ReloadPolicies`
