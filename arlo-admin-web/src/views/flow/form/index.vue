<template>
  <div class="page-container flow-form-page">
    <div v-show="!designing" class="flow-form-row">
      <aside class="flow-form-left">
        <group ref="groupRef" @event="groupEvent" />
      </aside>
      <section class="flow-form-right">
        <list ref="listRef" @event="listEvent" />
      </section>
    </div>
    <design-panel ref="designRef" @update:open="designing = $event" @success="designSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import Group from './components/group.vue'
import List from './components/list.vue'
import DesignPanel from './design.vue'

const listRef = ref<InstanceType<typeof List>>()
const groupRef = ref<InstanceType<typeof Group>>()
const designRef = ref<InstanceType<typeof DesignPanel>>()
const designing = ref(false)

function groupEvent(row: any) {
  listRef.value?.reload(row)
}

function listEvent(data: any) {
  if (data.type === 'design') {
    designRef.value?.init(data.row)
  } else if (data.type === 'refresh') {
    groupRef.value?.initList(data.row)
  }
}

function designSuccess() {
  designing.value = false
  groupRef.value?.initList()
}
</script>

<style lang="scss" scoped>
.flow-form-page {
  position: relative;
  height: 100%;
  min-height: 520px;
  background: var(--el-bg-color);
}
.flow-form-row {
  display: flex;
  align-items: stretch;
  height: 100%;
  min-height: 520px;
  gap: 16px;
}
.flow-form-left {
  width: 380px;
  flex: 0 0 380px;
  min-width: 380px;
  max-width: 380px;
  height: 100%;
}
.flow-form-right {
  flex: 1;
  min-width: 0;
  height: 100%;
}
</style>
