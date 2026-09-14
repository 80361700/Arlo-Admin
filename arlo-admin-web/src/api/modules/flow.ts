import request from '../request'

export interface FlowCategoryTreeItem {
  categoryId: number
  categoryName: string
  categoryRemark: string
  categorySort: number
  processList: FlowProcessBrief[]
}

export interface FlowCategoryOption {
  id: number
  name: string
}

export interface FlowProcessBrief {
  processId: number
  categoryId: number
  processKey: string
  processName: string
  processIcon: string
  processType: string
  processVersion: number
  processState: number
  remark: string
  createdAt: string
  updatedAt: string
}

export interface FlowProcessDetail {
  processId: number
  categoryId: number
  processKey: string
  processName: string
  processIcon: string
  processBgcolor: string
  processType: string
  processVersion: number
  processState: number
  remark: string
  modelContent: Record<string, any>
  processForm: Record<string, any>
  processSetting: Record<string, any>
  processPermissionList: Record<string, any>[]
  createdAt: string
  updatedAt: string
}

export function getFlowCategoryTree() {
  return request.get<FlowCategoryTreeItem[]>('/v1/flow/category/tree')
}

export function getFlowCategoryOptions() {
  return request.get<FlowCategoryOption[]>('/v1/flow/category/options')
}

export function createFlowCategory(data: { name: string; sort?: number; remark?: string }) {
  return request.post('/v1/flow/category', data)
}

export function updateFlowCategory(data: { id: number; name: string; sort?: number; remark?: string }) {
  return request.put('/v1/flow/category', data)
}

export function deleteFlowCategory(id: number) {
  return request.delete(`/v1/flow/category/${id}`)
}

export function getFlowProcess(id: number) {
  return request.get<FlowProcessDetail>(`/v1/flow/process/${id}`)
}

export function saveFlowProcess(data: Record<string, any>) {
  return request.post<FlowProcessDetail>('/v1/flow/process', data)
}

export function updateFlowProcessState(id: number, state: number) {
  return request.put(`/v1/flow/process/${id}/state`, { state })
}

/** 启用前配置体检（不改状态） */
export function checkFlowProcess(id: number) {
  return request.get<{ ok: boolean; message: string; issues?: string[] }>(`/v1/flow/process/${id}/check`)
}

export function cloneFlowProcess(id: number) {
  return request.post<FlowProcessDetail>(`/v1/flow/process/${id}/clone`)
}

export function deleteFlowProcess(id: number) {
  return request.delete(`/v1/flow/process/${id}`)
}

export interface FlowProcessHistoryBrief {
  historyId: number
  processId: number
  processName: string
  processIcon: string
  processVersion: number
  remark: string
  createdAt: string
}

export interface FlowProcessHistoryDetail extends FlowProcessDetail {
  historyId: number
}

export function getFlowProcessHistories(processId: number, params?: { page?: number; pageSize?: number }) {
  return request.get<{
    list: FlowProcessHistoryBrief[]
    total: number
    page: number
    pageSize: number
  }>(`/v1/flow/process/${processId}/histories`, params)
}

export function getFlowProcessHistory(processId: number, historyId: number) {
  return request.get<FlowProcessHistoryDetail>(`/v1/flow/process/${processId}/histories/${historyId}`)
}

export function checkoutFlowProcessHistory(processId: number, historyId: number) {
  return request.post<FlowProcessDetail>(`/v1/flow/process/${processId}/histories/${historyId}/checkout`)
}

export interface FlowProcessOption {
  id: number
  name: string
  key: string
}

/** 子流程等下拉：仅启用中的流程；excludeId 排除当前流程自身，避免自调用死循环 */
export function getFlowProcessOptions(excludeId?: number) {
  return request.get<FlowProcessOption[]>(
    '/v1/flow/process/options',
    excludeId && excludeId > 0 ? { excludeId } : undefined,
  )
}

// ---------- 表单模板 ----------

export interface FlowFormCategoryTreeItem {
  categoryId: number
  categoryName: string
  categoryRemark: string
  categorySort: number
  formList: FlowFormBrief[]
}

export interface FlowFormCategoryOption {
  id: number
  name: string
}

export interface FlowFormBrief {
  formId: number
  categoryId: number
  name: string
  code: string
  formType: number
  pcUrl?: string
  status: number
  sort: number
  remark: string
  bound: boolean
  createdAt: string
  updatedAt: string
}

