<template>
  <div class="icon-picker">
    <div class="icon-picker-trigger" @click="visible = true">
      <el-icon v-if="modelValue" :size="18">
        <component :is="modelValue" />
      </el-icon>
      <span v-else class="icon-picker-placeholder">{{ placeholder }}</span>
      <el-icon class="icon-picker-arrow"><ArrowDown /></el-icon>
    </div>

    <el-dialog
      v-model="visible"
      title="选择图标"
      width="720px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-input
        v-model="search"
        placeholder="搜索图标名称..."
        clearable
        class="icon-search"
      />
      <div class="icon-grid" v-if="filteredIcons.length">
        <div
          v-for="name in filteredIcons"
          :key="name"
          class="icon-cell"
          :class="{ active: name === modelValue }"
          :title="name"
          @click="select(name)"
        >
          <el-icon :size="22">
            <component :is="name" />
          </el-icon>
          <span class="icon-label">{{ name }}</span>
        </div>
      </div>
      <el-empty v-else description="没有匹配的图标" />
      <template #footer>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="visible = false">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, shallowRef } from 'vue'
import { ArrowDown } from '@element-plus/icons-vue'
import * as Icons from '@element-plus/icons-vue'

interface Props {
  modelValue: string
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '请选择图标',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const visible = ref(false)
const search = ref('')

// 所有图标名（排除内部工具函数）
const allIcons = shallowRef(
  Object.keys(Icons).filter(
    k => k !== 'default' && !k.startsWith('_') && k[0] === k[0].toUpperCase()
  )
)

const filteredIcons = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return allIcons.value
  return allIcons.value.filter(name => name.toLowerCase().includes(q))
})

function select(name: string) {
  emit('update:modelValue', name)
}
</script>

<style scoped lang="scss">
.icon-picker-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  height: 32px;
  padding: 0 12px;
  border: 1px solid var(--el-border-color);
  border-radius: var(--el-border-radius-base);
  cursor: pointer;
  box-sizing: border-box;
  transition: border-color 0.2s;

  &:hover {
    border-color: var(--el-color-primary);
  }

  .icon-picker-placeholder {
    color: var(--el-text-color-placeholder);
    font-size: 14px;
  }

  .icon-picker-arrow {
    margin-left: auto;
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
}

.icon-search {
  margin-bottom: 16px;
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 6px;
  max-height: 380px;
  overflow-y: auto;
}

.icon-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 8px 4px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;

  &:hover {
    border-color: var(--el-color-primary);
    color: var(--el-color-primary);
  }

  &.active {
    border-color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
  }

  .icon-label {
    font-size: 10px;
    text-align: center;
    line-height: 1.2;
    word-break: break-all;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 1;
    -webkit-box-orient: vertical;
  }
}
</style>
