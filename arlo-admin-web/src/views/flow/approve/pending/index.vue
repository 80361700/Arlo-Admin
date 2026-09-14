<template>
  <div class="page-container pending-page" v-loading="listLoading">
    <div class="pending-row">
      <aside class="pending-left">
        <div class="left-pane">
          <div class="left-header">
            <div class="left-title">
              <span>待办列表</span>
              <div class="left-title-actions">
                <el-button
                  class="header-action-btn"
                  size="small"
                  :title="batchMode ? '取消批量' : '批量处理'"
                  @click="toggleBatchMode"
                >
                  {{ batchMode ? '取消' : '批量' }}
                </el-button>
                <el-button class="header-action-btn" size="small" title="委托设置" @click="openDelegate">
                  委托
                </el-button>
                <el-button class="header-action-btn is-icon" size="small" title="刷新" @click="reloadList">
                  <el-icon :size="14"><Refresh /></el-icon>
                </el-button>
                <ApproveAdvancedFilter v-model="advancedFilter" @search="applySearch" />
              </div>
            </div>
            <el-input
              v-model="keyword"
              clearable
              placeholder="搜索流程名称"
              :prefix-icon="Search"
              @keyup.enter="applySearch"
              @clear="applySearch"
            />
            <div v-if="activeDelegate?.enabled && activeDelegate.toUserName" class="delegate-hint">
              <span>
                已委托给 <strong>{{ activeDelegate.toUserName }}</strong> 代为处理
              </span>
              <el-button link type="primary" @click="openDelegate">修改</el-button>
            </div>
            <div v-if="batchMode" class="batch-bar">
              <el-checkbox
                :model-value="isAllBatchChecked"
                :indeterminate="isBatchIndeterminate"
                @change="toggleCheckAll"
              >
                全选可批量项
              </el-checkbox>
              <div class="batch-ops">
                <el-button size="small" type="primary" :disabled="!checkedTaskIds.length" :loading="acting" @click="doBatchConsent">
                  同意({{ checkedTaskIds.length }})
                </el-button>
                <el-button size="small" type="danger" :disabled="!checkedTaskIds.length" :loading="acting" @click="doBatchReject">
                  驳回
                </el-button>
              </div>
            </div>
          </div>

          <div
            class="task-list"
            v-infinite-scroll="loadMore"
            :infinite-scroll-disabled="listLoadDisabled"
            :infinite-scroll-distance="48"
          >
            <div
              v-for="row in filteredList"
              :key="`${row.taskId}-${row.actorId}`"
              class="task-card"
              :class="{
                active: !batchMode && selected?.taskId === row.taskId,
                'is-batch': batchMode,
              }"
              @click="batchMode ? toggleCheck(row) : selectTask(row)"
            >
              <div class="card-top">
                <div class="card-title-wrap">
                  <el-checkbox
                    v-if="batchMode"
                    class="card-check"
                    :model-value="checkedTaskIds.includes(row.taskId)"
                    :disabled="!canBatchRow(row)"
                    @click.stop
                    @change="toggleCheck(row)"
                  />
                  <div class="card-title" :title="row.processName">{{ row.processName }}</div>
                </div>
                <div class="card-tags">
                  <el-tag v-if="row.isSubProcess" size="small" type="primary" effect="plain">子流程</el-tag>
                  <el-tag v-if="row.isDelegate" size="small" type="info" effect="plain">代批</el-tag>
                  <el-tag size="small" type="warning" effect="plain">
                    {{ (row.nodeName || '').includes('修改') ? '待修改' : '待审批' }}
                  </el-tag>
                </div>
              </div>
              <div class="card-body">
                <div class="card-node">当前节点：{{ row.nodeName || '-' }}</div>
                <div class="card-meta">
                  <el-avatar :size="22" class="avatar">{{ (row.createBy || '?').charAt(0) }}</el-avatar>
                  <span class="meta-text">{{ row.createBy || '-' }}</span>
                </div>
                <div class="card-time">到达于 {{ row.createdAt || '-' }}</div>
              </div>
            </div>
            <el-empty v-if="!filteredList.length && !listLoading" description="暂无待办" :image-size="72" />
            <div v-if="loadingMore" class="list-load-tip">
              <el-icon class="is-loading"><Loading /></el-icon>
              <span>加载中...</span>
            </div>
          </div>
        </div>
      </aside>

      <section class="pending-right">
        <div class="right-pane">
          <template v-if="selected">
            <div v-loading="detailLoading" class="detail-wrap">
              <div class="detail-head">
                <div class="detail-id-row">
                  <div class="detail-id">编号：{{ detail?.instanceId || selected.instanceId }}</div>
                  <el-button class="print-btn is-icon" size="small" title="打印" @click="printVisible = true">
                    <el-icon :size="14"><Printer /></el-icon>
                  </el-button>
                </div>
                <div class="detail-title-row">
                  <h2 class="detail-title">{{ detail?.processName || selected.processName }}</h2>
                  <el-tag
                    v-if="detail?.isSubProcess || selected.isSubProcess"
                    size="small"
                    type="primary"
                    effect="plain"
                  >
                    {{ detail?.subProcessName || selected.subProcessName || '子流程' }}
                  </el-tag>
                  <el-tag size="small" type="warning">{{ stateLabel(detail?.instanceState) }}</el-tag>
                </div>
                <div class="detail-sub">
                  <el-avatar :size="24" class="avatar">
                    {{ (detail?.createBy || selected.createBy || '?').charAt(0) }}
                  </el-avatar>
                  <span>
                    {{ detail?.createBy || selected.createBy }} 提交于
                    {{ detail?.createdAt || selected.createdAt }}
                  </span>
                  <span v-if="detail?.currentNodeName || selected.nodeName" class="detail-cur-node">
                    当前：{{
                      detail?.workNodeName
                        ? `${detail.currentNodeName || '子流程'} · ${detail.workNodeName}`
                        : detail?.currentNodeName || selected.nodeName
                    }}
                  </span>
                </div>
              </div>

              <el-tabs v-model="activeTab" class="detail-tabs">
                <el-tab-pane label="审批信息" name="info">
                  <div class="tab-scroll">
                    <div class="block">
                      <ApproveFormBlock
                        ref="formBlockRef"
                        :form-ready="formReady"
                        :loading="detailLoading"
                        :form-render-type="formRenderType"
                        :form-component="formComponent"
                        :form-data="formData"
                        :page-schema="pageSchema"
                        :field-states="fieldStates"
                        :render-key="formRenderKey"
                        :readonly="vueFormReadonly"
                        :disabled="!schemaFormEditable"
                      />
                    </div>
                  </div>
                </el-tab-pane>

                <el-tab-pane label="流转记录" name="timeline" lazy>
                  <div class="tab-scroll">
                    <div class="block">
                      <FlowTimeline
                        :items="detail?.timeline || []"
                        :comments="detail?.comments || []"
                      />
                    </div>
                  </div>
                </el-tab-pane>

                <el-tab-pane label="流程图" name="diagram" lazy>
                  <div class="diagram-wrap">
                    <div class="diagram-canvas">
                      <div class="diagram-legend">
                        <span><i class="dot done" />已执行</span>
                        <span><i class="dot active" />执行中</span>
                        <span><i class="dot pending" />未执行</span>
                      </div>
                      <div class="diagram-scroll">
                        <FlowProcess
                        v-if="flowModel"
                        :data="flowModel"
                        :flow="{}"
                        readonly
                        :node-states="nodeRunStates"
                        :node-actors="nodeRunActors"
                      />
                        <el-empty v-else description="暂无流程图" :image-size="64" />
                      </div>
                    </div>
                  </div>
                </el-tab-pane>
              </el-tabs>

              <div v-if="hasFooterActions" class="detail-footer">
                <div class="footer-actions">
                  <el-button
                    v-if="detail?.canComment"
                    :loading="acting"
                    @click="commentVisible = true"
                  >
                    评论
                  </el-button>
                  <el-button
                    v-if="detail?.canReject"
                    v-permission="'flow:approve:handle'"
                    type="danger"
                    :loading="acting"
                    @click="openReject"
                  >
                    {{ rejectBtnLabel }}
                  </el-button>
                  <el-button
                    v-if="detail?.canConsent"
                    v-permission="'flow:approve:handle'"
                    type="primary"
                    :loading="acting"
                    @click="openConsent"
                  >
                    同意
                  </el-button>
                  <el-button
                    v-if="detail?.canResubmit"
                    v-permission="'flow:approve:todo'"
                    type="primary"
                    :loading="acting"
                    @click="onResubmit"
                  >
                    重新提交
                  </el-button>
                  <el-dropdown
                    v-if="moreActions.length"
                    v-permission="'flow:approve:handle'"
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
              </div>
              <div v-else class="detail-footer tip">当前任务仅可查看</div>
            </div>
          </template>
          <el-empty v-else class="right-empty" description="请选择左侧待办查看详情" :image-size="96" />
        </div>
      </section>
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
    <ApprovePrintDialog
      v-model="printVisible"
      :process-name="detail?.processName || selected?.processName"
      :instance-id="detail?.instanceId || selected?.instanceId"
      :created-at="detail?.createdAt || selected?.createdAt"
      :form-ready="formReady"
      :form-render-type="formRenderType"
      :form-component="formComponent"
      :form-data="formData"
      :page-schema="pageSchema"
      :field-states="fieldStates"
      :render-key="formRenderKey"
    />

    <el-dialog v-model="delegateVisible" title="审批委托" width="480px" destroy-on-close append-to-body>
      <el-form label-width="88px">
        <el-form-item label="启用委托">
          <el-switch v-model="delegateForm.enabled" />
        </el-form-item>
        <template v-if="delegateForm.enabled">
          <el-form-item label="受托人" required>
            <UserSelect
              v-model="delegateForm.toUserId"
              placeholder="选择代批人"
              :exclude-ids="authStore.userInfo?.id ? [authStore.userInfo.id] : []"
              @change="onDelegateUserChange"
            />
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="delegateForm.remark" maxlength="100" show-word-limit placeholder="可选" />
          </el-form-item>
          <el-alert
            type="info"
            :closable="false"
            show-icon
            title="启用后，你的待办会出现在受托人的待办列表中，由其代批。流程设置关闭「允许委托」的实例不会进入代批，也无法代批办理。"
          />
        </template>
      </el-form>
      <template #footer>
        <el-button @click="delegateVisible = false">取消</el-button>
        <el-button type="primary" :loading="acting" @click="saveDelegate">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="batchOpinionVisible"
      :title="batchOpinionMode === 'consent' ? '批量同意' : '批量驳回'"
      width="520px"
      destroy-on-close
      append-to-body
    >
      <el-form label-position="top" @submit.prevent>
        <el-form-item required label="审批意见">
          <el-input
            v-model="batchOpinion"
            type="textarea"
            :rows="4"
            maxlength="64"
            show-word-limit
            placeholder="请输入内容"
          />
        </el-form-item>
        <el-form-item v-if="batchOpinionMode === 'consent'" label="推荐回复">
          <div class="batch-quick-replies">
            <button
              v-for="t in batchQuickReplies"
              :key="t"
              type="button"
              class="batch-quick-btn"
              @click="batchOpinion = t"
            >
              {{ t }}
            </button>
          </div>
          <el-alert
            style="margin-top: 10px"
            type="info"
            :closable="false"
            show-icon
            title="若某条待办在当前节点有未填的可写必填项，该条会失败，请单独打开填写后再批。"
          />
        </el-form-item>
        <el-alert
          v-if="batchOpinionMode === 'reject' && !batchRejectTerminate"
          type="info"
          :closable="false"
          show-icon
          title="默认按各节点驳回策略执行；若节点为「指定驳回目标」且未预置目标，该条会失败"
        />
        <el-form-item v-if="batchOpinionMode === 'reject'" label="">
          <el-checkbox v-model="batchRejectTerminate">统一终止流程</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchOpinionVisible = false">取消</el-button>
        <el-button
          :type="batchOpinionMode === 'consent' ? 'primary' : 'danger'"
          :loading="acting"
          @click="submitBatchOpinion"
        >
          确定（{{ checkedTaskIds.length }}）
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading, Printer, Refresh, Search } from '@element-plus/icons-vue'
import { type FieldStates, type PageSchema } from 'epic-designer'
import FlowProcess from '@/components/flowProcess/index.vue'
import FlowTimeline from '../components/FlowTimeline.vue'
import ApproveFormBlock from '../components/ApproveFormBlock.vue'
import ApproveConsentDialog from '../components/ApproveConsentDialog.vue'
import ApproveRejectDialog from '../components/ApproveRejectDialog.vue'
import ApproveTransferDialog from '../components/ApproveTransferDialog.vue'
import ApproveAppendDialog from '../components/ApproveAppendDialog.vue'
import ApproveRemoveDialog from '../components/ApproveRemoveDialog.vue'
import ApproveRollbackDialog from '../components/ApproveRollbackDialog.vue'
import ApproveCcDialog from '../components/ApproveCcDialog.vue'
import ApproveCommentDialog from '../components/ApproveCommentDialog.vue'
import ApproveAdvancedFilter from '../components/ApproveAdvancedFilter.vue'
import ApprovePrintDialog from '../components/ApprovePrintDialog.vue'
import UserSelect from '@/components/UserSelect.vue'
import { useAuthStore } from '@/stores/auth'
import {
  addApproveComment,
  appendApproveActor,
  batchConsentApproveTasks,
  batchRejectApproveTasks,
  consentApproveTask,
  getApproveDelegate,
  getApproveInstance,
  getPendingTasks,
  rejectApproveTask,
  removeApproveActor,
  resubmitApproveTask,
  rollbackApproveTask,
  runtimeCcApproveTask,
  setApproveDelegate,
  transferApproveTask,
  type InstanceDetail,
  type PendingTaskItem,
} from '@/api'
import {
  buildFieldStates,
  canEditApproveForm,
  normalizeProcessForm,
  pageSchemaHasFields,
} from '@/views/flow/process/components/formSchema'
import { waitAtLeast } from '@/utils/waitAtLeast'
import { buildNodeRunStates } from '../utils/nodeRunStates'
import { approveFilterParams, emptyApproveAdvancedFilter } from '../utils/dateRange'

