import request, { resolveApiUrl } from '../request'

// ===================== 文件管理 =====================
export interface FileItem {
  id: number
  accessKey: string
  name: string
  /** 统一访问地址 /api/v1/file/{accessKey} */
  url: string
  size: number
  mimeType: string
  category: string
  isPublic?: number
  md5: string
  uploaderId: number
  uploader: string
  createdAt: string
}

export interface FilePageResult {
  list: FileItem[]
  total: number
  page: number
  pageSize: number
}

export interface FileListQuery {
  name?: string
  mimeType?: string
  category?: string
  isPublic?: number
  page: number
  pageSize: number
}

const BASE = '/v1/file'

export function getFileList(params: FileListQuery) {
  return request.get<FilePageResult>(`${BASE}/list`, params)
}

export function uploadFile(
  file: File,
  options?: { public?: boolean; onUploadProgress?: (progressEvent: any) => void },
) {
  const formData = new FormData()
  formData.append('file', file)
  // 默认公开；仅显式 false 时传私有
  formData.append('public', options?.public === false ? '0' : '1')
  return request.post<FileItem>(`${BASE}/upload`, formData, {
    headers: { 'Content-Type': undefined },
    onUploadProgress: options?.onUploadProgress,
  })
}

export function setFilePublic(id: number, isPublic: boolean) {
  return request.put(`${BASE}/${id}/public`, { public: isPublic })
}

/** 统一访问 URL（不含 JWT；私有文件请用 acquireAuthFileUrl） */
export function getFileAccessUrl(accessKey: string) {
  return resolveApiUrl(`${BASE}/${accessKey}`)
}

/** @deprecated 使用 getFileAccessUrl(accessKey) */
export function getDownloadUrl(accessKey: string) {
  return getFileAccessUrl(accessKey)
}

export function deleteFile(id: number) {
  return request.delete(`${BASE}/${id}`)
}
