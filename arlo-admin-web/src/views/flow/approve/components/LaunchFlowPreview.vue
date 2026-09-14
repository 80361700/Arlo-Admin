<template>
  <div v-if="hasRoot" class="launch-flow-preview">
    <div class="section-title">审批流程</div>
    <div v-if="!readonly" class="section-tip">
      含条件/并行分支时，请切换分支完成各支路的自选人员，提交时会一并带上。
    </div>
    <div v-else class="section-tip">尚未发起，以下为绑定流程的设计预览。</div>
    <el-timeline class="timeline">
      <el-timeline-item v-for="(v, index) in processTimelineList" :key="`${v.nodeKey || 'n'}-${index}`">
        <template v-if="v.local_nodes?.length">
          <el-radio-group v-model="processChecked[v.nodeKey]" size="small">
            <el-radio-button v-for="c in v.local_nodes" :key="c.nodeKey" :value="c.nodeKey">
              {{ c.nodeName }}
            </el-radio-button>
          </el-radio-group>
        </template>

        <template v-else>
          <div class="node-name">{{ v.nodeName || nodeTypeLabel(v) }}</div>
          <div v-if="configSummary(v)" class="config-summary">{{ configSummary(v) }}</div>
          <div v-if="examineModeText(v)" class="mode-tag">
            <el-tag size="small" type="warning" effect="plain">{{ examineModeText(v) }}</el-tag>
          </div>

          <div v-if="assigneeMap[v.nodeKey]" class="people-row">
            <el-tooltip
              v-if="!readonly && !assigneeMap[v.nodeKey].disabled"
              :content="assigneeMap[v.nodeKey].type === 3 ? '添加角色' : '添加人员'"
              placement="top"
            >
              <button type="button" class="add-btn" @click="openSelect(v.nodeKey)">
                <el-icon :size="14"><User /></el-icon>
              </button>
            </el-tooltip>
            <FlowNodeAvatar
              v-for="(item, i) in assigneeMap[v.nodeKey].assignees"
              :key="`${item.id}-${i}`"
              :name="item.name"
            />
            <el-tag
              v-if="readonly && !assigneeMap[v.nodeKey].assignees?.length && !assigneeMap[v.nodeKey].disabled"
              type="info"
              effect="plain"
              size="small"
            >
              发起时自选
            </el-tag>
          </div>

          <div v-else-if="assigneeDesc[v.nodeKey]" class="desc">
            <el-tag type="info" effect="plain">{{ assigneeDesc[v.nodeKey] }}</el-tag>
          </div>
        </template>
      </el-timeline-item>
    </el-timeline>

    <template v-if="!readonly">
      <user-select ref="userRef" @success="onPicked" />
      <role-select ref="roleRef" @success="onPicked" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { User } from '@element-plus/icons-vue'
import FlowNodeAvatar from './FlowNodeAvatar.vue'
import userSelect from '@/components/flowProcess/components/common/user/index.vue'
import roleSelect from '@/components/flowProcess/components/common/role/index.vue'
import { useAuthStore } from '@/stores/auth'

type Person = { id: number | string; name: string }

type AssigneeItem = {
  type: 1 | 3
  /** 1=审批自选 2=抄送 */
  kind: 'assignee' | 'cc'
  nodeType: number
  assignees: Person[]
  disabled?: boolean
  selectMode?: number
  /** 设计器配置的候选范围；有值时选择器仅展示这些项 */
  candidates?: Person[]
}

const props = withDefaults(
  defineProps<{
    modelContent?: Record<string, any> | null
    /** 详情预览：只读，不可自选人员 */
    readonly?: boolean
    /** 暂存回填：审批人自选 */
    initialAssignees?: Record<string, { id: number; name: string }[]>
    /** 暂存回填：抄送自选 */
    initialCcUsers?: Record<string, { id: number; name: string }[]>
  }>(),
  { readonly: false },
)

const authStore = useAuthStore()
const userRef = ref<InstanceType<typeof userSelect>>()
const roleRef = ref<InstanceType<typeof roleSelect>>()
const activeNodeKey = ref('')

