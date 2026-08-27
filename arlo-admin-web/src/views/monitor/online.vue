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
      <el-table-column prop="username" label="用户名" width="120" />
      <el-table-column prop="userId" label="用户ID" width="90" align="center" />
      <el-table-column prop="ip" label="登录IP" width="140" />
      <el-table-column prop="browser" label="浏览器" min-width="120" show-overflow-tooltip />
      <el-table-column prop="os" label="操作系统" min-width="120" show-overflow-tooltip />
      <el-table-column prop="loginAt" label="登录时间" width="170" />
      <el-table-column label="操作" width="100" align="center" fixed="right">
        <template #default="{ row }">
          <el-button
            v-permission="'monitor:online:kick'"
            type="danger"
            link
            size="small"
            @click="handleKick(row as OnlineSessionItem)"
          >
            强退
          </el-button>
        </template>
      </el-table-column>
    </ProTable>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import { getOnlineList, kickOnlineUser, type OnlineSessionItem } from '@/api/modules/monitor'

const proTableRef = ref()
const tableData = ref<OnlineSessionItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, search: {} as Record<string, any> })

const searchFields = [
  { prop: 'username', label: '用户名' },
]

async function loadData(params: { page: number; pageSize: number; search?: Record<string, any> }) {
  loading.value = true
  try {
    const res = await getOnlineList({
      username: params.search?.username || undefined,
      page: params.page,
      pageSize: params.pageSize,
    })
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) {
  query.page = 1
  query.search = p
  loadData({ ...query, search: p })
}
function handleReset() {
  query.search = {}
  loadData({ page: 1, pageSize: 10 })
}
function handlePageChange(p: any) {
  query.page = p.page
  query.pageSize = p.pageSize
  loadData({ ...query })
}

async function handleKick(row: OnlineSessionItem) {
  await ElMessageBox.confirm(`确认强制下线用户「${row.username}」的该会话？`, '提示', { type: 'warning' })
  await kickOnlineUser({ userId: row.userId, sessionId: row.sessionId })
  ElMessage.success('已强制下线')
  loadData(query)
}

onMounted(() => loadData(query))
</script>
