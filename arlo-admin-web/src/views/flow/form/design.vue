<template>
  <div v-if="dialogShow" class="form-design-panel">
    <div class="dialog-header">
      <div class="dialog-header-left">
        <button type="button" class="back-btn" @click="close">
          <el-icon :size="16"><Back /></el-icon>
          <span>返回</span>
        </button>
        <span class="header-divider" />
        <span class="dialog-title">表单设计 · {{ title }}</span>
      </div>
      <div class="dialog-actions">
        <el-button type="primary" :loading="loading" @click="save">保存</el-button>
      </div>
    </div>
    <div v-loading="loading" class="designer-body">
      <form-designer v-if="schemaReady" ref="formRef" v-model="schema" class="designer-fill" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { nextTick, ref } from 'vue'
import { Back } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getFlowForm, saveFlowFormSchema } from '@/api'
import formDesigner from '../process/components/formDesigner.vue'
import { createEmptyFormSchema, normalizeProcessForm } from '../process/components/formSchema'

const emits = defineEmits<{ success: []; 'update:open': [v: boolean] }>()

const dialogShow = ref(false)
const loading = ref(false)
const schemaReady = ref(false)
const title = ref('')
const formId = ref<number>()
const schema = ref(createEmptyFormSchema())
const formRef = ref<InstanceType<typeof formDesigner>>()

async function init(row: any) {
  if (!row?.formId) return
  formId.value = row.formId
  title.value = row.name || ''
  dialogShow.value = true
  emits('update:open', true)
  loading.value = true
  schemaReady.value = false
  try {
    const res = await getFlowForm(row.formId)
    title.value = res.data?.name || title.value
    schema.value = normalizeProcessForm(res.data?.formSchema)
    schemaReady.value = true
    await nextTick()
    formRef.value?.setJson(schema.value)
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!formId.value) return
  loading.value = true
  try {
    const json = formRef.value?.getJson?.() ?? schema.value
    await saveFlowFormSchema(formId.value, json)
    ElMessage.success('保存成功')
    emits('success')
    close()
  } finally {
    loading.value = false
  }
}

function close() {
  dialogShow.value = false
  schemaReady.value = false
  emits('update:open', false)
}

defineExpose({ init })
</script>

<style lang="scss" scoped>
.form-design-panel {
  position: absolute;
  inset: 0;
  z-index: 20;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
}
.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.dialog-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}
.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: var(--el-text-color-regular);
  padding: 0;
}
.header-divider {
  width: 1px;
  height: 16px;
  background: var(--el-border-color);
}
.dialog-title {
  font-weight: 500;
  font-size: 15px;
}
.designer-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 16px;
}
.designer-fill {
  flex: 1;
  min-height: 0;
  height: 100%;
}
</style>
