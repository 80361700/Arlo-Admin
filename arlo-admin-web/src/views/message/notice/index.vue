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
        <el-button v-permission="'message:notice:add'" type="primary" @click="handleAdd">新增公告</el-button>
      </template>

      <el-table-column prop="id" label="ID" width="80" align="center" />
      <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
      <el-table-column label="类型" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.type === 2 ? 'warning' : 'info'" size="small">
            {{ noticeTypeLabel(row.type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="级别" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="levelTag(row.level)" size="small">
            {{ noticeLevelLabel(row.level) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90" align="center">
        <template #default="{ row }">
          <el-tag :type="statusTag(row.status)" size="small">
            {{ statusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="publisher" label="发布人" width="100">
        <template #default="{ row }">
          <span>{{ row.publisher || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="170" />

      <template #actions="{ row }">
        <!-- 草稿 / 已撤回：可编辑、发布 -->
        <template v-if="row.status === 0 || row.status === 2">
          <el-button v-permission="'message:notice:edit'" type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button v-permission="'message:notice:edit'" type="success" link size="small" @click="handlePublish(row)">
            {{ row.status === 2 ? '重新发布' : '发布' }}
          </el-button>
          <el-dropdown
            v-if="authStore.hasPermission('message:notice:view') || authStore.hasPermission('message:notice:delete')"
            trigger="click"
            @command="(cmd: string) => handleAction(row, cmd)"
          >
            <el-button type="info" link size="small">
              更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-if="authStore.hasPermission('message:notice:view')" command="detail">详情</el-dropdown-item>
                <el-dropdown-item v-if="authStore.hasPermission('message:notice:delete')" command="delete">删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <!-- 已发布：可撤回 -->
        <template v-else-if="row.status === 1">
          <el-button v-permission="'message:notice:edit'" type="warning" link size="small" @click="handleRevoke(row)">撤回</el-button>
          <el-button v-permission="'message:notice:view'" type="info" link size="small" @click="handleDetail(row)">详情</el-button>
          <el-dropdown v-if="authStore.hasPermission('message:notice:delete')" trigger="click" @command="(cmd: string) => handleAction(row, cmd)">
            <el-button type="info" link size="small">
              更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="delete">删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </template>
    </ProTable>

    <!-- 新增/编辑弹窗 -->
    <ProFormDialog
      ref="formDialogRef"
      width="1000px"
      v-model="dialogVisible"
      :title="dialogTitle"
      :model="form"
      :rules="formRules"
      :submitting="submitting"
      @submit="handleSubmit"
    >
      <el-form-item label="标题" prop="title">
        <el-input v-model="form.title" placeholder="请输入标题" maxlength="128" />
      </el-form-item>
      <el-form-item label="类型" prop="type">
        <el-radio-group v-model="form.type">
          <el-radio v-for="opt in noticeTypeOptions" :key="String(opt.value)" :value="opt.value">
            {{ opt.label }}
          </el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="级别" prop="level">
        <el-radio-group v-model="form.level">
          <el-radio v-for="opt in noticeLevelOptions" :key="String(opt.value)" :value="opt.value">
            {{ opt.label }}
          </el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="内容" prop="content">
        <RichEditor v-model="form.content" placeholder="请输入公告内容" />
      </el-form-item>
    </ProFormDialog>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="公告详情" width="1000px" :close-on-click-modal="false">
      <template v-if="detailData">
        <div class="detail-header">
          <h3>{{ detailData.title }}</h3>
          <div class="detail-meta">
            <el-tag :type="detailData.type === 2 ? 'warning' : 'info'" size="small">{{ noticeTypeLabel(detailData.type) }}</el-tag>
            <el-tag :type="levelTag(detailData.level)" size="small">{{ noticeLevelLabel(detailData.level) }}</el-tag>
            <el-tag :type="statusTag(detailData.status)" size="small">{{ statusLabel(detailData.status) }}</el-tag>
            <span>发布人：{{ detailData.publisher || '-' }}</span>
            <span>{{ detailData.createdAt }}</span>
          </div>
        </div>
        <div class="detail-body">
          <RichContent :html="detailData.content" />
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import { ArrowDown } from '@element-plus/icons-vue'
import type { FormRules } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import RichEditor from '@/components/RichEditor.vue'
import RichContent from '@/components/RichContent.vue'
import {
  getNoticeList, getNoticeDetail, createNotice, updateNotice, deleteNotice,
  publishNotice, revokeNotice,
  type NoticeItem, type NoticeListQuery, type NoticeFormParams,
} from '@/api/modules/notice'
import { useAuthStore } from '@/stores/auth'
import { useDict, DictCode } from '@/utils/useDict'

const authStore = useAuthStore()
const { options: noticeTypeOptions, getLabel: noticeTypeLabel } = useDict(DictCode.NoticeType)
const { options: noticeLevelOptions, getLabel: noticeLevelLabel } = useDict(DictCode.NoticeLevel)
const { options: noticeStatusOptions, getLabel: noticeStatusLabel } = useDict(DictCode.NoticeStatus)

// ==================== 查询 ====================
const proTableRef = ref()
const tableData = ref<NoticeItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive<NoticeListQuery>({ page: 1, pageSize: 10 })

const searchFields = computed(() => [
  { prop: 'title', label: '标题' },
  {
    prop: 'status', label: '状态', type: 'select' as const,
    options: noticeStatusOptions.value,
  },
  {
    prop: 'type', label: '类型', type: 'select' as const,
    options: noticeTypeOptions.value,
  },
])

function levelTag(l: number) {
  const m: Record<number, 'info' | 'warning' | 'danger'> = { 1: 'info', 2: 'warning', 3: 'danger' }
  return m[l] || 'info'
}
function statusTag(s: number) {
  const m: Record<number, 'info' | 'success' | 'warning'> = { 0: 'info', 1: 'success', 2: 'warning' }
  return m[s] || 'info'
}
function statusLabel(s: number) {
  return noticeStatusLabel(s)
}

async function loadData(params: NoticeListQuery) {
  loading.value = true
  try {
    const res = await getNoticeList(params)
    tableData.value = res.data.list || []
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) { query.page = 1; Object.assign(query, p); loadData(query) }
function handleReset() { loadData({ page: 1, pageSize: 10 }) }
function handlePageChange(p: any) { Object.assign(query, p); loadData(query) }

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const dialogTitle = ref('')
const isEdit = ref(false)
const submitting = ref(false)
const formDialogRef = ref()

const defaultForm: NoticeFormParams & { id: number } = {
  id: 0, title: '', content: '', type: 1, level: 1,
}
const form = reactive({ ...defaultForm })

const formRules: FormRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }],
}

function handleAdd() {
  isEdit.value = false
  dialogTitle.value = '新增公告'
  Object.assign(form, defaultForm)
  dialogVisible.value = true
}

async function handleEdit(row: NoticeItem) {
  isEdit.value = true
  dialogTitle.value = '编辑公告'
  try {
    const res = await getNoticeDetail(row.id)
    const d = res.data
    Object.assign(form, { id: d.id, title: d.title, content: d.content, type: d.type, level: d.level })
    dialogVisible.value = true
  } catch (err: any) {
    showRequestError(err, '获取详情失败')
  }
}

async function handleSubmit() {
  submitting.value = true
  try {
    const data: NoticeFormParams = { title: form.title, content: form.content, type: form.type, level: form.level }
    if (isEdit.value) {
      await updateNotice(form.id, data)
      ElMessage.success('更新成功')
    } else {
      await createNotice(data)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    loadData(query)
  } catch (err: any) {
    showRequestError(err, '操作失败')
  } finally {
    submitting.value = false
  }
}

// ==================== 发布/撤回 ====================
async function handlePublish(row: NoticeItem) {
  try {
    await ElMessageBox.confirm(`确认发布"${row.title}"？`, '发布确认', { type: 'warning' })
    await publishNotice(row.id)
    ElMessage.success('发布成功')
    loadData(query)
  } catch { /* 取消 */ }
}

async function handleRevoke(row: NoticeItem) {
  try {
    await ElMessageBox.confirm(`确认撤回"${row.title}"？`, '撤回确认', { type: 'warning' })
    await revokeNotice(row.id)
    ElMessage.success('撤回成功')
    loadData(query)
  } catch { /* 取消 */ }
}

// ==================== 删除 ====================
async function handleDelete(row: NoticeItem) {
  try {
    await ElMessageBox.confirm(`确认删除"${row.title}"？`, '删除确认', { type: 'warning' })
    await deleteNotice(row.id)
    ElMessage.success('删除成功')
    loadData(query)
  } catch { /* 取消 */ }
}

// ==================== 详情 ====================
const detailVisible = ref(false)
const detailData = ref<NoticeItem | null>(null)

async function handleDetail(row: NoticeItem) {
  try {
    const res = await getNoticeDetail(row.id)
    detailData.value = res.data
    detailVisible.value = true
  } catch (err: any) {
    showRequestError(err, '获取详情失败')
  }
}

function handleAction(row: NoticeItem, command: string) {
  if (command === 'detail') handleDetail(row)
  else if (command === 'delete') handleDelete(row)
}

onMounted(() => loadData(query))
</script>

<style scoped lang="scss">
.page-container { padding: 0; }

.detail-header {
  margin-bottom: 20px;
  h3 { margin: 0 0 12px; font-size: 18px; }
  .detail-meta {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 13px;
    color: #909399;
    span { line-height: 1; }
  }
}
.detail-body {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 4px;
}
</style>
