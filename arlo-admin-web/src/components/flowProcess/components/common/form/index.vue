<template>
  <el-dialog
    v-model="dialogShow"
    title="表单分类选择"
    width="640px"
    destroy-on-close
    append-to-body
    :before-close="dialogClose"
  >
    <el-input v-model="keyword" placeholder="搜索名称/编码" clearable style="margin-bottom: 12px" />
    <el-table
      v-loading="loading"
      :data="filtered"
      border
      highlight-current-row
      max-height="420"
      @current-change="onCurrentChange"
      @row-click="onCurrentChange"
    >
      <el-table-column width="50" align="center">
        <template #default="{ row }">
          <el-radio class="form-pick-radio" :model-value="selectedId" :value="row.id" />
        </template>
      </el-table-column>
      <el-table-column prop="name" label="模板名称" min-width="160" show-overflow-tooltip />
      <el-table-column prop="code" label="编码" min-width="120" show-overflow-tooltip />
      <el-table-column label="类型" width="100" align="center">
        <template #default="{ row }">
          <el-tag v-if="Number(row.formType) === 2" size="small" type="warning" effect="plain">系统</el-tag>
          <el-tag v-else size="small" effect="plain">设计</el-tag>
        </template>
      </el-table-column>
    </el-table>
    <template #footer>
      <el-button @click="dialogClose">取消</el-button>
      <el-button type="primary" @click="dialogSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getFlowFormOptions, type FlowFormOption } from '@/api'

const props = withDefaults(
  defineProps<{
    /** 仅展示指定类型：1 设计 2 系统；不传则全部 */
    formTypes?: number[]
  }>(),
  {
    formTypes: undefined,
  },
)

const emits = defineEmits<{
  success: [item: { id: number; name: string; code: string; formType?: number; pcUrl?: string }]
  close: []
}>()

const dialogShow = ref(false)
const loading = ref(false)
const keyword = ref('')
const list = ref<FlowFormOption[]>([])
const selectedId = ref<number>()
const current = ref<FlowFormOption>()

const filtered = computed(() => {
  const kw = keyword.value.trim()
  if (!kw) return list.value
  return list.value.filter(
    (item) => item.name.includes(kw) || item.code.includes(kw) || String(item.id).includes(kw),
  )
})

async function init(obj: { formId?: number | string } = {}) {
  dialogShow.value = true
  keyword.value = ''
  selectedId.value = obj.formId ? Number(obj.formId) : undefined
  current.value = undefined
  loading.value = true
  try {
    const res = await getFlowFormOptions()
    let rows = res.data || []
    if (props.formTypes?.length) {
      const allow = new Set(props.formTypes.map(Number))
      rows = rows.filter((i) => allow.has(Number(i.formType) || 1))
    }
    list.value = rows
    if (selectedId.value) {
      current.value = list.value.find((i) => i.id === selectedId.value)
    }
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

function onCurrentChange(row: FlowFormOption | null | undefined) {
  if (!row) return
  selectedId.value = row.id
  current.value = row
}

function dialogClose() {
  dialogShow.value = false
  emits('close')
}

function dialogSubmit() {
  const hit = current.value || list.value.find((i) => i.id === selectedId.value)
  if (!hit) {
    ElMessage.warning('请选择表单')
    return
  }
  dialogShow.value = false
  emits('success', {
    id: hit.id,
    name: hit.name,
    code: hit.code,
    formType: hit.formType,
    pcUrl: hit.pcUrl,
  })
}

defineExpose({ init })
</script>

<style scoped lang="scss">
/* 表格里只显示圆点，避免窄列把 label 截成 ... */
.form-pick-radio {
  height: auto;
  :deep(.el-radio__label) {
    display: none;
  }
}
</style>
