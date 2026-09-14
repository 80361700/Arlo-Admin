<template>
  <div v-if="dialogShow" class="flow-designer-panel">
    <div class="dialog-header">
      <div class="dialog-header-left">
        <button type="button" class="back-btn" @click="close">
          <el-icon :size="16"><Back /></el-icon>
          <span>返回</span>
        </button>
        <span class="header-divider" />
        <span class="dialog-title">{{ dialogTitle }}</span>
      </div>
      <div class="dialog-actions">
        <template v-if="!readonly">
          <el-button v-if="step > 1" @click="handleStep('up')">上一步</el-button>
          <el-button v-if="step < maxStep" type="primary" @click="handleStep('next')">下一步</el-button>
          <el-button v-else type="primary" :loading="loading" @click="dialogSubmit">保存流程</el-button>
        </template>
        <template v-else>
          <el-button v-if="step > 1" @click="handleStep('up')">上一步</el-button>
          <el-button v-if="step < maxStep" type="primary" @click="handleStep('next')">下一步</el-button>
        </template>
      </div>
    </div>

    <div v-loading="loading" class="designer-container">
      <div class="designer-container-header">
        <el-steps :active="step" finish-status="success" simple>
          <el-step
            v-for="(s, i) in wizardSteps"
            :key="s.key"
            :title="s.title"
            @click="handleStep('to', i + 1)"
          />
        </el-steps>
      </div>

      <div
        v-if="flowForm.flow"
        class="designer-container-content"
        :class="{ 'is-readonly': readonly }"
      >
        <basic
          v-show="currentStepKey === 'basic'"
          ref="basicRef"
          v-model="flowForm"
          :readonly="readonly"
          class="step-pane"
        />
        <form-designer
          v-if="hasFormStep"
          v-show="currentStepKey === 'form'"
          ref="formRef"
          v-model="flowForm"
          class="step-pane step-pane--fill"
        />
        <process-designer
          v-show="currentStepKey === 'process'"
          ref="processRef"
          v-model="flowForm"
          :readonly="readonly"
          class="step-pane step-pane--fill"
        />
        <setting
          v-if="hasSettingStep"
          v-show="currentStepKey === 'setting'"
          ref="settingRef"
          v-model="flowForm"
          class="step-pane"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, nextTick, provide, ref } from 'vue'
import { Back } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getFlowProcess, getFlowProcessHistory, saveFlowProcess } from '@/api'
import config from '@/components/flowProcess/config'
import { syncNodesFormConfig } from '@/components/flowProcess/components/common/utils'
import basic from './components/basic.vue'
import formDesigner from './components/formDesigner.vue'
import processDesigner from './components/processDesigner.vue'
import setting from './components/setting.vue'
import { createEmptyFormSchema, normalizeProcessForm } from './components/formSchema'
import {
  normalizeProcessType,
  processTypeLabel,
  wizardStepsFor,
  type ProcessType,
} from './processType'

const emits = defineEmits<{ success: []; close: []; 'update:open': [v: boolean] }>()

const dialogShow = ref(false)
const dialogData = ref<any>({})
const loading = ref(false)
const readonly = ref(false)
const step = ref(1)
const formRef = ref<InstanceType<typeof formDesigner>>()
const basicRef = ref<InstanceType<typeof basic>>()
const processRef = ref()
const settingRef = ref()
const flowForm = ref<any>({})

const processType = computed(() => normalizeProcessType(flowForm.value?.flow?.processType))
const wizardSteps = computed(() => wizardStepsFor(processType.value))
const maxStep = computed(() => wizardSteps.value.length)
const currentStepKey = computed(() => wizardSteps.value[step.value - 1]?.key)
const hasFormStep = computed(() => wizardSteps.value.some((s) => s.key === 'form'))
const hasSettingStep = computed(() => wizardSteps.value.some((s) => s.key === 'setting'))
const dialogTitle = computed(() => {
  const name = flowForm.value?.flow?.processName
  const typeLabel = processTypeLabel(processType.value)
  const ver = flowForm.value?.flow?.processVersion
  if (readonly.value) {
    const base = name ? `${typeLabel} · ${name}` : typeLabel
    return ver ? `预览 · ${base} · V${ver}` : `预览 · ${base}`
  }
  if (name) return `${typeLabel} · ${name}`
  return typeLabel
})

