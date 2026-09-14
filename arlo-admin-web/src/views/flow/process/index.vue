<template>
  <div class="page-container flow-process-page">
    <div v-show="!designing" class="flow-process-row">
      <aside class="flow-process-left">
        <group ref="groupRef" @event="groupEvent" />
      </aside>
      <section class="flow-process-right">
        <list ref="listRef" @event="listEvent" />
      </section>
    </div>
    <flow-dialog ref="flowRef" @update:open="designing = $event" @success="flowSuccess" />
    <history-dialog ref="historyRef" @preview="onHistoryPreview" @checkout="onHistoryCheckout" />
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import Group from './components/group.vue'
import List from './components/list.vue'
import HistoryDialog from './components/history.vue'
import FlowDialog from './flow.vue'
import type { FlowProcessHistoryBrief } from '@/api'

const listRef = ref<InstanceType<typeof List>>()
const groupRef = ref<InstanceType<typeof Group>>()
const flowRef = ref<InstanceType<typeof FlowDialog>>()
const historyRef = ref<InstanceType<typeof HistoryDialog>>()
const designing = ref(false)

function groupEvent(row: any) {
  listRef.value?.reload(row)
}

function listEvent(data: any) {
  if (data.type === 'edit') {
    flowRef.value?.init(data.row)
  } else if (data.type === 'history') {
    historyRef.value?.init(data.row)
  } else if (data.type === 'refresh') {
    groupRef.value?.initList(data.row)
  }
}

function onHistoryPreview(row: FlowProcessHistoryBrief) {
  flowRef.value?.init({
    processId: row.processId,
    historyId: row.historyId,
    preview: true,
  })
}

function onHistoryCheckout() {
  groupRef.value?.initList()
}

function flowSuccess() {
  designing.value = false
  groupRef.value?.initList()
}
</script>

<style lang="scss" scoped>
.flow-process-page {
  position: relative;
  height: 100%;
  min-height: 520px;
  background: var(--el-bg-color);
}
.flow-process-row {
  display: flex;
  align-items: stretch;
  height: 100%;
  min-height: 520px;
  gap: 16px;
}
.flow-process-left {
  width: 380px;
  flex: 0 0 380px;
  min-width: 380px;
  max-width: 380px;
  height: 100%;
}
.flow-process-right {
  flex: 1;
  min-width: 0;
  height: 100%;
}
</style>