const assigneeMap = ref<Record<string, AssigneeItem>>({})
const assigneeDesc = ref<Record<string, string>>({})
const processChecked = reactive<Record<string, string>>({})
const processTimelineList = ref<any[]>([])

const root = computed(() => {
  const mc = props.modelContent
  if (!mc || typeof mc !== 'object') return null
  return (mc.nodeConfig || mc.childNode || mc) as any
})

const hasRoot = computed(() => {
  const r = root.value
  return !!(r && (r.nodeKey || r.type !== undefined || r.nodeName))
})

function branchList(config: any): any[] | null {
  // type 23 routeNodes 是跳转目标而非子树分支，不在时间线里用 radio 展开
  const list = config?.conditionNodes || config?.parallelNodes || config?.inclusiveNodes
  return Array.isArray(list) && list.length ? list : null
}

function normalizeAssignees(list: any): Person[] {
  if (!Array.isArray(list)) return []
  return list.map((u: any) => ({
    id: u.id ?? u.key ?? '',
    name: String(u.name || u.label || '-'),
  }))
}

function ensureAssignee(
  nodeKey: string,
  partial: Partial<AssigneeItem> & { assignees?: any; candidates?: any },
) {
  if (!nodeKey) return
  if (assigneeMap.value[nodeKey]) return
  const nodeType = Number(partial.nodeType) || 0
  const kind = partial.kind || (nodeType === 2 ? 'cc' : 'assignee')
  assigneeMap.value[nodeKey] = {
    type: partial.type || 1,
    kind,
    nodeType,
    assignees: normalizeAssignees(partial.assignees),
    disabled: partial.disabled,
    selectMode: partial.selectMode,
    candidates: normalizeAssignees(partial.candidates),
  }
}

function applyInitialPicks() {
  const asg = props.initialAssignees || {}
  const cc = props.initialCcUsers || {}
  for (const [key, list] of Object.entries(asg)) {
    const item = assigneeMap.value[key]
    if (!item || item.disabled) continue
    item.assignees = normalizeAssignees(list)
  }
  for (const [key, list] of Object.entries(cc)) {
    const item = assigneeMap.value[key]
    if (!item || item.disabled) continue
    item.assignees = normalizeAssignees(list)
  }
}

function fillNodeMeta(config: any) {
  const type = Number(config?.type)
  switch (type) {
    case 0: {
      const u = authStore.userInfo
      ensureAssignee(config.nodeKey, {
        type: 1,
        kind: 'assignee',
        nodeType: 0,
        assignees: u ? [{ id: u.id, name: u.name || u.username || '发起人' }] : [],
        disabled: true,
      })
      break
    }
    case 1: {
      if (!Reflect.has(config, 'setType')) break
      const setType = Number(config.setType)
      const level = Number(config.examineLevel) || 1
      switch (setType) {
        case 1:
          ensureAssignee(config.nodeKey, {
            type: 1,
            kind: 'assignee',
            nodeType: 1,
            assignees: config.nodeAssigneeList,
            disabled: true,
          })
          break
        case 2:
          if (!assigneeDesc.value[config.nodeKey]) {
            assigneeDesc.value[config.nodeKey] =
              level === 1 ? '直接主管' : `发起人的第${level}级主管`
          }
          break
        case 3:
          ensureAssignee(config.nodeKey, {
            type: 3,
            kind: 'assignee',
            nodeType: 1,
            assignees: config.nodeAssigneeList,
            disabled: true,
          })
          break
        case 4: {
          const selectMode = Number(config.selectMode) || 1
          const candidates = config.nodeCandidate?.assignees || []
          ensureAssignee(config.nodeKey, {
            type: selectMode === 3 ? 3 : 1,
            kind: 'assignee',
            nodeType: 1,
            assignees: [],
            disabled: false,
            selectMode,
            candidates,
          })
          break
        }
        case 5:
          if (!assigneeDesc.value[config.nodeKey]) {
            assigneeDesc.value[config.nodeKey] = '发起人自己'
          }
          break
        case 6: {
          const endLevel = Number(config.directorLevel) || Number(config.examineLevel) || 1
          if (!assigneeDesc.value[config.nodeKey]) {
            assigneeDesc.value[config.nodeKey] =
              Number(config.directorMode) === 0
                ? '连续多级主管（直到最上层）'
                : `连续多级主管（直到第${endLevel}级）`
          }
          break
        }
      }
      break
    }
    case 2:
      ensureAssignee(config.nodeKey, {
        type: 1,
        kind: 'cc',
        nodeType: 2,
        assignees: config.nodeAssigneeList,
        disabled: !config.allowSelection,
      })
      break
    default:
      break
  }
}