const authStore = useAuthStore()
const listLoading = ref(false)
const loadingMore = ref(false)
const detailLoading = ref(false)
const acting = ref(false)
const list = ref<PendingTaskItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const appliedKeyword = ref('')
const advancedFilter = ref(emptyApproveAdvancedFilter())
const selected = ref<PendingTaskItem | null>(null)
const batchMode = ref(false)
const checkedTaskIds = ref<number[]>([])
const batchOpinionVisible = ref(false)
const batchOpinionMode = ref<'consent' | 'reject'>('consent')
const batchOpinion = ref('')
const batchRejectTerminate = ref(false)
const batchQuickReplies = ['同意', '已阅', '收到', '已核对', '合格', '情况属实', '确认']
const delegateVisible = ref(false)
const delegateForm = ref({
  toUserId: undefined as number | undefined,
  toUserName: '',
  enabled: false,
  remark: '',
})
/** 当前生效的委托（用于列表提示） */
const activeDelegate = ref<{ enabled: boolean; toUserId?: number; toUserName: string } | null>(null)
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
const activeTab = ref('info')
const consentVisible = ref(false)
const rejectVisible = ref(false)
const transferVisible = ref(false)
const appendVisible = ref(false)
const removeVisible = ref(false)
const rollbackVisible = ref(false)
const ccVisible = ref(false)
const commentVisible = ref(false)
const printVisible = ref(false)

