<template>
  <div class="rich-content" v-html="normalizedHtml" />
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  html?: string | null
}>()

const normalizedHtml = computed(() => {
  const raw = props.html ?? ''
  // 纯文本换行兼容；已是 HTML 时不影响标签结构
  if (!/[<>]/.test(raw)) return raw.replace(/\n/g, '<br>')
  return raw
})
</script>

<style scoped lang="scss">
.rich-content {
  line-height: 1.8;
  font-size: 14px;
  color: #303133;
  word-break: break-word;

  :deep(p) {
    margin: 10px 0;
  }

  :deep(h1),
  :deep(h2),
  :deep(h3),
  :deep(h4),
  :deep(h5) {
    margin: 16px 0 10px;
    font-weight: 600;
    line-height: 1.4;
  }

  :deep(img) {
    max-width: 100%;
    height: auto;
  }

  :deep(video) {
    max-width: 100%;
  }

  :deep(blockquote) {
    margin: 12px 0;
    padding: 8px 12px;
    border-left: 4px solid #dcdfe6;
    background: #f5f7fa;
    color: #606266;
  }

  :deep(ul),
  :deep(ol) {
    padding-left: 1.5em;
    margin: 10px 0;
  }

  :deep(code) {
    padding: 2px 6px;
    border-radius: 3px;
    background: #f5f7fa;
    font-family: monospace;
  }

  :deep(pre) {
    margin: 12px 0;
    padding: 12px;
    overflow-x: auto;
    border-radius: 4px;
    background: #f5f7fa;
  }

  /* 对齐 wangEditor 表格样式（编辑器 CSS 仅作用于 .w-e-text-container） */
  :deep(.table-container) {
    width: 100%;
    margin: 12px 0;
    padding: 10px;
    overflow-x: auto;
    border: 1px dashed #ccc;
    border-radius: 5px;
  }

  :deep(table) {
    width: 100%;
    border-collapse: collapse;
  }

  :deep(td),
  :deep(th) {
    min-width: 30px;
    padding: 6px 10px;
    border: 1px solid #ccc;
    line-height: 1.5;
    text-align: left;
  }

  :deep(th) {
    background-color: #f5f2f0;
    font-weight: 700;
    text-align: center;
  }
}
</style>
