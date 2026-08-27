<template>
  <div class="page-container welcome-page">
    <div class="welcome-bg" aria-hidden="true">
      <div class="bg-base" />

      <div class="bg-code">
        <div
          v-for="(col, i) in codeColumns"
          :key="i"
          class="code-col"
        >
          <div
            class="code-track"
            :style="{ animationDuration: `${6 + i * 5}s`, animationDelay: `${-i * 6}s` }"
          >
            <pre>{{ col }}{{ '\n' }}{{ col }}</pre>
          </div>
        </div>
      </div>

      <div class="code-stack">
        <div
          v-for="(win, i) in windows"
          :key="win.title"
          class="code-window"
          :class="`code-window--${i}`"
        >
          <div class="code-window-bar">
            <span class="dot dot-r" />
            <span class="dot dot-y" />
            <span class="dot dot-g" />
            <span class="code-window-title">{{ win.title }}</span>
          </div>
          <div class="code-window-body">
            <pre>{{ win.code }}</pre>
          </div>
        </div>
      </div>
    </div>

    <div class="welcome-block">
      <p class="meta">{{ todayText }}</p>

      <h1 class="title">
        <span class="title-greet">{{ dayPart }}，</span>{{ nickname }}
      </h1>

      <p class="lead">
        欢迎回到 {{ appStore.systemName }}。通过左侧菜单进入各业务模块，处理消息、账号与系统配置。
      </p>

      <blockquote class="story">
        <p class="story-label">科普知识</p>
        <p class="story-body">{{ dailyStory }}</p>
      </blockquote>

      <p class="version">{{ appStore.systemName }} · v{{ appStore.systemVersion }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const authStore = useAuthStore()
const appStore = useAppStore()

const nickname = computed(
  () => authStore.userInfo?.nickname || authStore.userInfo?.username || '用户',
)

const dayPart = computed(() => {
  const h = new Date().getHours()
  if (h < 5) return '夜深了'
  if (h < 11) return '早上好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const todayText = computed(() => {
  const d = new Date()
  const week = ['日', '一', '二', '三', '四', '五', '六'][d.getDay()]
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日 星期${week}`
})

const stories = [
  'HTTP 状态码 418 是「I\'m a teapot」：服务器拒绝泡咖啡，因为它是茶壶。这是 1998 年愚人节 RFC 留下的彩蛋。',
  'Git 的默认分支曾叫 master。现在许多仓库改用 main，只是命名约定变化，功能完全一样。',
  'JWT 由三部分组成：Header、Payload、Signature，中间用点分隔。Payload 默认只是 Base64，不是加密。',
  'Unix 时间戳从 1970-01-01 00:00:00 UTC 起算秒数。2038 年问题来自 32 位有符号整数会溢出。',
  'CSS 的 z-index 只在定位元素（relative / absolute / fixed / sticky）上生效，且受层叠上下文限制。',
  'TCP 三次握手：客户端 SYN → 服务端 SYN+ACK → 客户端 ACK。目的是双方确认收发能力并协商序号。',
  'REST 里 PUT 通常表示完整替换资源，PATCH 表示部分更新。很多项目实际混用，需要看接口约定。',
  'SQL 的 NULL 表示「未知」。任何与 NULL 的比较结果都不是 true，要用 IS NULL / IS NOT NULL。',
  '浏览器同源策略：协议、域名、端口三者相同才算同源。跨域请求常靠 CORS 显式放行。',
  'B 树适合磁盘索引：扇出大、高度低，一次查找只需少量页读取，所以数据库爱用它。',
  '幂等：同一操作执行一次和多次，效果相同。GET、PUT、DELETE 常被设计为幂等，POST 通常不是。',
  'UUID v4 几乎全是随机数。碰撞概率极低，但不是绝对零；高并发仍建议结合业务唯一约束。',
  'Gzip / Brotli 压缩的是传输体积，不会减轻服务端算力。CPU 忙时，压缩本身也可能成为瓶颈。',
  '哈希不等于加密。哈希难逆向，用于校验与摘要；加密可解密，用于保密传输与存储。',
  'DNS 把域名解析成 IP。本地有缓存，CDN / 负载均衡常靠短 TTL 或智能解析把流量导向不同节点。',
  '事务 ACID：原子性、一致性、隔离性、持久性。隔离级别越高，并发越保守，脏读幻读越少。',
  'CSP（内容安全策略）用白名单限制脚本与资源来源，是缓解 XSS 的重要浏览器侧防线。',
  '浮点数用二进制近似十进制，0.1 + 0.2 往往不等于 0.3。金额计算应优先用整数分或专用十进制类型。',
  'OAuth 2.0 解决「授权」：让第三方在不拿到密码的情况下，用令牌访问受保护资源。',
  'WebSocket 在一次 HTTP 握手后升级为全双工长连接，适合推送、聊天、协同等实时场景。',
  '索引能加速查询，但会增加写入成本与存储。并不是列越多越好，要匹配真实查询模式。',
  'HTTPS = HTTP + TLS。证书证明服务器身份，握手后用对称密钥加密后续流量。',
  '消息队列削峰填谷：高峰先入队，消费者按能力消化。代价是最终一致性与运维复杂度上升。',
  'CDN 把静态资源缓存在离用户更近的节点，减少源站压力与首屏延迟，尤其利于图片与前端资源。',
  '正则回溯过多时可能极慢，甚至被恶意输入拖垮（ReDoS）。复杂匹配要警惕灾难性回溯。',
  '容器共享宿主机内核，比虚拟机更轻；虚拟机有独立客户机内核，隔离更强但更重。',
  'Bloom Filter 能快速判断「一定不存在」或「可能存在」，省空间，但有误判，不能当精确集合。',
  '分页深翻页很贵：OFFSET 越大，数据库通常要跳过越多行。游标 / seek 分页更适合海量数据。',
  'SameSite Cookie 可限制跨站携带，是缓解 CSRF 的重要手段；仍需配合服务端校验等措施。',
  '连接池复用数据库连接，避免每次请求都建连。池过小会排队，过大则浪费连接与内存。',
  '语义化版本 MAJOR.MINOR.PATCH：不兼容改主版本，兼容新功能改次版本，兼容修复改正版本。',
]

const dailyStory = stories[Math.floor(Math.random() * stories.length)]

const snippets = [
  'const token = await auth.refresh()',
  'if (!user?.roles.length) return',
  'router.push("/dashboard")',
  'SELECT id, name FROM sys_user',
  'casbin.enforce(sub, obj, act)',
  'redis.set(key, jti, ttl)',
  'fn handleLogin(c *gin.Context)',
  'export function useAuthStore()',
  'jwt.ParseWithClaims(token)',
  'permission.includes("sys:user")',
  'await db.Where("status = 1")',
  'return response.OK(c, data)',
  'watchEffect(() => loadList())',
  'go func() { worker.Run() }()',
  'type UserInfo struct {',
  '  ID uint64 `json:"id"`',
  '}',
  'npm run build && go build',
  'middleware.JWTAuth()',
  'v-permission="user:add"',
  '0x7f 0x3a 0x91 0x0c',
  '>>> system.ready = true',
  'ssh arlo@prod-01',
  'curl -H "Authorization: Bearer"',
  'encrypt.password(bcrypt)',
  'queue.push(job.log_cleanup)',
  'span.SetAttribute("user.id")',
  'catch (err) { showError(err) }',
  'COMMIT; -- migration 025',
  'echo $ACCESS_TOKEN | pbcopy',
]

function buildColumn(seed: number): string {
  const lines: string[] = []
  for (let i = 0; i < 28; i++) {
    lines.push(snippets[(seed * 7 + i * 3) % snippets.length])
  }
  return lines.join('\n')
}

const codeColumns = Array.from({ length: 5 }, (_, i) => buildColumn(i + 1))

const windows = [
  {
    title: 'router.go — go-admin',
    code: [
      'api.POST("/login", auth.Login)',
      'api.GET("/info", auth.Info)',
      'sys.GET("/user/list", user.List)',
      'sys.POST("/role/assignMenus", role.Assign)',
      'msg.GET("/unread-count", message.Unread)',
    ].join('\n'),
  },
  {
    title: 'auth.service.go — go-admin',
    code: [
      'func (s *AuthService) Login(req) {',
      '  user, err := s.repo.FindByName(req.User)',
      '  if err != nil { return err }',
      '  if !bcrypt.Compare(user.Pwd, req.Pwd) {',
      '    return errors.Unauthorized',
      '  }',
      '  return s.issueTokens(user)',
      '}',
    ].join('\n'),
  },
  {
    title: 'message.ts — arlo-admin-web',
    code: [
      'export const useMessageStore = defineStore(',
      "  'message',",
      '  () => {',
      '    const unreadCount = ref(0)',
      '    async function fetchUnreadCount() {',
      '      const res = await getUnreadCount()',
      '      unreadCount.value = res.data.count',
      '    }',
      '    return { unreadCount, fetchUnreadCount }',
      '  },',
      ')',
    ].join('\n'),
  },
  {
    title: 'arlo-admin — zsh',
    code: [
      '$ go run cmd/server/main.go',
      '',
      '➜  listening on :8090',
      '➜  casbin policies loaded',
      '➜  redis connected',
      '',
      '$ curl -s localhost:8090/health',
      '{ "status": "ok" }',
      '',
      '>>> permission.check("sys:user")',
      'true',
    ].join('\n'),
  },
]
</script>

<style scoped lang="scss">
.welcome-page {
  position: relative;
  isolation: isolate;
  overflow: hidden;
  /* 铺满 layout-main（含抵消 padding 的负边距），勿用 100vh，否则开 tags 会多出滚动条 */
  min-height: calc(100% + 32px);
  height: calc(100% + 32px);
  margin: -16px;
  padding: 0 !important;
  border-radius: 0 !important;
  background: #fff !important;
}

.welcome-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
}

.bg-base {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 55% 50% at 88% 88%, rgba(48, 65, 86, 0.06), transparent 70%),
    linear-gradient(180deg, #ffffff 0%, #f6f8fa 100%);
}

.bg-code {
  position: absolute;
  left: 40px;
  right: 0;
  bottom: 0;
  height: 50%;
  z-index: 0;
  display: flex;
  align-items: stretch;
  justify-content: flex-start;
  gap: 36px;
  padding: 0 30px 0 0;
  overflow: hidden;
  color: #304156;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  line-height: 1.7;
  mask-image: linear-gradient(
    to top,
    rgba(0, 0, 0, 0.85) 0%,
    rgba(0, 0, 0, 0.5) 50%,
    rgba(0, 0, 0, 0.16) 80%,
    transparent 100%
  ),
  linear-gradient(
    to right,
    rgba(0, 0, 0, 0.85) 0%,
    rgba(0, 0, 0, 0.55) 40%,
    rgba(0, 0, 0, 0.22) 75%,
    transparent 100%
  );
  -webkit-mask-image: linear-gradient(
    to top,
    rgba(0, 0, 0, 0.85) 0%,
    rgba(0, 0, 0, 0.5) 50%,
    rgba(0, 0, 0, 0.16) 80%,
    transparent 100%
  ),
  linear-gradient(
    to right,
    rgba(0, 0, 0, 0.85) 0%,
    rgba(0, 0, 0, 0.55) 40%,
    rgba(0, 0, 0, 0.22) 75%,
    transparent 100%
  );
  mask-composite: intersect;
  -webkit-mask-composite: source-in;
}

.code-col {
  flex: 0 0 auto;
  width: 240px;
  height: 100%;
  overflow: hidden;
  opacity: 0.42;

  &:nth-child(2n) { opacity: 0.34; }
  &:nth-child(3n) { opacity: 0.48; }
}

.code-track {
  animation: code-rise linear infinite;

  pre {
    margin: 0;
    padding: 0;
    border: 0;
    font: inherit;
    color: inherit;
    white-space: pre;
    line-height: inherit;
  }
}

@keyframes code-rise {
  from {
    transform: translateY(0);
  }
  to {
    transform: translateY(-50%);
  }
}

@media (prefers-reduced-motion: reduce) {
  .code-track {
    animation: none;
  }
}

.code-stack {
  position: absolute;
  right: -10px;
  bottom: -10px;
  z-index: 1;
  width: min(70vw, 640px);
  height: min(86%, 700px);
  mask-image: radial-gradient(
    ellipse 130% 120% at 100% 100%,
    #000 0%,
    #000 42%,
    rgba(0, 0, 0, 0.85) 58%,
    rgba(0, 0, 0, 0.55) 72%,
    rgba(0, 0, 0, 0.28) 84%,
    rgba(0, 0, 0, 0.1) 92%,
    transparent 100%
  );
  -webkit-mask-image: radial-gradient(
    ellipse 130% 120% at 100% 100%,
    #000 0%,
    #000 42%,
    rgba(0, 0, 0, 0.85) 58%,
    rgba(0, 0, 0, 0.55) 72%,
    rgba(0, 0, 0, 0.28) 84%,
    rgba(0, 0, 0, 0.1) 92%,
    transparent 100%
  );
}

.code-window {
  position: absolute;
  display: flex;
  flex-direction: column;
  border-radius: 11px;
  border: 1px solid rgba(48, 65, 86, 0.1);
  background: rgba(255, 255, 255, 0.86);
  box-shadow:
    0 14px 32px rgba(31, 42, 55, 0.07),
    0 2px 8px rgba(31, 42, 55, 0.04);
  overflow: hidden;
}

.code-window--0 {
  right: 136px;
  top: 2%;
  width: 58%;
  height: 34%;
  z-index: 1;
  opacity: 0.48;
  transform: rotate(-3.8deg);
  background: rgba(255, 255, 255, 0.7);
}

.code-window--1 {
  right: 70px;
  top: 18%;
  width: 68%;
  height: 38%;
  z-index: 2;
  opacity: 0.62;
  transform: rotate(2.6deg);
  background: rgba(255, 255, 255, 0.78);
}

.code-window--2 {
  right: 28px;
  top: 40%;
  width: 74%;
  height: 40%;
  z-index: 3;
  opacity: 0.8;
  transform: rotate(-1.6deg);
  background: rgba(255, 255, 255, 0.86);
}

.code-window--3 {
  right: 0;
  bottom: 0;
  width: 82%;
  height: 44%;
  z-index: 4;
  opacity: 0.97;
}

.code-window-bar {
  display: flex;
  align-items: center;
  gap: 7px;
  height: 34px;
  padding: 0 12px;
  background: rgba(48, 65, 86, 0.04);
  border-bottom: 1px solid rgba(48, 65, 86, 0.08);
  flex-shrink: 0;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  opacity: 0.55;
}

.dot-r { background: #e5a3a3; }
.dot-y { background: #e2c48a; }
.dot-g { background: #9dcca8; }

.code-window-title {
  margin-left: 8px;
  font-size: 11px;
  color: #909399;
  letter-spacing: 0.02em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.code-window-body {
  flex: 1;
  overflow: hidden;
  padding: 12px 14px 16px;
  color: #304156;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11.5px;
  line-height: 1.65;
  opacity: 0.78;

  pre {
    margin: 0;
    font: inherit;
    color: inherit;
    white-space: pre;
  }
}

.welcome-block {
  position: relative;
  z-index: 1;
  padding: 36px 40px 48px;
  max-width: 640px;
}

.meta {
  margin: 0 0 10px;
  font-size: 13px;
  line-height: 1.4;
  color: #909399;
}

.title {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  line-height: 1.35;
  color: #303133;
}

.title-greet {
  font-weight: 500;
  color: #606266;
}

.lead {
  margin: 14px 0 0;
  font-size: 15px;
  line-height: 1.7;
  color: #606266;
}

.story {
  margin: 24px 0 0;
  padding: 4px 0 4px 16px;
  max-width: 420px;
  border: 0;
  border-left: 3px solid #dcdfe6;
  background: transparent;
}

.story-label {
  margin: 0 0 6px;
  font-size: 12px;
  font-weight: 400;
  line-height: 1.4;
  color: #c0c4cc;
}

.story-body {
  margin: 0;
  font-size: 14px;
  line-height: 1.85;
  color: #909399;
  font-style: italic;
}

.version {
  margin: 28px 0 0;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
  font-size: 12px;
  color: #c0c4cc;
}

@media (max-width: 900px) {
  .code-stack {
    width: min(84vw, 480px);
    right: 20px;
    bottom: 20px;
    height: min(70%, 480px);
  }
}
</style>
