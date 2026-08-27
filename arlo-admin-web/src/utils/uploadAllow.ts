/** 与后端 storage.allowedExts 默认白名单保持一致（前端仅作选择器辅助，真正校验在服务端） */
export const ALLOWED_UPLOAD_EXTS = [
  'jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'ico', 'svg',
  'mp3', 'wav', 'mp4', 'webm', 'mov', 'avi',
  'pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt', 'csv', 'md',
  'zip', 'rar', '7z',
] as const

export const UPLOAD_ACCEPT =
  ALLOWED_UPLOAD_EXTS.map((e) => `.${e}`).join(',')

export function isAllowedUploadFile(file: File | { name: string }): boolean {
  const name = file.name || ''
  const i = name.lastIndexOf('.')
  if (i < 0) return false
  const ext = name.slice(i + 1).toLowerCase()
  return (ALLOWED_UPLOAD_EXTS as readonly string[]).includes(ext)
}