const filteredList = computed(() => list.value)

const listLoadDisabled = computed(
  () => listLoading.value || loadingMore.value || list.value.length >= total.value,
)

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
  return list
})

const hasFooterActions = computed(() => {
  const d = detail.value
  if (!d) return false
  return !!(d.canConsent || d.canReject || d.canResubmit || d.canComment || moreActions.value.length)
})

/** 同意弹窗：下一节点预览（流转记录风格） */
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

const flowModel = computed(() => {
  const mc = detail.value?.modelContent
  if (!mc || typeof mc !== 'object') return null
  if (mc.nodeConfig) return mc
  return { nodeConfig: mc }
})

/** 流程图节点运行态：已执行 / 执行中 / 未执行 */
const nodeRunStates = computed(() => buildNodeRunStates(detail.value))

/** 流程图办理人：当前节点按任务 weight 排序（前加签在前）；其它节点按流转记录 */
const nodeRunActors = computed(() => {
  const map: Record<string, { name: string; actorType?: number }[]> = {}
  const d = detail.value
  if (!d) return map

  const push = (key: string, name: string, actorType?: number) => {
    if (!key || !name) return
    if (!map[key]) map[key] = []
    const hit = map[key].find((x) => x.name === name)
    if (hit) {
      if (actorType != null) hit.actorType = actorType
      return
    }
    map[key].push({ name, actorType })
  }

  for (const t of d.timeline || []) {
    if (Number(t.nodeType) === 0 || Number(t.nodeType) === -1) continue
    if (d.currentNodeKey && t.nodeKey === d.currentNodeKey) continue
    push(t.nodeKey, t.actorName)
  }

  if (d.currentNodeKey) {
    const actors = [...(d.taskActors || [])].sort(
      (a, b) => Number(a.weight) - Number(b.weight) || Number(a.id) - Number(b.id),
    )
    for (const a of actors) {
      if (Number(a.actorState) === 4) continue // skip
      push(d.currentNodeKey, a.actorName, a.actorType)
    }
  }

  return map
})

