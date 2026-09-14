<template>
  <div style="padding-top: 12px;" :class="{ 'is-readonly': readonly }">
    <el-form ref="formRef" :model="modelValue.flow" :rules="rules" label-width="140px" :disabled="readonly">
      <el-form-item label="图标" prop="processIcon">
        <IconPicker v-model="modelValue.flow.processIcon" placeholder="点击选择图标" />
        <el-color-picker v-model="modelValue.flow.processBgcolor" show-alpha :predefine="predefineColors" style="margin-left: 10px;" />
      </el-form-item>
      <el-form-item label="唯一标识key" prop="processKey">
        <el-input v-model="modelValue.flow.processKey" maxlength="64" :disabled="!!modelValue.flow.processId" />
      </el-form-item>
      <el-form-item label="名称" prop="processName">
        <el-input v-model="modelValue.flow.processName" maxlength="128" />
      </el-form-item>

      <el-form-item v-if="isBusiness" label="流程表单" prop="bindFormId">
        <div class="selectBtn">
          <el-button :icon="Plus" round type="primary" @click="handleForm">选择流程表单</el-button>
        </div>
        <div v-if="modelValue.flow.bindFormId" class="tags">
          <el-tag closable @close="clearForm">
            {{ modelValue.flow.bindFormName || modelValue.flow.bindFormId }}
          </el-tag>
        </div>
      </el-form-item>

      <el-form-item label="说明" prop="remark">
        <el-input v-model="modelValue.flow.remark" type="textarea" maxlength="200" show-word-limit />
      </el-form-item>
      <el-form-item label="分组" prop="categoryId">
        <el-select v-model="modelValue.flow.categoryId" placeholder="请选择分组" style="width: 100%">
          <el-option v-for="item in categoryOptions" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
      </el-form-item>
      <el-form-item prop="processPermissionList">
        <template #label>
          <span class="label-with-tip">
            流程管理员
            <el-tooltip
              placement="top"
              content="指定后：创建人、所选人员、超管可改流程、启停；进行中实例的终止/转办请到「工作流 → 流程监控」。"
            >
              <el-icon class="label-tip-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </span>
        </template>
        <div class="selectBtn">
          <el-button :icon="Plus" type="primary" @click="handleUser">选择人员</el-button>
        </div>
        <div class="tags">
          <el-tag
            v-for="(item, index) in modelValue.flow.processPermissionList"
            :key="index"
            closable
            @close="userDelete(index)"
          >
            {{ item.userName || item.name }}
          </el-tag>
        </div>
      </el-form-item>
    </el-form>

    <user-select ref="userRef" @success="userSuccess" />
    <form-select ref="formSelectRef" @success="formSuccess" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { Plus, QuestionFilled } from '@element-plus/icons-vue'
import IconPicker from '@/components/IconPicker.vue'
import userSelect from '@/components/flowProcess/components/common/user/index.vue'
import formSelect from '@/components/flowProcess/components/common/form/index.vue'
import { getFlowCategoryOptions, getFlowForm } from '@/api'
import { normalizeProcessForm, createEmptyFormSchema } from './formSchema'
import { normalizeProcessType } from '../processType'

const props = defineProps<{ modelValue: any; readonly?: boolean }>()

const formRef = ref<FormInstance>()
const categoryOptions = ref<{ id: number; name: string }[]>([])
const userRef = ref<InstanceType<typeof userSelect>>()
const formSelectRef = ref<InstanceType<typeof formSelect>>()
const predefineColors = [
  '#ff4500',
  '#ff8c00',
  '#ffd700',
  '#90ee90',
  '#00ced1',
  '#1e90ff',
  '#c71585',
  'rgba(255, 69, 0, 0.68)',
  'rgb(255, 120, 0)',
  'hsv(51, 100, 98)',
  'hsva(120, 40, 94, 0.5)',
  'hsl(181, 100%, 37%)',
  'hsla(209, 100%, 56%, 0.73)',
  '#c7158577',
]

