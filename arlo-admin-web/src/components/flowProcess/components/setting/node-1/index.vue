<template>
  <div>
    <el-form ref="formRef" :model="props.data.node" :rules="rules" label-width="auto" label-position="top">
      <el-form-item label="节点名称" prop="nodeName">
        <el-input v-model="props.data.node.nodeName"></el-input>
      </el-form-item>
      <el-form-item label="节点key" prop="nodeKey">
        <el-input v-model="props.data.node.nodeKey"></el-input>
      </el-form-item>
      <div>
        <el-tabs
          v-model="tabs"
          style="width: 100%;"
        >
          <el-tab-pane label="设置审批人" name="1" />
          <el-tab-pane label="表单权限" name="2" />
          <el-tab-pane label="操作权限" name="3" />
        </el-tabs>
      </div>

      <!-- 设置审批人 -->
      <template v-if="tabs == '1'">
        <el-form-item label="审批人员类型" prop="setType">
          <el-radio-group v-model="props.data.node.setType">
            <el-radio :value="1" style="width: 25%;">指定成员</el-radio>
            <el-radio :value="2" style="width: 25%;">主管</el-radio>
            <el-radio :value="3" style="width: 25%;">角色</el-radio>
            <el-radio v-if="!isChildProcess" :value="4" style="width: 25%;">发起人自选</el-radio>
            <el-radio :value="5" style="width: 25%;">发起人自己</el-radio>
            <el-radio :value="6" style="width: 25%;">连续多级主管</el-radio>
          </el-radio-group>
        </el-form-item>
        <!-- 指定成员 -->
        <template v-if="props.data.node.setType == 1">
          <el-form-item label="选择人员" prop="nodeAssigneeList">
            <div class="selectBtn">
              <el-button icon="Plus" round type="primary" @click="handleUser">选择人员</el-button>
            </div>
            
            <div class="tags">
              <el-tag v-for="(item, index) in props.data?.node?.nodeAssigneeList" :key="index" closable @close="userDelete(index)">{{item.name}}</el-tag>
            </div>
          </el-form-item>
        </template>
        
        <!-- 主管 -->
        <template v-if="props.data.node.setType == 2">
          <el-form-item label="指定主管" prop="examineLevel">
            <el-input-number v-model="props.data.node.examineLevel" :min="1" :max="100" :value-on-clear="1" />
            <div style="padding-top: 10px; width: 100%;">
              <el-alert :title="`发起人的第${props.data.node.examineLevel || 1}级主管`" type="info" show-icon :closable="false" />
            </div>
          </el-form-item>
        </template>
        
        <!-- 角色 -->
        <template v-if="props.data.node.setType == 3">
          <el-form-item label="全员参与审批" prop="groupStrategy">
            <el-switch
              v-model="props.data.node.groupStrategy"
              :active-value="1"
              :inactive-value="0"
            />
          </el-form-item>
          <el-form-item label="选择角色" prop="nodeAssigneeList">
            <div class="selectBtn">
              <el-button icon="Plus" round type="primary" @click="handleRole">选择角色</el-button>
            </div>
            
            <div class="tags">
              <el-tag v-for="(item, index) in props.data?.node?.nodeAssigneeList" :key="index" closable @close="roleDelete(index)">{{item.name}}</el-tag>
            </div>
          </el-form-item>
        </template>

        <!-- 发起人自选 -->
        <template v-if="props.data.node.setType == 4">
          <el-form-item label="发起人自选" prop="selectMode">
            <el-radio-group v-model="props.data.node.selectMode">
              <el-radio :value="1">自选一个人</el-radio>
              <el-radio :value="2">自选多个人</el-radio>
              <el-radio :value="3">自选角色</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item v-if="props.data.node.selectMode == 3" label="全员参与审批" prop="groupStrategy">
            <el-switch
              v-model="props.data.node.groupStrategy"
              :active-value="1"
              :inactive-value="0"
            />
          </el-form-item>

          <el-form-item v-if="props.data.node.selectMode == 3" label="候选角色">
            <div class="selectBtn">
              <el-button icon="Plus" round type="primary" @click="handleRole">选择角色</el-button>
            </div>
            
            <div class="tags">
              <el-tag v-for="(item, index) in props.data?.node?.nodeCandidate?.assignees" :key="index" closable @close="roleDelete(index)">{{item.name}}</el-tag>
            </div>
          </el-form-item>

          <el-form-item v-else label="候选人员">
            <div class="selectBtn">
              <el-button icon="Plus" round type="primary" @click="handleUser">选择人员</el-button>
            </div>
            
            <div class="tags">
              <el-tag v-for="(item, index) in props.data?.node?.nodeCandidate?.assignees" :key="index" closable @close="userDelete(index)">{{item.name}}</el-tag>
            </div>
          </el-form-item>
        </template>

        <!-- 发起人自己 -->
        <!-- 连续多级主管 -->
        <template v-if="props.data.node.setType == 6">
          <el-form-item label="连续主管审批终点" prop="directorMode">
            <el-radio-group v-model="props.data.node.directorMode">
              <el-radio :value="0">直到最上层主管</el-radio>
              <el-radio :value="1">自定义审批终点</el-radio>
            </el-radio-group>

            <div v-if="props.data.node.directorMode == 1" style="padding-top: 5px; width: 100%;">
              <el-input-number v-model="props.data.node.directorLevel" :min="1" :max="100" :value-on-clear="1" />
              <div style="padding-top: 10px; width: 100%;">
                <el-alert :title="`直到发起人的第${props.data.node.directorLevel || 1}级主管`" type="info" show-icon :closable="false" />
              </div>
            </div>
          </el-form-item>
        </template>


        <!-- 超时处理 -->
        <el-form-item label="超时配置" prop="termAuto">
          <el-checkbox v-model="props.data.node.termAuto" label="超时自动审批"></el-checkbox>
        </el-form-item>
        <el-form-item v-if="props.data.node.termAuto" label="审批期限（为 0 则不生效）" prop="term">
          <el-input-number v-model="props.data.node.term" :min="1" :max="100" :value-on-clear="1" />
          <span style="padding-left: 10px;">小时</span>
        </el-form-item>
        <el-form-item v-if="props.data.node.termAuto" label="审批期限超时后执行" prop="termMode">
          <el-radio-group v-model="props.data.node.termMode">
            <el-radio :value="0">自动通过</el-radio>
            <el-radio :value="1">自动拒绝</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 审批提醒 -->
        <el-form-item label="审批提醒" prop="remind">
          <el-checkbox v-model="props.data.node.remind" label="审批提醒"></el-checkbox>
        </el-form-item>
        <el-form-item v-if="props.data.node.remind" label="延时时间" prop="delayType">
          <el-radio-group v-model="props.data.node.delayType">
            <el-radio :value="1">固定时长</el-radio>
            <el-radio :value="2">自动计算</el-radio>
          </el-radio-group>

          <div v-if="props.data.node.delayType == 1" style="width: 100%; padding-top: 10px;">
            <el-input v-model="number" clearable>
              <template #append>
                <el-select v-model="timer" placeholder="单位" style="width: 100px">
                  <el-option label="天" value="d" />
                  <el-option label="小时" value="h" />
                  <el-option label="分钟" value="m" />
                </el-select>
              </template>
            </el-input>
            <div v-if="number" style="padding-top: 10px; width: 100%;">
              <el-alert :title="`${number}${timer == 'd' ? '天' : (timer == 'h' ? '小时' : '分钟')}后提醒审批人`" type="info" show-icon :closable="false" />
            </div>
          </div>

          <div v-else style="width: 100%; padding-top: 10px;">
            <el-time-picker v-model="props.data.node.extendConfig.remindTime" placeholder="选择时间" value-format="HH:mm:ss" style="width: 100%" />
            <div v-if="props.data.node.extendConfig.remindTime" style="padding-top: 10px; width: 100%;">
              <el-alert :title="`每天 ${props.data.node.extendConfig.remindTime} 提醒审批人（待办未办完前每日一次；已过则从次日该时刻起）`" type="info" show-icon :closable="false" />
            </div>
          </div>
        </el-form-item>

        <!-- 多人审批时审批方式 -->
        <el-form-item label="多人审批时审批方式" prop="examineMode">
          <el-radio-group v-model="props.data.node.examineMode">
            <el-radio :value="1">按顺序依次审批</el-radio>
            <el-radio :value="2">会签 (可同时审批，每个人必须审批通过)</el-radio>
            <el-radio :value="3">或签 (有一人审批通过即可)</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 审批人与提交人为同一人时 -->
        <el-form-item label="审批人与提交人为同一人时" prop="approveSelf">
          <el-radio-group v-model="props.data.node.approveSelf">
            <el-radio :value="0">由发起人对自己审批</el-radio>
            <el-radio :value="1">自动跳过</el-radio>
            <el-radio :value="2">转交给直接上级审批</el-radio>
            <el-radio :value="3">转交给部门负责人审批</el-radio>
          </el-radio-group>
        </el-form-item>

      </template>

      <!-- 表单权限 -->
      <template v-if="tabs == '2'">
        <el-form-item label="添加子表单" prop="actionUrl">
          <div class="selectBtn">
            <el-button icon="Plus" round type="primary" @click="handleSubForm">选择子表单</el-button>
          </div>
          <div v-if="subFormTag" class="sub-form-row">
            <el-tag closable @close="clearSubForm">{{ subFormTag.name }}</el-tag>
            <el-button link type="primary" :icon="View" :loading="previewLoading" @click="previewSubForm">
              表单预览
            </el-button>
          </div>
        </el-form-item>

        <el-form-item label="表单权限">
          <div class="table">
            <el-table :data="props.data?.node?.extendConfig?.formConfig" border style="width: 100%">
              <el-table-column prop="label" label="表单字段" />
              <el-table-column prop="opera" label="操作权限" width="210">
                <template #default="scope">
                  <el-radio-group v-model="scope.row.opera" size="small">
                    <el-radio :value="0">只读</el-radio>
                    <el-radio :value="1">编辑</el-radio>
                    <el-radio :value="2">隐藏</el-radio>
                  </el-radio-group>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-form-item>
      </template>

      <!-- 操作权限 -->
      <template v-if="tabs == '3'">
        <el-form-item label="操作权限">
          <el-checkbox v-model="props.data.node.allowTransfer" label="允许转交"></el-checkbox>
          <el-checkbox v-model="props.data.node.allowAppendNode" label="允许加签/减签"></el-checkbox>
          <el-checkbox v-model="props.data.node.allowRollback" label="允许回退"></el-checkbox>
          <el-checkbox v-model="props.data.node.allowCc" label="允许抄送"></el-checkbox>
        </el-form-item>

        <!-- 驳回策略 -->
        <el-form-item label="驳回策略" prop="rejectStrategy">
          <el-radio-group v-model="props.data.node.rejectStrategy">
            <el-radio :value="1" style="width: 22%;">驳回到发起人</el-radio>
            <el-radio :value="2" style="width: 30%;">驳回到上一审批节点</el-radio>
            <el-radio :value="3" style="width: 25%;">驳回到指定节点</el-radio>
            <el-radio :value="4" style="width: 22%;">终止流程</el-radio>
            <el-radio :value="5" style="width: 30%;">驳回到模型父节点</el-radio>
          </el-radio-group>
          <div style="padding-top: 8px; width: 100%;">
            <el-alert
              title="「上一审批节点」按本单实际流转路径回退；「模型父节点」按流程图结构取父级（条件分支可能不是审批人）。"
              type="info"
              show-icon
              :closable="false"
            />
          </div>
        </el-form-item>

        <el-form-item
          v-if="props.data.node.rejectStrategy == 3"
          label="指定驳回节点"
          prop="rejectNodeKey"
        >
          <el-select
            v-model="props.data.node.extendConfig.rejectNodeKey"
            clearable
            filterable
            placeholder="选择可驳回的目标节点"
            style="width: 100%"
          >
            <el-option
              v-for="n in rejectableNodes"
              :key="n.nodeKey"
              :label="`${n.nodeName}（${n.nodeKey}）`"
              :value="n.nodeKey"
            />
          </el-select>
          <div style="padding-top: 8px; width: 100%;">
            <el-alert
              title="未指定时，办理人驳回时可从已走过的节点中选择"
              type="info"
              show-icon
              :closable="false"
            />
          </div>
        </el-form-item>

        <!-- 驳回后重新审批 -->
        <el-form-item label="驳回后重新审批" prop="rejectStart">
          <el-radio-group v-model="props.data.node.rejectStart">
            <el-radio :value="1">从发起后继续往下</el-radio>
            <el-radio :value="2">回到原驳回节点</el-radio>
          </el-radio-group>
          <div style="padding-top: 8px; width: 100%;">
            <el-alert
              title="发起人改单重提后：选「继续往下」则按流程从头往后走；选「回到原驳回节点」则直接回到当初驳回你的那个节点。"
              type="info"
              show-icon
              :closable="false"
            />
          </div>
        </el-form-item>

        <el-form-item label="提示">
          <el-alert :title="`* 连续多级上级 / 连续多级部门负责人节点不支持加签、减签。`" type="info" :closable="false" style="margin-bottom: 10px;" />
          <el-alert :title="`* 依次审批支持加签：加签人按插入位置加入审批顺序；会签/或签加签后立即参与本节点。`" type="info" :closable="false" />
        </el-form-item>
      </template>
    </el-form>

    <role-select ref="roleRef" @success="roleSuccess"></role-select>
    <user-select ref="userRef" @success="userSuccess"></user-select>
    <form-select ref="formSelectRef" :form-types="[1]" @success="subFormSuccess"></form-select>

    <el-dialog
      v-model="previewVisible"
      title="表单预览"
      width="720px"
      destroy-on-close
      append-to-body
    >
      <div v-loading="previewLoading" class="sub-form-preview">
        <EBuilder v-if="previewSchema" :page-schema="previewSchema" disabled />
        <el-empty v-else-if="!previewLoading" description="暂无表单内容" :image-size="72" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">