function stateLabel(s?: number) {
  if (s === undefined || s === null) return '待审批'
  return ['审批中', '已通过', '已拒绝', '已撤销', '已终止', '已超时'][s] || '待审批'
}

function isReviseRow(row: PendingTaskItem) {
  return (row.nodeName || '').includes('修改')
}

function canBatchRow(row: PendingTaskItem) {
  return !!(row.allowBatchOperate && row.canHandle && !isReviseRow(row))
}

const batchableList = computed(() => filteredList.value.filter(canBatchRow))

const isAllBatchChecked = computed(() => {
  const ids = batchableList.value.map((r) => r.taskId)
  return ids.length > 0 && ids.every((id) => checkedTaskIds.value.includes(id))
})

const isBatchIndeterminate = computed(() => {
  const ids = batchableList.value.map((r) => r.taskId)
  const n = ids.filter((id) => checkedTaskIds.value.includes(id)).length
  return n > 0 && n < ids.length
})

function toggleBatchMode() {
  batchMode.value = !batchMode.value
  checkedTaskIds.value = []
}

function toggleCheck(row: PendingTaskItem) {
  if (!canBatchRow(row)) return
  const id = row.taskId
  const i = checkedTaskIds.value.indexOf(id)
  if (i >= 0) checkedTaskIds.value.splice(i, 1)
  else checkedTaskIds.value.push(id)
}

