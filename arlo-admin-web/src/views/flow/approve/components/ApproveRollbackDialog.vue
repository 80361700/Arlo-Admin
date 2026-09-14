<template>
  <el-dialog
    :model-value="modelValue"
    title="回退"
    width="480px"
    destroy-on-close
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-position="top" @submit.prevent>
      <el-form-item required label="回退到">
        <el-select v-model="targetNodeKey" placeholder="选择已走过的节点" style="width: 100%">
          <el-option
            v-for="t in targets"
            :key="t.nodeKey"
            :label="t.nodeName"
            :value="t.nodeKey"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="回退意见">
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
      <el-button type="primary" :loading="loading" @click="onConfirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    loading?: boolean
    targets?: { nodeKey: string; nodeName: string }[]
  }>(),
  {
    loading: false,
    targets: () => [],
  },
)

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  confirm: [payload: { targetNodeKey: string; opinion: string }]
}>()

const targetNodeKey = ref('')
const opinion = ref('')

watch(
  () => props.modelValue,
  (v) => {
    if (!v) return
    opinion.value = ''
    targetNodeKey.value = props.targets?.[0]?.nodeKey || ''
  },
)

function onConfirm() {
  if (!targetNodeKey.value) {
    ElMessage.warning('请选择回退目标节点')
    return
  }
  emit('confirm', {
    targetNodeKey: targetNodeKey.value,
    opinion: opinion.value.trim(),
  })
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
