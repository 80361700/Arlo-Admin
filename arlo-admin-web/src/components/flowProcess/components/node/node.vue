<template>
  <!-- 发起节点 默认 -->
  <div v-if="props.data.type == 0" class="node-wrap">
    <div class="node-wrap-box start-node" :class="{ 'is-readonly': isReadonly }" @click="handleSetting(props.data)">
      <div class="title" :style="{ background: titleBg(props.data) }">
        <el-icon class="icon"><Avatar /></el-icon>
        <span class="title_label">
          <span>{{ props.data.nodeName }}</span>
          <el-icon v-if="!isReadonly"><Edit /></el-icon>
        </span>
      </div>
      <div class="content">
        <span v-if="props.data?.nodeAssigneeList?.length">
          {{ names }}
        </span>
        <span v-else>所有人</span>
      </div>
    </div>
    <div class="add-node-btn-box" :class="{ 'is-readonly': isReadonly }">
      <div v-if="!isReadonly" class="add-node-btn">
        <node-tabs @onEvent="e => { addNode(e, props.data) }" />
      </div>
    </div>
  </div>

  <!-- 结束节点 固定 -->
  <div v-else-if="props.data.type == -1" class="node-wrap">
    <div class="node-wrap-box end-node" :class="{ 'is-readonly': isReadonly }">
      <div class="title" :style="{ background: titleBg(props.data) }">
        <span class="title_label">
          结束
        </span>
      </div>
      <div class="content">
        <span>
          流程结束
        </span>
      </div>
    </div>
  </div>

  <!-- 有分支 -->
  <div v-else-if="[4,8,9].includes(props.data.type)" class="branch-wrap">
    <div class="branch-box-wrap">
      <div class="branch-box" :class="{ 'is-readonly': isReadonly }">
        <el-button
          v-if="!isReadonly"
          type="default"
          round
          class="add-branch"
          @click="addBranch(getNode())"
        >
          <template v-if="props.data.type == 4">添加条件分支</template>
          <template v-if="props.data.type == 8">添加并行分支</template>
          <template v-if="props.data.type == 9">添加包容分支</template>
        </el-button>

        <div v-for="(node, index) in getNode()" class="col-box">
          <!-- 分支内容 -->
          <div class="condition-node">
            <div class="condition-node-box">
              <div
                class="auto-judge"
                :class="{ 'is-readonly': isReadonly }"
                :style="conditionBorder(node)"
                @click="handleSetting(node, props.data)"
              >
                <div class="title" :style="conditionTitleStyle(node)">
                  {{ node.nodeName }}
                  <div class="level">
                    优先级{{ node.priorityLevel }}
                  </div>
                  <div
                    v-if="!isReadonly && (props.data.type == 8 || getNode().length - 1 != index)"
                    class="close"
                  >
                    <el-icon size="14" @click.stop="deleteNodeBranch(props.parent, getNode(), index)"><Close /></el-icon>
                  </div>
                </div>
                <div class="content">
                  <!-- 条件分支 | 包容分支 -->
                  <template v-if="(props.data.type == 4 || props.data.type == 9) && (node.conditionList && node.conditionList.length)">
                    <template v-for="(item, index) in node.conditionList" :key="index">
                      <i v-if="index > 0" class="success">或</i>
                      <div>
                        <template v-for="(it, i) in item" :key="i">
                          <span>
                            <i v-if="i > 0" class="success"> 且</i>
                            {{ formatConditionText(it) }}
                          </span>
                        </template>
                      </div>
                    </template>
                  </template>
                  <div v-else-if="node.nodeName == '默认条件'">
                    未满足条件时，将进入默认流程
                  </div>
                  <div v-else class="placeholder">
                    {{ getNodeInfo().placeholder }}
                  </div>
                </div>
              </div>
              <div class="add-node-btn-box" :class="{ 'is-readonly': isReadonly }">
                <div v-if="!isReadonly" class="add-node-btn">
                  <node-tabs @onEvent="e => { addNode(e, node) }" />
                </div>
              </div>
            </div>
          </div>

          <!-- 下一个节点 -->
          <node v-if="node.childNode" :parent="node" :data="node.childNode" @onEvent="handleSetting" />

          <!-- 线 -->
          <div v-if="index == 0" class="top-left-cover-line" />
          <div v-if="index == 0" class="bottom-left-cover-line" />
          <div v-if="index == getNode().length - 1" class="top-right-cover-line" />
          <div v-if="index == getNode().length - 1" class="bottom-right-cover-line" />
        </div>

        <!-- 分支ICON -->
        <div class="svg-icon-box">
          <el-icon color="#626aef" size="18">
            <Share v-if="props.data.type == 4" />
            <Operation v-if="props.data.type == 8" />
            <CopyDocument v-if="props.data.type == 9" />
          </el-icon>
        </div>
      </div>

      <!-- 添加分支 -->
      <div class="add-node-btn-box" :class="{ 'is-readonly': isReadonly }">
        <div v-if="!isReadonly" class="add-node-btn">
          <node-tabs @onEvent="e => { addNode(e, props.data) }" />
        </div>
      </div>
    </div>
  </div>

  <!-- 无分支 -->
  <div v-else class="node-wrap">
    <div class="node-wrap-box" :class="{ 'is-readonly': isReadonly }" @click="handleSetting(props.data)">
      <div class="title" :style="{ background: titleBg(props.data) }">
        <el-icon class="icon"><component :is="getNodeInfo().icon"/></el-icon>
        <span class="title_label">
          {{ props.data.nodeName }}
          <el-icon v-if="!isReadonly"><Edit /></el-icon>
        </span>
        <div v-if="!isReadonly" class="option">
          <el-icon size="14" @click.stop="deleteNode(props.parent)"><Close /></el-icon>
        </div>
      </div>
      <div class="content">
        <!-- 运行态：展示实际办理人（含加签） -->
        <template v-if="runtimeActorText">
          <span>{{ runtimeActorText }}</span>
        </template>
        <!-- 审批人 -->
        <template v-else-if="props.data.type == 1">
          <template v-if="props.data.setType == 1">
            <span v-if="props.data.nodeAssigneeList.length">
              {{ props.data.nodeAssigneeList.map((item: any) => item.name).join(', ') }}
            </span>
            <span v-else class="placeholder">请选择人员</span>
          </template>
          <span v-if="props.data.setType == 2">
            {{ `发起人的第${props.data.examineLevel}级主管` }}
          </span>
          <template v-if="props.data.setType == 3">
            <span v-if="props.data.nodeAssigneeList.length">
              {{ props.data.nodeAssigneeList.map((item: any) => item.name).join(', ') }}
            </span>
            <span v-else class="placeholder">请选择角色</span>
          </template>
          <template v-if="props.data.setType == 4">
            <span v-if="props.data.nodeCandidate.assignees.length">
              {{ props.data.nodeCandidate.assignees.map((item: any) => item.name).join(', ') }}
            </span>
            <span v-else class="placeholder">请选择{{props.data.nodeCandidate.type == 0 ? "人员" : "角色"  }}</span>
          </template>
          <span v-if="props.data.setType == 5">
            发起人自己
          </span>
          <span v-if="props.data.setType == 6">
            {{ props.data.directorMode == 0 ? "直到最上层主管" : `直到第${props.data.directorLevel || props.data.examineLevel || 1}级主管` }}
          </span>
        </template>
        <!-- 抄送节点 -->
        <template v-else-if="props.data.type == 2 && props.data.nodeAssigneeList.length">
          <span>
            {{ props.data.nodeAssigneeList.map((item: any) => item.name).join(', ') }}
          </span>
        </template>
        <!-- 子流程 -->
        <template v-else-if="props.data.type == 5 && props.data.callProcess">
          <span>
            {{ props.data.callProcess.split(":")[1] }}
          </span>
        </template>
        <!-- 延迟等待 -->
        <template v-else-if="props.data.type == 6 && props.data.extendConfig.time">
          <span v-if="props.data.delayType == '1'">
            {{ props.data.extendConfig.time.split(":")[0] }}{{ `${props.data.extendConfig.time.split(":")[1] == 'd' ? '天' : (props.data.extendConfig.time.split(":")[1] == 'h' ? '小时' : '分钟')}后进入下一步` }}
          </span>
          <span v-else>
            {{ props.data.extendConfig.time }}后进入下一步
          </span>
        </template>
        <!-- 延迟等待 -->
        <template v-else-if="props.data.type == 7">
          <span v-if="props.data.triggerType == '1'">
            立即执行
          </span>
          <span v-else>
            延迟执行，
            <template v-if="props.data.delayType == '1'">
              等待{{ props.data.extendConfig.time.split(":")[0] }}{{ `${props.data.extendConfig.time.split(":")[1] == 'd' ? '天' : (props.data.extendConfig.time.split(":")[1] == 'h' ? '小时' : '分钟')}` }}
            </template>
            <template v-else>
              至当天{{ props.data.extendConfig.time }}
            </template>
          </span>
        </template>
        <!-- 路由分支 -->
        <template v-else-if="props.data.type == 23 && props.data.routeNodes.length">
          {{ `${props.data.routeNodes.length}条动态路由` }}
        </template>
        <!-- 自动通过，自动拒绝 -->
        <template v-else-if="props.data.type == 30 || props.data.type == 31">
          <span>
            {{ getNodeInfo().placeholder }}
          </span>
        </template>
        <span v-else class="placeholder">
          {{ getNodeInfo().placeholder }}
        </span>
      </div>
    </div>
    <!-- 添加分支 -->
    <div class="add-node-btn-box" :class="{ 'is-readonly': isReadonly }">
      <div v-if="!isReadonly" class="add-node-btn">
        <node-tabs @onEvent="e => { addNode(e, props.data) }" />
      </div>
    </div>
  </div>

  <!-- 加载下一个节点 -->
  <node v-if="props.data.childNode" :parent="props.data" :data="props.data.childNode" @onEvent="handleSetting" />