const isBusiness = computed(
  () => normalizeProcessType(props.modelValue?.flow?.processType) === 'business',
)

const rules = computed<FormRules>(() => {
  const base: FormRules = {
    processIcon: [{ required: true, message: '请选择图标', trigger: 'change' }],
    processKey: [{ required: true, message: '请输入唯一标识key', trigger: 'blur' }],
    processName: [{ required: true, message: '请输入名称', trigger: 'blur' }],
    categoryId: [{ required: true, message: '请选择分组', trigger: 'change' }],
  }
  if (isBusiness.value) {
    base.bindFormId = [{ required: true, message: '请选择流程表单', trigger: 'change' }]
  }
  return base
})

onMounted(async () => {
  const res = await getFlowCategoryOptions()
  categoryOptions.value = res.data || []
})

function validate() {
  return formRef.value
}

function handleUser() {
  userRef.value?.init({
    selectData: (props.modelValue.flow.processPermissionList || []).map((item: any) => ({
      id: item.userId || item.id,
      name: item.userName || item.name,
    })),
  })
}

function userSuccess(e: any[]) {
  props.modelValue.flow.processPermissionList = []
  e?.forEach((row) => {
    props.modelValue.flow.processPermissionList.push({
      id: row.id,
      name: row.name,
      operateApproval: 1,
      operateData: 1,
      operateOwner: 1,
      userId: row.id,
      userName: row.name,
    })
  })
}

function userDelete(i: number) {
  props.modelValue.flow.processPermissionList.splice(i, 1)
}

function handleForm() {
  formSelectRef.value?.init({ formId: props.modelValue.flow.bindFormId })
}

async function formSuccess(item: {
  id: number
  name: string
  code: string
  formType?: number
  pcUrl?: string
}) {
  props.modelValue.flow.bindFormId = item.id
  props.modelValue.flow.bindFormName = item.name
  props.modelValue.flow.bindFormCode = item.code
  try {
    const res = await getFlowForm(item.id)
    const detail = res.data
    if (Number(detail?.formType) === 2 || Number(item.formType) === 2) {
      // 系统表单：无 epic schema，流程设计里用空壳即可
      props.modelValue.flow.processForm = createEmptyFormSchema()
      props.modelValue.flow.bindFormName = detail?.pcUrl
        ? `${detail.name}（${detail.pcUrl}）`
        : detail?.name || item.name
    } else {
      props.modelValue.flow.processForm = normalizeProcessForm(detail?.formSchema)
    }
  } catch {
    ElMessage.warning('已绑定表单，但加载失败，请稍后重试')
    props.modelValue.flow.processForm = createEmptyFormSchema()
  }
  formRef.value?.validateField?.('bindFormId')
}

function clearForm() {
  props.modelValue.flow.bindFormId = undefined
  props.modelValue.flow.bindFormName = ''
  props.modelValue.flow.bindFormCode = ''
  props.modelValue.flow.processForm = createEmptyFormSchema()
}

defineExpose({ validate })
</script>

<style lang="scss" scoped>
.viewIcon-item {
  width: 70px;
  height: 70px;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 8px;
  position: relative;
  .color-mask {
    position: absolute;
    left: 0;
    top: 0;
  }
  :deep(.el-color-picker__trigger) {
    width: 70px;
    height: 70px;
    opacity: 0;
  }
}
.color-tip {
  color: #999;
}
.tags {
  width: 100%;
  padding-top: 10px;
  :deep(.el-tag) {
    margin-right: 10px;
    margin-bottom: 6px;
  }
}
.selectBtn {
  display: flex;
}
.label-with-tip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.label-tip-icon {
  color: var(--el-text-color-secondary);
  cursor: help;
  font-size: 14px;
}
.is-readonly {
  pointer-events: none;
}
</style>