function toggleCheckAll(val: boolean | string | number) {
  if (val) {
    checkedTaskIds.value = batchableList.value.map((r) => r.taskId)
  } else {
    checkedTaskIds.value = []
  }
}

function doBatchConsent() {
  if (!checkedTaskIds.value.length) return
  batchOpinionMode.value = 'consent'
  batchOpinion.value = '同意'
  batchOpinionVisible.value = true
}

function doBatchReject() {
  if (!checkedTaskIds.value.length) return
  batchOpinionMode.value = 'reject'
  batchOpinion.value = '驳回'
  batchRejectTerminate.value = false
  batchOpinionVisible.value = true
}

async function submitBatchOpinion() {
  const opinion = batchOpinion.value.trim()
  if (!opinion) {
    ElMessage.warning('请填写审批意见')
    return
  }
  const ids = [...checkedTaskIds.value]
  if (!ids.length) return
  acting.value = true
  try {
    const res =
      batchOpinionMode.value === 'consent'
        ? await batchConsentApproveTasks({ taskIds: ids, opinion })
        : await batchRejectApproveTasks({
            taskIds: ids,
            opinion,
            rejectStrategy: batchRejectTerminate.value ? 4 : 0,
          })
    const data = res.data
    const ok = data?.success || 0
    const fail = data?.failed || 0
    const action = batchOpinionMode.value === 'consent' ? '同意' : '驳回'
    if (fail) {
      ElMessage.warning(`成功 ${ok} 条，失败 ${fail} 条${data?.errors?.[0] ? `：${data.errors[0]}` : ''}`)
    } else {
      ElMessage.success(`已批量${action} ${ok} 条`)
    }
    batchOpinionVisible.value = false
    checkedTaskIds.value = []
    batchMode.value = false
    await loadList(false)
  } finally {
    acting.value = false
  }
}