</template>

<script setup lang="ts">
import { computed, inject, type ComputedRef } from 'vue'
import nodeTabs from '../tabs/index.vue'
import config from '../../config'
import { formatConditionText } from '../common/utils'

type FlowNodeRunState = 'done' | 'active' | 'pending'

const STATE_COLOR: Record<FlowNodeRunState, string> = {
  done: '#67c23a',
  active: '#e6a23c',
  pending: '#c0c4cc',
}

const props = defineProps<{
  data: any
  parent: any
}>()

const isReadonly = inject<ComputedRef<boolean>>(
  'flowReadonly',
  computed(() => false),
)
const nodeStates = inject<ComputedRef<Record<string, FlowNodeRunState>>>(
  'flowNodeStates',
  computed(() => ({})),
)
const nodeActors = inject<
  ComputedRef<Record<string, { name: string; actorType?: number }[]>>
>('flowNodeActors', computed(() => ({})))

const names = computed(() => {
  return props.data?.nodeAssigneeList?.map((item: any) => item.name).join(', ') || ''
})

/** 只读流程图：用运行时办理人覆盖设计态名单（加签可见） */
const runtimeActorText = computed(() => {
  if (!isReadonly.value) return ''
  const key = props.data?.nodeKey
  if (!key) return ''
  const list = nodeActors.value[key]
  if (!list?.length) return ''
  return list
    .map((a) => (Number(a.actorType) === 1 ? `${a.name}(加签)` : a.name))
    .join(', ')
})

