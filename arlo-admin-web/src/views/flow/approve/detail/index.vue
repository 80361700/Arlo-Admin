<template>
  <div class="page-container detail-page">
    <div class="toolbar">
      <el-button text type="primary" @click="goBack">返回</el-button>
      <span v-if="detail?.instanceId" class="detail-id">编号：{{ detail.instanceId }}</span>
      <span class="title">{{ detail?.processName || '审批详情' }}</span>
      <el-tag
        v-if="detail?.isSubProcess"
        size="small"
        type="primary"
        effect="plain"
        class="state-tag"
      >
        {{ detail.subProcessName || '子流程' }}
      </el-tag>
      <el-tag v-if="detail" size="small" class="state-tag">{{ stateLabel(detail.instanceState) }}</el-tag>
    </div>

    <div v-loading="loading" class="body">
      <el-row :gutter="16">
        <el-col :span="16">
          <el-card shadow="never" header="表单信息">
            <ApproveFormBlock
              ref="formBlockRef"
              :form-ready="formReady"
              :loading="loading"
              :form-render-type="formRenderType"
              :form-component="formComponent"
              :form-data="formData"
              :page-schema="pageSchema"
              :field-states="fieldStates"
              :render-key="formRenderKey"
              :readonly="vueFormReadonly"
              :disabled="!schemaFormEditable"
            />
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="never" header="流转记录">
            <FlowTimeline
              :items="detail?.timeline || []"
              :comments="detail?.comments || []"
              :image-size="64"
            />
          </el-card>
        </el-col>
      </el-row>
    </div>

    <div v-if="hasFooterActions" class="footer">
      <el-button
        v-if="detail?.canComment"
        :loading="acting"
        @click="commentVisible = true"
      >
        评论
      </el-button>
      <el-button
        v-if="detail?.canReject"
        type="danger"
        :loading="acting"
        @click="openReject"
      >
        {{ rejectBtnLabel }}
      </el-button>
      <el-button
        v-if="detail?.canConsent"
        type="primary"
        :loading="acting"
        @click="openConsent"
      >
        同意
      </el-button>
      <el-button
        v-if="detail?.canResubmit"
        type="primary"
        :loading="acting"
        @click="onResubmit"
      >
        重新提交
      </el-button>
      <el-dropdown
        v-if="moreActions.length"
        trigger="click"
        @command="onMoreCommand"
      >
        <el-button :loading="acting">更多</el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item
              v-for="a in moreActions"
              :key="a.cmd"
              :command="a.cmd"
            >
              {{ a.label }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <ApproveConsentDialog
      v-model="consentVisible"
      :loading="acting"
      :next-node-name="nextPreview.name"
      :next-node-type="nextPreview.type"
      :next-actors="nextPreview.actors"
      :next-hint="nextPreview.hint"
      @confirm="onConsentConfirm"
    />
    <ApproveRejectDialog
      v-model="rejectVisible"
      :loading="acting"
      :need-reject-target="needRejectTarget"
      :reject-targets="detail?.rejectTargets || []"
      :default-reject-node-key="detail?.rejectNodeKey || ''"
      @confirm="onRejectConfirm"
    />
    <ApproveTransferDialog
      v-model="transferVisible"
      :loading="acting"
      @confirm="onTransferConfirm"
    />
    <ApproveAppendDialog
      v-model="appendVisible"
      :loading="acting"
      @confirm="onAppendConfirm"
    />
    <ApproveRemoveDialog
      v-model="removeVisible"
      :loading="acting"
      :actors="detail?.taskActors || []"
      :current-user-id="authStore.userInfo?.id"
      @confirm="onRemoveConfirm"
    />
    <ApproveRollbackDialog
      v-model="rollbackVisible"
      :loading="acting"
      :targets="detail?.rollbackTargets || detail?.rejectTargets || []"
      @confirm="onRollbackConfirm"
    />
    <ApproveRevokeDialog
      v-model="revokeVisible"
      :loading="acting"
      @confirm="onRevokeConfirm"
    />
    <ApproveCcDialog
      v-model="ccVisible"
      :loading="acting"
      @confirm="onCcConfirm"
    />
    <ApproveCommentDialog
      v-model="commentVisible"
      :loading="acting"
      @confirm="onCommentConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { type FieldStates, type PageSchema } from 'epic-designer'
import FlowTimeline from '../components/FlowTimeline.vue'
import ApproveFormBlock from '../components/ApproveFormBlock.vue'
import ApproveConsentDialog from '../components/ApproveConsentDialog.vue'
import ApproveRejectDialog from '../components/ApproveRejectDialog.vue'
import ApproveTransferDialog from '../components/ApproveTransferDialog.vue'
import ApproveAppendDialog from '../components/ApproveAppendDialog.vue'
import ApproveRemoveDialog from '../components/ApproveRemoveDialog.vue'
import ApproveRollbackDialog from '../components/ApproveRollbackDialog.vue'
import ApproveRevokeDialog from '../components/ApproveRevokeDialog.vue'
import ApproveCcDialog from '../components/ApproveCcDialog.vue'
import ApproveCommentDialog from '../components/ApproveCommentDialog.vue'
import { useAuthStore } from '@/stores/auth'
import {
  addApproveComment,
  appendApproveActor,
  consentApproveTask,
  getApproveInstance,
  rejectApproveTask,
  removeApproveActor,
  resubmitApproveTask,
  revokeApproveInstance,
  rollbackApproveTask,
  runtimeCcApproveTask,
  transferApproveTask,
  urgeApproveInstance,
  type InstanceDetail,
} from '@/api'
import {
  buildFieldStates,
  canEditApproveForm,
  normalizeProcessForm,
  pageSchemaHasFields,
} from '@/views/flow/process/components/formSchema'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const acting = ref(false)
const detail = ref<InstanceDetail | null>(null)
const pageSchema = ref<PageSchema | null>(null)
const formData = ref<Record<string, any>>({})
const fieldStates = ref<FieldStates>([])
const formReady = ref(false)
const formEditable = ref(false)
const schemaFormEditable = ref(false)
const vueFormReadonly = ref(true)
const formRenderKey = ref(0)
const formBlockRef = ref<InstanceType<typeof ApproveFormBlock>>()
const formRenderType = ref('designer')
const formComponent = ref('')
const consentVisible = ref(false)
const rejectVisible = ref(false)
const transferVisible = ref(false)
const appendVisible = ref(false)
const removeVisible = ref(false)
const rollbackVisible = ref(false)
const revokeVisible = ref(false)
const ccVisible = ref(false)
const commentVisible = ref(false)

const needRejectTarget = computed(() => {
  if (!detail.value?.canReject) return false
  const s = Number(detail.value.rejectStrategy || 0) || 3
  if (s !== 3) return false
  return !detail.value.rejectNodeKey
})

const rejectBtnLabel = computed(() => {
  const s = Number(detail.value?.rejectStrategy || 0) || 3
  if (s === 4) return '拒绝'
  return '驳回'
})

const moreActions = computed(() => {
  const d = detail.value
  if (!d) return [] as { cmd: string; label: string }[]
  const list: { cmd: string; label: string }[] = []
  if (d.canTransfer) list.push({ cmd: 'transfer', label: '转交' })
  if (d.canAppend) list.push({ cmd: 'append', label: '加签' })
  if (d.canRemove) list.push({ cmd: 'remove', label: '减签' })
  if (d.canRollback) list.push({ cmd: 'rollback', label: '回退' })
  if (d.canCc) list.push({ cmd: 'cc', label: '抄送' })
  if (d.canUrge) list.push({ cmd: 'urge', label: '催办' })
  if (d.canRevoke) list.push({ cmd: 'revoke', label: '撤销' })
  return list
})

const hasFooterActions = computed(() => {
  const d = detail.value
  if (!d) return false
  return !!(d.canConsent || d.canReject || d.canResubmit || d.canComment || moreActions.value.length)
})

const nextPreview = computed(() => {
  const empty = { name: '结束', type: -1, actors: [] as string[], hint: '' }
  const d = detail.value
  if (!d) return empty
  const root = (d.modelContent as any)?.nodeConfig
  if (!root) return empty
  const cur = findNodeByKey(root, d.currentNodeKey)
  const next = cur?.childNode
  if (!next || next.type === -1) return empty
  return {
    name: next.nodeName || '下一节点',
    type: Number(next.type ?? 1),
    actors: describeNextActors(next).actors,
    hint: describeNextActors(next).hint,
  }
})

function describeNextActors(node: any): { actors: string[]; hint: string } {
  const list = Array.isArray(node?.nodeAssigneeList) ? node.nodeAssigneeList : []
  const names = list.map((x: any) => x?.name).filter(Boolean)
  if (names.length) return { actors: names, hint: '' }
  const setType = Number(node?.setType || 0)
  if (setType === 2) return { actors: [], hint: `发起人的第${node.examineLevel || 1}级主管` }
  if (setType === 5) return { actors: [], hint: '发起人自己' }
  if (setType === 6) {
    const end = node.directorLevel || node.examineLevel || 1
    return {
      actors: [],
      hint: node.directorMode == 0 ? '直到最上层主管' : `连续多级主管（直到第${end}级）`,
    }
  }
  if (setType === 4) {
    const cand = node.nodeCandidate?.assignees || []
    const cn = cand.map((x: any) => x?.name).filter(Boolean)
    if (cn.length) return { actors: cn, hint: '' }
    return { actors: [], hint: '发起人自选' }
  }
  if (setType === 3 && !names.length) return { actors: [], hint: '按角色审批' }
  return { actors: [], hint: '' }
}

function findNodeByKey(node: any, key?: string): any {
  if (!node || !key) return null
  if (node.nodeKey === key) return node
  if (node.childNode) {
    const hit = findNodeByKey(node.childNode, key)
    if (hit) return hit
  }
  for (const list of [node.conditionNodes, node.parallelNodes, node.inclusiveNodes, node.routeNodes]) {
    if (!Array.isArray(list)) continue
    for (const n of list) {
      const hit = findNodeByKey(n, key)
      if (hit) return hit
    }
  }
  return null
}

function stateLabel(s: number) {
  return ['审批中', '已通过', '已拒绝', '已撤销', '已终止', '已超时'][s] || String(s)
}

function goBack() {
  router.back()
}

/** 同意前校验可编辑表单（含节点子表单必填） */
async function validateApproveForm(): Promise<boolean> {
  if (!(detail.value?.canConsent || detail.value?.canResubmit)) return true
  const res = await formBlockRef.value?.validate({
    vue: !vueFormReadonly.value,
    schema: schemaFormEditable.value,
  })
  if (res && !res.ok) {
    ElMessage.warning(res.message)
    return false
  }
  return true
}

async function openConsent() {
  if (!(await validateApproveForm())) return
  closeActionDialogs()
  consentVisible.value = true
}

function openReject() {
  closeActionDialogs()
  rejectVisible.value = true
}

function closeActionDialogs() {
  consentVisible.value = false
  rejectVisible.value = false
  transferVisible.value = false
  appendVisible.value = false
  removeVisible.value = false
  rollbackVisible.value = false
  revokeVisible.value = false
  ccVisible.value = false
  commentVisible.value = false
}

async function onMoreCommand(cmd: string) {
  closeActionDialogs()
  if (cmd === 'transfer') transferVisible.value = true
  else if (cmd === 'append') appendVisible.value = true
  else if (cmd === 'remove') removeVisible.value = true
  else if (cmd === 'rollback') rollbackVisible.value = true
  else if (cmd === 'cc') ccVisible.value = true
  else if (cmd === 'revoke') revokeVisible.value = true
  else if (cmd === 'urge') await doUrge()
}

async function doUrge() {
  if (!detail.value?.instanceId) return
  acting.value = true
  try {
    await urgeApproveInstance({ instanceId: detail.value.instanceId })
    ElMessage.success('已催办')
  } finally {
    acting.value = false
  }
}

async function load() {
  const instanceId = Number(route.query.instanceId)
  const taskId = Number(route.query.taskId || 0)
  if (!instanceId) return
  loading.value = true
  formReady.value = false
  formEditable.value = false
  schemaFormEditable.value = false
  vueFormReadonly.value = true
  pageSchema.value = null
  formData.value = {}
  fieldStates.value = []
  formRenderType.value = 'designer'
  formComponent.value = ''
  try {
    const res = await getApproveInstance(instanceId, taskId || undefined)
    detail.value = res.data || null
    if (detail.value) {
      formData.value = { ...(detail.value.formData || {}) }
      formRenderType.value = detail.value.formRenderType === 'vue' ? 'vue' : 'designer'
      formComponent.value = detail.value.formComponent || ''
      pageSchema.value = normalizeProcessForm(detail.value.formSchema)
      let cfg = detail.value.formConfig || []
      if (
        detail.value.canResubmit &&
        cfg.length &&
        cfg.every((c) => Number(c.opera) === 0)
      ) {
        cfg = []
      }
      schemaFormEditable.value = canEditApproveForm(cfg, {
        canConsent: !!detail.value.canConsent,
        canResubmit: !!detail.value.canResubmit,
      })
      const hasSub = pageSchemaHasFields(pageSchema.value)
      vueFormReadonly.value =
        formRenderType.value !== 'vue'
          ? true
          : detail.value.canResubmit
            ? false
            : hasSub
              ? true
              : !(detail.value.canConsent || detail.value.canResubmit)
      formEditable.value =
        formRenderType.value === 'vue'
          ? !vueFormReadonly.value || schemaFormEditable.value
          : schemaFormEditable.value
      fieldStates.value = buildFieldStates(cfg, schemaFormEditable.value, pageSchema.value)
      formRenderKey.value += 1
      formReady.value = true
    }
  } finally {
    loading.value = false
  }
}

async function collectFormData() {
  return formBlockRef.value?.collectData({ ...(detail.value?.formData || {}) }) || {
    ...(detail.value?.formData || {}),
  }
}

async function confirmSecondOperate(action: string) {
  if (!detail.value?.secondOperatePrompt) return true
  try {
    await ElMessageBox.confirm(`确定要${action}吗？`, '二次确认', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    })
    return true
  } catch {
    return false
  }
}

