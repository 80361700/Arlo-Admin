import request from '../request'

// ===================== 日志管理 =====================
export interface LoginLogItem {
  id: number
  username: string
  ip: string
  location: string
  browser: string
  os: string
  status: number
  msg: string
  createdAt: string
}

export interface OperationLogItem {
  id: number
  userId: number
  username: string
  module: string
  action: string
  method: string
  url: string
  ip: string
  userAgent: string
  params: string
  result?: string
  costTime: number
  status: number
  errorMsg: string
  createdAt: string
}

export interface LogPageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export interface LoginLogQuery {
  username?: string
  status?: number
  startTime?: string
  endTime?: string
  page: number
  pageSize: number
}

export interface OperationLogQuery {
  username?: string
  module?: string
  url?: string
  startTime?: string
  endTime?: string
  status?: number
  page: number
  pageSize: number
}

export function getLoginLogList(params: LoginLogQuery) {
  return request.get<LogPageResult<LoginLogItem>>('/v1/log/login/list', params)
}

export function getOperationLogList(params: OperationLogQuery) {
  return request.get<LogPageResult<OperationLogItem>>('/v1/log/operation/list', params)
}

export function exportLoginLogs(params?: Partial<LoginLogQuery>) {
  return import('@/utils/download').then(({ downloadFile }) =>
    downloadFile('/v1/log/login/export', params, 'login_logs.xlsx'),
  )
}

export function exportOperationLogs(params?: Partial<OperationLogQuery>) {
  return import('@/utils/download').then(({ downloadFile }) =>
    downloadFile('/v1/log/operation/export', params, 'operation_logs.xlsx'),
  )
}