/** 整棵树注册自选节点，切换分支时人选保留，提交时一并带上 */
function registerTreeMeta(data: any) {
  if (!data || typeof data !== 'object') return
  fillNodeMeta(data)
  if (data.childNode) registerTreeMeta(data.childNode)
  for (const list of [data.conditionNodes, data.parallelNodes, data.inclusiveNodes]) {
    if (!Array.isArray(list)) continue
    for (const n of list) {
      registerTreeMeta(n)
      if (n?.childNode) registerTreeMeta(n.childNode)
    }
  }
}

/** 对齐飞龙：沿 childNode 展开；网关切换分支后递归推进该支路 */
function packageProcess(data: any, list: any[] = []): any[] {
  if (!data || typeof data !== 'object') return list

  const chain: any[] = [data]
  let cur = data
  while (cur?.childNode) {
    chain.push(cur.childNode)
    cur = cur.childNode
  }

  return chain.reduce((_list: any[], config: any) => {
    const nodes = branchList(config)
    if (nodes) {
      _list.push({ ...config, local_nodes: nodes })
      if (!processChecked[config.nodeKey]) {
        processChecked[config.nodeKey] = nodes[0].nodeKey
      }
      const selected =
        nodes.find((n: any) => n.nodeKey === processChecked[config.nodeKey]) || nodes[0]
      if (selected?.childNode) {
        packageProcess(selected.childNode, _list)
      }
      return _list
    }
    _list.push(config)
    fillNodeMeta(config)
    return _list
  }, list)
}

function rebuildTimeline() {
  registerTreeMeta(root.value)
  applyInitialPicks()
  const list = packageProcess(root.value, [])
  if (!list.length) {
    processTimelineList.value = []
    return
  }
  const last = list[list.length - 1]
  processTimelineList.value =
    Number(last?.type) === -1
      ? list
      : [...list, { nodeKey: '__end__', nodeName: '结束', type: -1 }]
}

watch(
  [root, processChecked],
  () => rebuildTimeline(),
  { immediate: true, deep: true },
)

watch(
  () => props.modelContent,
  () => {
    Object.keys(processChecked).forEach((k) => delete processChecked[k])
    assigneeMap.value = {}
    assigneeDesc.value = {}
    rebuildTimeline()
  },
)

watch(
  () => [props.initialAssignees, props.initialCcUsers] as const,
  () => applyInitialPicks(),
  { deep: true },
)

function nodeTypeLabel(v: any): string {
  const map: Record<number, string> = {
    0: '发起人',
    1: '审核人',
    2: '抄送人',
    3: '条件',
    4: '条件路由',
    5: '子流程',
    6: '延时处理',
    7: '触发器',
    8: '并行路由',
    9: '包容路由',
    23: '路由分支',
    30: '自动通过',
    31: '自动拒绝',
    [-1]: '结束',
  }
  return map[Number(v?.type)] || '节点'
}

function formatDelayTime(raw: string | undefined, delayType: any): string {
  if (!raw) return '延时等待'
  if (String(delayType) === '1' || delayType === 1) {
    const [n, unit] = String(raw).split(':')
    const unitText = unit === 'd' ? '天' : unit === 'h' ? '小时' : '分钟'
    return `${n || ''}${unitText}后进入下一步`
  }
  return `到达每天 ${raw}（已过则次日）后进入下一步`
}

