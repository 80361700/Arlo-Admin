<template>
  <div class="conditionList">
    <div v-for="(group, groupIndex) in modelValue" :key="groupIndex" class="item">
      <div v-if="groupIndex > 0" class="tips">或满足</div>
      <div class="item-body">
        <div class="title">
          <span>{{ `条件组${groupIndex + 1}` }}</span>
          <el-button icon="Delete" circle type="primary" plain size="small" @click="deleteGroup(groupIndex)" />
        </div>

        <div class="rows">
          <div v-for="(row, rowIndex) in group" :key="rowIndex" class="row">
            <span class="when">{{ rowIndex === 0 ? '当' : '且' }}</span>

            <!-- 发起人条件 -->
            <template v-if="row.type === 'initiator'">
              <el-select model-value="initiator" disabled class="field-select">
                <el-option label="发起人" value="initiator" />
              </el-select>
              <el-select v-model="row.operator" class="op-select" placeholder="运算符">
                <el-option label="属于" value="belong" />
                <el-option label="不属于" value="notbelong" />
              </el-select>
              <div class="value-box">
                <el-tag
                  v-for="(tag, tagIndex) in row.valueList || []"
                  :key="`${tag.kind}-${tag.id}`"
                  size="small"
                  closable
                  @close="removeMember(row, tagIndex)"
                >
                  {{ tag.name }}
                </el-tag>
                <el-button link type="primary" size="small" @click="openOrgPicker(row)">添加</el-button>
              </div>
            </template>

            <!-- 表单条件 -->
            <template v-else>
              <el-select
                :model-value="row.field"
                class="field-select"
                placeholder="选择表单字段"
                filterable
                @change="(val: string) => onFormFieldChange(row, val)"
              >
                <el-option
                  v-for="f in formFields"
                  :key="f.id"
                  :label="f.label"
                  :value="f.id"
                />
              </el-select>
              <el-select v-model="row.operator" class="op-select" placeholder="运算符">
                <el-option label="等于" value="==" />
                <el-option label="不等于" value="!=" />
                <el-option label="大于" value=">" />
                <el-option label="大于等于" value=">=" />
                <el-option label="小于" value="<" />
                <el-option label="小于等于" value="<=" />
                <el-option label="包含" value="include" />
                <el-option label="不包含" value="notinclude" />
              </el-select>
              <el-input v-model="row.value" class="value-input" placeholder="值" />
            </template>

            <el-button link type="primary" size="small" @click="deleteRow(group, rowIndex)">删除</el-button>
          </div>
        </div>

        <div class="footer">
          <el-dropdown trigger="click" @command="(cmd: string) => addCondition(group, cmd)">
            <el-button link type="primary" icon="Plus" size="small">添加条件</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="form">表单条件</el-dropdown-item>
                <el-dropdown-item command="initiator">发起人条件</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </div>

    <div class="actions">
      <el-button icon="Plus" type="primary" plain @click="pushGroup">添加条件分组</el-button>
      <slot name="extra" />
    </div>

    <org-member-select ref="orgRef" @success="onOrgPicked" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  createFormCondition,
  createInitiatorCondition,
  getFormFields,
} from '../utils'
import orgMemberSelect from '../orgMember/index.vue'

type OrgMemberItem = { id: string | number; name: string; kind: 'dept' | 'role' | 'user' }

const props = defineProps<{
  modelValue: any[][]
  processForm?: any
}>()

const emit = defineEmits<{
  'update:modelValue': [value: any[][]]
}>()

const formFields = computed(() => getFormFields(props.processForm))
const orgRef = ref()
const editingRow = ref<any>(null)

function ensureList() {
  if (!Array.isArray(props.modelValue)) {
    emit('update:modelValue', [])
  }
}

function syncValueText(row: any) {
  if (row.type === 'initiator') {
    row.value = (row.valueList || []).map((v: OrgMemberItem) => v.name).join('、')
  }
}

function onFormFieldChange(row: any, fieldId: string) {
  const field = formFields.value.find((f) => f.id === fieldId)
  row.type = 'form'
  row.field = fieldId
  row.label = field?.label || fieldId
}

function addCondition(group: any[], cmd: string) {
  if (cmd === 'initiator') {
    group.push(createInitiatorCondition())
  } else {
    group.push(createFormCondition())
  }
}

function deleteRow(group: any[], index: number) {
  group.splice(index, 1)
}

function pushGroup() {
  ensureList()
  const list = props.modelValue || []
  list.push([createFormCondition()])
  emit('update:modelValue', list)
}

function deleteGroup(index: number) {
  props.modelValue.splice(index, 1)
}

function openOrgPicker(row: any) {
  editingRow.value = row
  if (!Array.isArray(row.valueList)) row.valueList = []
  orgRef.value?.init({ selectData: row.valueList })
}

function onOrgPicked(list: OrgMemberItem[]) {
  if (!editingRow.value) return
  editingRow.value.valueList = list
  syncValueText(editingRow.value)
  editingRow.value = null
}

function removeMember(row: any, index: number) {
  row.valueList.splice(index, 1)
  syncValueText(row)
}

/** 兼容旧数据：custom / 空 type 视为表单条件 */
function normalize(list: any[][]) {
  if (!Array.isArray(list)) return
  list.forEach((group) => {
    if (!Array.isArray(group)) return
    group.forEach((row) => {
      if (!row) return
      if (row.type === 'initiator') {
        row.label = row.label || '发起人'
        row.field = row.field || 'initiator'
        if (!row.operator || !['belong', 'notbelong'].includes(row.operator)) {
          row.operator = row.operator === 'notinclude' ? 'notbelong' : 'belong'
        }
        if (!Array.isArray(row.valueList)) row.valueList = []
        syncValueText(row)
      } else {
        row.type = 'form'
        if (row.field && !row.label) {
          const f = formFields.value.find((it) => it.id === String(row.field))
          row.label = f?.label || row.field
        }
      }
    })
  })
}

normalize(props.modelValue)
</script>

<style lang="scss" scoped>
.conditionList {
  width: 100%;
  box-sizing: border-box;

  .item {
    margin-bottom: 10px;
    width: 100%;

    .item-body {
      width: 100%;
      box-sizing: border-box;
      border: 1px solid var(--el-border-color-lighter);
      border-radius: 4px;
      overflow: hidden;
    }

    .tips {
      font-size: 14px;
      color: #999;
      margin-bottom: 6px;
    }

    .title {
      padding: 5px 12px;
      font-size: 14px;
      background: #f4f4f5;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .rows {
      padding: 8px 12px;
      width: 100%;
      box-sizing: border-box;
    }

    .row {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 8px;
      margin-bottom: 8px;
      width: 100%;

      .when {
        color: var(--el-text-color-secondary);
        font-size: 13px;
        flex-shrink: 0;
      }

      .field-select {
        width: 120px;
      }

      .op-select {
        width: 110px;
      }

      .value-input {
        flex: 1;
        min-width: 100px;
      }

      .value-box {
        flex: 1;
        min-width: 140px;
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 4px;
        padding: 2px 8px;
        min-height: 32px;
        border: 1px solid var(--el-border-color);
        border-radius: var(--el-border-radius-base);
        background: var(--el-fill-color-blank);
      }
    }

    .footer {
      padding: 5px 12px 10px;
    }
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }
}
</style>
