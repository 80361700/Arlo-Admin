<template>
  <el-dialog
    v-model="visible"
    title="历史版本"
    width="860px"
    destroy-on-close
    append-to-body
    @closed="onClosed"
  >
    <el-table v-loading="loading" :data="list" border stripe style="width: 100%">
      <el-table-column label="图标" width="80" align="center">
        <template #default="{ row }">
          <div class="viewIcon-item" :style="{ backgroundColor: iconColor(asHistory(row)) }">
            <el-icon v-if="iconName(asHistory(row))" :size="22" color="#fff">
              <component :is="iconName(asHistory(row))" />
            </el-icon>
            <span v-else class="icon-fallback">流</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="processName" label="流程名称" min-width="180" show-overflow-tooltip />
      <el-table-column label="流程版本" width="110" align="center">
        <template #default="{ row }">V{{ asHistory(row).processVersion }}</template>
      </el-table-column>
      <el-table-column label="备注" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ asHistory(row).remark || '-' }}</template>
      </el-table-column>
      <el-table-column prop="createdAt" label="保存时间" width="170" />
      <el-table-column label="操作" width="140" fixed="right" align="center">
        <template #default="{ row }">
          <el-button
            v-permission="'flow:process:edit'"
            type="primary"
            link
            size="small"
            @click="handleCheckout(asHistory(row))"
          >
            迁出
          </el-button>
          <el-button
            v-permission="'flow:process:list'"
            type="primary"
            link
            size="small"
            @click="handlePreview(asHistory(row))"
          >
            预览
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        background
        layout="total, prev, pager, next"
        :total="total"
        @current-change="load"
        @size-change="load"
      />
    </div>
  </el-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  checkoutFlowProcessHistory,
  getFlowProcessHistories,
  type FlowProcessHistoryBrief,
} from '@/api'

const emits = defineEmits<{
  preview: [row: FlowProcessHistoryBrief]
  checkout: []
}>()

const visible = ref(false)
const loading = ref(false)
const processId = ref(0)
const list = ref<FlowProcessHistoryBrief[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

function asHistory(row: unknown): FlowProcessHistoryBrief {
  return row as FlowProcessHistoryBrief
}

function parseIcon(row: FlowProcessHistoryBrief): { icon?: string; color?: string } {
  const raw = row.processIcon
  if (!raw) return {}
  try {
    if (typeof raw === 'string' && raw.startsWith('{')) return JSON.parse(raw)
    return { icon: raw }
  } catch {
    return { icon: String(raw) }
  }
}

function iconName(row: FlowProcessHistoryBrief) {
  return parseIcon(row).icon || ''
}

function iconColor(row: FlowProcessHistoryBrief) {
  return parseIcon(row).color || 'rgba(30, 144, 255, 1)'
}

async function load() {
  if (!processId.value) return
  loading.value = true
  try {
    const res = await getFlowProcessHistories(processId.value, {
      page: page.value,
      pageSize: pageSize.value,
    })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

function init(row: { processId: number }) {
  processId.value = row.processId
  page.value = 1
  visible.value = true
  load()
}

function onClosed() {
  list.value = []
  processId.value = 0
}

function handlePreview(row: FlowProcessHistoryBrief) {
  visible.value = false
  emits('preview', row)
}

function handleCheckout(row: FlowProcessHistoryBrief) {
  ElMessageBox.confirm(
    `确定迁出 V${row.processVersion} 覆盖当前流程内容？迁出后将生成新版本。`,
    '迁出历史版本',
    { type: 'warning', confirmButtonText: '确定迁出', cancelButtonText: '取消' },
  )
    .then(async () => {
      await checkoutFlowProcessHistory(row.processId, row.historyId)
      ElMessage.success('迁出成功')
      visible.value = false
      emits('checkout')
    })
    .catch(() => {})
}

defineExpose({ init })
</script>

<style lang="scss" scoped>
.viewIcon-item {
  width: 40px;
  height: 40px;
  margin: 0 auto;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 6px;
}
.icon-fallback {
  color: #fff;
  font-size: 14px;
}
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