async function onConsentConfirm(opinion: string) {
  if (!detail.value?.taskId) return
  if (!(await validateApproveForm())) return
  if (!(await confirmSecondOperate('同意该审批'))) return
  acting.value = true
  try {
    await consentApproveTask({
      taskId: detail.value.taskId,
      opinion,
      formData: await collectFormData(),
    })
    ElMessage.success('已同意')
    consentVisible.value = false
    router.replace('/flow/approve/pending')
  } finally {
    acting.value = false
  }
}

async function onResubmit() {
  if (!detail.value?.taskId) return
  if (!(await validateApproveForm())) return
  acting.value = true
  try {
    await resubmitApproveTask({
      taskId: detail.value.taskId,
      formData: await collectFormData(),
    })
    ElMessage.success('已重新提交')
    router.replace('/flow/approve/pending')
  } finally {
    acting.value = false
  }
}

async function onRejectConfirm(payload: {
  opinion: string
  rejectStrategy: number
  rejectNodeKey?: string
}) {
  if (!detail.value?.taskId) return
  if (!(await confirmSecondOperate('驳回该审批'))) return
  acting.value = true
  try {
    await rejectApproveTask({
      taskId: detail.value.taskId,
      opinion: payload.opinion,
      formData: await collectFormData(),
      rejectStrategy: payload.rejectStrategy,
      rejectNodeKey: payload.rejectNodeKey,
    })
    ElMessage.success('已处理')
    rejectVisible.value = false
    router.replace('/flow/approve/pending')
  } finally {
    acting.value = false
  }
}

