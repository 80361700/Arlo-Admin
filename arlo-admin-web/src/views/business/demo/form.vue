<template>
  <el-form
    ref="formRef"
    :model="form"
    :rules="rules"
    label-width="100px"
    :disabled="readonly"
    class="biz-demo-form"
  >
    <el-alert
      type="info"
      :closable="false"
      show-icon
      title="示例系统表单（business/demo/form）。业务页按此约定实现 validate / getData / setData。"
      style="margin-bottom: 16px"
    />
    <el-form-item label="标题" prop="title">
      <el-input v-model="form.title" maxlength="64" placeholder="请输入标题" />
    </el-form-item>
    <el-form-item label="金额" prop="amount">
      <el-input-number v-model="form.amount" :min="0" :precision="2" style="width: 100%" />
    </el-form-item>
    <el-form-item label="说明" prop="remark">
      <el-input v-model="form.remark" type="textarea" :rows="3" maxlength="200" show-word-limit />
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

const props = withDefaults(
  defineProps<{
    formData?: Record<string, any>
    readonly?: boolean
  }>(),
  { formData: () => ({}), readonly: false },
)

const formRef = ref<FormInstance>()
const form = reactive({
  title: '',
  amount: 0 as number,
  remark: '',
})

const rules: FormRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
}

function apply(data?: Record<string, any>) {
  const d = data || {}
  form.title = String(d.title ?? '')
  form.amount = Number(d.amount) || 0
  form.remark = String(d.remark ?? '')
}

watch(
  () => props.formData,
  (v) => apply(v),
  { immediate: true, deep: true },
)

async function validate() {
  try {
    await formRef.value?.validate()
    return true
  } catch {
    return false
  }
}

function getData() {
  return {
    title: form.title,
    amount: form.amount,
    remark: form.remark,
  }
}

function setData(data: Record<string, any>) {
  apply(data)
}

defineExpose({ validate, getData, setData })
</script>