import { ref, defineProps, defineExpose, watch, computed } from "vue"
import { View } from "@element-plus/icons-vue"
import { EBuilder, type PageSchema } from "epic-designer"
import { ElMessage } from "element-plus"
import roleSelect from "../../common/role/index.vue"
import userSelect from "../../common/user/index.vue"
import formSelect from "../../common/form/index.vue"
import config from "../../../config"
import { getMergedFormConfig } from "../../common/utils"
import { getFlowForm } from "@/api"
import { normalizeProcessForm } from "@/views/flow/process/components/formSchema"

const props = defineProps<{ 
  data: any, 
  flow: any,
}>()

/** 子流程只能被自动启动，无法在发起页选人 */
const isChildProcess = computed(() => String(props.flow?.processType || '') === 'child')

/** 可指定驳回的目标：发起人 + 其他审批节点（不含当前） */
const rejectableNodes = computed(() => {
  const root = props.flow?.modelContent?.nodeConfig
  const curKey = props.data?.node?.nodeKey
  const out: { nodeKey: string; nodeName: string }[] = []
  const walk = (n: any) => {
    if (!n || typeof n !== 'object') return
    const t = Number(n.type)
    if ((t === 0 || t === 1) && n.nodeKey && n.nodeKey !== curKey) {
      out.push({
        nodeKey: String(n.nodeKey),
        nodeName: String(n.nodeName || n.nodeKey),
      })
    }
    if (n.childNode) walk(n.childNode)
    for (const list of [n.conditionNodes, n.parallelNodes, n.inclusiveNodes]) {
      if (!Array.isArray(list)) continue
      for (const c of list) {
        walk(c)
        if (c?.childNode) walk(c.childNode)
      }
    }
  }
  walk(root)
  return out
})