async function onTransferConfirm(payload: { toUserId: number; toName: string; opinion: string }) {
  if (!detail.value?.taskId) return
  acting.value = true
  try {
    await transferApproveTask({
      taskId: detail.value.taskId,
      toUserId: payload.toUserId,
      toName: payload.toName,
      opinion: payload.opinion,
    })
    ElMessage.success('已转交')
    transferVisible.value = false
    router.replace('/flow/approve/pending')
  } finally {
    acting.value = false
  }
}

async function onAppendConfirm(payload: { toUserId: number; toName: string; position: number }) {
  if (!detail.value?.taskId) return
  acting.value = true
  try {
    await appendApproveActor({
      taskId: detail.value.taskId,
      toUserId: payload.toUserId,
      toName: payload.toName,
      position: payload.position,
    })
    ElMessage.success('已加签')
    appendVisible.value = false
    await load()
  } finally {
    acting.value = false
  }
}

async function onRemoveConfirm(payload: { actorId: number }) {
  if (!detail.value?.taskId) return
  acting.value = true
  try {
    await removeApproveActor({
      taskId: detail.value.taskId,
      actorId: payload.actorId,
    })
    ElMessage.success('已减签')
    removeVisible.value = false
    await load()
  } finally {
    acting.value = false
  }
}

