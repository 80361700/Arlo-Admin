<template>
  <div class="page-container server-monitor" v-loading="loading">
    <div class="toolbar">
      <span class="hint">主机 / 进程 / 依赖健康</span>
      <div class="toolbar-right">
        <el-switch v-model="autoRefresh" active-text="自动刷新" />
        <el-select v-model="refreshSec" style="width: 110px" :disabled="!autoRefresh">
          <el-option :value="5" label="每 5 秒" />
          <el-option :value="10" label="每 10 秒" />
          <el-option :value="30" label="每 30 秒" />
        </el-select>
        <el-button type="primary" @click="loadData">刷新</el-button>
      </div>
    </div>

    <template v-if="info">
      <el-row :gutter="16" class="stat-row" type="flex">
        <el-col :xs="24" :sm="12" :lg="6">
          <div class="card gauge-card">
            <div class="card-title">CPU</div>
            <el-progress type="dashboard" :percentage="roundPct(info.cpu.usagePct)" :color="progressColor" />
            <div class="meta">核心 {{ info.cpu.cores }}　负载 {{ fmtLoad(info.cpu.load1) }} / {{ fmtLoad(info.cpu.load5) }} / {{ fmtLoad(info.cpu.load15) }}</div>
          </div>
        </el-col>
        <el-col :xs="24" :sm="12" :lg="6">
          <div class="card gauge-card">
            <div class="card-title">内存</div>
            <el-progress type="dashboard" :percentage="roundPct(info.mem.usagePct)" :color="progressColor" />
            <div class="meta">已用 {{ formatBytes(info.mem.used) }}　总量 {{ formatBytes(info.mem.total) }}</div>
          </div>
        </el-col>
        <el-col :xs="24" :sm="12" :lg="6">
          <div class="card">
            <div class="card-title">应用</div>
            <el-descriptions :column="1" size="small" class="desc">
              <el-descriptions-item label="名称">{{ info.app.name }}</el-descriptions-item>
              <el-descriptions-item label="模式">{{ info.app.mode || '-' }}</el-descriptions-item>
              <el-descriptions-item label="启动">{{ info.app.startTime || '-' }}</el-descriptions-item>
              <el-descriptions-item label="运行">{{ formatDuration(info.app.runSeconds) }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-col>
        <el-col :xs="24" :sm="12" :lg="6">
          <div class="card">
            <div class="card-title">系统</div>
            <el-descriptions :column="1" size="small" class="desc">
              <el-descriptions-item label="主机">{{ info.sys.hostname || '-' }}</el-descriptions-item>
              <el-descriptions-item label="系统">{{ info.sys.os }}/{{ info.sys.arch }}</el-descriptions-item>
              <el-descriptions-item label="开机">{{ formatDuration(info.sys.uptime) }}</el-descriptions-item>
              <el-descriptions-item label="Go">{{ info.go.version }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :xs="24" :lg="12">
          <div class="card">
            <div class="card-title">Go 运行时</div>
            <el-descriptions :column="2" size="small" class="desc">
              <el-descriptions-item label="Goroutine">{{ info.go.goroutines }}</el-descriptions-item>
              <el-descriptions-item label="GOMAXPROCS">{{ info.go.gomaxprocs }}</el-descriptions-item>
              <el-descriptions-item label="Heap Alloc">{{ formatBytes(info.go.heapAlloc) }}</el-descriptions-item>
              <el-descriptions-item label="Heap Sys">{{ formatBytes(info.go.heapSys) }}</el-descriptions-item>
              <el-descriptions-item label="GC 次数">{{ info.go.numGC }}</el-descriptions-item>
              <el-descriptions-item label="最近 GC">{{ info.go.lastGC || '-' }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-col>
        <el-col :xs="24" :lg="12">
          <div class="card">
            <div class="card-title">依赖健康</div>
            <div class="deps">
              <div class="dep">
                <el-tag :type="statusType(info.db.status)" size="small">MySQL {{ info.db.status }}</el-tag>
                <span>Ping {{ formatPing(info.db.pingMs) }}</span>
                <span>连接 {{ info.db.inUse }}/{{ info.db.open }}（空闲 {{ info.db.idle }}，最大 {{ info.db.maxOpen }}）</span>
              </div>
              <div class="dep">
                <el-tag :type="statusType(info.redis.status)" size="small">Redis {{ info.redis.status }}</el-tag>
                <span>Ping {{ formatPing(info.redis.pingMs) }}</span>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>

      <div class="card">
        <div class="card-title">磁盘</div>
        <el-table :data="info.disk || []" border stripe empty-text="暂无磁盘数据">
          <el-table-column prop="mount" label="挂载点" min-width="120" />
          <el-table-column label="总量" width="120">
            <template #default="{ row }">{{ formatBytes(row.total) }}</template>
          </el-table-column>
          <el-table-column label="已用" width="120">
            <template #default="{ row }">{{ formatBytes(row.used) }}</template>
          </el-table-column>
          <el-table-column label="可用" width="120">
            <template #default="{ row }">{{ formatBytes(row.free) }}</template>
          </el-table-column>
          <el-table-column label="使用率" min-width="180">
            <template #default="{ row }">
              <el-progress :percentage="roundPct(row.usagePct)" :color="progressColor" />
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { getServerMonitor, type ServerMonitorInfo } from '@/api/modules/monitor'

const loading = ref(false)
const info = ref<ServerMonitorInfo | null>(null)
const autoRefresh = ref(true)
const refreshSec = ref(10)
let timer: ReturnType<typeof setInterval> | null = null

function roundPct(v: number) {
  if (!Number.isFinite(v)) return 0
  return Math.min(100, Math.max(0, Math.round(v * 10) / 10))
}
function fmtLoad(v: number) {
  return Number.isFinite(v) ? v.toFixed(2) : '-'
}
function formatPing(ms: number) {
  if (!Number.isFinite(ms) || ms < 0) return '-'
  if (ms < 1) return `${ms.toFixed(2)} ms`
  if (ms < 10) return `${ms.toFixed(1)} ms`
  return `${Math.round(ms)} ms`
}
function formatBytes(n: number) {
  if (!n || n <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let val = n
  while (val >= 1024 && i < units.length - 1) {
    val /= 1024
    i++
  }
  return `${val.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}
function formatDuration(sec: number) {
  if (!sec || sec <= 0) return '-'
  const d = Math.floor(sec / 86400)
  const h = Math.floor((sec % 86400) / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = Math.floor(sec % 60)
  if (d > 0) return `${d}天${h}小时`
  if (h > 0) return `${h}小时${m}分`
  if (m > 0) return `${m}分${s}秒`
  return `${s}秒`
}
function progressColor(pct: number) {
  if (pct >= 90) return '#f56c6c'
  if (pct >= 70) return '#e6a23c'
  return '#67c23a'
}
function statusType(status: string) {
  if (status === 'up') return 'success'
  if (status === 'down') return 'danger'
  return 'info'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getServerMonitor()
    info.value = res.data
  } finally {
    loading.value = false
  }
}
function clearTimer() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}
function setupTimer() {
  clearTimer()
  if (!autoRefresh.value) return
  timer = setInterval(loadData, refreshSec.value * 1000)
}

watch([autoRefresh, refreshSec], setupTimer)
onMounted(() => {
  loadData()
  setupTimer()
})
onUnmounted(clearTimer)
</script>

<style scoped>
.server-monitor {
  padding: 16px;
}
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}
.hint {
  color: #909399;
  font-size: 13px;
}
.stat-row {
  display: flex;
  flex-wrap: wrap;
}
.stat-row :deep(.el-col) {
  display: flex;
}
.card {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px 18px;
  margin-bottom: 16px;
  box-sizing: border-box;
  width: 100%;
  min-height: 240px;
  display: flex;
  flex-direction: column;
}
.gauge-card {
  align-items: center;
}
.gauge-card .el-progress {
  margin: auto 0;
}
.card-title {
  width: 100%;
  font-weight: 600;
  margin-bottom: 12px;
  color: #303133;
  flex-shrink: 0;
}
.meta {
  margin-top: auto;
  padding-top: 8px;
  font-size: 13px;
  color: #606266;
  text-align: center;
  line-height: 1.6;
  flex-shrink: 0;
}
.desc {
  flex: 1;
}
.desc :deep(.el-descriptions__label) {
  width: 96px;
  color: #909399;
}
.desc :deep(.el-descriptions__content) {
  color: #606266;
}
.deps {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-top: 4px;
}
.dep {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 16px;
  align-items: center;
  font-size: 13px;
  color: #606266;
}
</style>
