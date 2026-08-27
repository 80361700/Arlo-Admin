import request from '../request'

// ===================== 站内信 =====================
export interface MessageItem {
  id: number
  title: string
  content: string
  type: number       // 1=系统消息 2=通知 3=私信
  senderId: number
  sender: string
  receiverId: number  // 0=全部用户
  receiverCount: number // 接收者数量（发送记录聚合用）
  isRead: number      // 0=未读 1=已读
  readAt: string | null
  createdAt: string
}

export interface MessagePageResult {
  list: MessageItem[]
  total: number
  page: number
  pageSize: number
}

export interface MessageListQuery {
  isRead?: number
  direction?: number  // 0=全部 1=我收到的 2=我发送的
  page: number
  pageSize: number
}

export interface SendMessageParams {
  title: string
  content: string
  type: number
  /** 接收者 ID 列表；省略或空数组 = 广播全部用户 */
  receiverIds?: number[]
}

export interface UnreadCountResult {
  count: number
}

const BASE = '/v1/message'

export function getMessageList(params: MessageListQuery) {
  return request.get<MessagePageResult>(`${BASE}/list`, params)
}

export function sendMessage(data: SendMessageParams) {
  return request.post(`${BASE}/send`, data)
}

export function markMessageRead(id: number) {
  return request.put(`${BASE}/${id}/read`)
}

export function markAllMessageRead() {
  return request.put(`${BASE}/read-all`)
}

export function deleteMessage(id: number, side: 'sent' | 'received' = 'received') {
  return request.delete(`${BASE}/${id}`, { side })
}

export function getUnreadCount() {
  // 无消息菜单权限时会 403；轮询场景由 store 静默忽略，勿弹全局 toast
  return request.get<UnreadCountResult>(`${BASE}/unread-count`, undefined, { skipErrorToast: true })
}
