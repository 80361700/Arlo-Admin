import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

/** 带鉴权下载文件（Excel 等二进制响应） */
export async function downloadFile(url: string, params?: Record<string, any>, filename?: string) {
  const authStore = useAuthStore()
  const baseURL = import.meta.env.VITE_API_BASE_URL || ''
  const res = await axios.get(url, {
    baseURL,
    params,
    responseType: 'blob',
    headers: {
      Authorization: authStore.accessToken ? `Bearer ${authStore.accessToken}` : '',
    },
  })

  let name = filename || 'download.xlsx'
  const disposition = res.headers['content-disposition'] as string | undefined
  if (disposition) {
    const m = /filename\*=UTF-8''([^;]+)|filename="?([^";]+)"?/i.exec(disposition)
    const raw = decodeURIComponent((m?.[1] || m?.[2] || '').trim())
    if (raw) name = raw
  }

  const blob = new Blob([res.data])
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = name
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(link.href)
}
