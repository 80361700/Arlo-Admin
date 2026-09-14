<template>
  <el-dialog
    :model-value="modelValue"
    title="拒绝审批"
    width="520px"
    destroy-on-close
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-position="top" @submit.prevent>
      <el-form-item required label="审批意见">
        <el-input
          v-model="opinion"
          type="textarea"
          :rows="4"
          maxlength="64"
          show-word-limit
          placeholder="请输入内容"
        />
      </el-form-item>

      <el-form-item v-if="needRejectTarget && !terminate && rejectTargets.length" label="驳回到">
        <el-select v-model="rejectNodeKey" placeholder="选择已走过的节点" style="width: 100%">
          <el-option
            v-for="t in rejectTargets"
            :key="t.nodeKey"
            :label="t.nodeName"
            :value="t.nodeKey"
          />
        </el-select>
      </el-form-item>

      <el-checkbox v-model="terminate">终止流程</el-checkbox>
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
    needRejectTarget?: boolean
    rejectTargets?: { nodeKey: string; nodeName: string }[]
    defaultRejectNodeKey?: string
  }>(),
  {
    loading: false,
    needRejectTarget: false,
    rejectTargets: () => [],
    defaultRejectNodeKey: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  confirm: [payload: { opinion: string; rejectStrategy: number; rejectNodeKey?: string }]
}>()

const opinion = ref('')
const terminate = ref(false)
const rejectNodeKey = ref('')

watch(
  () => props.modelValue,
  (v) => {
    if (v) {
      opinion.value = ''
      terminate.value = false
      rejectNodeKey.value = props.defaultRejectNodeKey || ''
    }
  },
)

function onConfirm() {
  const text = opinion.value.trim()
  if (!text) {
    ElMessage.warning('请填写审批意见')
    return
  }
  if (!terminate.value && props.needRejectTarget && !rejectNodeKey.value) {
    ElMessage.warning('请选择驳回目标节点')
    return
  }
  emit('confirm', {
    opinion: text,
    rejectStrategy: terminate.value ? 4 : 0,
    rejectNodeKey: terminate.value
      ? undefined
      : rejectNodeKey.value || props.defaultRejectNodeKey || undefined,
  })
}
</script>

<style scoped lang="scss">
:deep(.el-dialog__footer) {
  .el-button + .el-button {
    margin-left: 0;
  }
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
