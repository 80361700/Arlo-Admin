<template>
  <el-dialog
    :model-value="modelValue"
    title="撤销申请"
    width="480px"
    destroy-on-close
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-position="top" @submit.prevent>
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        title="撤销后流程将终止，且不可继续审批。"
        style="margin-bottom: 16px"
      />
      <el-form-item label="撤销说明">
        <el-input
          v-model="opinion"
          type="textarea"
          :rows="3"
          maxlength="64"
          show-word-limit
          placeholder="选填"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="danger" :loading="loading" @click="onConfirm">确定撤销</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{ modelValue: boolean; loading?: boolean }>(),
  { loading: false },
)

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  confirm: [opinion: string]
}>()

const opinion = ref('')

watch(
  () => props.modelValue,
  (v) => {
    if (v) opinion.value = ''
  },
)

function onConfirm() {
  emit('confirm', opinion.value.trim())
}
</script>

<style scoped lang="scss">
:deep(.el-dialog__footer) {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  .el-button + .el-button {
    margin-left: 0;
  }
}
</style>
