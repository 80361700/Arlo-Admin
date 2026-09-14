<template>
  <el-drawer
    :model-value="modelValue"
    :title="title"
    size="720px"
    destroy-on-close
    class="biz-instance-detail-drawer"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="drawer-body">
      <section class="biz-section">
        <slot name="business">
          <el-empty description="暂无业务详情" :image-size="64" />
        </slot>
      </section>

      <template v-if="instanceId">
        <div v-loading="loading" class="flow-block">
          <el-tabs v-model="flowTab" class="flow-tabs">
            <el-tab-pane label="审批信息" name="form">
              <div class="pane-body">
                <ApproveFormBlock
                  :form-ready="formReady"
                  :loading="loading"
                  :form-render-type="formRenderType"
                  :form-component="formComponent"
                  :form-data="formData"
                  :page-schema="pageSchema"
                  :field-states="fieldStates"
                  :render-key="formRenderKey"
                  :readonly="true"
                  :disabled="true"
                />
              </div>
            </el-tab-pane>

            <el-tab-pane label="流转记录" name="timeline" lazy>
              <div class="pane-body">
                <FlowTimeline
                  :items="detail?.timeline || []"
                  :comments="detail?.comments || []"
                />
              </div>
            </el-tab-pane>

            <el-tab-pane label="流程图" name="diagram" lazy>
              <div class="diagram-wrap">
                <div class="diagram-legend">
                  <span><i class="dot done" />已执行</span>
                  <span><i class="dot active" />执行中</span>
                  <span><i class="dot pending" />未执行</span>
                </div>
                <div class="diagram-scroll">
                  <FlowProcess
                    v-if="flowModel"
                    :data="flowModel"
                    :flow="{}"
                    readonly
                    :node-states="nodeRunStates"
                    :node-actors="nodeRunActors"
                  />
                  <el-empty v-else description="暂无流程图" :image-size="64" />
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </template>

      <div v-else-if="canPreview" v-loading="previewLoading" class="preview-block">
        <LaunchFlowPreview v-if="previewModel" :model-content="previewModel" readonly />
        <el-empty v-else-if="!previewLoading" description="暂无流程预览" :image-size="64" />
      </div>

      <el-alert
        v-else
        type="info"
        :closable="false"
        show-icon
        title="尚未发起审批，仅展示业务详情"
        style="margin-top: 12px"
      />
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { type FieldStates, type PageSchema } from 'epic-designer'
import FlowProcess from '@/components/flowProcess/index.vue'
import FlowTimeline from '@/views/flow/approve/components/FlowTimeline.vue'
import LaunchFlowPreview from '@/views/flow/approve/components/LaunchFlowPreview.vue'
import ApproveFormBlock from '@/views/flow/approve/components/ApproveFormBlock.vue'
import { getApproveInstance, type InstanceDetail } from '@/api'
import {
  buildFieldStates,
  normalizeProcessForm,
} from '@/views/flow/process/components/formSchema'
import { buildNodeRunStates } from '@/views/flow/approve/utils/nodeRunStates'
import { showRequestError } from '@/utils/requestError'

export type ProcessPreviewPayload = {
  modelContent?: Record<string, any> | null
}

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    instanceId?: number | null
    /** 未发起时：业务单据 ID，配合 loadProcessPreview */
    bizId?: number | null
    /** 业务模块提供的流程预览（避免走流程管理权限） */
    loadProcessPreview?: (bizId: number) => Promise<ProcessPreviewPayload>
    title?: string
  }>(),
  {
    instanceId: 0,
    bizId: 0,
    title: '详情',
  },
)

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
}>()

const flowTab = ref('form')
const loading = ref(false)
const detail = ref<InstanceDetail | null>(null)
const pageSchema = ref<PageSchema | null>(null)
const formData = ref<Record<string, any>>({})
const fieldStates = ref<FieldStates>([])
const formReady = ref(false)
const formRenderKey = ref(0)
const formRenderType = ref('designer')
const formComponent = ref('')

const previewLoading = ref(false)
const previewModel = ref<Record<string, any> | null>(null)

const canPreview = computed(
  () => !!(Number(props.bizId) > 0 && typeof props.loadProcessPreview === 'function'),
)

const flowModel = computed(() => {
  const mc = detail.value?.modelContent
  if (!mc || typeof mc !== 'object') return null
  if (mc.nodeConfig) return mc
  return { nodeConfig: mc }
})

const nodeRunStates = computed(() => buildNodeRunStates(detail.value))

