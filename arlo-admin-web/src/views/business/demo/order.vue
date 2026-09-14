<template>
  <div class="page-container">
    <ProTable
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
        <el-button v-permission="'business:purchaseOrder:add'" type="primary" @click="handleAdd">
          新增
        </el-button>
        <el-link type="danger" :underline="false" class="demo-tip">
          演示：创建时绑定业务流程；「发起」受设计器发起人权限控制。真实业务可固定自己的流程 KEY。
        </el-link>
      </template>

      <el-table-column prop="id" label="ID" width="80" align="center" />
      <el-table-column prop="title" label="名称" min-width="140" show-overflow-tooltip />
      <el-table-column prop="content" label="内容" min-width="160" show-overflow-tooltip />
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.status === 1" type="warning" size="small">审批中</el-tag>
          <el-tag v-else-if="row.status === 2" type="success" size="small">已通过</el-tag>
          <el-tag v-else-if="row.status === 3" type="danger" size="small">已拒绝</el-tag>
          <el-tag v-else type="info" size="small">待审批</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="业务流程" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.processKey">
            {{ row.processName || row.processKey }}
            <span class="key-muted">（{{ row.processKey }}）</span>
          </span>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="instanceId" label="实例ID" width="100" align="center">
        <template #default="{ row }">
          {{ row.instanceId || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="createBy" label="创建人" width="100" />
      <el-table-column prop="createdAt" label="创建时间" min-width="170" />

      <template #actions="{ row }">
        <el-button
          v-if="row.status === 0 && row.canLaunch"
          v-permission="'business:purchaseOrder:launch'"
          type="primary"
          link
          size="small"
          @click="openLaunch(row)"
        >
          发起
        </el-button>
        <el-tooltip
          v-else-if="row.status === 0 && row.processId && !row.canLaunch"
          content="当前账号不符合该流程设计器中的发起人权限"
          placement="top"
        >
          <el-button type="info" link size="small" disabled>发起</el-button>
        </el-tooltip>
        <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
        <el-popconfirm
          v-if="row.status !== 1 && authStore.hasPermission('business:purchaseOrder:delete')"
          title="确认删除该单据？"
          @confirm="handleDelete(row.id)"
        >
          <template #reference>
            <el-button type="danger" link size="small">删除</el-button>
          </template>
        </el-popconfirm>
      </template>
    </ProTable>

    <ProFormDialog
      v-model="dialogVisible"
      title="新增业务单据"
      :model="form"
      :rules="formRules"
      :submitting="submitting"
      @submit="handleSubmit"
    >
      <template #default>
        <el-form-item label="业务流程" prop="processId">
          <el-select
            v-model="form.processId"
            placeholder="请选择业务流程"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="p in processOptions"
              :key="p.processId"
              :label="`${p.processName}（${p.processKey}）`"
              :value="p.processId"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" prop="title">
          <el-input v-model="form.title" placeholder="请输入名称" maxlength="128" />
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="3"
            placeholder="请输入内容"
            maxlength="512"
            show-word-limit
          />
        </el-form-item>
      </template>
    </ProFormDialog>

    <OrderLaunchDrawer v-model="launchVisible" :row="launchRow" @success="loadList" />

    <BizInstanceDetailDrawer
      v-model="detailVisible"
      :instance-id="detailRow?.instanceId || 0"
      :biz-id="detailRow?.id || 0"
      :load-process-preview="loadDetailProcessPreview"
      :title="detailRow ? `详情 · ${detailRow.title}` : '详情'"
    >
      <template #business>
        <el-descriptions v-if="detailRow" :column="1" border>
          <el-descriptions-item label="名称">{{ detailRow.title }}</el-descriptions-item>
          <el-descriptions-item label="内容">{{ detailRow.content }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag v-if="detailRow.status === 1" type="warning" size="small">审批中</el-tag>
            <el-tag v-else-if="detailRow.status === 2" type="success" size="small">已通过</el-tag>
            <el-tag v-else-if="detailRow.status === 3" type="danger" size="small">已拒绝</el-tag>
            <el-tag v-else type="info" size="small">待审批</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="业务流程">
            <template v-if="detailRow.processKey">
              {{ detailRow.processName || detailRow.processKey }}
              （{{ detailRow.processKey }}）
            </template>
            <template v-else>-</template>
          </el-descriptions-item>
          <el-descriptions-item label="实例ID">
            {{ detailRow.instanceId || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建人">{{ detailRow.createBy || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ detailRow.createdAt || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </BizInstanceDetailDrawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { showRequestError } from '@/utils/requestError'
import {
  createDemoPurchaseOrder,
  deleteDemoPurchaseOrders,
  getDemoProcessOptions,
  getDemoProcessPreview,
  getDemoPurchaseOrderList,
  type DemoProcessOption,
  type DemoPurchaseOrderItem,
} from '@/api'
import OrderLaunchDrawer from './OrderLaunchDrawer.vue'
import BizInstanceDetailDrawer from '@/views/flow/approve/components/BizInstanceDetailDrawer.vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'

const authStore = useAuthStore()

const loading = ref(false)
const tableData = ref<DemoPurchaseOrderItem[]>([])
const total = ref(0)
const query = reactive({
  page: 1,
  pageSize: 10,
  title: '',
  status: undefined as number | undefined,
})

const searchFields = [
  { prop: 'title', label: '名称' },
  {
    prop: 'status',
    label: '状态',
    type: 'select' as const,
    options: [
      { label: '待审批', value: 0 },
      { label: '审批中', value: 1 },
      { label: '已通过', value: 2 },
      { label: '已拒绝', value: 3 },
    ],
  },
]

const dialogVisible = ref(false)
const submitting = ref(false)
const processOptions = ref<DemoProcessOption[]>([])
const form = reactive({ title: '', content: '', processId: undefined as number | undefined })
const formRules: FormRules = {
  processId: [{ required: true, message: '请选择业务流程', trigger: 'change' }],
  title: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }],
}

const launchVisible = ref(false)
const launchRow = ref<DemoPurchaseOrderItem | null>(null)
const detailVisible = ref(false)
const detailRow = ref<DemoPurchaseOrderItem | null>(null)

async function loadDetailProcessPreview(bizId: number) {
  const res = await getDemoProcessPreview(bizId)
  return { modelContent: res.data?.modelContent || null }
}

async function loadProcessOptions() {
  try {
    const res = await getDemoProcessOptions()
    processOptions.value = res.data || []
  } catch {
    processOptions.value = []
  }
}

async function loadList() {
  loading.value = true
  try {
    const res = await getDemoPurchaseOrderList({
      page: query.page,
      pageSize: query.pageSize,
      title: query.title || undefined,
      status: query.status,
    })
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    showRequestError(e)
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) {
  query.page = 1
  query.title = p.title || ''
  query.status = p.status === '' || p.status === undefined || p.status === null ? undefined : Number(p.status)
  loadList()
}

function handleReset() {
  query.page = 1
  query.pageSize = 10
  query.title = ''
  query.status = undefined
  loadList()
}

function handlePageChange(p: any) {
  Object.assign(query, p)
  loadList()
}

async function handleAdd() {
  form.title = ''
  form.content = ''
  form.processId = undefined
  await loadProcessOptions()
  if (!processOptions.value.length) {
    ElMessage.warning('暂无可用业务流程，请先启用类型为「业务」的流程')
    return
  }
  if (processOptions.value.length === 1) {
    form.processId = processOptions.value[0].processId
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!form.processId) {
    ElMessage.warning('请选择业务流程')
    return
  }
  submitting.value = true
  try {
    await createDemoPurchaseOrder({
      title: form.title,
      content: form.content,
      processId: form.processId,
    })
    ElMessage.success('新增成功')
    dialogVisible.value = false
    loadList()
  } catch (e) {
    showRequestError(e)
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await deleteDemoPurchaseOrders([id])
    ElMessage.success('删除成功')
    loadList()
  } catch (e) {
    showRequestError(e)
  }
}

function openLaunch(row: DemoPurchaseOrderItem) {
  launchRow.value = row
  launchVisible.value = true
}

function openDetail(row: DemoPurchaseOrderItem) {
  detailRow.value = row
  detailVisible.value = true
}

onMounted(loadList)
</script>

<style scoped lang="scss">
.demo-tip {
  margin-left: 12px;
  vertical-align: middle;
}
.key-muted {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>