export interface FlowFormDetail {
  formId: number
  categoryId: number
  name: string
  code: string
  formType: number
  pcUrl?: string
  status: number
  sort: number
  remark: string
  formSchema: Record<string, any>
  bound: boolean
  createdAt: string
  updatedAt: string
}

export interface FlowFormOption {
  id: number
  name: string
  code: string
  formType?: number
  pcUrl?: string
}

export function getFlowFormCategoryTree() {
  return request.get<FlowFormCategoryTreeItem[]>('/v1/flow/form-category/tree')
}

export function getFlowFormCategoryOptions() {
  return request.get<FlowFormCategoryOption[]>('/v1/flow/form-category/options')
}

export function createFlowFormCategory(data: { name: string; sort?: number; remark?: string }) {
  return request.post('/v1/flow/form-category', data)
}

export function updateFlowFormCategory(data: { id: number; name: string; sort?: number; remark?: string }) {
  return request.put('/v1/flow/form-category', data)
}

export function deleteFlowFormCategory(id: number) {
  return request.delete(`/v1/flow/form-category/${id}`)
}

export function getFlowForm(id: number) {
  return request.get<FlowFormDetail>(`/v1/flow/form/${id}`)
}

export function saveFlowForm(data: Record<string, any>) {
  return request.post<FlowFormDetail>('/v1/flow/form', data)
}

export function updateFlowFormState(id: number, status: number) {
  return request.put(`/v1/flow/form/${id}/state`, { status })
}

export function saveFlowFormSchema(id: number, formSchema: Record<string, any>) {
  return request.put<FlowFormDetail>(`/v1/flow/form/${id}/schema`, { formSchema })
}

export function deleteFlowForm(id: number) {
  return request.delete(`/v1/flow/form/${id}`)
}

/** 节点选子表单：仅启用中的设计表单 */
export function getFlowFormOptions() {
  return request.get<FlowFormOption[]>('/v1/flow/form/options')
}

// ===================== 审批运行时 =====================

export interface LaunchProcessItem {
  processId: number
  categoryId: number
  categoryName: string
  processKey: string
  processName: string
  processIcon: string
  processBgcolor: string
  processType: string
  remark: string
}

export function getLaunchProcessList(name?: string) {
  return request.get<LaunchProcessItem[]>('/v1/flow/approve/launch/list', { name })
}

export function getLaunchProcessForm(processId: number) {
  return request.get<LaunchFormDetail>(`/v1/flow/approve/launch/${processId}`)
}

export interface LaunchSelectionNode {
  nodeKey: string
  nodeName: string
  kind: 'assignee' | 'cc' | string
  selectMode: number
  required: boolean
}

export interface LaunchFormDetail {
  processId: number
  processKey: string
  processName: string
  processType: string
  processForm: Record<string, any>
  modelContent: Record<string, any>
  selectionNodes: LaunchSelectionNode[]
  /** designer | vue */
  formRenderType?: string
  /** vue 时：相对 views 的组件路径 */
  formComponent?: string
}

export function launchFlowProcess(data: {
  processId: number
  formData?: Record<string, any>
  nodeAssignees?: Record<string, { id: number; name: string }[]>
  nodeCcUsers?: Record<string, { id: number; name: string }[]>
  saveAsDraft?: boolean
}) {
  return request.post<{ instanceId: number }>('/v1/flow/approve/launch', data)
}

export function activateApproveDraft(data: {
  instanceId: number
  formData?: Record<string, any>
  nodeAssignees?: Record<string, { id: number; name: string }[]>
  nodeCcUsers?: Record<string, { id: number; name: string }[]>
}) {
  return request.post('/v1/flow/approve/draft/activate', data)
}

export function updateApproveDraft(data: {
  instanceId: number
  formData?: Record<string, any>
  nodeAssignees?: Record<string, { id: number; name: string }[]>
  nodeCcUsers?: Record<string, { id: number; name: string }[]>
}) {
  return request.post('/v1/flow/approve/draft/update', data)
}

export function deleteApproveDraft(data: { instanceId: number }) {
  return request.post('/v1/flow/approve/draft/delete', data)
}