const emit = defineEmits<{
  (e: 'onEvent', value: any, parent: any): void
}>()

const getNode = () => {
  const type = props.data.type
  if (type == 4) return props.data.conditionNodes
  if (type == 8) return props.data.parallelNodes
  if (type == 9) return props.data.inclusiveNodes
}
const getNodeInfo = () => {
  const obj = config.nodes.find((item) => item.type == props.data.type) as any
  return obj || {}
}

function runState(node: any): FlowNodeRunState {
  const key = node?.nodeKey
  if (!key) return 'pending'
  return nodeStates.value[key] || 'pending'
}

function titleBg(node: any) {
  if (!isReadonly.value) {
    if (node?.type === 0 || node?.type === -1) return 'var(--el-color-info)'
    return getNodeInfoFor(node).color || '#909399'
  }
  return STATE_COLOR[runState(node)]
}

function getNodeInfoFor(node: any) {
  return (config.nodes.find((item) => item.type == node?.type) as any) || {}
}

function conditionBorder(node: any) {
  if (!isReadonly.value) return undefined
  const color = STATE_COLOR[runState(node)]
  // 内描边，避免与顶部连接箭头重叠出现「小口」
  return { boxShadow: `inset 0 0 0 1px ${color}, 0 2px 5px #0000001a` }
}

