<template>
  <div class="avatar" :style="wrapStyle">
    <el-avatar :size="size" class="icon">{{ name?.charAt(0) || '-' }}</el-avatar>
    <div v-if="showName" class="name" :title="name">{{ name }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    name?: string
    size?: number
    showName?: boolean
  }>(),
  { name: '', size: 24, showName: true },
)

const wrapStyle = computed(() => {
  const padding = props.showName ? 4 : 0
  return {
    height: `${props.size + padding * 2}px`,
    borderRadius: `${(props.size + padding * 2) / 2}px`,
    padding: `0 ${padding}px`,
  }
})
</script>

<style lang="scss" scoped>
.avatar {
  background: var(--el-fill-color-light);
  display: inline-flex;
  align-items: center;
  width: fit-content;

  .icon {
    flex-shrink: 0;
    background: var(--el-color-primary);
  }

  .name {
    user-select: none;
    color: var(--el-text-color-regular);
    margin: 0 4px;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    max-width: 72px;
    font-size: 12px;
  }
}
</style>
