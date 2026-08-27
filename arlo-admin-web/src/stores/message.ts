import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUnreadCount } from '@/api/modules/message'
import { useAuthStore } from '@/stores/auth'

export const useMessageStore = defineStore('message', () => {
  const unreadCount = ref(0)
  let socket: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let closedByUser = false
  let reconnectAttempt = 0

  async function fetchUnreadCount() {
    try {
      const res = await getUnreadCount()
      unreadCount.value = Number(res.data?.count) || 0
    } catch {
      /* 无权限或未登录时忽略 */
    }
  }

  function buildWsUrl(token: string) {
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    // 开发走 Vite 代理同源 /api；生产同域反代 /api
    return `${proto}//${window.location.host}/api/v1/ws?token=${encodeURIComponent(token)}`
  }

  function connect() {
    const authStore = useAuthStore()
    const token = authStore.accessToken
    if (!token) return

    stopRealtime()
    closedByUser = false
    void fetchUnreadCount()

    try {
      socket = new WebSocket(buildWsUrl(token))
    } catch {
      scheduleReconnect()
      return
    }

    socket.onopen = () => {
      reconnectAttempt = 0
    }

    socket.onmessage = (ev) => {
      try {
        const data = JSON.parse(String(ev.data || '{}')) as { type?: string }
        if (data.type === 'unread_changed') {
          void fetchUnreadCount()
        }
      } catch {
        /* ignore bad frame */
      }
    }

    socket.onclose = () => {
      socket = null
      if (!closedByUser) scheduleReconnect()
    }

    socket.onerror = () => {
      /* onclose 会跟着重连 */
    }
  }

  function scheduleReconnect() {
    if (closedByUser || reconnectTimer) return
    const delay = Math.min(30_000, 1000 * 2 ** Math.min(reconnectAttempt, 4))
    reconnectAttempt += 1
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, delay)
  }

  function stopRealtime() {
    closedByUser = true
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (socket) {
      socket.onclose = null
      socket.close()
      socket = null
    }
  }

  /** @deprecated 兼容旧调用名，内部改为 WebSocket */
  function startPolling() {
    connect()
  }

  /** @deprecated */
  function stopPolling() {
    stopRealtime()
  }

  function clear() {
    stopRealtime()
    unreadCount.value = 0
  }

  return {
    unreadCount,
    fetchUnreadCount,
    connect,
    stopRealtime,
    startPolling,
    stopPolling,
    clear,
  }
})
