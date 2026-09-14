<template>
  <el-popover
    v-model:visible="visible"
    placement="right-start"
    :width="280"
    trigger="click"
    :show-arrow="true"
    :hide-after="0"
    popper-class="approve-advanced-filter-popper"
    @show="syncDraft"
  >
    <template #reference>
      <el-button
        class="header-action-btn is-icon"
        size="small"
        title="高级筛选"
        :class="{ 'is-active': hasActive || visible }"
      >
        <el-icon :size="14"><Filter /></el-icon>
      </el-button>
    </template>
    <div class="af-wrap" @click.stop>
      <div class="af-title">高级筛选</div>
      <div class="af-row">
        <el-input
          v-model="draft.createBy"
          clearable
          placeholder="请输入创建人名称"
        />
      </div>
      <div class="af-row">
        <el-select
          v-model="draft.instanceState"
          clearable
          placeholder="流程状态"
          style="width: 100%"
          :teleported="false"
          :value-on-clear="undefined"
        >
          <el-option
            v-for="item in stateOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </div>
      <div class="af-row">
        <el-date-picker
          v-model="draft.beginTime"
          type="date"
          value-format="YYYY-MM-DD"
          placeholder="开始时间"
          style="width: 100%"
          :disabled-date="disabledStart"
          :teleported="false"
        />
      </div>
      <div class="af-row">
        <el-date-picker
          v-model="draft.endTime"
          type="date"
          value-format="YYYY-MM-DD"
          placeholder="结束时间"
          style="width: 100%"
          :disabled-date="disabledEnd"
          :teleported="false"
        />
      </div>
      <div class="af-footer">
        <el-button plain @click="onReset">重置</el-button>
        <el-button type="primary" plain @click="onSearch">检索</el-button>
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Filter } from '@element-plus/icons-vue'
import type { ApproveAdvancedFilterValue } from '../utils/dateRange'
import { defaultApproveDateRange } from '../utils/dateRange'

const props = withDefaults(
  defineProps<{
    modelValue: ApproveAdvancedFilterValue
    /** 页面默认近 N 天；与默认一致时不点亮按钮。0/不传表示默认无时间条件 */
    defaultDays?: number
    /** 页面默认流程状态；与默认一致时不点亮按钮。不传表示默认无状态条件 */
    defaultInstanceState?: number
  }>(),
  { defaultDays: 0 },
)

const emit = defineEmits<{
  'update:modelValue': [v: ApproveAdvancedFilterValue]
  search: []
}>()

const visible = ref(false)
const draft = reactive<{
  createBy: string
  /** 用字符串避免 el-select 把 0 当成空值误匹配 */
  instanceState?: string
  beginTime?: string
  endTime?: string
}>({
  createBy: '',
  instanceState: undefined,
  beginTime: undefined,
  endTime: undefined,
})

const stateOptions = [
  { value: '0', label: '审批中' },
  { value: '1', label: '已通过' },
  { value: '2', label: '已拒绝' },
  { value: '3', label: '已撤销' },
  { value: '4', label: '已终止' },
  { value: '5', label: '已超时' },
  { value: '-1', label: '暂存待审' },
]

function isBaselineTime(v: ApproveAdvancedFilterValue) {
  const begin = v.beginTime || undefined
  const end = v.endTime || undefined
  if (!props.defaultDays || props.defaultDays <= 0) {
    return !begin && !end
  }
  const [b, e] = defaultApproveDateRange(props.defaultDays)
  return begin === b && end === e
}

function isBaselineState(v: ApproveAdvancedFilterValue) {
  const cur =
    v.instanceState !== undefined && v.instanceState !== null ? Number(v.instanceState) : undefined
  if (props.defaultInstanceState === undefined || props.defaultInstanceState === null) {
    return cur === undefined
  }
  return cur === Number(props.defaultInstanceState)
}

const hasActive = computed(() => {
  const v = props.modelValue
  if (v.createBy?.trim()) return true
  if (!isBaselineState(v)) return true
  return !isBaselineTime(v)
})

function syncDraft() {
  draft.createBy = props.modelValue.createBy || ''
  draft.instanceState =
    props.modelValue.instanceState !== undefined && props.modelValue.instanceState !== null
      ? String(props.modelValue.instanceState)
      : undefined
  draft.beginTime = props.modelValue.beginTime || undefined
  draft.endTime = props.modelValue.endTime || undefined
}

function onReset() {
  Object.assign(draft, {
    createBy: '',
    instanceState:
      props.defaultInstanceState !== undefined && props.defaultInstanceState !== null
        ? String(props.defaultInstanceState)
        : undefined,
    beginTime: undefined,
    endTime: undefined,
  })
}

function onSearch() {
  let instanceState: number | undefined
  if (draft.instanceState !== undefined && draft.instanceState !== null && draft.instanceState !== '') {
    const n = Number(draft.instanceState)
    if (Number.isFinite(n)) instanceState = n
  }
  emit('update:modelValue', {
    createBy: draft.createBy?.trim() || '',
    instanceState,
    beginTime: draft.beginTime || undefined,
    endTime: draft.endTime || undefined,
  })
  visible.value = false
  emit('search')
}

function disabledStart(time: Date) {
  if (!draft.endTime) return false
  return time.getTime() > new Date(`${draft.endTime} 23:59:59`).getTime()
}

function disabledEnd(time: Date) {
  if (!draft.beginTime) return false
  return time.getTime() < new Date(`${draft.beginTime} 00:00:00`).getTime()
}

watch(
  () => props.modelValue,
  () => {
    if (visible.value) syncDraft()
  },
  { deep: true },
)
</script>

<style scoped lang="scss">
.header-action-btn {
  margin: 0;
  height: 28px;
  width: 28px;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--el-text-color-regular);
  background: #fff;
  border: 1px solid var(--el-border-color);
  &:hover,
  &:focus {
    color: var(--el-color-primary);
    border-color: var(--el-color-primary-light-5);
    background: var(--el-color-primary-light-9);
  }
  &.is-active {
    color: var(--el-color-primary);
    border-color: var(--el-color-primary-light-5);
    background: var(--el-color-primary-light-9);
  }
}
.af-wrap {
  display: flex;
  flex-direction: column;
  gap: 10px;
  .af-title {
    border-bottom: 1px solid var(--el-border-color-light);
    margin: -12px -12px 0;
    padding: 10px 12px;
    font-size: 15px;
    font-weight: 500;
    color: var(--el-text-color-primary);
  }
  .af-row {
    width: 100%;
  }
  .af-footer {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: 8px;
    padding-top: 2px;
    .el-button {
      margin: 0;
    }
  }
}
</style>

<style lang="scss">
.approve-advanced-filter-popper {
  padding: 12px !important;
  overflow: visible !important;
}
.approve-advanced-filter-popper .af-wrap {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.approve-advanced-filter-popper .af-row {
  width: 100%;
}
.approve-advanced-filter-popper .af-row .el-select,
.approve-advanced-filter-popper .af-row .el-date-editor,
.approve-advanced-filter-popper .af-row .el-input {
  width: 100% !important;
}
.approve-advanced-filter-popper .af-footer {
  display: flex !important;
  justify-content: flex-end !important;
  align-items: center;
  gap: 8px !important;
}
.approve-advanced-filter-popper .af-footer .el-button {
  margin: 0 !important;
  flex: none !important;
}
</style>