export interface PendingTaskItem {
  actorId: number
  actorUserId?: number
  taskId: number
  instanceId: number
  parentInstanceId?: number
  nodeName: string
  workNodeName?: string
  nodeKey: string
  examineMode: number
  processName: string
  /** 子实例自身流程名（展示标题已换成主流程名） */
  subProcessName?: string
  isSubProcess?: boolean
  createBy: string
  createId: number
  createdAt: string
  canHandle: boolean
  allowBatchOperate?: boolean
  isDelegate?: boolean
}

export type ApproveListQuery = {
  page?: number
  pageSize?: number
  keyword?: string
  createBy?: string
  instanceState?: number
  beginTime?: string
  endTime?: string
}

export function getPendingTasks(params?: ApproveListQuery) {
  return request.get<{ list: PendingTaskItem[]; total: number; page: number; pageSize: number }>(
    '/v1/flow/approve/pending',
    params,
  )
}

export interface ApproveInstanceListItem {
  /** 历史参与人行 id（已办/抄送列表唯一键） */
  id?: number
  instanceId: number
  processName: string
  subProcessName?: string
  isSubProcess?: boolean
  processKey: string
  currentNodeName?: string
  instanceState: number
  createBy: string
  createdAt: string
  finishTime?: string
  nodeName?: string
  taskId?: number
  hisTaskId?: number
  actorState?: number
  /** 当前用户作为受托人代批完成 */
  isDelegate?: boolean
  /** 抄送是否已读（我收到的） */
  read?: boolean
}

export function getMyApplications(params?: ApproveListQuery) {
  return request.get<{ list: ApproveInstanceListItem[]; total: number; page: number; pageSize: number }>(
    '/v1/flow/approve/mine',
    params,
  )
}

export function getMonitorInstances(params?: ApproveListQuery & { onlyActive?: number }) {
  return request.get<{ list: ApproveInstanceListItem[]; total: number; page: number; pageSize: number }>(
    '/v1/flow/approve/monitor',
    params,
  )
}

export function getReceivedApprovals(params?: ApproveListQuery) {
  return request.get<{ list: ApproveInstanceListItem[]; total: number; page: number; pageSize: number }>(
    '/v1/flow/approve/received',
    params,
  )
}

/** 标记抄送已读（我收到的列表行 id） */
export function markReceivedApprovalRead(id: number) {
  return request.post('/v1/flow/approve/received/read', { id })
}

export function getApprovedList(params?: ApproveListQuery) {
  return request.get<{ list: ApproveInstanceListItem[]; total: number; page: number; pageSize: number }>(
    '/v1/flow/approve/approved',
    params,
  )
}

export function getClaimableTasks(params?: ApproveListQuery) {
  return request.get<{ list: ApproveInstanceListItem[]; total: number; page: number; pageSize: number }>(
    '/v1/flow/approve/claim',
    params,
  )
}

export function claimApproveTask(data: { taskId: number }) {
  return request.post('/v1/flow/approve/claim', data)
}

export interface InstanceDetail {
  instanceId: number
  workInstanceId?: number
  parentInstanceId?: number
  processId: number
  processName: string
  subProcessName?: string
  isSubProcess?: boolean
  processKey: string
  processVersion: number
  instanceState: number
  currentNodeKey: string
  currentNodeName: string
  workNodeKey?: string
  workNodeName?: string
  createId: number
  createBy: string
  createdAt: string
  finishTime: string
  formSchema: Record<string, any>
  formData: Record<string, any>
  formConfig: { id: string; label: string; opera: number }[]
  /** designer | vue */
  formRenderType?: string
  formComponent?: string
  modelContent: Record<string, any>
  taskId: number
  canConsent: boolean
  canReject: boolean
  canResubmit: boolean
  canActivateDraft?: boolean
  canUpdateDraft?: boolean
  canDeleteDraft?: boolean
  selectionNodes?: LaunchSelectionNode[]
  launchSelection?: {
    assignees?: Record<string, { id: number; name: string }[]>
    ccUsers?: Record<string, { id: number; name: string }[]>
  }
  canTransfer: boolean
  canAppend: boolean
  canRemove: boolean
  canRollback: boolean
  canRevoke: boolean
  canUrge: boolean
  canTerminate?: boolean
  canAdminTransfer?: boolean
  canCc: boolean
  canComment: boolean
  secondOperatePrompt: boolean
  allowDelegate?: boolean
  taskNodeType: number
  rejectStrategy: number
  rejectStart: number
  rejectNodeKey: string
  rejectTargets: { nodeKey: string; nodeName: string }[]
  rollbackTargets: { nodeKey: string; nodeName: string }[]
  allowTransfer: boolean
  allowAppendNode: boolean
  allowRollback: boolean
  allowCc: boolean
  taskActors: {
    id: number
    actorId: number
    actorName: string
    actorType: number
    actorState: number
    weight: number
  }[]
  timeline: {
    nodeName: string
    nodeKey: string
    nodeType: number
    taskState: number
    actorName: string
    agentName?: string
    actorState: number
    opinion: string
    finishTime: string
    createdAt: string
    childInstanceId?: number
    childProcessName?: string
    childInstanceState?: number
  }[]
  comments: {
    id: number
    userId: number
    userName: string
    content: string
    createdAt: string
  }[]
}

