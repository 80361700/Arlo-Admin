<template>
  <el-dialog
    :model-value="modelValue"
    title="同意审批"
    width="560px"
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

      <el-form-item label="下一节点审批人">
        <div class="next-preview">
          <div class="np-rail">
            <div class="np-icon-wrap">
              <div class="np-icon" :class="isEnd ? 'is-end' : 'is-approve'">
                <el-icon :size="18">
                  <CircleCheck v-if="isEnd" />
                  <Avatar v-else />
                </el-icon>
              </div>
              <div v-if="!isEnd" class="np-badge">
                <el-icon :size="10"><Clock /></el-icon>
              </div>
            </div>
          </div>
          <div class="np-body">
            <div class="np-name">{{ nextLabel }}</div>
            <div v-if="actors.length" class="np-actors">
              <FlowNodeAvatar v-for="(a, i) in actors" :key="`${a}-${i}`" :name="a" :size="22" />
            </div>
            <div v-else-if="!isEnd && nextHint" class="np-hint">{{ nextHint }}</div>
          </div>
        </div>
      </el-form-item>

      <el-form-item label="推荐回复">
        <div class="quick-replies">
          <button
            v-for="t in quickReplies"
            :key="t"
            type="button"
            class="quick-btn"
            @click="opinion = t"
          >
            {{ t }}
          </button>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="loading" @click="onConfirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Avatar, CircleCheck, Clock } from '@element-plus/icons-vue'
import FlowNodeAvatar from './FlowNodeAvatar.vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    loading?: boolean
    /** 下一节点展示名；空则显示「结束」 */
    nextNodeName?: string
    /** 下一节点类型，-1/结束节点用结束样式 */
    nextNodeType?: number
    /** 下一节点已知审批人 */
    nextActors?: string[]
    /** 无法枚举人员时的说明，如「发起人的第1级主管」 */
    nextHint?: string
  }>(),
  {
    loading: false,
    nextNodeName: '',
    nextNodeType: undefined,
    nextActors: () => [],
    nextHint: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  confirm: [opinion: string]
}>()

const opinion = ref('')
const quickReplies = [
  '同意',
  '已阅',
  '收到',
  '已核对',
  '合格',
  '情况属实',
  '确认',
  '已复核',
  '知悉',
  '辛苦了',
  '已安排',
]

const nextLabel = computed(() => props.nextNodeName?.trim() || '结束')
const isEnd = computed(() => {
  if (Number(props.nextNodeType) === -1) return true
  return nextLabel.value === '结束'
})
const actors = computed(() => (props.nextActors || []).filter(Boolean))

watch(
  () => props.modelValue,
  (v) => {
    if (v) opinion.value = ''
  },
)

function onConfirm() {
  const text = opinion.value.trim()
  if (!text) {
    ElMessage.warning('请填写审批意见')
    return
  }
  emit('confirm', text)
}
</script>

<style scoped lang="scss">
.next-preview {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 4px 0;
}
.np-rail {
  flex-shrink: 0;
}
.np-icon-wrap {
  position: relative;
  width: 40px;
  height: 40px;
}
.np-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: var(--el-color-primary);
  &.is-end {
    background: #909399;
  }
}
.np-badge {
  position: absolute;
  right: -2px;
  bottom: -2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: #67c23a;
  border: 2px solid var(--el-bg-color);
  box-sizing: border-box;
}
.np-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.np-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  line-height: 1.4;
}
.np-name + .np-actors,
.np-name + .np-hint {
  margin-top: 8px;
}
.np-actors {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.np-hint {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.quick-replies {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.quick-btn {
  border: none;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-regular);
  font-size: 12px;
  line-height: 1;
  padding: 8px 10px;
  border-radius: 4px;
  cursor: pointer;
  &:hover {
    background: var(--el-fill-color);
    color: var(--el-color-primary);
  }
}
:deep(.el-dialog__footer) {
  .el-button + .el-button {
    margin-left: 0;
  }
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
