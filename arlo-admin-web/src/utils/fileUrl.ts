import { resolveApiUrl } from '@/api/request'

/** 32 位 hex access_key */
const ACCESS_KEY_RE = /^[a-f0-9]{32}$/i

/** 从业务引用解析 access_key（支持纯 key、统一 /file/{key}） */
export function parseAccessKey(ref: string | number | null | undefined): string | null {
  const s = String(ref ?? '').trim()
  if (!s) return null
  if (ACCESS_KEY_RE.test(s)) return s.toLowerCase()

  const unified = s.match(/\/file\/([a-f0-9]{32})(?:\?|$)/i)
  if (unified) return unified[1].toLowerCase()

  return null
}

/** 选文件后写入业务字段：存 accessKey */
export function toFileRef(file: { accessKey: string } | string): string {
  if (typeof file === 'string') return file
  return file.accessKey
}

/**
 * 统一访问地址 /v1/file/{accessKey}（不含 JWT）。
 * 公开文件可直接用于 <img>；私有文件请用 acquireAuthFileUrl（Bearer → blob）。
 */
export function getFileAccessUrl(accessKey: string, inline = true): string {
  const path = resolveApiUrl(`/v1/file/${accessKey}`)
  return inline ? `${path}?inline=1` : path
}

/**
 * 解析为可直接访问的 URL（无鉴权头）。
 * 登录页 Logo 等公开资源用此方法；后台私有预览请用 useAuthFileSrc / AuthAvatar。
 */
export function resolveFileSrc(ref: string | number | null | undefined): string {
  const s = String(ref ?? '').trim()
  if (!s) return ''

  if (/^https?:\/\//i.test(s) && !s.includes('/file/')) return s

  const key = parseAccessKey(s)
  if (!key) {
    if (/^https?:\/\//i.test(s)) return s
    return s
  }

  return getFileAccessUrl(key, true)
}