function conditionTitleStyle(node: any) {
  if (!isReadonly.value) return undefined
  return { color: STATE_COLOR[runState(node)] }
}

function handleSetting(node: any, parent?: any) {
  if (isReadonly.value) return
  if (node?.nodeName == '默认条件') return
  emit('onEvent', node, parent ?? props.parent)
}

// 添加节点
const addNode = (e: any, node: any) => {
  if (isReadonly.value) return
  const newNode = {
    nodeName: e.defaultName,
    nodeKey: config.nodeKey(),
    type: e.type,
    childNode: node.childNode,
    ...e.config,
  }
  node.childNode = JSON.parse(JSON.stringify(newNode))
}
// 删除节点
const deleteNode = (node: any) => {
  if (isReadonly.value) return
  node.childNode = node.childNode.childNode
}

// 添加分支
const addBranch = (nodes: any) => {
  if (isReadonly.value) return
  const type = props.data.type
  if (type == 4) {
    nodes.splice(nodes.length - 1, 0, {
      nodeName: `条件${nodes.length}`,
      nodeKey: config.nodeKey(),
      type: 3,
      priorityLevel: 1,
      conditionList: [],
    })
  }
  if (type == 8) {
    nodes.push({
      nodeName: `并行分支${nodes.length + 1}`,
      nodeKey: config.nodeKey(),
      type: 3,
      priorityLevel: 1,
      conditionList: [],
    })
  }
  if (type == 9) {
    nodes.splice(nodes.length - 1, 0, {
      nodeName: `包容条件${nodes.length}`,
      nodeKey: config.nodeKey(),
      type: 3,
      priorityLevel: 1,
      conditionList: [],
    })
  }

  updateLevel(nodes)
}
// 删除分支
const deleteNodeBranch = (parent: any, nodes: any, index: number) => {
  if (isReadonly.value) return
  if (nodes.length <= 2) {
    parent.childNode = parent.childNode.childNode
  } else {
    nodes.splice(index, 1)
  }
  updateLevel(nodes)
}

// 更新分支优先级
const updateLevel = (nodes: any) => {
  nodes.forEach((node: any, index: number) => {
    node.priorityLevel = index + 1
  })
}
</script>

<style lang="scss" scoped>

