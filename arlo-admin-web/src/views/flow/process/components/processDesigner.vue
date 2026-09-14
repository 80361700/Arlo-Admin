<template>
  <div class="process-designer-page" :class="{ 'is-readonly': readonly }">
    <process-designer :data="modelValue.flow.modelContent" :flow="modelValue.flow" :readonly="readonly" />
  </div>
</template>

<script setup lang="ts">
import processDesigner from '@/components/flowProcess/index.vue'

withDefaults(
  defineProps<{
    modelValue: any
    readonly?: boolean
  }>(),
  { readonly: false },
)
</script>

<style lang="scss" scoped>
.process-designer-page {
  flex: 1;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;

  &.is-readonly {
    pointer-events: none;
  }
}
</style>
