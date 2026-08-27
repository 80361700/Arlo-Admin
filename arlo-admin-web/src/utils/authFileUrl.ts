import { resolveApiUrl } from '@/api/request'
import { parseAccessKey, getFileAccessUrl } from '@/utils/fileUrl'
import { useAuthStore } from '@/stores/auth'

/** accessKey → blob: URL（会话内缓存，避免列表重复拉取） */
const cache = new Map<string, string>()
const inflight = new Map<string, Promise<string>>()

function currentToken(): string {
  try {
    const store = useAuthStore()
    if (store.accessToken) return store.accessToken
  } catch {
    /* pinia 未就绪时走 localStorage */
  }
  return localStorage.getItem('accessToken') || ''
}

/**
 * 带 Authorization 拉取文件并转为 blob URL。
 * 公开文件未登录也可直接用 getFileAccessUrl；后台有登录态时统一走此方法。
 */
export async function acquireAuthFileUrl(
  ref: string | number | null | undefined,
): Promise<string> {
  const s = String(ref ?? '').trim()
  if (!s) return ''

  if (/^https?:\/\//i.test(s) && !s.includes('/file/')) return s

  const key = parseAccessKey(s)
  if (!key) return s

  const hit = cache.get(key)
  if (hit) return hit

  const pending = inflight.get(key)
  if (pending) return pending

  const task = (async () => {
    const token = currentToken()
    const url = getFileAccessUrl(key, true)
    const res = await fetch(url, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
    if (!res.ok) {
      throw new Error(`文件加载失败 (${res.status})`)
    }
    const blob = await res.blob()
    const obj = URL.createObjectURL(blob)
    cache.set(key, obj)
    return obj
  })()

  inflight.set(key, task)
  try {
    return await task
  } finally {
    inflight.delete(key)
  }
}

/** 下载（attachment）：拉流后触发浏览器下载，不依赖 ?token= */
export async function downloadAuthFile(
  ref: string | number | null | undefined,
  filename?: string,
): Promise<void> {
  const key = parseAccessKey(ref)
  if (!key) throw new Error('无效文件引用')

  const token = currentToken()
  const url = resolveApiUrl(`/v1/file/${key}`)
  const res = await fetch(url, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) throw new Error(`下载失败 (${res.status})`)
  const blob = await res.blob()
  const obj = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = obj
  a.download = filename || key
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(obj)
}

/** 登出时释放全部 blob */
export function clearAuthFileCache() {
  for (const url of cache.values()) {
    URL.revokeObjectURL(url)
  }
  cache.clear()
  inflight.clear()
}
