<template>
  <div class="flow-process-root" :class="{ 'is-readonly': readonly }">
    <nodeView :data="props.data" @onEvent="nodeEvent" />
    <node-setting v-if="!readonly" ref="settingRef" :flow="props.flow" @success="settingSuccess" />
  </div>
</template>

<script setup lang="ts">
import { computed, provide, ref } from 'vue'
import nodeView from './components/node/index.vue'
import nodeSetting from './components/setting/index.vue'

export type FlowNodeRunState = 'done' | 'active' | 'pending'

export type FlowNodeActor = {
  name: string
  /** 1=加签 */
  actorType?: number
}

const props = withDefaults(
  defineProps<{
    data: any
    flow: any
    readonly?: boolean
    /** 运行态着色：nodeKey -> done|active|pending */
    nodeStates?: Record<string, FlowNodeRunState>
    /** 运行态办理人（含加签），覆盖设计态展示 */
    nodeActors?: Record<string, FlowNodeActor[]>
  }>(),
  { readonly: false, nodeStates: () => ({}), nodeActors: () => ({}) },
)

provide(
  'flowReadonly',
  computed(() => !!props.readonly),
)
provide(
  'flowNodeStates',
  computed(() => props.nodeStates || {}),
)
provide(
  'flowNodeActors',
  computed(() => props.nodeActors || {}),
)

const settingRef = ref()
const nodeEvent = (e: any, parent: any) => {
  if (props.readonly) return
  settingRef.value?.init(e, parent)
}
const settingSuccess = (_e: any) => {}

defineExpose({})
</script>

<style lang="scss" scoped>
.flow-process-root {
  flex: 1;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>
