import request from '../request'
import type { LaunchFormDetail } from './flow'

export interface DemoPurchaseOrderItem {
  id: number
  title: string
  content: string
  status: number
  instanceId: number
  processId: number
  processKey: string
  processName?: string
  /** 当前用户是否符合设计器发起人权限 */
  canLaunch?: boolean
  createBy: string
  createdAt: string
}

export interface DemoProcessOption {
  processId: number
  processKey: string
  processName: string
  processType: string
  remark: string
}

export function getDemoPurchaseOrderList(params: {
  page: number
  pageSize: number
  title?: string
  status?: number
}) {
  return request.get<{
    list: DemoPurchaseOrderItem[]
    total: number
    page: number
    pageSize: number
  }>('/v1/business/purchase-order/list', params)
}

export function createDemoPurchaseOrder(data: {
  title: string
  content: string
  processId: number
}) {
  return request.post<{ id: number }>('/v1/business/purchase-order', data)
}

export function deleteDemoPurchaseOrders(ids: number[]) {
  return request.post('/v1/business/purchase-order/delete', { ids })
}

export function getDemoProcessOptions() {
  return request.get<DemoProcessOption[]>('/v1/business/purchase-order/process-options')
}

export function getDemoProcessPreview(orderId: number) {
  return request.get<{
    processId: number
    processKey: string
    processName: string
    modelContent: Record<string, any>
  }>(`/v1/business/purchase-order/${orderId}/process-preview`)
}

export function getDemoPurchaseOrderLaunchForm(orderId: number) {
  return request.get<LaunchFormDetail>(`/v1/business/purchase-order/${orderId}/launch-form`)
}

export function launchDemoPurchaseOrder(
  id: number,
  data: {
    formData?: Record<string, any>
    nodeAssignees?: Record<string, { id: number; name: string }[]>
    nodeCcUsers?: Record<string, { id: number; name: string }[]>
  },
) {
  return request.post<{ instanceId: number }>(`/v1/business/purchase-order/${id}/launch`, data)
}