function initForm(type: ProcessType = 'main') {
  return {
    flow: {
      processId: undefined as number | undefined,
      categoryId: undefined as number | undefined,
      processIcon: '',
      processBgcolor: 'rgba(30, 144, 255, 1)',
      processType: type,
      processKey: '',
      processName: '',
      remark: '',
      processPermissionList: [] as any[],
      bindFormId: undefined as number | undefined,
      bindFormName: '',
      bindFormCode: '',
      processForm: createEmptyFormSchema(),
      modelContent: {
        key: '',
        name: '',
        nodeConfig: {
          nodeName: '发起人',
          nodeKey: config.nodeKey(),
          type: 0,
          childNode: {
            nodeName: '结束',
            nodeKey: config.nodeKey(),
            type: -1,
          },
          nodeAssigneeList: [],
          extendConfig: {
            formConfig: [],
          },
        },
      },
      processSetting: {
        allowRevocation: true,
        allowDelegate: true,
        allowBatchOperate: true,
        secondOperatePrompt: true,
        repeatOperateSkip: true,
        bindFormId: undefined as number | undefined,
        bindFormName: '',
        bindFormCode: '',
      },
    },
  }
}

function hydrateBindForm(flow: any) {
  const setting = flow.processSetting || {}
  flow.bindFormId = flow.bindFormId ?? setting.bindFormId
  flow.bindFormName = flow.bindFormName || setting.bindFormName || ''
  flow.bindFormCode = flow.bindFormCode || setting.bindFormCode || ''
}

function persistBindForm(flow: any) {
  if (!flow.processSetting || typeof flow.processSetting !== 'object') {
    flow.processSetting = {}
  }
  if (normalizeProcessType(flow.processType) === 'business') {
    flow.processSetting.bindFormId = flow.bindFormId
    flow.processSetting.bindFormName = flow.bindFormName || ''
    flow.processSetting.bindFormCode = flow.bindFormCode || ''
  } else {
    delete flow.processSetting.bindFormId
    delete flow.processSetting.bindFormName
    delete flow.processSetting.bindFormCode
  }
  delete flow.bindFormId
  delete flow.bindFormName
  delete flow.bindFormCode
}

async function init(obj: any = {}) {
  dialogData.value = obj
  dialogShow.value = true
  emits('update:open', true)
  step.value = 1
  readonly.value = !!obj.preview

  if (obj.processId && obj.historyId) {
    loading.value = true
    try {
      const res = await getFlowProcessHistory(obj.processId, obj.historyId)
      await applyProcessDetail(res.data || {})
    } finally {
      loading.value = false
    }
  } else if (obj.processId) {
    loading.value = true
    try {
      const res = await getFlowProcess(obj.processId)
      await applyProcessDetail(res.data || {})
    } finally {
      loading.value = false
    }
  } else {
    readonly.value = false
    const type = normalizeProcessType(obj.processType)
    flowForm.value = initForm(type)
    if (obj.categoryId && obj.categoryId !== '-1') {
      flowForm.value.flow.categoryId = Number(obj.categoryId)
    }
    await nextTick()
    if (hasFormStep.value) {
      flowForm.value.flow.processForm = formRef.value?.getJson() || createEmptyFormSchema()
    }
  }
}

async function applyProcessDetail(raw: any) {
  const resData = JSON.parse(JSON.stringify(raw || {}))
  if (!resData.modelContent?.nodeConfig) {
    resData.modelContent = initForm().flow.modelContent
  }
  resData.processType = normalizeProcessType(resData.processType)
  resData.processForm = normalizeProcessForm(resData.processForm)
  if (!resData.processSetting) {
    resData.processSetting = initForm().flow.processSetting
  }
  if (!resData.processPermissionList) resData.processPermissionList = []
  if (!resData.processBgcolor) resData.processBgcolor = 'rgba(30, 144, 255, 1)'
  hydrateBindForm(resData)
  flowForm.value = { flow: resData }
  await nextTick()
  if (hasFormStep.value) formRef.value?.setJson()
}

