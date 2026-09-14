<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="720px"
    top="6vh"
    destroy-on-close
    append-to-body
    class="sub-process-dialog"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div v-loading="loading" class="sub-body">
      <!-- 子流程通常共用主表单且只读，审批信息与主流程重复，只保留流转与流程图 -->
      <el-tabs v-model="tab">
        <el-tab-pane label="流转记录" name="timeline">
          <div class="sub-scroll">
            <FlowTimeline
              :items="detail?.timeline || []"
              :comments="detail?.comments || []"
              :image-size="48"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane label="流程图" name="diagram" lazy>
          <div class="sub-diagram-wrap">
            <div class="diagram-legend">
              <span><i class="dot done" />已执行</span>
              <span><i class="dot active" />执行中</span>
              <span><i class="dot pending" />未执行</span>
            </div>
            <div class="sub-diagram">
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
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { getApproveInstance, type InstanceDetail } from '@/api'
import FlowProcess from '@/components/flowProcess/index.vue'
import type { FlowNodeActor } from '@/components/flowProcess/index.vue'
import FlowTimeline from './FlowTimeline.vue'
import { buildNodeRunStates } from '../utils/nodeRunStates'

const props = defineProps<{
  modelValue: boolean
  instanceId?: number
  processName?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [boolean] }>()

const loading = ref(false)
const tab = ref('timeline')
const detail = ref<InstanceDetail | null>(null)

const title = computed(() => props.processName || detail.value?.subProcessName || detail.value?.processName || '子流程')

const flowModel = computed(() => {
  const mc = detail.value?.modelContent as any
  return mc?.nodeConfig ? mc : null
})

/** 流程图节点运行态：已执行 / 执行中 / 未执行 */
const nodeRunStates = computed(() => buildNodeRunStates(detail.value))

const nodeRunActors = computed(() => {
  const map: Record<string, FlowNodeActor[]> = {}
  const d = detail.value
  if (!d) return map

  const push = (key: string, name: string, actorType?: number) => {
    if (!key || !name || name === '系统') return
    if (!map[key]) map[key] = []
    if (map[key].some((x) => x.name === name)) return
    map[key].push({ name, actorType })
  }

  for (const t of d.timeline || []) {
    if (!t.nodeKey || !t.actorName) continue
    push(t.nodeKey, t.actorName)
  }

  for (const a of d.taskActors || []) {
    if (d.currentNodeKey && a.actorName) {
      push(d.currentNodeKey, a.actorName, a.actorType)
    }
  }

  return map
})

watch(
  () => [props.modelValue, props.instanceId] as const,
  async ([open, id]) => {
    if (!open || !id) return
    tab.value = 'timeline'
    loading.value = true
    detail.value = null
    try {
      const res = await getApproveInstance(id)
      detail.value = res.data || null
    } finally {
      loading.value = false
    }
  },
)
</script>

<style scoped lang="scss">
.sub-body {
  min-height: 360px;
}
.sub-scroll {
  max-height: 62vh;
  overflow: auto;
  padding-right: 4px;
}
.sub-diagram-wrap {
  position: relative;
  min-height: 360px;
  height: 62vh;
  max-height: 62vh;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  overflow: hidden;
}
.sub-diagram {
  height: 100%;
  overflow: hidden;
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
</style>
