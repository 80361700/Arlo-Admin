<template>
  <div class="arlo-rich-editor" :class="{ 'is-disabled': disabled }">
    <RichEditor
      v-model="model"
      :placeholder="placeholder"
      :height="height"
    />
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import RichEditor from '@/components/RichEditor.vue'

withDefaults(
  defineProps<{
    placeholder?: string
    height?: number
    disabled?: boolean
  }>(),
  {
    placeholder: '请输入内容...',
    height: 280,
  },
)

const model = defineModel<string>({ default: '' })
const emit = defineEmits<{ change: [value: string] }>()

watch(model, (v) => emit('change', v ?? ''))
</script>

<style scoped>
.arlo-rich-editor {
  width: 100%;
}
.arlo-rich-editor.is-disabled {
  pointer-events: none;
  opacity: 0.7;
}
</style>
