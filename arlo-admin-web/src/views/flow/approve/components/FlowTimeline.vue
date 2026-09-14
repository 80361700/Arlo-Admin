<template>
  <div v-if="merged.length" class="rich-timeline">
    <div v-for="(t, i) in merged" :key="rowKey(t, i)" class="rt-item">
      <div class="rt-rail">
        <div class="rt-icon-wrap">
          <div class="rt-icon" :class="iconTone(t)">
            <el-icon :size="18"><component :is="mainIcon(t)" /></el-icon>
          </div>
          <div class="rt-badge" :class="badgeTone(t)">
            <el-icon :size="10"><component :is="badgeIcon(t)" /></el-icon>
          </div>
        </div>
        <div v-if="i < merged.length - 1" class="rt-line" />
      </div>

      <div class="rt-body">
        <div class="rt-node">{{ t.nodeName || '-' }}</div>
        <div v-if="t.actorName" class="rt-actor">
          <div class="rt-actor-row">
            <FlowNodeAvatar :name="t.actorName" :size="22" />
            <span v-if="t.agentName && t.agentName !== t.actorName" class="rt-agent">（{{ t.agentName }}代批）</span>
          </div>
        </div>
        <div v-if="isSubProcess(t) && t.childInstanceId" class="rt-sub">
          <span>{{ t.opinion || '子流程' }}</span>
          <button type="button" class="rt-sub-link" @click="openSub(t)">
            （{{ t.childProcessName || t.nodeName || '查看子流程' }}）
          </button>
        </div>
        <div v-else-if="t.opinion" class="rt-opinion" :class="{ 'is-comment': isComment(t) }">
          {{ t.opinion }}
        </div>
        <div v-if="displayTime(t)" class="rt-time">{{ displayTime(t) }}</div>
      </div>
    </div>
  </div>
  <el-empty v-else description="暂无记录" :image-size="imageSize" />

  <SubProcessDialog
    v-model="subVisible"
    :instance-id="subInstanceId"
    :process-name="subProcessName"
  />
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, ref } from 'vue'
import {
  Avatar,
  ChatDotRound,
  CircleCheck,
  CircleClose,
  Clock,
  Connection,
  Message,
  Promotion,
  Right,
  Switch,
} from '@element-plus/icons-vue'
import FlowNodeAvatar from './FlowNodeAvatar.vue'

const SubProcessDialog = defineAsyncComponent(() => import('./SubProcessDialog.vue'))

/** 评论在时间线中的虚拟节点类型（仅前端展示用，非流程引擎节点） */
const TIMELINE_COMMENT_TYPE = 100
/** 引擎子流程节点 type */
const NODE_SUB_PROCESS = 5

export type TimelineRow = {
  nodeName?: string
  nodeKey?: string
  nodeType?: number
  taskState?: number
  actorName?: string
  /** 代批人；有值时展示「委托人（受托人代批）」 */
  agentName?: string
  actorState?: number
  opinion?: string
  finishTime?: string
  createdAt?: string
  /** 评论行稳定 key */
  commentId?: number
  childInstanceId?: number
  childProcessName?: string
  childInstanceState?: number
}

export type CommentRow = {
  id: number
  userId?: number
  userName?: string
  content?: string
  createdAt?: string
}

const props = withDefaults(
  defineProps<{
    items?: TimelineRow[]
    comments?: CommentRow[]
    imageSize?: number
  }>(),
  {
    items: () => [],
    comments: () => [],
    imageSize: 56,
  },
)

const subVisible = ref(false)
const subInstanceId = ref<number>()
const subProcessName = ref('')

const merged = computed(() => {
  const rows: TimelineRow[] = (props.items || []).map((t) => ({ ...t }))
  for (const c of props.comments || []) {
    rows.push({
      commentId: c.id,
      nodeName: '评论',
      nodeType: TIMELINE_COMMENT_TYPE,
      actorName: c.userName || '-',
      actorState: 1,
      opinion: c.content || '',
      finishTime: c.createdAt || '',
      createdAt: c.createdAt || '',
    })
  }
  return rows.sort((a, b) => {
    const ap = isPendingFlow(a) ? 1 : 0
    const bp = isPendingFlow(b) ? 1 : 0
    if (ap !== bp) return ap - bp
    return timeKey(a).localeCompare(timeKey(b))
  })
})

function timeKey(t: TimelineRow) {
  return t.finishTime || t.createdAt || ''
}

function isPendingFlow(t: TimelineRow) {
  if (isComment(t)) return false
  const taskState = Number(t.taskState ?? -1)
  if (taskState === 1 || taskState === 2 || taskState === 3 || taskState === 4) return false
  return Number(t.actorState ?? 0) === 0
}