.node-wrap {
  display: inline-flex;
  width: 100%;
  flex-flow: column wrap;
  justify-content: flex-start;
  align-items: center;
  padding: 0 50px;
  position: relative;
  z-index: 1;
  box-sizing: border-box;
}
.node-wrap-box {
  display: inline-flex;
  flex-direction: column;
  position: relative;
  width: 220px;
  min-height: 72px;
  flex-shrink: 0;
  background: #fff;
  border-radius: 4px;
  cursor: pointer;
  box-shadow: 0 2px 5px #0000001a;
  &::after {
    pointer-events: none;
    content: "";
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 2;
    border-radius: 4px;
    transition: all .1s;
  }
  &::before {
    content: "";
    position: absolute;
    top: -12px;
    left: 50%;
    transform: translate(-50%);
    width: 0px;
    border-style: solid;
    border-width: 8px 6px 4px;
    border-color: rgb(202, 202, 202) transparent transparent;
    background: #f6f8f9;
  }
  .title {
    height: 24px;
    line-height: 24px;
    font-size: 12px;
    color: #fff;
    padding-left: 16px;
    padding-right: 30px;
    border-radius: 4px 4px 0 0;
    position: relative;
    display: flex;
    align-items: center;
    .icon {
      margin-right: 5px;
    }
    .title_label {
      display: flex;
      align-items: center;
      i {
        margin-left: 5px;
      }
      &:hover {
        text-decoration: underline;
      }
    }
    .option {
      position: absolute;
      right: 10px;
      display: none;
      align-items: center;
    }
  }
  .content {
    position: relative;
    padding: 15px;
    font-size: 12px;
    color: #666;
    .placeholder {
      color: #999;
    }
  }
  &:hover {
    .title {
      .option {
        display: flex;
      }
    }
    &::after {
      // border: 1px solid var(--el-color-primary);
      // box-shadow: 0 0 6px 0 var(--el-color-primary-light-5)
      box-shadow: 0 0px 6px #0000002a;
    }
  }
}
.add-node-btn-box {
  display: inline-flex;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
  &::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: -1;
    margin: auto;
    width: 2px;
    height: 100%;
    background-color: #cacaca;
  }
  &.is-readonly {
    height: 48px;
    width: 240px;
    /* 竖线止于下沿内侧，避免穿出下一节点横线 */
    &::before {
      height: calc(100% - 2px);
    }
  }
  .add-node-btn {
    -webkit-user-select: none;
    -moz-user-select: none;
    user-select: none;
    width: 240px;
    padding: 20px 0 32px;
    display: flex;
    justify-content: center;
    flex-shrink: 0;
    flex-grow: 1;
  }
}
.node-wrap-box.is-readonly,
.auto-judge.is-readonly {
  cursor: default;
  /* 只读预览：无悬停高亮 / 无下划线（非设计模式） */
  &:hover::after {
    border: none;
    box-shadow: none;
  }
  .title .title_label:hover {
    text-decoration: none;
  }
}
.node-wrap-box.is-readonly:hover .title .option {
  display: none;
}
.auto-judge.is-readonly:hover .title {
  width: 140px;
  .close {
    display: none;
  }
  .level {
    display: block;
  }
}
.start-node {
  &::before {
    display: none;
  }
}
.end-node {
  cursor: default;
  .title {
    .title_label {
      &:hover {
        text-decoration: none;
      }
    }
  }
}

