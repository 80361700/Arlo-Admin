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
        <el-button v-permission="'log:login:export'" @click="handleExport">导出</el-button>
      </template>
      <el-table-column prop="username" label="用户名" width="100" />
      <el-table-column prop="ip" label="IP" width="140" />
      <el-table-column prop="browser" label="浏览器" min-width="120" show-overflow-tooltip />
      <el-table-column prop="os" label="操作系统" min-width="120" show-overflow-tooltip />
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="msg" label="消息" min-width="150" show-overflow-tooltip />
      <el-table-column prop="createdAt" label="时间" width="160" />
    </ProTable>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import ProTable from '@/components/ProTable.vue'
import { getLoginLogList, exportLoginLogs, type LoginLogItem } from '@/api/modules/log'

const proTableRef = ref()
const tableData = ref<LoginLogItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive({
  page: 1,
  pageSize: 10,
  username: undefined as string | undefined,
  status: undefined as number | undefined,
  startTime: undefined as string | undefined,
  endTime: undefined as string | undefined,
})

const searchFields = [
  { prop: 'username', label: '用户名' },
  { prop: 'status', label: '状态', type: 'select' as const, options: [{ label: '成功', value: 1 }, { label: '失败', value: 0 }] },
  { prop: 'timeRange', label: '登录时间', type: 'datetimerange' as const },
]

function applySearch(p: any) {
  query.username = p.username || undefined
  query.status = p.status !== undefined && p.status !== '' ? p.status : undefined
  if (Array.isArray(p.timeRange) && p.timeRange.length === 2) {
    query.startTime = p.timeRange[0]
    query.endTime = p.timeRange[1]
  } else {
    query.startTime = undefined
    query.endTime = undefined
  }
}

async function loadData() {
  loading.value = true
  try {
    const res = await getLoginLogList({
      username: query.username,
      status: query.status,
      startTime: query.startTime,
      endTime: query.endTime,
      page: query.page,
      pageSize: query.pageSize,
    })
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) { query.page = 1; applySearch(p); loadData() }
function handleReset() {
  Object.assign(query, { page: 1, pageSize: 10, username: undefined, status: undefined, startTime: undefined, endTime: undefined })
  loadData()
}
function handlePageChange(p: any) { query.page = p.page; query.pageSize = p.pageSize; loadData() }

async function handleExport() {
  try {
    await exportLoginLogs({
      username: query.username,
      status: query.status,
      startTime: query.startTime,
      endTime: query.endTime,
    })
    ElMessage.success('导出成功')
  } catch (err: any) {
    showRequestError(err, '导出失败')
  }
}

onMounted(() => loadData())
</script>
