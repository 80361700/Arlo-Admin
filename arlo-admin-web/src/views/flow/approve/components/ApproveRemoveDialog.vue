<template>
  <el-dialog
    :model-value="modelValue"
    title="减签"
    width="480px"
    destroy-on-close
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="hint">只能移除尚未处理的加签人，原审批人不可减，且至少保留一名待审批人。</div>
    <el-radio-group v-if="candidates.length" v-model="selectedId" class="actor-list">
      <el-radio
        v-for="a in candidates"
        :key="a.id"
        :value="a.id"
        class="actor-item"
      >
        <span class="name">{{ a.actorName || '-' }}</span>
        <el-tag size="small" type="warning" effect="plain">加签</el-tag>
      </el-radio>
    </el-radio-group>
    <el-empty v-else description="当前没有可减签的人员" :image-size="64" />
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button
        type="primary"
        :loading="loading"
        :disabled="!candidates.length"
        @click="onConfirm"
      >
        确定
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

export type RemoveActorCandidate = {
  id: number
  actorId: number
  actorName: string
  actorType: number
  actorState: number
}

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    loading?: boolean
    actors?: RemoveActorCandidate[]
    /** 当前操作人用户 id，用于排除自己 */
    currentUserId?: number
  }>(),
  {
    loading: false,
    actors: () => [],
    currentUserId: 0,
  },
)

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  confirm: [payload: { actorId: number }]
}>()

const selectedId = ref<number>()

const candidates = computed(() => {
  const pending = (props.actors || []).filter((a) => Number(a.actorState) === 0)
  if (pending.length <= 1) return [] as RemoveActorCandidate[]
  return pending.filter(
    (a) => Number(a.actorType) === 1 && Number(a.actorId) !== Number(props.currentUserId || 0),
  )
})

watch(
  () => props.modelValue,
  (v) => {
    if (!v) return
    selectedId.value = candidates.value[0]?.id
  },
)

function onConfirm() {
  if (!selectedId.value) {
    ElMessage.warning('请选择要减签的人员')
    return
  }
  emit('confirm', { actorId: selectedId.value })
}
</script>

<style scoped lang="scss">
.hint {
  margin-bottom: 14px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}
.actor-list {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
  width: 100%;
}
.actor-item {
  margin: 0;
  padding: 10px 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  height: auto;
  :deep(.el-radio__label) {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }
  .name {
    font-size: 14px;
    color: var(--el-text-color-primary);
  }
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