const nodeRunActors = computed(() => {
  const map: Record<string, { name: string; actorType?: number }[]> = {}
  const d = detail.value
  if (!d) return map
  const push = (key: string, name: string, actorType?: number) => {
    if (!key || !name) return
    if (!map[key]) map[key] = []
    if (map[key].some((x) => x.name === name)) return
    map[key].push({ name, actorType })
  }
  for (const t of d.timeline || []) {
    if (Number(t.nodeType) === 0 || Number(t.nodeType) === -1) continue
    if (d.currentNodeKey && t.nodeKey === d.currentNodeKey) continue
    push(t.nodeKey, t.actorName)
  }
  if (d.currentNodeKey) {
    for (const a of d.taskActors || []) {
      if (Number(a.actorState) === 4) continue
      push(d.currentNodeKey, a.actorName, a.actorType)
    }
  }
  return map
})

function resetFlow() {
  detail.value = null
  formReady.value = false
  pageSchema.value = null
  formData.value = {}
  fieldStates.value = []
  formRenderType.value = 'designer'
  formComponent.value = ''
}

function resetPreview() {
  previewModel.value = null
}

async function loadInstance() {
  const id = Number(props.instanceId) || 0
  if (!id) {
    resetFlow()
    return
  }
  loading.value = true
  formReady.value = false
  try {
    const res = await getApproveInstance(id)
    detail.value = res.data || null
    if (detail.value) {
      formData.value = { ...(detail.value.formData || {}) }
      formRenderType.value = detail.value.formRenderType === 'vue' ? 'vue' : 'designer'
      formComponent.value = detail.value.formComponent || ''
      pageSchema.value = normalizeProcessForm(detail.value.formSchema)
      fieldStates.value = buildFieldStates(detail.value.formConfig || [], false, pageSchema.value)
      formRenderKey.value += 1
      formReady.value = true
    }
  } catch (e) {
    resetFlow()
    showRequestError(e)
  } finally {
    loading.value = false
  }
}

async function loadPreview() {
  const id = Number(props.bizId) || 0
  if (!id || !props.loadProcessPreview) {
    resetPreview()
    return
  }
  previewLoading.value = true
  try {
    const data = await props.loadProcessPreview(id)
    const mc = data?.modelContent
    previewModel.value =
      mc && typeof mc === 'object' && Object.keys(mc).length ? mc : null
  } catch (e) {
    resetPreview()
    showRequestError(e)
  } finally {
    previewLoading.value = false
  }
}

function onOpen() {
  flowTab.value = 'form'
  if (Number(props.instanceId) > 0) {
    resetPreview()
    loadInstance()
  } else {
    resetFlow()
    loadPreview()
  }
}

// 合并监听，避免 open + id 双触发导致重复请求/双 toast
watch(
  () => [props.modelValue, Number(props.instanceId) || 0, Number(props.bizId) || 0] as const,
  ([open]) => {
    if (!open) return
    onOpen()
  },
)
</script>

<style scoped lang="scss">
.drawer-body {
  height: 100%;
  overflow: auto;
  padding-bottom: 8px;
}
.biz-section {
  margin-bottom: 12px;
}
.flow-block {
  min-height: 200px;
}
.preview-block {
  min-height: 120px;
  padding-top: 4px;
}
.flow-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 12px;
  }
}
.pane-body {
  padding: 4px 0 8px;
}
.diagram-wrap {
  position: relative;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  background: #f6f8f9;
  overflow: visible;
}
.diagram-legend {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 5;
  display: flex;
  gap: 16px;
  padding: 6px 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  pointer-events: none;
  .dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 6px;
    &.done {
      background: #67c23a;
    }
    &.active {
      background: #e6a23c;
    }
    &.pending {
      background: #c0c4cc;
    }
  }
}
.diagram-scroll {
  overflow: visible;
  :deep(.flow-process-root) {
    height: auto;
    min-height: 0;
    flex: none;
  }
  :deep(.create-approval-main) {
    height: auto;
    min-height: 0;
    overflow: visible;
  }
  :deep(.canvas-scroll) {
    height: auto;
    overflow: visible;
  }
  :deep(.sc-workflow-design),
  :deep(.box-scale) {
    min-height: 0;
  }
}
</style>

<style lang="scss">
.biz-instance-detail-drawer.el-drawer {
  .el-drawer__header {
    margin-bottom: 8px;
    padding-bottom: 12px;
  }
  .el-drawer__body {
    padding-top: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
}
</style>
