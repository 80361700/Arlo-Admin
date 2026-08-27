import request from '../request'

export interface JobItem {
  id: number
  name: string
  handler: string
  cron: string
  params: string
  status: number
  remark: string
  lastRunAt: string
  lastStatus: number | null
  nextRunAt: string
  createdAt: string
  updatedAt: string
}

export interface JobHandlerItem {
  code: string
  name: string
  description: string
}

export interface JobLogItem {
  id: number
  jobId: number
  jobName: string
  handler: string
  triggerType: number
  status: number
  result: string
  errorMsg: string
  durationMs: number
  createdAt: string
}

export function getJobList(params: {
  page: number
  pageSize: number
  name?: string
  handler?: string
  status?: number
}) {
  return request.get<{ list: JobItem[]; total: number; page: number; pageSize: number }>(
    '/v1/monitor/job/list',
    params,
  )
}

export function getJobDetail(id: number) {
  return request.get<JobItem>(`/v1/monitor/job/${id}`)
}

export function getJobHandlers() {
  return request.get<JobHandlerItem[]>('/v1/monitor/job/handlers')
}

export function createJob(data: {
  name: string
  handler: string
  cron: string
  params?: string
  status?: number
  remark?: string
}) {
  return request.post<JobItem>('/v1/monitor/job', data)
}

export function updateJob(id: number, data: {
  name: string
  cron: string
  params?: string
  remark?: string
}) {
  return request.put(`/v1/monitor/job/${id}`, data)
}

export function updateJobStatus(id: number, status: number) {
  return request.put(`/v1/monitor/job/${id}/status`, { status })
}

export function deleteJob(id: number) {
  return request.delete(`/v1/monitor/job/${id}`)
}

export function runJob(id: number) {
  return request.post<{ ok: boolean; msg: string }>(`/v1/monitor/job/${id}/run`)
}

export function getJobLogList(params: {
  page: number
  pageSize: number
  jobId?: number
  status?: number
  triggerType?: number
}) {
  return request.get<{ list: JobLogItem[]; total: number; page: number; pageSize: number }>(
    '/v1/monitor/job/log/list',
    params,
  )
}

export function getJobLogDetail(logId: number) {
  return request.get<JobLogItem>(`/v1/monitor/job/log/${logId}`)
}