function rowKey(t: TimelineRow, i: number) {
  if (t.commentId) return `c-${t.commentId}`
  return `t-${t.nodeKey || ''}-${t.actorName || ''}-${t.finishTime || t.createdAt || ''}-${i}`
}

function isComment(t: TimelineRow) {
  return Number(t.nodeType) === TIMELINE_COMMENT_TYPE
}

function isSubProcess(t: TimelineRow) {
  return Number(t.nodeType) === NODE_SUB_PROCESS
}

function openSub(t: TimelineRow) {
  if (!t.childInstanceId) return
  subInstanceId.value = t.childInstanceId
  subProcessName.value = t.childProcessName || t.nodeName || '子流程'
  subVisible.value = true
}

function mainIcon(t: TimelineRow) {
  if (isComment(t)) return ChatDotRound
  const typ = Number(t.nodeType ?? 1)
  if (typ === 0) return Promotion
  if (typ === 2) return Message
  if (typ === NODE_SUB_PROCESS) return Connection
  return Avatar
}

function iconTone(t: TimelineRow) {
  if (isComment(t)) return 'is-comment'
  const typ = Number(t.nodeType ?? 1)
  if (typ === 0) return 'is-start'
  if (typ === 2) return 'is-cc'
  if (typ === NODE_SUB_PROCESS) return 'is-sub'
  return 'is-approve'
}

function badgeIcon(t: TimelineRow) {
  if (isComment(t)) return ChatDotRound
  switch (Number(t.actorState ?? 0)) {
    case 1:
      return CircleCheck
    case 2:
      return CircleClose
    case 3:
      return Right
    case 4:
      return Switch
    default:
      return Clock
  }
}

function badgeTone(t: TimelineRow) {
  if (isComment(t)) return 'comment'
  switch (Number(t.actorState ?? 0)) {
    case 1:
      return 'ok'
    case 2:
      return 'fail'
    case 3:
      return 'transfer'
    case 4:
      return 'skip'
    default:
      return 'pending'
  }
}

function displayTime(t: TimelineRow) {
  if (isComment(t)) return t.finishTime || t.createdAt || ''
  if (Number(t.actorState ?? 0) === 0) return ''
  return t.finishTime || t.createdAt || ''
}
</script>

<style scoped lang="scss">
.rich-timeline {
  padding: 4px 0 8px;
}

.rt-item {
  display: flex;
  gap: 14px;
  min-height: 72px;
}

.rt-rail {
  position: relative;
  width: 40px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.rt-icon-wrap {
  position: relative;
  width: 40px;
  height: 40px;
  flex-shrink: 0;
}

.rt-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: var(--el-color-primary);

  &.is-cc {
    background: #409eff;
  }
  &.is-start {
    background: var(--el-color-primary);
  }
  &.is-approve {
    background: var(--el-color-primary);
  }
  &.is-sub {
    background: #8b5cf6;
  }
  &.is-comment {
    background: #67c23a;
  }
}

.rt-badge {
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
  border: 2px solid var(--el-bg-color);

  &.ok {
    background: #67c23a;
  }
  &.fail {
    background: #f56c6c;
  }
  &.pending {
    background: #e6a23c;
  }
  &.transfer {
    background: #409eff;
  }
  &.skip {
    background: #909399;
  }
  &.comment {
    background: #67c23a;
  }
}

.rt-line {
  flex: 1;
  width: 2px;
  min-height: 24px;
  background: var(--el-border-color-lighter);
}

.rt-body {
  flex: 1;
  min-width: 0;
  padding-bottom: 20px;
}

.rt-node {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  line-height: 1.4;
  margin-bottom: 8px;
}

.rt-actor {
  margin-bottom: 8px;
}

.rt-actor-row {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.rt-agent {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
}

.rt-sub {
  margin-bottom: 8px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  line-height: 1.5;
}

.rt-sub-link {
  border: 0;
  padding: 0;
  margin: 0;
  background: transparent;
  color: var(--el-color-primary);
  cursor: pointer;
  font-size: inherit;
  line-height: inherit;

  &:hover {
    text-decoration: underline;
  }
}

.rt-opinion {
  display: inline-block;
  max-width: 100%;
  padding: 6px 10px;
  margin-bottom: 8px;
  border-radius: 4px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-regular);
  font-size: 13px;
  line-height: 1.5;
  word-break: break-word;

  &.is-comment {
    background: rgba(103, 194, 58, 0.12);
  }
}

.rt-time {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