function syncProcessForm() {
  if (!hasFormStep.value || !formRef.value || !flowForm.value?.flow) return
  flowForm.value.flow.processForm = formRef.value.getJson() || createEmptyFormSchema()
  // 表单字段变更后同步刷新各节点 formConfig，避免打开节点才看到旧字段
  const nodeConfig = flowForm.value.flow.modelContent?.nodeConfig
  if (nodeConfig) {
    syncNodesFormConfig(nodeConfig, flowForm.value.flow.processForm)
  }
}

provide('syncFlowProcessForm', syncProcessForm)

function close() {
  dialogShow.value = false
  readonly.value = false
  emits('update:open', false)
  emits('close')
}

function stepIndexOf(key: string) {
  return wizardSteps.value.findIndex((s) => s.key === key) + 1
}

async function dialogSubmit() {
  if (readonly.value) return
  const form = basicRef.value?.validate()
  if (!form) return
  try {
    await form.validate()
  } catch {
    handleStep('to', stepIndexOf('basic') || 1)
    ElMessage.warning('请完善基础信息')
    return
  }

  if (!flowForm.value.flow.modelContent?.nodeConfig) {
    ElMessage.warning('请完成流程设计')
    handleStep('to', stepIndexOf('process') || maxStep.value)
    return
  }

  flowForm.value.flow.modelContent.key = flowForm.value.flow.processKey
  flowForm.value.flow.modelContent.name = flowForm.value.flow.processName
  syncProcessForm()

  const formData = JSON.parse(JSON.stringify(flowForm.value.flow))
  formData.processType = normalizeProcessType(formData.processType)
  persistBindForm(formData)
  if (!formData.processId) delete formData.processId

  loading.value = true
  try {
    await saveFlowProcess(formData)
    ElMessage.success('保存成功')
    close()
    emits('success')
  } finally {
    loading.value = false
  }
}

function handleStep(type: string, index = 0) {
  if (currentStepKey.value === 'form') syncProcessForm()

  if (type === 'up') {
    if (step.value <= 1) return
    step.value--
  } else if (type === 'next') {
    if (step.value >= maxStep.value) return
    step.value++
  } else {
    if (index < 1 || index > maxStep.value) return
    step.value = index
  }

  // 进入流程设计步时再同步一次（防止仅点步骤条时字段未落盘）
  if (currentStepKey.value === 'process') syncProcessForm()

  if (currentStepKey.value === 'form') {
    nextTick(() => formRef.value?.setJson())
  }
}

defineExpose({ init })
</script>

<style lang="scss" scoped>
.flow-designer-panel {
  position: absolute;
  inset: 0;
  z-index: 20;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
  border-radius: inherit;
}
.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
  flex-shrink: 0;
}
.dialog-header-left {
  display: flex;
  align-items: center;
  min-width: 0;
}
.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--el-text-color-regular);
  font-size: 14px;
  line-height: 22px;
  cursor: pointer;
  &:hover {
    color: var(--el-color-primary);
  }
}
.header-divider {
  display: inline-block;
  width: 1px;
  height: 16px;
  margin: 0 12px;
  background: var(--el-border-color);
  flex-shrink: 0;
}
.dialog-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 24px;
}
.designer-container {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 16px;
  background: var(--el-bg-color);
  &-header {
    flex-shrink: 0;
    background-color: var(--el-bg-color);
    z-index: 99;

    :deep(.el-steps--simple) {
      padding: 13px 10%;
    }
    /* 保持 EP 默认均分；仅禁止标题折行 */
    :deep(.el-steps--simple .el-step__title) {
      white-space: nowrap;
    }
  }
  &-content {
    flex: 1;
    min-height: 0;
    overflow: auto;
    margin-top: 16px;
    display: flex;
    flex-direction: column;

    &.is-readonly {
      pointer-events: none;
      user-select: none;
    }
  }
}
:deep(.step-pane--fill) {
  flex: 1;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
}
:deep(.el-step) {
  cursor: pointer;
}
</style>
