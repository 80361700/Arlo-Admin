import request from '../request'

// ===================== 通知公告 =====================
export interface NoticeItem {
  id: number
  title: string
  content: string
  type: number        // 1=通知 2=公告
  level: number       // 1=普通 2=重要 3=紧急
  status: number      // 0=草稿 1=已发布 2=已撤回
  publisherId: number
  publisher: string
  createdAt: string
  updatedAt: string
}

export interface NoticePageResult {
  list: NoticeItem[]
  total: number
  page: number
  pageSize: number
}

export interface NoticeListQuery {
  title?: string
  status?: number
  type?: number
  page: number
  pageSize: number
}

export interface NoticeFormParams {
  title: string
  content: string
  type: number
  level: number
}

const BASE = '/v1/message/notice'

export function getNoticeList(params: NoticeListQuery) {
  return request.get<NoticePageResult>(`${BASE}/list`, params)
}

export function getNoticeDetail(id: number) {
  return request.get<NoticeItem>(`${BASE}/${id}`)
}

export function createNotice(data: NoticeFormParams) {
  return request.post<NoticeItem>(BASE, data)
}

export function updateNotice(id: number, data: NoticeFormParams) {
  return request.put<NoticeItem>(`${BASE}/${id}`, data)
}

export function deleteNotice(id: number) {
  return request.delete(`${BASE}/${id}`)
}

export function publishNotice(id: number) {
  return request.put(`${BASE}/${id}/publish`)
}

export function revokeNotice(id: number) {
  return request.put(`${BASE}/${id}/revoke`)
}
