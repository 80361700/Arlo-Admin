<template>
  <div class="page-container pending-page" v-loading="listLoading">
    <div class="pending-row">
      <aside class="pending-left">
        <div class="left-pane">
          <div class="left-header">
            <div class="left-title">
              <span>已审批</span>
              <div class="left-title-actions">
                <el-button class="header-action-btn is-icon" size="small" title="刷新" @click="reloadList">
                  <el-icon :size="14"><Refresh /></el-icon>
                </el-button>
                <ApproveAdvancedFilter v-model="advancedFilter" :default-days="90" @search="applySearch" />
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
          </div>

          <div
            class="task-list"
            v-infinite-scroll="loadMore"
            :infinite-scroll-disabled="listLoadDisabled"
            :infinite-scroll-distance="48"
          >
            <div
              v-for="row in filteredList"
              :key="rowKey(row)"
              class="task-card"
              :class="{ active: isSelected(row) }"
              @click="selectRow(row)"
            >
              <div class="card-top">
                <div class="card-title" :title="row.processName">{{ row.processName }}</div>
                <div class="card-tags">
                  <el-tag v-if="row.isSubProcess" size="small" type="primary" effect="plain">子流程</el-tag>
                  <el-tag v-if="row.isDelegate" size="small" type="info" effect="plain">代批</el-tag>
                  <el-tag size="small" :type="stateTagType(row.instanceState)" effect="plain">
                    {{ stateLabel(row.instanceState) }}
                  </el-tag>
                </div>
              </div>
              <div class="card-node">处理节点：{{ row.nodeName || '-' }}</div>
              <div class="card-meta">
                <el-avatar :size="22" class="avatar">{{ (row.createBy || '?').charAt(0) }}</el-avatar>
                <span class="meta-text">{{ row.createBy || '-' }}</span>
              </div>
              <div class="card-time">
                {{ row.finishTime ? `处理于 ${row.finishTime}` : `提交于 ${row.createdAt || '-'}` }}
              </div>
            </div>
            <el-empty v-if="!filteredList.length && !listLoading" description="暂无已审批" :image-size="72" />
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
                  <el-tag size="small" :type="stateTagType(detail?.instanceState ?? selected.instanceState)">
                    {{ stateLabel(detail?.instanceState ?? selected.instanceState) }}
                  </el-tag>
                </div>
                <div class="detail-sub">
                  <el-avatar :size="24" class="avatar">
                    {{ (detail?.createBy || selected.createBy || '?').charAt(0) }}
                  </el-avatar>
                  <span>
                    {{ detail?.createBy || selected.createBy }} 提交于
                    {{ detail?.createdAt || selected.createdAt }}
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
                        :readonly="true"
                        :disabled="true"
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

              <div v-if="detail?.canComment" class="detail-footer">
                <div class="footer-actions">
                  <el-button :loading="acting" @click="commentVisible = true">评论</el-button>
                </div>
              </div>
            </div>
          </template>
          <el-empty v-else class="right-empty" description="请选择左侧记录查看详情" :image-size="96" />
        </div>
      </section>
    </div>

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
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading, Printer, Refresh, Search } from '@element-plus/icons-vue'
import { type FieldStates, type PageSchema } from 'epic-designer'
import FlowProcess from '@/components/flowProcess/index.vue'
import FlowTimeline from '../components/FlowTimeline.vue'
import ApproveFormBlock from '../components/ApproveFormBlock.vue'
import ApproveAdvancedFilter from '../components/ApproveAdvancedFilter.vue'
import ApproveCommentDialog from '../components/ApproveCommentDialog.vue'
import ApprovePrintDialog from '../components/ApprovePrintDialog.vue'
import {
  addApproveComment,
  getApproveInstance,
  getApprovedList,
  type ApproveInstanceListItem,
  type InstanceDetail,
} from '@/api'
import {
  buildFieldStates,
  normalizeProcessForm,
} from '@/views/flow/process/components/formSchema'
import { waitAtLeast } from '@/utils/waitAtLeast'
import { buildNodeRunStates } from '../utils/nodeRunStates'
import {
  approveFilterParams,
  defaultApproveAdvancedFilter,
} from '../utils/dateRange'