export function getApproveInstance(instanceId: number, taskId?: number) {
  return request.get<InstanceDetail>(`/v1/flow/approve/instance/${instanceId}`, {
    taskId: taskId || undefined,
  })
}

export function consentApproveTask(data: { taskId: number; opinion?: string; formData?: Record<string, any> }) {
  return request.post('/v1/flow/approve/consent', data)
}

/** 发起人改单后重新提交 */
export function resubmitApproveTask(data: { taskId: number; formData?: Record<string, any> }) {
  return request.post('/v1/flow/approve/resubmit', data)
}

export function rejectApproveTask(data: {
  taskId: number
  opinion?: string
  formData?: Record<string, any>
  rejectStrategy?: number
  rejectNodeKey?: string
}) {
  return request.post('/v1/flow/approve/reject', data)
}

export function revokeApproveInstance(data: { instanceId: number; opinion?: string }) {
  return request.post('/v1/flow/approve/revoke', data)
}

export function terminateApproveInstance(data: { instanceId: number; opinion?: string }) {
  return request.post('/v1/flow/approve/terminate', data)
}

export function transferApproveTask(data: {
  taskId: number
  toUserId: number
  toName?: string
  opinion?: string
}) {
  return request.post('/v1/flow/approve/transfer', data)
}

export function adminTransferApproveTask(data: {
  taskId: number
  toUserId: number
  toName?: string
  opinion?: string
  fromActorId?: number
}) {
  return request.post('/v1/flow/approve/admin-transfer', data)
}

export function rollbackApproveTask(data: {
  taskId: number
  opinion?: string
  targetNodeKey?: string
}) {
  return request.post('/v1/flow/approve/rollback', data)
}

export function urgeApproveInstance(data: { instanceId: number }) {
  return request.post('/v1/flow/approve/urge', data)
}

export function appendApproveActor(data: {
  taskId: number
  toUserId: number
  toName?: string
  position?: number
}) {
  return request.post('/v1/flow/approve/append', data)
}

export function removeApproveActor(data: { taskId: number; actorId: number }) {
  return request.post('/v1/flow/approve/remove-actor', data)
}

export function runtimeCcApproveTask(data: { taskId: number; users: { id: number; name: string }[] }) {
  return request.post('/v1/flow/approve/cc', data)
}

export function addApproveComment(data: {
  instanceId: number
  content: string
  mentionUserIds?: number[]
}) {
  return request.post('/v1/flow/approve/comment', data)
}

export function batchConsentApproveTasks(data: { taskIds: number[]; opinion?: string }) {
  return request.post<{ success: number; failed: number; errors: string[] }>(
    '/v1/flow/approve/batch-consent',
    data,
  )
}

export function batchRejectApproveTasks(data: {
  taskIds: number[]
  opinion?: string
  /** 0=按节点配置；4=统一终止 */
  rejectStrategy?: number
}) {
  return request.post<{ success: number; failed: number; errors: string[] }>(
    '/v1/flow/approve/batch-reject',
    data,
  )
}

export function getApproveDelegate() {
  return request.get<{ toUserId: number; toUserName: string; enabled: boolean; remark: string }>(
    '/v1/flow/approve/delegate',
  )
}

export function setApproveDelegate(data: {
  toUserId?: number
  toUserName?: string
  enabled?: boolean
  remark?: string
  clear?: boolean
}) {
  return request.put('/v1/flow/approve/delegate', data)
}