async function loadActiveDelegate() {
  try {
    const res = await getApproveDelegate()
    const d = res.data
    if (d?.enabled && d.toUserId) {
      activeDelegate.value = {
        enabled: true,
        toUserId: d.toUserId,
        toUserName: d.toUserName || `用户#${d.toUserId}`,
      }
    } else {
      activeDelegate.value = null
    }
  } catch {
    activeDelegate.value = null
  }
}

async function openDelegate() {
  delegateVisible.value = true
  try {
    const res = await getApproveDelegate()
    const d = res.data
    const enabled = !!(d?.enabled && d.toUserId)
    delegateForm.value = {
      toUserId: enabled ? d.toUserId : undefined,
      toUserName: enabled ? d.toUserName || '' : '',
      enabled,
      remark: d?.remark || '',
    }
  } catch {
    delegateForm.value = { toUserId: undefined, toUserName: '', enabled: false, remark: '' }
  }
}

function onDelegateUserChange(id: number | undefined, name?: string) {
  delegateForm.value.toUserName = name || ''
  if (!id) delegateForm.value.toUserId = undefined
}

async function saveDelegate() {
  if (delegateForm.value.enabled && !delegateForm.value.toUserId) {
    ElMessage.warning('请选择受托人')
    return
  }
  if (
    delegateForm.value.enabled &&
    authStore.userInfo?.id &&
    delegateForm.value.toUserId === authStore.userInfo.id
  ) {
    ElMessage.warning('不能委托给自己')
    return
  }
  acting.value = true
  try {
    if (!delegateForm.value.enabled) {
      await setApproveDelegate({ clear: true })
      ElMessage.success('已关闭并清除委托')
      activeDelegate.value = null
    } else {
      await setApproveDelegate({
        toUserId: delegateForm.value.toUserId,
        toUserName: delegateForm.value.toUserName,
        enabled: true,
        remark: delegateForm.value.remark,
      })
      ElMessage.success('委托已保存')
      await loadActiveDelegate()
    }
    delegateVisible.value = false
    await loadList(false)
  } finally {
    acting.value = false
  }
}

function applySearch() {
  appliedKeyword.value = keyword.value
  reloadList()
}

async function loadList(append = false) {
  const started = Date.now()
  if (append) {
    loadingMore.value = true
    await nextTick()
  } else {
    listLoading.value = true
  }
  try {
    if (!append) page.value = 1
    const res = await getPendingTasks({
      page: page.value,
      pageSize: pageSize.value,
      keyword: appliedKeyword.value.trim() || undefined,
      ...approveFilterParams(advancedFilter.value),
    })
    const rows = res.data?.list || []
    total.value = res.data?.total || 0
    list.value = append ? [...list.value, ...rows] : rows
    if (!append) {
      if (list.value.length) {
        const keep = selected.value
          ? list.value.find((r) => r.taskId === selected.value?.taskId)
          : null
        await selectTask(keep || list.value[0])
      } else {
        selected.value = null
        detail.value = null
        pageSchema.value = null
        formData.value = {}
        fieldStates.value = []
        formReady.value = false
      }
    }
  } finally {
    if (append) {
      await waitAtLeast(started, 400)
      loadingMore.value = false
    } else {
      listLoading.value = false
    }
  }
}

function reloadList() {
  loadList(false)
}

function loadMore() {
  if (listLoadDisabled.value) return
  page.value += 1
  loadList(true)
}