const rules = ref({
  nodeName: [
    { required: true, message: '请输入节点名称', trigger: 'blur' },
  ],
  nodeKey: [
    { required: true, message: '请输入节点key', trigger: 'blur' },
  ],
})

const formRef = ref()
const getFormRef = () => {
  return formRef.value
}

defineExpose({
  getFormRef,
})

const number = ref("")
const timer = ref("m")
const tabs = ref('1')

watch(() => props.data.node.setType, (n , o) => {
  props.data.node.nodeAssigneeList = []
  props.data.node.nodeCandidate.type = 0;
  props.data.node.nodeCandidate.assignees = []
})
watch(() => props.data.node.delayType, (n , o) => {
  if (n == 1) {
    props.data.node.extendConfig.remindTime = "1:m"
    number.value = '1'
    timer.value = 'm'
  }
  else {
    props.data.node.extendConfig.remindTime = ""
  }
})
watch(() => number.value, (n, o) => {
  if (props.data.node.delayType == 1 && n) {
    props.data.node.extendConfig.remindTime = `${n}:${timer.value}`
  }
})
watch(() => timer.value, (n, o) => {
  if (props.data.node.delayType == 1) {
    props.data.node.extendConfig.remindTime = `${number.value}:${n}`
  }
})
watch(() => props.data.node.selectMode, (n , o) => {
  props.data.node.nodeCandidate.assignees = []
  if (n == 3) {
    props.data.node.nodeCandidate.type = 1;
  }
  else {
    props.data.node.nodeCandidate.type = 0;
  }
})


