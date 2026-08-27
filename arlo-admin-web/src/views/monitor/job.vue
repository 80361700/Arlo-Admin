<template>
  <div class="page-container">
    <ProTable
      ref="proTableRef"
      :data="tableData"
      :loading="loading"
      :total="total"
      :search-fields="searchFields"
      :show-index="false"
      :action-width="200"
      @search="handleSearch"
      @reset="handleReset"
      @page-change="handlePageChange"
    >
      <template #toolbar>
        <el-button v-permission="'monitor:job:edit'" type="primary" @click="openCreate">新增任务</el-button>
      </template>

      <el-table-column prop="name" label="任务名称" min-width="140" show-overflow-tooltip />
      <el-table-column prop="handler" label="处理器" width="140" show-overflow-tooltip />
      <el-table-column prop="cron" label="Cron" width="130" />
      <el-table-column label="状态" width="90" align="center">
        <template #default="{ row }">
          <el-switch
            v-if="authStore.hasPermission('monitor:job:status')"
            :model-value="row.status === 1"
            :disabled="statusSavingId === row.id"
            @change="(v: boolean | string | number) => handleStatusChange(row as JobItem, Boolean(v))"
          />
          <el-tag v-else :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '启用' : '暂停' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="上次执行" width="170">
        <template #default="{ row }">
          <div>{{ row.lastRunAt || '-' }}</div>
          <el-tag
            v-if="row.lastStatus !== null && row.lastStatus !== undefined"
            :type="row.lastStatus === 1 ? 'success' : 'danger'"
            size="small"
          >
            {{ row.lastStatus === 1 ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="nextRunAt" label="下次执行" width="170">
        <template #default="{ row }">{{ row.nextRunAt || '-' }}</template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />

      <template #actions="{ row }">
        <el-button v-permission="'monitor:job:run'" type="primary" link size="small" @click="handleRun(row)">
          执行
        </el-button>
        <el-button v-permission="'monitor:job:log'" type="primary" link size="small" @click="openLogs(row)">
          日志
        </el-button>
        <el-dropdown
          v-if="authStore.hasPermission('monitor:job:edit')"
          trigger="click"
          @command="(cmd: string) => handleAction(row, cmd)"
        >
          <el-button type="info" link size="small">
            更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="edit">编辑</el-dropdown-item>
              <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
    </ProTable>

    <ProFormDialog
      v-model="formVisible"
      :title="formMode === 'create' ? '新增任务' : '编辑任务'"
      :model="form"
      :rules="formRules"
      width="560px"
      :submitting="formSubmitting"
      @submit="handleFormSubmit"
    >
      <el-form-item label="任务名称" prop="name">
        <el-input v-model="form.name" maxlength="64" placeholder="请输入任务名称" />
      </el-form-item>
      <el-form-item v-if="formMode === 'create'" label="处理器" prop="handler">
        <el-select v-model="form.handler" placeholder="请选择处理器" style="width: 100%">
          <el-option
            v-for="h in handlers"
            :key="h.code"
            :label="`${h.name}（${h.code}）`"
            :value="h.code"
          />
        </el-select>
      </el-form-item>
      <el-form-item v-else label="处理器">
        <el-input :model-value="form.handler" disabled />
      </el-form-item>
      <el-form-item label="Cron" prop="cron">
        <el-input v-model="form.cron" placeholder="分 时 日 月 周，如 0 3 * * *" />
        <div class="form-tip">标准 5 段：分 时 日 月 周（0=周日）。例：每天 03:00 → 0 3 * * *</div>
      </el-form-item>
      <el-form-item label="参数" prop="params">
        <el-input
          v-model="form.params"
          type="textarea"
          :rows="3"
          placeholder='JSON，如 {"retainDays":90}'
        />
      </el-form-item>
      <el-form-item label="备注" prop="remark">
        <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="255" />
      </el-form-item>
    </ProFormDialog>

    <el-drawer v-model="logVisible" :title="logTitle" size="720px" destroy-on-close>
      <div class="log-toolbar">
        <el-select v-model="logStatus" clearable placeholder="执行状态" style="width: 120px" @change="loadLogs">
          <el-option label="成功" :value="1" />
          <el-option label="失败" :value="0" />
        </el-select>
        <el-button @click="loadLogs">刷新</el-button>
      </div>
      <el-table :data="logData" v-loading="logLoading" border size="small">
        <el-table-column prop="createdAt" label="时间" width="160" />
        <el-table-column label="触发" width="70" align="center">
          <template #default="{ row }">{{ row.triggerType === 1 ? '手动' : '调度' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="durationMs" label="耗时(ms)" width="90" align="center" />
        <el-table-column prop="result" label="结果" min-width="160" show-overflow-tooltip />
        <el-table-column label="操作" width="70" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showLogDetail(row as JobLogItem)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="log-pager">
        <el-pagination
          v-model:current-page="logQuery.page"
          v-model:page-size="logQuery.pageSize"
          :total="logTotal"
          layout="total, prev, pager, next"
          small
          @current-change="loadLogs"
        />
      </div>
    </el-drawer>

    <el-dialog v-model="logDetailVisible" title="执行日志详情" width="640px" destroy-on-close>
      <el-descriptions v-if="currentLog" :column="2" border>
        <el-descriptions-item label="任务">{{ currentLog.jobName }}</el-descriptions-item>
        <el-descriptions-item label="处理器">{{ currentLog.handler }}</el-descriptions-item>
        <el-descriptions-item label="触发">
          {{ currentLog.triggerType === 1 ? '手动' : '调度' }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentLog.status === 1 ? 'success' : 'danger'" size="small">
            {{ currentLog.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="耗时">{{ currentLog.durationMs }} ms</el-descriptions-item>
        <el-descriptions-item label="时间">{{ currentLog.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="结果" :span="2">
          <div class="pre-wrap">{{ currentLog.result || '-' }}</div>
        </el-descriptions-item>
        <el-descriptions-item v-if="currentLog.errorMsg" label="错误" :span="2">
          <div class="text-danger pre-wrap">{{ currentLog.errorMsg }}</div>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormRules } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { useAuthStore } from '@/stores/auth'
import {
  getJobList,
  getJobHandlers,
  createJob,
  updateJob,
  updateJobStatus,
  deleteJob,
  runJob,
  getJobLogList,
  type JobItem,
  type JobHandlerItem,
  type JobLogItem,
} from '@/api/modules/job'

const authStore = useAuthStore()
const proTableRef = ref()
const tableData = ref<JobItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, search: {} as Record<string, any> })
const statusSavingId = ref<number | null>(null)

const handlers = ref<JobHandlerItem[]>([])

const searchFields = computed(() => [
  { prop: 'name', label: '任务名称' },
  {
    prop: 'handler',
    label: '处理器',
    type: 'select' as const,
    options: handlers.value.map((h) => ({ label: h.name, value: h.code })),
  },
  {
    prop: 'status',
    label: '状态',
    type: 'select' as const,
    options: [
      { label: '启用', value: 1 },
      { label: '暂停', value: 0 },
    ],
  },
])

async function loadHandlers() {
  try {
    const res = await getJobHandlers()
    handlers.value = res.data || []
  } catch {
    handlers.value = []
  }
}

async function loadData(params: { page: number; pageSize: number; search?: Record<string, any> }) {
  loading.value = true
  try {
    const res = await getJobList({
      page: params.page,
      pageSize: params.pageSize,
      name: params.search?.name || undefined,
      handler: params.search?.handler || undefined,
      status: params.search?.status !== undefined && params.search?.status !== ''
        ? Number(params.search.status)
        : undefined,
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
  loadData({ page: 1, pageSize: query.pageSize })
}
function handlePageChange(p: any) {
  query.page = p.page
  query.pageSize = p.pageSize
  loadData({ ...query })
}

const formVisible = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const formSubmitting = ref(false)
const editingId = ref(0)
const form = reactive({
  name: '',
  handler: '',
  cron: '0 3 * * *',
  params: '',
  remark: '',
})
const formRules: FormRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  handler: [{ required: true, message: '请选择处理器', trigger: 'change' }],
  cron: [{ required: true, message: '请输入 Cron', trigger: 'blur' }],
}

function resetForm() {
  form.name = ''
  form.handler = handlers.value[0]?.code || ''
  form.cron = '0 3 * * *'
  form.params = ''
  form.remark = ''
}

function openCreate() {
  formMode.value = 'create'
  resetForm()
  formVisible.value = true
}

function openEdit(row: JobItem) {
  formMode.value = 'edit'
  editingId.value = row.id
  form.name = row.name
  form.handler = row.handler
  form.cron = row.cron
  form.params = row.params || ''
  form.remark = row.remark || ''
  formVisible.value = true
}

async function handleFormSubmit() {
  formSubmitting.value = true
  try {
    if (formMode.value === 'create') {
      await createJob({
        name: form.name,
        handler: form.handler,
        cron: form.cron,
        params: form.params,
        remark: form.remark,
        status: 1,
      })
      ElMessage.success('创建成功')
    } else {
      await updateJob(editingId.value, {
        name: form.name,
        cron: form.cron,
        params: form.params,
        remark: form.remark,
      })
      ElMessage.success('更新成功')
    }
    formVisible.value = false
    loadData(query)
  } finally {
    formSubmitting.value = false
  }
}

async function handleStatusChange(row: JobItem, enabled: boolean) {
  const next = enabled ? 1 : 0
  const prev = row.status
  row.status = next
  statusSavingId.value = row.id
  try {
    await updateJobStatus(row.id, next)
    ElMessage.success(enabled ? '已启用' : '已暂停')
    loadData(query)
  } catch {
    row.status = prev
  } finally {
    statusSavingId.value = null
  }
}

async function handleRun(row: JobItem) {
  await ElMessageBox.confirm(`确认立即执行「${row.name}」？`, '提示', { type: 'warning' })
  const res = await runJob(row.id)
  if (res.data?.ok === false) {
    ElMessage.warning(res.data.msg || '执行失败，详见日志')
  } else {
    ElMessage.success(res.data?.msg || '执行完成')
  }
  loadData(query)
}

async function handleDelete(row: JobItem) {
  await ElMessageBox.confirm(`确认删除任务「${row.name}」？`, '提示', { type: 'warning' })
  await deleteJob(row.id)
  ElMessage.success('已删除')
  loadData(query)
}

function handleAction(row: JobItem, cmd: string) {
  if (cmd === 'edit') openEdit(row)
  else if (cmd === 'delete') void handleDelete(row)
}

const logVisible = ref(false)
const logLoading = ref(false)
const logData = ref<JobLogItem[]>([])
const logTotal = ref(0)
const logJobId = ref<number | null>(null)
const logJobName = ref('')
const logStatus = ref<number | undefined>(undefined)
const logQuery = reactive({ page: 1, pageSize: 10 })
const logTitle = computed(() => (logJobName.value ? `执行日志 - ${logJobName.value}` : '执行日志'))

function openLogs(row: JobItem) {
  logJobId.value = row.id
  logJobName.value = row.name
  logStatus.value = undefined
  logQuery.page = 1
  logVisible.value = true
  loadLogs()
}

async function loadLogs() {
  if (!logJobId.value) return
  logLoading.value = true
  try {
    const res = await getJobLogList({
      page: logQuery.page,
      pageSize: logQuery.pageSize,
      jobId: logJobId.value,
      status: logStatus.value,
    })
    logData.value = res.data?.list || []
    logTotal.value = res.data?.total || 0
  } finally {
    logLoading.value = false
  }
}

const logDetailVisible = ref(false)
const currentLog = ref<JobLogItem | null>(null)
function showLogDetail(row: JobLogItem) {
  currentLog.value = row
  logDetailVisible.value = true
}

onMounted(async () => {
  await loadHandlers()
  loadData(query)
})
</script>

<style scoped lang="scss">
.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
}
.log-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.log-pager {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
.pre-wrap {
  white-space: pre-wrap;
  word-break: break-all;
}
.text-danger {
  color: var(--el-color-danger);
}
</style>