.branch-wrap {
  display: inline-flex;
  width: 100%;
} 
.branch-box-wrap {
  display: flex;
  flex-flow: column wrap;
  align-items: center;
  min-height: 270px;
  width: 100%;
  flex-shrink: 0;
}
.branch-box {
  display: flex;
  overflow: visible;
  min-height: 180px;
  height: auto;
  border-bottom: 2px solid #ccc;
  border-top: 2px solid #ccc;
  position: relative;
  margin-top: 15px;
  margin-bottom: 15px;
  &.is-readonly {
    margin-top: 0;
    /* 遮住上方竖线可能多出的 1~2px，保证 T 字交汇干净 */
    &::after {
      content: '';
      position: absolute;
      top: 2px;
      left: 50%;
      z-index: 3;
      width: 6px;
      height: 3px;
      margin-left: -3px;
      background: #f6f8f9;
      pointer-events: none;
    }
  }
}
.col-box {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  background: #f6f8f9;
  &::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 0;
    margin: auto;
    width: 2px;
    height: 100%;
    background-color: #cacaca;
  }
}
.add-branch {
  justify-content: center;
  padding: 0 10px;
  position: absolute;
  top: -16px;
  left: 50%;
  transform: translate(-50%);
  transform-origin: center center;
  z-index: 1;
  display: inline-flex;
  align-items: center;
}
.condition-node {
  display: inline-flex;
  flex-direction: column;
  min-height: 220px;
}
.condition-node-box {
  padding-top: 30px;
  padding-right: 50px;
  padding-left: 50px;
  justify-content: center;
  align-items: center;
  flex-grow: 1;
  position: relative;
  display: inline-flex;
  flex-direction: column;
  &::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    margin: auto;
    width: 2px;
    height: 100%;
    background-color: #cacaca;
  }
}
.auto-judge {
  position: relative;
  width: 220px;
  min-height: 72px;
  background: #fff;
  border-radius: 4px;
  padding: 15px;
  cursor: pointer;
  box-shadow: 0 2px 5px #0000001a;
  box-sizing: border-box;
  &::before {
    content: "";
    position: absolute;
    top: -12px;
    left: 50%;
    transform: translate(-50%);
    width: 0px;
    border-style: solid;
    border-width: 8px 6px 4px;
    border-color: rgb(202, 202, 202) transparent transparent;
    background: #f6f8f9;
  }
  &::after {
    pointer-events: none;
    content: "";
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 2;
    border-radius: 4px;
    transition: all .1s;
  }
  &:hover {
    .title {
      width: 170px;
      .close {
        display: block;
      }
      .level {
        display: none;
      }
    }
    &::after {
      border: 1px solid var(--el-color-primary);
      box-shadow: 0 0 6px 0 var(--el-color-primary-light-5)
    }
  }

  .title {
    line-height: 16px;
    font-size: 14px;
    color: #606266;
    width: 140px;
    white-space: nowrap;      /* 禁止换行 */
    overflow: hidden;         /* 隐藏超出部分 */
    text-overflow: ellipsis;  /* 显示省略号 */
    .close {
      font-size: 15px;
      position: absolute;
      top: 15px;
      right: 15px;
      color: #999;
      display: none;
    }
    .level {
      font-size: 12px;
      position: absolute;
      top: 15px;
      right: 15px;
      color: #999;
    }
  }
  .content {
    position: relative;
    padding-top: 15px;
    color: #666;
    font-size: 12px;
    .placeholder {
      color: #999;
    }
    i {
      font-style: normal;
      &.success {
        color: var(--el-color-success)
      }
    }
  }
  .sort-left {
    position: absolute;
    top: 0;
    bottom: 0;
    z-index: 1;
    left: 0;
    display: none;
    justify-content: center;
    align-items: center;
    flex-direction: column;
  }
  .sort-right {
    position: absolute;
    top: 0;
    bottom: 0;
    z-index: 1;
    right: 0;
    display: none;
    justify-content: center;
    align-items: center;
    flex-direction: column;
  }
}

.top-left-cover-line, 
.top-right-cover-line {
  position: absolute;
  height: 3px;
  width: 50%;
  background-color: #f6f8f9;
  top: -2px;
}
.bottom-left-cover-line, 
.bottom-right-cover-line {
  position: absolute;
  height: 3px;
  width: 50%;
  background-color: #f6f8f9;
  bottom: -2px;
}
.top-left-cover-line {
  left: -1px;
}
.bottom-left-cover-line {
  left: -1px;
}
.top-right-cover-line {
  right: -1px;
}
.bottom-right-cover-line {
  right: -1px;
}

.svg-icon-box {
  position: absolute;
  bottom: -20px;
  right: 50%;
  background-color: #fff;
  border-radius: 50%;
  margin-right: -18px;
  z-index: 99;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
}

</style>