props.data.node = Object.assign(JSON.parse(JSON.stringify(config.nodes.find((item: any) => item.type == props.data.node.type)?.config)), props.data.node)
if (isChildProcess.value && Number(props.data.node.setType) === 4) {
  props.data.node.setType = 1
  if (props.data.node.nodeCandidate) {
    props.data.node.nodeCandidate.type = 0
    props.data.node.nodeCandidate.assignees = []
  }
}
if (!props.data.node.extendConfig || typeof props.data.node.extendConfig !== 'object') {
  props.data.node.extendConfig = {}
}
if (props.data.node.extendConfig.rejectNodeKey == null) {
  props.data.node.extendConfig.rejectNodeKey = ''
}
// 兼容旧数据：连续主管终点曾误写入 examineLevel
if (props.data.node.setType == 6 && !(Number(props.data.node.directorLevel) > 0) && Number(props.data.node.examineLevel) > 0) {
  props.data.node.directorLevel = props.data.node.examineLevel
}
function parseActionUrl(url: string) {
  const raw = String(url || '').trim()
  if (!raw) return null
  const idx = raw.indexOf(':')
  if (idx < 0) {
    const id = Number(raw)
    return Number.isFinite(id) && id > 0 ? { id, name: raw } : null
  }
  const id = Number(raw.slice(0, idx))
  const name = raw.slice(idx + 1) || String(id)
  if (!Number.isFinite(id) || id <= 0) return null
  return { id, name }
}