async function onRollbackConfirm(payload: { targetNodeKey: string; opinion: string }) {
  if (!detail.value?.taskId) return
  acting.value = true
  try {
    await rollbackApproveTask({
      taskId: detail.value.taskId,
      targetNodeKey: payload.targetNodeKey,
      opinion: payload.opinion,
    })
    ElMessage.success('已回退')
    rollbackVisible.value = false
    router.replace('/flow/approve/pending')
  } finally {
    acting.value = false
  }
}

async function onRevokeConfirm(opinion: string) {
  if (!detail.value?.instanceId) return
  acting.value = true
  try {
    await revokeApproveInstance({
      instanceId: detail.value.instanceId,
      opinion,
    })
    ElMessage.success('已撤销')
    revokeVisible.value = false
    router.replace('/flow/approve/mine')
  } finally {
    acting.value = false
  }
}

async function onCcConfirm(payload: { users: { id: number; name: string }[] }) {
  if (!detail.value?.taskId) return
  acting.value = true
  try {
    await runtimeCcApproveTask({ taskId: detail.value.taskId, users: payload.users })
    ElMessage.success('已抄送')
    ccVisible.value = false
    await load()
  } finally {
    acting.value = false
  }
}

async function onCommentConfirm(payload: { content: string; mentionUserIds?: number[] }) {
  if (!detail.value?.instanceId) return
  acting.value = true
  try {
    await addApproveComment({
      instanceId: detail.value.instanceId,
      content: payload.content,
      mentionUserIds: payload.mentionUserIds,
    })
    ElMessage.success('评论已发表')
    commentVisible.value = false
    await load()
  } finally {
    acting.value = false
  }
}

onMounted(load)
</script>

<style lang="scss" scoped>
.detail-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 520px;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  .detail-id {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }
  .title {
    font-size: 16px;
    font-weight: 600;
  }
}
.body {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
.footer {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color-lighter);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  :deep(.el-button + .el-button) {
    margin-left: 0;
  }
}
</style>
