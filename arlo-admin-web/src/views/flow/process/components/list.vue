<template>
  <div class="flow-list">
    <ProTable
      :data="data"
      :loading="loading"
      :total="data.length"
      :search-fields="searchFields"
      :show-index="false"
      :show-pagination="false"
      :action-width="180"
      @search="handleSearch"
      @reset="handleReset"
    >
      <template #toolbar>
        <el-dropdown v-permission="'flow:process:add'" trigger="click" @command="onCreate">
          <el-button type="primary">
            创建流程
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="t in PROCESS_CREATE_TYPES"
                :key="t"
                :command="t"
              >
                {{ PROCESS_TYPE_META[t].createLabel }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>

      <el-table-column label="流程名称" min-width="280">
        <template #default="{ row }">
          <div class="name-cell">
            <div class="viewIcon-item" :style="{ backgroundColor: iconColor(row) }">
              <el-icon v-if="iconName(row)" :size="28" color="#fff">
                <component :is="iconName(row)" />
              </el-icon>
              <span v-else class="icon-fallback">流</span>
            </div>
            <div>
              <div>{{ row.processName }}</div>
              <div class="key-text">{{ row.processKey }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="110" align="center">
        <template #default="{ row }">
          <el-tag size="small" effect="plain" :type="typeTagType(row.processType)">
            {{ processTypeLabel(row.processType) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="流程版本" min-width="140">
        <template #default="{ row }">
          V{{ row.processVersion }}
          <el-tag v-if="row.processState == 0" type="danger" size="small" style="margin-left: 6px">已禁用</el-tag>
          <el-tag v-else type="success" size="small" style="margin-left: 6px">已启用</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">{{ row.remark || '-' }}</template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" min-width="170" />

      <template #actions="{ row }">
        <el-button
          v-if="row.processState == 0"
          v-permission="'flow:process:edit'"
          type="primary"
          link
          size="small"
          @click="handleState(row)"
        >
          启用
        </el-button>
        <el-button
          v-else
          v-permission="'flow:process:edit'"
          type="danger"
          link
          size="small"
          @click="handleState(row)"
        >
          禁用
        </el-button>
        <el-button v-permission="'flow:process:edit'" type="primary" link size="small" @click="handleEdit(row)">
          编辑
        </el-button>
        <el-dropdown
          v-if="
            authStore.hasPermission('flow:process:add') ||
            authStore.hasPermission('flow:process:list') ||
            authStore.hasPermission('flow:process:delete')
          "
          trigger="click"
          @command="(cmd: string) => handleMore(row, cmd)"
        >
          <el-button type="info" link size="small">
            更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-if="authStore.hasPermission('flow:process:add')"
                command="copy"
              >
                复制
              </el-dropdown-item>
              <el-dropdown-item
                v-if="authStore.hasPermission('flow:process:list')"
                command="history"
              >
                历史版本
              </el-dropdown-item>
              <el-dropdown-item
                v-if="authStore.hasPermission('flow:process:edit')"
                command="check"
              >
                配置检查
              </el-dropdown-item>
              <el-dropdown-item
                v-if="authStore.hasPermission('flow:process:delete')"
                command="delete"
                divided
              >
                删除
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
    </ProTable>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue'
import { ArrowDown } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import { checkFlowProcess, cloneFlowProcess, deleteFlowProcess, updateFlowProcessState } from '@/api'
import { useAuthStore } from '@/stores/auth'
import {
  PROCESS_CREATE_TYPES,
  PROCESS_TYPE_META,
  normalizeProcessType,
  processTypeLabel,
  type ProcessType,
} from '../processType'

const emits = defineEmits<{ event: [data: any] }>()
const authStore = useAuthStore()

const loading = ref(false)
const defaultData = ref<any>({ processList: [] })
const data = ref<any[]>([])
const filters = ref<{ processName?: string; processType?: string; processState?: number | '' }>({})

const searchFields = computed(() => [
  { prop: 'processName', label: '流程名称', placeholder: '请输入' },
  {
    prop: 'processType',
    label: '类型',
    type: 'select' as const,
    options: [
      { label: '审批流程', value: 'main' },
      { label: '业务流程', value: 'business' },
      { label: '子流程', value: 'child' },
    ],
  },
  {
    prop: 'processState',
    label: '状态',
    type: 'select' as const,
    options: [
      { label: '已启用', value: 1 },
      { label: '已禁用', value: 0 },
    ],
  },
])

function onCreate(type: ProcessType) {
  handleEdit({ processType: type })
}

function typeTagType(type: string) {
  const t = normalizeProcessType(type)
  if (t === 'child') return 'warning'
  if (t === 'business') return 'success'
  return 'primary'
}

function parseIcon(row: any): { icon?: string; color?: string } {
  const raw = row.processIcon
  if (!raw) return {}
  try {
    if (typeof raw === 'string' && raw.startsWith('{')) return JSON.parse(raw)
    return { icon: raw }
  } catch {
    return { icon: String(raw) }
  }
}

function iconName(row: any) {
  return parseIcon(row).icon || ''
}

function iconColor(row: any) {
  return parseIcon(row).color || 'rgba(30, 144, 255, 1)'
}

function applyFilter() {
  let list = defaultData.value.processList || []
  const name = (filters.value.processName || '').trim()
  if (name) list = list.filter((item: any) => String(item.processName).includes(name))
  if (filters.value.processType) {
    list = list.filter(
      (item: any) => normalizeProcessType(item.processType) === filters.value.processType,
    )
  }
  if (filters.value.processState === 0 || filters.value.processState === 1) {
    list = list.filter((item: any) => item.processState === filters.value.processState)
  }
  data.value = list
}

function handleSearch(p: Record<string, any>) {
  filters.value = {
    processName: p.processName,
    processType: p.processType,
    processState: p.processState,
  }
  applyFilter()
}

function handleReset() {
  filters.value = {}
  applyFilter()
}

function handleDelete(row: any) {
  ElMessageBox.confirm('确定将选择数据删除?', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await deleteFlowProcess(row.processId)
      ElMessage.success('操作成功!')
      emits('event', { type: 'refresh', row: defaultData.value })
    })
    .catch(() => {})
}

function handleEdit(row: any = {}) {
  const payload = { ...row }
  if (!payload.categoryId && defaultData.value?.id && defaultData.value.id !== '-1') {
    payload.categoryId = defaultData.value.id
  }
  emits('event', { type: 'edit', row: payload })
}

async function handleState(row: any) {
  const state = row.processState == 0 ? 1 : 0
  await updateFlowProcessState(row.processId, state)
  ElMessage.success('操作成功')
  emits('event', { type: 'refresh', row: defaultData.value })
}

function handleCopy(row: any) {
  ElMessageBox.confirm('确定将复制一条相同流程?', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await cloneFlowProcess(row.processId)
      ElMessage.success('复制成功!')
      emits('event', { type: 'refresh', row: defaultData.value })
    })
    .catch(() => {})
}

function handleHistory(row: any) {
  emits('event', { type: 'history', row })
}

async function handleCheck(row: any) {
  try {
    const res = await checkFlowProcess(row.processId)
    const data = res?.data
    if (data?.ok) {
      ElMessage.success(data.message || '配置检查通过')
    } else {
      ElMessageBox.alert(data?.message || '配置检查未通过', '配置检查', { type: 'warning' })
    }
  } catch {
    /* request 已 toast */
  }
}

function handleMore(row: any, cmd: string) {
  if (cmd === 'copy') handleCopy(row)
  else if (cmd === 'history') handleHistory(row)
  else if (cmd === 'check') handleCheck(row)
  else if (cmd === 'delete') handleDelete(row)
}

function reload(row: any) {
  defaultData.value = row || { processList: [] }
  defaultData.value.processList = defaultData.value.processList || []
  applyFilter()
}

defineExpose({ reload })
</script>

<style lang="scss" scoped>
.flow-list {
  height: 100%;
  overflow: auto;
}
.name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.viewIcon-item {
  width: 56px;
  height: 56px;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 8px;
  flex-shrink: 0;
}
.icon-fallback {
  color: #fff;
  font-size: 18px;
}
.key-text {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}
</style>