/** 主表单字段 + 子表单字段一并进权限表（子表新字段默认可编辑，便于补录） */
async function syncFormConfigWithSubForm() {
  if (!props.data?.node?.extendConfig) return
  const sources: { form: any; defaultOpera: 0 | 1 | 2 }[] = [
    { form: props.flow.processForm, defaultOpera: 0 },
  ]
  const sub = parseActionUrl(props.data.node.actionUrl)
  if (sub?.id) {
    try {
      const res = await getFlowForm(sub.id)
      sources.push({ form: res.data?.formSchema, defaultOpera: 1 })
    } catch {
      /* 保留主表单权限 */
    }
  }
  props.data.node.extendConfig.formConfig = getMergedFormConfig(
    sources,
    props.data.node.extendConfig.formConfig,
  )
}

void syncFormConfigWithSubForm()

watch(
  () => [props.flow?.processForm, props.data?.node?.actionUrl],
  () => {
    void syncFormConfigWithSubForm()
  },
  { deep: true },
)

if (props.data.node.delayType == 1 && props.data.node.extendConfig.remindTime) {
  let strs = props.data.node.extendConfig.remindTime.split(':')
  number.value = strs[0]
  timer.value = strs[1]
}

const subFormTag = computed(() => parseActionUrl(props.data?.node?.actionUrl))

