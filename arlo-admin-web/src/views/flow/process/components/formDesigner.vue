<template>
  <div class="form-designer">
    <EDesigner
      ref="designerRef"
      form-mode
      hidden-header
      disabled-zoom
      :draggable="false"
      title="表单设计"
      :default-schema="defaultSchema"
      @ready="onReady"
    />
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { EDesigner, type PageSchema } from 'epic-designer'
import { createEmptyFormSchema, normalizeProcessForm } from './formSchema'

function pickSchema(modelValue: any): unknown {
  if (!modelValue) return undefined
  // 流程向导：{ flow: { processForm } }
  if (modelValue.flow && 'processForm' in modelValue.flow) {
    return modelValue.flow.processForm
  }
  // 表单模板：直接传 PageSchema
  return modelValue
}

const props = defineProps<{ modelValue: any }>()

const designerRef = ref<InstanceType<typeof EDesigner>>()
const ready = ref(false)
const defaultSchema = createEmptyFormSchema()
let pendingSchema: PageSchema | null = null

function applySchema(schema: PageSchema) {
  const pageSchema = normalizeProcessForm(schema)
  try {
    designerRef.value?.setData?.(pageSchema)
  } catch (err) {
    console.warn('[formDesigner] setData failed', err)
  }
}

function onReady() {
  ready.value = true
  const schema = pendingSchema ?? normalizeProcessForm(pickSchema(props.modelValue))
  pendingSchema = null
  nextTick(() => applySchema(schema))
}

function getJson(): PageSchema {
  if (!ready.value) {
    return normalizeProcessForm(pickSchema(props.modelValue))
  }
  try {
    return normalizeProcessForm(designerRef.value?.getData?.())
  } catch {
    return createEmptyFormSchema()
  }
}

function setJson(schema?: PageSchema) {
  const pageSchema = normalizeProcessForm(schema ?? pickSchema(props.modelValue))
  if (!ready.value) {
    pendingSchema = pageSchema
    return
  }
  nextTick(() => applySchema(pageSchema))
}

defineExpose({ getJson, setJson })
</script>

<style lang="scss" scoped>
.form-designer {
  flex: 1;
  min-height: 480px;
  height: 100%;
  overflow: hidden;
  // border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;

  :deep(.ep-designer-main) {
    height: 100%;
  }

  /* 流程向导自带保存；隐藏 epic 工具栏里的保存按钮 */
  :deep(.ep-edit-toolbar .ep-action-item:has(.icon--epic--save-outline-rounded)) {
    display: none !important;
  }

  /* 表单步不需要画布横向平移，去掉底部无用的横滚条 */
  :deep(.ep-edit-screen-container) {
    overflow-x: hidden !important;
  }
}
</style>
