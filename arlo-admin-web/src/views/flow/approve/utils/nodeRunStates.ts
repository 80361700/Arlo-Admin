import type { InstanceDetail } from '@/api/modules/flow'

export type FlowNodeRunState = 'done' | 'active' | 'pending'

/** 从详情时间线 + 实例状态推导流程图节点运行态（含条件分支、结束节点） */
export function buildNodeRunStates(detail: InstanceDetail | null | undefined): Record<string, FlowNodeRunState> {
  const map: Record<string, FlowNodeRunState> = {}
  if (!detail) return map

  for (const t of detail.timeline || []) {
    const key = t.nodeKey
    if (!key) continue
    if (Number(t.actorState) === 0) map[key] = 'active'
    else if (map[key] !== 'active') map[key] = 'done'
  }

  if (detail.currentNodeKey) {
    const st = Number(detail.instanceState)
    // 审批中 / 暂存：当前节点为执行中；已结束：当前节点算已执行
    map[detail.currentNodeKey] = st === 0 || st === -1 ? 'active' : 'done'
  }

  const root = resolveRoot(detail.modelContent)
  const startKey = root?.nodeKey
  if (startKey && !map[startKey]) map[startKey] = 'done'

  // 走过的条件/并行/包容分支条：子树内有已执行或执行中节点时标为已执行
  if (root) markReachedBranches(root, map)

  // 已通过：结束节点标为已执行（引擎 finish 时常落在上一业务节点，结束本身不进时间线）
  if (Number(detail.instanceState) === 1 && root) {
    markEndNodes(root, map)
  }

  return map
}

function resolveRoot(modelContent: Record<string, any> | undefined | null): any | null {
  if (!modelContent || typeof modelContent !== 'object') return null
  return (modelContent as any).nodeConfig || modelContent
}

function subtreeHasVisited(node: any, map: Record<string, FlowNodeRunState>): boolean {
  if (!node) return false
  if (node.nodeKey && map[node.nodeKey] && map[node.nodeKey] !== 'pending') return true
  if (node.childNode && subtreeHasVisited(node.childNode, map)) return true
  for (const list of [node.conditionNodes, node.parallelNodes, node.inclusiveNodes, node.routeNodes]) {
    if (!Array.isArray(list)) continue
    for (const b of list) {
      if (subtreeHasVisited(b, map)) return true
    }
  }
  return false
}

function markReachedBranches(node: any, map: Record<string, FlowNodeRunState>) {
  if (!node) return
  for (const list of [node.conditionNodes, node.parallelNodes, node.inclusiveNodes]) {
    if (!Array.isArray(list)) continue
    for (const branch of list) {
      // 分支条本身 type=3，执行引擎不写时间线；子路径有访问则标绿
      if (branch?.nodeKey && subtreeHasVisited(branch, map) && map[branch.nodeKey] !== 'active') {
        map[branch.nodeKey] = 'done'
      }
      if (branch?.childNode) markReachedBranches(branch.childNode, map)
    }
  }
  if (node.childNode) markReachedBranches(node.childNode, map)
  if (Array.isArray(node.routeNodes)) {
    for (const r of node.routeNodes) {
      if (r?.childNode) markReachedBranches(r.childNode, map)
    }
  }
}

function markEndNodes(node: any, map: Record<string, FlowNodeRunState>) {
  if (!node) return
  if (Number(node.type) === -1 && node.nodeKey) {
    map[node.nodeKey] = 'done'
  }
  if (node.childNode) markEndNodes(node.childNode, map)
  for (const list of [node.conditionNodes, node.parallelNodes, node.inclusiveNodes, node.routeNodes]) {
    if (!Array.isArray(list)) continue
    for (const b of list) {
      if (b?.childNode) markEndNodes(b.childNode, map)
      if (Number(b?.type) === -1 && b?.nodeKey) map[b.nodeKey] = 'done'
    }
  }
}