const listLoading = ref(false)
const loadingMore = ref(false)
const detailLoading = ref(false)
const acting = ref(false)
const list = ref<ApproveInstanceListItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const appliedKeyword = ref('')
const advancedFilter = ref(defaultApproveAdvancedFilter(90))
const selected = ref<ApproveInstanceListItem | null>(null)
const detail = ref<InstanceDetail | null>(null)
const pageSchema = ref<PageSchema | null>(null)
const formData = ref<Record<string, any>>({})
const fieldStates = ref<FieldStates>([])
const formReady = ref(false)
const formRenderKey = ref(0)
const formBlockRef = ref<InstanceType<typeof ApproveFormBlock>>()
const formRenderType = ref('designer')
const formComponent = ref('')
const activeTab = ref('info')
const commentVisible = ref(false)
const printVisible = ref(false)

const filteredList = computed(() => list.value)

const listLoadDisabled = computed(
  () => listLoading.value || loadingMore.value || list.value.length >= total.value,
)

const flowModel = computed(() => {
  const mc = detail.value?.modelContent
  if (!mc || typeof mc !== 'object') return null
  if (mc.nodeConfig) return mc
  return { nodeConfig: mc }
})

const nodeRunStates = computed(() => buildNodeRunStates(detail.value))

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
      if (Number(a.actorState) === 4) continue
      push(d.currentNodeKey, a.actorName, a.actorType)
    }
  }
  return map
})

function rowKey(row: ApproveInstanceListItem) {
  // 历史办结 vs 节点未结束但本人已办：表不同，id 可能撞号
  if (row.hisTaskId) return `h-${row.hisTaskId}-${row.id || 0}`
  if (row.taskId && row.id) return `t-${row.taskId}-${row.id}`
  if (row.id) return `a-${row.id}`
  return `${row.instanceId}-${row.taskId || 0}-${row.nodeName || ''}`
}

function isSelected(row: ApproveInstanceListItem) {
  if (!selected.value) return false
  return rowKey(row) === rowKey(selected.value)
}

function stateLabel(s?: number) {
  if (s === undefined || s === null) return '-'
  return ['审批中', '已通过', '已拒绝', '已撤销', '已终止', '已超时'][s] || '-'
}

function stateTagType(s?: number): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  switch (Number(s)) {
    case 0:
      return 'primary'
    case 1:
      return 'success'
    case 2:
    case 4:
      return 'danger'
    case 3:
      return 'warning'
    default:
      return 'info'
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
    const res = await getApprovedList({
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
          ? list.value.find((r) => rowKey(r) === rowKey(selected.value!))
          : null
        await selectRow(keep || list.value[0])
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

async function selectRow(row: ApproveInstanceListItem, opts?: { keepTab?: boolean }) {
  selected.value = row
  if (!opts?.keepTab) activeTab.value = 'info'
  commentVisible.value = false
  formReady.value = false
  pageSchema.value = null
  formData.value = {}
  fieldStates.value = []
  formRenderType.value = 'designer'
  formComponent.value = ''
  detailLoading.value = true
  try {
    const res = await getApproveInstance(row.instanceId)
    detail.value = res.data || null
    if (detail.value) {
      formData.value = { ...(detail.value.formData || {}) }
      formRenderType.value = detail.value.formRenderType === 'vue' ? 'vue' : 'designer'
      formComponent.value = detail.value.formComponent || ''
      pageSchema.value =
        normalizeProcessForm(detail.value.formSchema)
      fieldStates.value = buildFieldStates(detail.value.formConfig || [], false, pageSchema.value)
      formRenderKey.value += 1
      formReady.value = true
    }
  } finally {
    detailLoading.value = false
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
    if (selected.value) await selectRow(selected.value, { keepTab: true })
  } finally {
    acting.value = false
  }
}

watch(filteredList, (rows) => {
  if (selected.value && !rows.some((r) => rowKey(r) === rowKey(selected.value!))) {
    if (rows[0]) selectRow(rows[0])
    else {
      selected.value = null
      detail.value = null
    }
  }
})

let skipKeepAliveActivate = true
onMounted(() => loadList(false))
onActivated(() => {
  if (skipKeepAliveActivate) {
    skipKeepAliveActivate = false
    return
  }
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
  height: calc(100% - 8px);
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
}
.card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
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
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.4;
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
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
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
}
.footer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  :deep(.el-button + .el-button) {
    margin-left: 0;
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