async function selectTask(row: PendingTaskItem, opts?: { keepTab?: boolean }) {
  selected.value = row
  if (!opts?.keepTab) activeTab.value = 'info'
  consentVisible.value = false
  rejectVisible.value = false
  transferVisible.value = false
  appendVisible.value = false
  removeVisible.value = false
  rollbackVisible.value = false
  ccVisible.value = false
  commentVisible.value = false
  // 先卸表单，避免切换时把上一单 merge 进来
  formReady.value = false
  formEditable.value = false
  schemaFormEditable.value = false
  vueFormReadonly.value = true
  pageSchema.value = null
  formData.value = {}
  fieldStates.value = []
  formRenderType.value = 'designer'
  formComponent.value = ''
  detailLoading.value = true
  try {
    const res = await getApproveInstance(row.instanceId, row.taskId)
    detail.value = res.data || null
    if (detail.value) {
      formData.value = { ...(detail.value.formData || {}) }
      formRenderType.value = detail.value.formRenderType === 'vue' ? 'vue' : 'designer'
      formComponent.value = detail.value.formComponent || ''
      // 系统表单(vue)时仍保留 formSchema：节点子表单（设计表单）需单独渲染
      pageSchema.value = normalizeProcessForm(detail.value.formSchema)
      // 发起人改单：历史全只读配置视为可编辑（与发起页一致）
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
      // 系统主表：有子表单时审批只读（改单可写）；无子表单时沿用可办状态
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
    } else {
      pageSchema.value = null
      formData.value = {}
      fieldStates.value = []
      formEditable.value = false
      schemaFormEditable.value = false
      vueFormReadonly.value = true
    }
  } finally {
    detailLoading.value = false
  }
}

async function collectFormData() {
  return formBlockRef.value?.collectData({ ...(detail.value?.formData || {}) }) || {
    ...(detail.value?.formData || {}),
  }
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
  ccVisible.value = false
  commentVisible.value = false
}

function onMoreCommand(cmd: string) {
  closeActionDialogs()
  if (cmd === 'transfer') transferVisible.value = true
  else if (cmd === 'append') appendVisible.value = true
  else if (cmd === 'remove') removeVisible.value = true
  else if (cmd === 'rollback') rollbackVisible.value = true
  else if (cmd === 'cc') ccVisible.value = true
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
    await loadList(false)
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
    await loadList(false)
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
    await loadList(false)
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
    await loadList(false)
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
    await selectTask(selected.value!)
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
    await selectTask(selected.value!)
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
    await loadList(false)
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
    await selectTask(selected.value!)
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
    if (selected.value) await selectTask(selected.value, { keepTab: true })
  } finally {
    acting.value = false
  }
}

watch(filteredList, (rows) => {
  if (selected.value && !rows.some((r) => r.taskId === selected.value?.taskId)) {
    if (rows[0]) selectTask(rows[0])
    else {
      selected.value = null
      detail.value = null
    }
  }
})

let skipKeepAliveActivate = true
onMounted(() => {
  loadActiveDelegate()
  loadList(false)
})
onActivated(() => {
  if (skipKeepAliveActivate) {
    skipKeepAliveActivate = false
    return
  }
  loadActiveDelegate()
  loadList(false)
})
</script>

<style lang="scss" scoped>
.pending-page {
  position: relative;
  height: 100%;
  min-height: 560px;
  background: var(--el-bg-color);
}
.pending-row {
  display: flex;
  align-items: stretch;
  height: 100%;
  min-height: 560px;
  gap: 16px;
}
.pending-left {
  width: 380px;
  flex: 0 0 380px;
  min-width: 380px;
  max-width: 380px;
  height: 100%;
}
.pending-right {
  flex: 1;
  min-width: 0;
  height: 100%;
}