const formSelectRef = ref()
function handleSubForm() {
  formSelectRef.value?.init({ formId: subFormTag.value?.id })
}

function clearSubForm() {
  props.data.node.actionUrl = ''
  void syncFormConfigWithSubForm()
}

function subFormSuccess(item: { id: number; name: string; formType?: number }) {
  if (Number(item.formType) === 2) {
    ElMessage.warning('子表单仅支持设计表单')
    return
  }
  props.data.node.actionUrl = `${item.id}:${item.name}`
  void syncFormConfigWithSubForm()
}

const previewVisible = ref(false)
const previewLoading = ref(false)
const previewSchema = ref<PageSchema | null>(null)

async function previewSubForm() {
  const sub = subFormTag.value
  if (!sub?.id) return
  previewVisible.value = true
  previewLoading.value = true
  previewSchema.value = null
  try {
    const res = await getFlowForm(sub.id)
    previewSchema.value = normalizeProcessForm(res.data?.formSchema)
  } catch {
    ElMessage.error('加载子表单失败')
    previewVisible.value = false
  } finally {
    previewLoading.value = false
  }
}

const roleRef = ref()
const handleRole = () => {
  if (props.data.node.setType == 4) {
    roleRef.value.init({
      selectData: props.data.node.nodeCandidate.assignees
    })
    return;
  }

  roleRef.value.init({
    selectData: props.data.node.nodeAssigneeList
  })
}
const roleSuccess = (e: any) => {
  if (props.data.node.setType == 4) {
    props.data.node.nodeCandidate.assignees = e
    return;
  }

  props.data.node.nodeAssigneeList = e
}
const roleDelete = (i: number) => {
  if (props.data.node.setType == 4) {
    props.data.node.nodeCandidate.assignees.splice(i, 1)
    return;
  }

  props.data.node.nodeAssigneeList.splice(i, 1)
}

const userRef = ref()
const handleUser = () => {
  if (props.data.node.setType == 4) {
    userRef.value.init({
      selectData: props.data.node.nodeCandidate.assignees
    })
    return;
  }

  userRef.value.init({
    selectData: props.data.node.nodeAssigneeList
  })
}
const userSuccess = (e: any) => {
  if (props.data.node.setType == 4) {
    if (props.data.node.selectMode == 1 && e.length > 1) {
      ElMessage.warning("最多只能选择1个人员")
      return;
    }
    props.data.node.nodeCandidate.assignees = e
    return;
  }

  props.data.node.nodeAssigneeList = e
}
const userDelete = (i: number) => {
  if (props.data.node.setType == 4) {
    props.data.node.nodeCandidate.assignees.splice(i, 1)
    return;
  }

  props.data.node.nodeAssigneeList.splice(i, 1)
}

</script>

<style lang="scss" scoped>

.tags {
  width: 100%;
  padding-top: 10px;
  :deep(.el-tag) {
    margin-right: 10px;
  }
  .tips {
    font-size: 14px;
    color: #999;
  }
}
.sub-form-row {
  width: 100%;
  padding-top: 10px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.sub-form-preview {
  min-height: 200px;
}
.table {
  width: 100%;
  :deep(.el-radio) {
    margin-right: 15px;
  }
}
.selectBtn {
  display: flex;
  .tips {
    font-size: 14px;
    padding-left: 10px;
    color: #999;
  }
}


</style>