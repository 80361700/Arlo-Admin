<template>
  <el-drawer
    :model-value="modelValue"
    :title="`发起 · ${processName || row?.processName || row?.processKey || '业务流程'}`"
    size="640px"
    destroy-on-close
    class="order-launch-drawer"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div v-loading="formLoading" class="launch-drawer-body">
      <CustomFlowFormHost
        v-if="formRenderType === 'vue' && formComponent"
        ref="customFormRef"
        :component-path="formComponent"
        :form-data="seedFormData"
      />
      <EBuilder
        v-else-if="pageSchema"
        ref="builderRef"
        :page-schema="pageSchema"
        :field-states="fieldStates"
      />
      <el-empty v-else-if="!formLoading && !modelContent" description="暂无表单，可直接提交" />
      <LaunchFlowPreview v-if="modelContent" ref="previewRef" :model-content="modelContent" />
    </div>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">提交</el-button>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { EBuilder, type FieldStates, type PageSchema } from 'epic-designer'
import CustomFlowFormHost from '@/views/flow/approve/components/CustomFlowFormHost.vue'
import LaunchFlowPreview from '@/views/flow/approve/components/LaunchFlowPreview.vue'
import {
  getDemoPurchaseOrderLaunchForm,
  launchDemoPurchaseOrder,
  type DemoPurchaseOrderItem,
  type LaunchSelectionNode,
} from '@/api'
import {
  buildFieldStates,
  normalizeProcessForm,
  readBuilderFormData,
} from '@/views/flow/process/components/formSchema'

const props = defineProps<{
  modelValue: boolean
  row: DemoPurchaseOrderItem | null
}>()

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  success: []
}>()

const formLoading = ref(false)
const submitting = ref(false)
const processName = ref('')
const pageSchema = ref<PageSchema | null>(null)
const fieldStates = ref<FieldStates>([])
const modelContent = ref<Record<string, any> | null>(null)
const selectionNodes = ref<LaunchSelectionNode[]>([])
const formRenderType = ref('designer')
const formComponent = ref('')
const builderRef = ref<InstanceType<typeof EBuilder>>()
const customFormRef = ref<InstanceType<typeof CustomFlowFormHost>>()
const previewRef = ref<InstanceType<typeof LaunchFlowPreview>>()

const seedFormData = computed(() => ({
  title: props.row?.title || '',
  content: props.row?.content || '',
  remark: props.row?.content || '',
  amount: 0,
}))

function startFormConfigOf(mc: Record<string, any> | null | undefined) {
  const root = (mc?.nodeConfig || mc) as any
  const raw = root?.extendConfig?.formConfig
  if (!Array.isArray(raw) || !raw.length) return []
  if (raw.every((c: any) => Number(c.opera) === 0)) return []
  return raw
}

async function loadForm() {
  if (!props.row?.id) return
  formLoading.value = true
  pageSchema.value = null
  fieldStates.value = []
  modelContent.value = null
  selectionNodes.value = []
  formRenderType.value = 'designer'
  formComponent.value = ''
  processName.value = ''
  try {
    const res = await getDemoPurchaseOrderLaunchForm(props.row.id)
    processName.value = res.data?.processName || ''
    formRenderType.value = res.data?.formRenderType === 'vue' ? 'vue' : 'designer'
    formComponent.value = res.data?.formComponent || ''
    if (formRenderType.value === 'vue') {
      pageSchema.value = null
    } else {
      const schema = normalizeProcessForm(res.data?.processForm)
      pageSchema.value = schema?.schemas?.length ? schema : null
    }
    const mc = res.data?.modelContent
    modelContent.value = mc && typeof mc === 'object' && Object.keys(mc).length ? mc : null
    fieldStates.value = buildFieldStates(startFormConfigOf(modelContent.value), true)
    selectionNodes.value = res.data?.selectionNodes || []
  } catch (e: any) {
    ElMessage.error(e?.msg || e?.message || '加载发起表单失败')
    emit('update:modelValue', false)
  } finally {
    formLoading.value = false
  }
}

watch(
  () => props.modelValue,
  (v) => {
    if (v && props.row) loadForm()
  },
)

async function submit() {
  if (!props.row) return
  const picks = previewRef.value?.getSelectionPicks?.() || { nodeAssignees: {}, nodeCcUsers: {} }
  for (const node of selectionNodes.value) {
    const bag = node.kind === 'cc' ? picks.nodeCcUsers : picks.nodeAssignees
    if (node.required && !(bag[node.nodeKey]?.length)) {
      ElMessage.warning(`请为「${node.nodeName}」选择${node.kind === 'cc' ? '抄送人' : '审批人'}`)
      return
    }
  }
  if (formRenderType.value === 'vue') {
    const ok = await customFormRef.value?.validate?.()
    if (!ok) {
      ElMessage.warning('请完善表单必填项')
      return
    }
  } else if (builderRef.value) {
    try {
      await (builderRef.value as any).validate?.()
    } catch {
      ElMessage.warning('请完善表单必填项')
      return
    }
  }
  submitting.value = true
  try {
    const formData =
      formRenderType.value === 'vue'
        ? customFormRef.value?.getData?.() || {}
        : await readBuilderFormData(builderRef.value)
    await launchDemoPurchaseOrder(props.row.id, {
      formData,
      nodeAssignees: picks.nodeAssignees,
      nodeCcUsers: picks.nodeCcUsers,
    })
    ElMessage.success('发起成功')
    emit('update:modelValue', false)
    emit('success')
  } catch (e: any) {
    ElMessage.error(e?.msg || e?.message || '发起失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.launch-drawer-body {
  min-height: 200px;
}
</style>

<style lang="scss">
.order-launch-drawer.el-drawer {
  .el-drawer__header {
    margin-bottom: 8px;
    padding-bottom: 12px;
  }
  .el-drawer__body {
    padding-top: 0;
  }
}
</style>
