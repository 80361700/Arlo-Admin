<template>
  <el-dialog
    v-model="visible"
    :title="title"
    :width="width"
    :close-on-click-modal="false"
    destroy-on-close
    @closed="handleClosed"
  >
    <el-form
      ref="formRef"
      :model="model"
      :rules="rules"
      :label-width="labelWidth"
      @submit.prevent="handleSubmit"
    >
      <slot />
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ submitText }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

const props = withDefaults(defineProps<{
  modelValue: boolean
  model?: Record<string, any>
  title?: string
  width?: string
  labelWidth?: string
  rules?: FormRules
  submitText?: string
  submitting?: boolean
}>(), {
  title: '表单',
  width: '560px',
  labelWidth: '100px',
  model: () => ({}),
  rules: () => ({}),
  submitText: '确 定',
  submitting: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: []
}>()

const visible = ref(props.modelValue)
const formRef = ref<FormInstance>()

watch(() => props.modelValue, (val) => {
  visible.value = val
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  emit('submit')
}

function handleClosed() {
  formRef.value?.resetFields()
}

defineExpose({
  formRef,
  close() {
    visible.value = false
  },
})
</script>

<style scoped lang="scss">
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;

  :deep(.el-button + .el-button) {
    margin-left: 0;
  }
}
</style>