function configSummary(v: any): string {
  const t = Number(v?.type)
  if (t === 5) {
    const name = String(v.callProcess || v.subProcessValue || '').split(':')[1]
    if (name) return `调用子流程［${name}］`
    return v.callProcess ? `调用子流程［${v.callProcess}］` : '子流程'
  }
  if (t === 6) return formatDelayTime(v.extendConfig?.time, v.delayType)
  if (t === 7) {
    if (String(v.triggerType) === '1' || v.triggerType === 1) return '立即执行'
    const delay = formatDelayTime(v.extendConfig?.time, v.delayType)
    return `延迟执行，${delay.replace('后进入下一步', '')}`
  }
  if (t === 23) {
    const n = Array.isArray(v.routeNodes) ? v.routeNodes.length : 0
    return n ? `${n}条动态路由` : '请设置路由节点'
  }
  if (t === 30) return '自动通过'
  if (t === 31) return '自动拒绝'
  return ''
}

function examineModeText(v: any): string {
  if (Number(v?.type) !== 1) return ''
  const people = assigneeMap.value[v.nodeKey]?.assignees || v.nodeAssigneeList || []
  if (!Array.isArray(people) || people.length < 2) return ''
  const mode = Number(v.examineMode) || 1
  if (mode === 2) return '会签'
  if (mode === 3) return '或签'
  return '依次审批'
}

function openSelect(nodeKey: string) {
  const item = assigneeMap.value[nodeKey]
  if (!item || item.disabled) return
  activeNodeKey.value = nodeKey
  nextTick(() => {
    const payload = {
      selectData: item.assignees,
      candidates: item.candidates,
    }
    if (item.type === 3) roleRef.value?.init(payload)
    else userRef.value?.init(payload)
  })
}

function onPicked(list: { id: string | number; name: string }[]) {
  const key = activeNodeKey.value
  const item = assigneeMap.value[key]
  if (!item) return
  let next = list.map((u) => ({ id: u.id, name: u.name }))
  if (item.selectMode === 1 && next.length > 1) next = next.slice(0, 1)
  item.assignees = next
  activeNodeKey.value = ''
}

function getSelectionPicks() {
  const nodeAssignees: Record<string, { id: number; name: string }[]> = {}
  const nodeCcUsers: Record<string, { id: number; name: string }[]> = {}
  // 遍历整棵树注册过的自选，切换分支后的人选也会带上
  for (const [nodeKey, map] of Object.entries(assigneeMap.value)) {
    if (!map || map.disabled) continue
    const people = (map.assignees || [])
      .map((u) => ({ id: Number(u.id), name: u.name }))
      .filter((u) => u.id > 0)
    if (!people.length) continue
    if (map.kind === 'cc' || map.nodeType === 2) nodeCcUsers[nodeKey] = people
    else if (map.kind === 'assignee' || map.nodeType === 1) nodeAssignees[nodeKey] = people
  }
  return { nodeAssignees, nodeCcUsers }
}

defineExpose({ getSelectionPicks, assigneeMap })
</script>

<style lang="scss" scoped>
.launch-flow-preview {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--el-border-color-lighter);
}
.section-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--el-text-color-primary);
}
.section-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 12px;
  line-height: 1.5;
}
.timeline {
  padding-left: 8px;
}
.node-name {
  padding-bottom: 4px;
  font-size: 14px;
  color: var(--el-text-color-primary);
}
.config-summary {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}
.mode-tag {
  margin-bottom: 6px;
}
.people-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  min-height: 32px;

  :deep(.el-tooltip__trigger) {
    display: inline-flex;
    align-items: center;
  }
}
.add-btn {
  box-sizing: border-box;
  width: 30px;
  height: 30px;
  margin: 0;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  background: var(--el-bg-color);
  color: var(--el-text-color-regular);
  cursor: pointer;
  flex-shrink: 0;
  line-height: 1;

  &:hover {
    border-color: var(--el-color-primary);
    color: var(--el-color-primary);
  }
}
.desc {
  margin-top: 2px;
}
</style>
