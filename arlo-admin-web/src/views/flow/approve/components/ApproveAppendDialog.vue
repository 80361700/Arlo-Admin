<template>
  <el-dialog
    :model-value="modelValue"
    title="加签"
    width="480px"
    destroy-on-close
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-position="top" @submit.prevent>
      <el-form-item required label="加签人">
        <UserSelect v-model="toUserId" placeholder="请选择加签人" />
      </el-form-item>
      <el-form-item label="加签方式">
        <el-radio-group v-model="position">
          <el-radio :value="1">前加签</el-radio>
          <el-radio :value="2">后加签</el-radio>
        </el-radio-group>
        <div class="hint">前加签：插到当前审批人之前先处理；后加签：插到当前审批人之后再处理</div>
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
import UserSelect from '@/components/UserSelect.vue'
import { getAllUsers } from '@/api'

const props = withDefaults(
  defineProps<{ modelValue: boolean; loading?: boolean }>(),
  { loading: false },
)

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  confirm: [payload: { toUserId: number; toName: string; position: number }]
}>()

const toUserId = ref<number>()
const position = ref(1)
const nameMap = ref<Record<number, string>>({})

watch(
  () => props.modelValue,
  async (v) => {
    if (!v) return
    toUserId.value = undefined
    position.value = 1
    try {
      const res = await getAllUsers()
      const map: Record<number, string> = {}
      for (const u of res.data || []) {
        map[u.id] = u.name || u.username
      }
      nameMap.value = map
    } catch {
      /* ignore */
    }
  },
)

function onConfirm() {
  if (!toUserId.value) {
    ElMessage.warning('请选择加签人')
    return
  }
  emit('confirm', {
    toUserId: toUserId.value,
    toName: nameMap.value[toUserId.value] || '',
    position: position.value,
  })
}
</script>

<style scoped lang="scss">
.hint {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
}
:deep(.el-dialog__footer) {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  .el-button + .el-button {
    margin-left: 0;
  }
}
</style>