.left-pane {
  height: 100%;
  padding-right: 16px;
  border-right: 1px solid var(--el-border-color-lighter);
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
}
.left-header {
  position: sticky;
  top: 0;
  z-index: 2;
  padding-bottom: 16px;
  background: var(--el-bg-color);
}
.left-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  span {
    font-size: 16px;
    font-weight: 500;
  }
  .left-title-actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .header-action-btn {
    margin: 0;
    height: 28px;
    padding: 0 10px;
    font-size: 13px;
    font-weight: 400;
    color: var(--el-text-color-regular);
    background: #fff;
    border: 1px solid var(--el-border-color);
    &:hover,
    &:focus {
      color: var(--el-color-primary);
      border-color: var(--el-color-primary-light-5);
      background: var(--el-color-primary-light-9);
    }
    &.is-icon {
      width: 28px;
      padding: 0;
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }
  }
}
.batch-bar {
  margin-top: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  .batch-ops {
    display: flex;
    gap: 6px;
  }
}
.delegate-hint {
  margin-top: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 10px;
  font-size: 13px;
  line-height: 1.4;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  border-radius: 6px;
  strong {
    font-weight: 600;
  }
}
.task-list {
  flex: 1;
  overflow: auto;
  padding-right: 4px;
}
.task-card {
  padding: 14px 16px;
  margin-bottom: 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;

  &:hover {
    background: rgba(96, 98, 102, 0.04);
  }
  &.active {
    background: rgba(64, 158, 255, 0.08);
    border-color: var(--el-color-primary-light-5);
  }
  &.is-batch .card-body {
    /* 勾选框 14px + 间距 8px，与标题文字左缘对齐 */
    padding-left: 22px;
  }
}
.card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}
.card-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex: 1;
}
.card-check {
  flex-shrink: 0;
  height: 20px;
  :deep(.el-checkbox__inner) {
    vertical-align: middle;
  }
}
.card-tags {
  display: flex;
  flex-direction: row;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 4px;
  flex-shrink: 0;
}
.card-title {
  flex: 1;
  min-width: 0;
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.4;
}
.card-body {
  min-width: 0;
}
.card-node {
  margin-top: 10px;
  font-size: 13px;
  color: var(--el-text-color-regular);
}
.card-meta {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.meta-text {
  font-size: 13px;
  color: var(--el-text-color-regular);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.card-time {
  margin-top: 6px;
  padding-left: 30px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.avatar {
  background: var(--el-color-primary);
  color: #fff;
  font-size: 11px;
  flex-shrink: 0;
}
.list-load-tip {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 0 14px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.right-pane {
  height: 100%;
  min-width: 0;
  display: flex;
  flex-direction: column;
  padding-left: 4px;
}
.right-empty {
  margin: auto;
}
.detail-wrap {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.detail-head {
  padding: 8px 0px 16px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}
.detail-id-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.detail-id {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.print-btn {
  margin: 0;
  height: 28px;
  width: 28px;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--el-color-primary);
  background: #fff;
  border: 1px solid var(--el-border-color);
  &:hover,
  &:focus {
    color: var(--el-color-primary);
    border-color: var(--el-color-primary-light-5);
    background: var(--el-color-primary-light-9);
  }
}
.detail-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 10px;
}
.detail-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.detail-sub {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.detail-cur-node {
  margin-left: 4px;
  color: var(--el-text-color-regular);
}
.detail-tabs {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  margin-top: 8px;

  :deep(.el-tabs__header) {
    margin-bottom: 0;
  }
  :deep(.el-tabs__content) {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }
  :deep(.el-tab-pane) {
    height: 100%;
  }
}
.tab-scroll {
  height: 100%;
  overflow: auto;
  padding: 24px 0px;
}
.block {
  margin-bottom: 0;
}
.diagram-wrap {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 16px 0px;
}
.diagram-canvas {
  position: relative;
  flex: 1;
  min-height: 400px;
  border: 1px solid var(--el-border-color-lighter);
  // border-radius: 8px;
  overflow: hidden;
  background: var(--el-fill-color-blank);
}
.diagram-scroll {
  height: 100%;
  overflow: auto;
}
.diagram-legend {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 5;
  display: flex;
  gap: 16px;
  padding: 6px 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  pointer-events: none;
  .dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 6px;
    &.done {
      background: #67c23a;
    }
    &.active {
      background: #e6a23c;
    }
    &.pending {
      background: #c0c4cc;
    }
  }
}
.detail-footer {
  border-top: 1px solid var(--el-border-color-lighter);
  padding: 16px 0px 0;
  background: #fff;
  &.tip {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }
}
.footer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  :deep(.el-button + .el-button) {
    margin-left: 0;
  }
}
.batch-quick-replies {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.batch-quick-btn {
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

@media (max-width: 992px) {
  .pending-row {
    flex-direction: column;
  }
  .pending-left {
    width: 100%;
    flex: none;
    min-width: 0;
    max-width: none;
    height: auto;
  }
  .pending-right {
    width: 100%;
  }
  .left-pane {
    height: auto;
    max-height: 320px;
    padding-right: 0;
    border-right: 0;
    margin-bottom: 16px;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }
}
</style>
