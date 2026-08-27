<template>
  <div class="page-container">
    <ProTable
      ref="proTableRef"
      :data="tableData"
      :loading="loading"
      :total="total"
      :search-fields="searchFields"
      :show-index="false"
      @search="handleSearch"
      @reset="handleReset"
      @page-change="handlePageChange"
    >
      <template #toolbar>
        <el-button v-permission="'log:operation:export'" @click="handleExport">导出</el-button>
      </template>
      <el-table-column prop="username" label="操作人" width="100" />
      <el-table-column prop="module" label="操作模块" width="120" />
      <el-table-column prop="action" label="操作类型" width="120" />
      <el-table-column label="方法" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="methodColor(row.method)" size="small">{{ row.method }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="url" label="请求URL" min-width="180" show-overflow-tooltip />
      <el-table-column prop="ip" label="IP" width="140" />
      <el-table-column prop="costTime" label="耗时(ms)" width="90" align="center" />
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="时间" width="160" />
      <el-table-column label="操作" width="80" align="center" fixed="right">
        <template #default="{ row }">
          <el-button v-permission="'log:operation:view'" type="primary" link size="small" @click="handleDetail(row as OperationLogItem)">
            详情
          </el-button>
        </template>
      </el-table-column>
    </ProTable>

    <el-dialog v-model="detailVisible" title="操作日志详情" width="700px" align-center destroy-on-close>
      <el-descriptions :column="2" border v-if="currentRow">
        <el-descriptions-item label="操作人">{{ currentRow.username || '-' }}</el-descriptions-item>
        <el-descriptions-item label="操作模块">{{ currentRow.module || '-' }}</el-descriptions-item>
        <el-descriptions-item label="操作类型">{{ currentRow.action || '-' }}</el-descriptions-item>
        <el-descriptions-item label="请求方法">
          <el-tag :type="methodColor(currentRow.method)" size="small">{{ currentRow.method }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="请求URL" :span="2">{{ currentRow.url || '-' }}</el-descriptions-item>
        <el-descriptions-item label="请求IP">{{ currentRow.ip || '-' }}</el-descriptions-item>
        <el-descriptions-item label="耗时">{{ currentRow.costTime }} ms</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentRow.status === 1 ? 'success' : 'danger'" size="small">
            {{ currentRow.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="操作时间">{{ currentRow.createdAt || '-' }}</el-descriptions-item>
        <el-descriptions-item label="用户代理" :span="2">
          <div class="break-all">{{ currentRow.userAgent || '-' }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="错误信息" :span="2" v-if="currentRow.status === 0 && currentRow.errorMsg">
          <div class="text-danger">{{ currentRow.errorMsg }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="请求数据" :span="2">
          <div class="pre-wrap">{{ formatParams(currentRow.params) || '-' }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="响应摘要" :span="2" v-if="currentRow.result">
          <div class="pre-wrap">{{ formatParams(currentRow.result) || currentRow.result }}</div>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import { getOperationLogList, exportOperationLogs, type OperationLogItem } from '@/api/modules/log'
import { showRequestError } from '@/utils/requestError'

const proTableRef = ref()
const tableData = ref<OperationLogItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive({
  page: 1,
  pageSize: 10,
  username: undefined as string | undefined,
  module: undefined as string | undefined,
  url: undefined as string | undefined,
  startTime: undefined as string | undefined,
  endTime: undefined as string | undefined,
  status: undefined as number | undefined,
})
const detailVisible = ref(false)
const currentRow = ref<OperationLogItem | null>(null)

const searchFields = [
  { prop: 'username', label: '操作人' },
  { prop: 'module', label: '模块' },
  { prop: 'url', label: '请求地址', placeholder: '如 /api/v1/system/user' },
  { prop: 'timeRange', label: '操作时间', type: 'datetimerange' as const },
  { prop: 'status', label: '状态', type: 'select' as const, options: [{ label: '成功', value: 1 }, { label: '失败', value: 0 }] },
]

function methodColor(method: string): 'success' | 'primary' | 'warning' | 'danger' | 'info' {
  const map: Record<string, 'success' | 'primary' | 'warning' | 'danger'> = {
    GET: 'success', POST: 'primary', PUT: 'warning', DELETE: 'danger',
  }
  return map[method] || 'info'
}

function formatParams(params: string): string {
  if (!params) return ''
  try {
    return JSON.stringify(JSON.parse(params), null, 2)
  } catch {
    return params
  }
}

function applySearch(p: Record<string, any> = {}) {
  const range = Array.isArray(p.timeRange) ? p.timeRange : []
  query.username = p.username || undefined
  query.module = p.module || undefined
  query.url = p.url || undefined
  query.startTime = range[0] || undefined
  query.endTime = range[1] || undefined
  query.status = p.status !== undefined && p.status !== null && p.status !== '' ? p.status : undefined
}

async function loadData() {
  loading.value = true
  try {
    const res = await getOperationLogList({
      username: query.username,
      module: query.module,
      url: query.url,
      startTime: query.startTime,
      endTime: query.endTime,
      status: query.status,
      page: query.page,
      pageSize: query.pageSize,
    })
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) {
  query.page = 1
  applySearch(p)
  loadData()
}

function handleReset() {
  query.page = 1
  applySearch({})
  loadData()
}

function handlePageChange(p: any) {
  query.page = p.page
  query.pageSize = p.pageSize
  applySearch(p)
  loadData()
}

function handleDetail(row: OperationLogItem) {
  currentRow.value = row
  detailVisible.value = true
}

async function handleExport() {
  try {
    await exportOperationLogs({
      username: query.username,
      module: query.module,
      url: query.url,
      startTime: query.startTime,
      endTime: query.endTime,
      status: query.status,
    })
    ElMessage.success('导出成功')
  } catch (err: any) {
    showRequestError(err, '导出失败')
  }
}

onMounted(() => loadData())
</script>

<style scoped>
.page-container {
  padding: 16px;
}

.break-all {
  word-break: break-all;
  line-height: 1.5;
}

.pre-wrap {
  white-space: pre-wrap;
  word-break: break-all;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  background: #f5f7fa;
  padding: 8px 12px;
  border-radius: 4px;
  max-height: 300px;
  overflow-y: auto;
}

.text-danger {
  color: #f56c6c;
  word-break: break-all;
  line-height: 1.5;
}
</style>
