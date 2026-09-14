<template>
  <div class="approve-form-block">
    <CustomFlowFormHost
      v-if="showVueForm"
      :key="`${renderKey}-vue`"
      ref="customFormRef"
      :component-path="formComponent"
      :form-data="formData"
      :readonly="readonly"
    />
    <div v-if="showVueForm && showSchemaForm" class="subform-divider">节点子表单</div>
    <EBuilder
      v-if="showSchemaForm"
      :key="`${renderKey}-schema`"
      ref="builderRef"
      :page-schema="pageSchema!"
      :form-data="formData"
      :field-states="fieldStates"
      :disabled="disabled"
      @ready="onBuilderReady"
    />
    <el-empty
      v-else-if="showEmpty && !loading && formReady && !hasAnyForm"
      description="无表单数据"
      :image-size="64"
    />
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { EBuilder, type FieldStates, type PageSchema } from 'epic-designer'
import CustomFlowFormHost from './CustomFlowFormHost.vue'
import {
  pageSchemaHasFields,
  readBuilderFormData,
  writeBuilderFormData,
} from '@/views/flow/process/components/formSchema'

const props = withDefaults(
  defineProps<{
    formReady?: boolean
    loading?: boolean
    formRenderType?: string
    formComponent?: string
    formData?: Record<string, any>
    pageSchema?: PageSchema | null
    fieldStates?: FieldStates
    renderKey?: number | string
    /** 系统表单只读 */
    readonly?: boolean
    /** 设计表单（含节点子表）禁用 */
    disabled?: boolean
    showEmpty?: boolean
  }>(),
  {
    formReady: false,
    loading: false,
    formRenderType: 'designer',
    formComponent: '',
    formData: () => ({}),
    pageSchema: null,
    fieldStates: () => [],
    renderKey: 0,
    readonly: true,
    disabled: true,
    showEmpty: true,
  },
)

const customFormRef = ref<InstanceType<typeof CustomFlowFormHost>>()
const builderRef = ref<InstanceType<typeof EBuilder>>()

const showVueForm = computed(
  () =>
    !!(props.formReady && props.formRenderType === 'vue' && String(props.formComponent || '').trim()),
)
const showSchemaForm = computed(
  () => !!(props.formReady && pageSchemaHasFields(props.pageSchema)),
)
const hasAnyForm = computed(() => showVueForm.value || showSchemaForm.value)

function onBuilderReady() {
  writeBuilderFormData(builderRef.value, props.formData || {})
}

function getCustomForm() {
  return customFormRef.value
}

function getBuilder() {
  return builderRef.value
}

async function collectData(base: Record<string, any> = {}) {
  const data: Record<string, any> = { ...base }
  if (props.formRenderType === 'vue') {
    Object.assign(data, customFormRef.value?.getData?.() || {})
  }
  if (showSchemaForm.value && builderRef.value) {
    Object.assign(data, await readBuilderFormData(builderRef.value))
  }
  return data
}

async function validate(opts?: { vue?: boolean; schema?: boolean }) {
  const vue = opts?.vue === true
  const schema = opts?.schema === true
  if (vue && customFormRef.value) {
    const ok = !!(await customFormRef.value.validate?.())
    if (!ok) return { ok: false as const, message: '请完善表单必填项' }
  }
  if (schema && builderRef.value && showSchemaForm.value) {
    try {
      await (builderRef.value as any).validate?.()
    } catch {
      return { ok: false as const, message: '请完善节点子表单必填项' }
    }
  }
  return { ok: true as const }
}

defineExpose({
  customFormRef,
  builderRef,
  getCustomForm,
  getBuilder,
  collectData,
  validate,
})
</script>

<style scoped lang="scss">
.approve-form-block {
  .subform-divider {
    margin: 16px 0 12px;
    padding-top: 12px;
    border-top: 1px dashed var(--el-border-color);
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }
}
</style>
