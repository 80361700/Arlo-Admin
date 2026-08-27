import axios, { type AxiosInstance, type AxiosRequestConfig, type InternalAxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

declare module 'axios' {
  export interface AxiosRequestConfig {
    /** 为 true 时业务/HTTP 错误不弹全局 toast，交由调用方自行处理 */
    skipErrorToast?: boolean
  }
}

// 统一响应格式（适配后端：code=200 成功，msg 字段）
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

// 分页响应
export interface PageData<T = any> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

/** 拼接带基础路径的完整 API URL（用于 img/src、下载链等不走 axios 的场景） */
export function resolveApiUrl(path: string) {
  const base = String(import.meta.env.VITE_API_BASE_URL || '').replace(/\/$/, '')
  const p = path.startsWith('/') ? path : `/${path}`
  if (base && (p === base || p.startsWith(`${base}/`))) return p
  return `${base}${p}`
}

const instance: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

// 是否正在刷新 token
let isRefreshing = false
// 等待刷新的请求队列
let refreshQueue: Array<{
  resolve: (token: string) => void
  reject: (err: any) => void
}> = []
// 避免并发 401 重复弹窗 / 重复跳转
let isRedirectingToLogin = false

function isAuthExemptUrl(url = '') {
  return url.includes('/auth/login')
    || url.includes('/auth/refresh')
    || url.includes('/auth/captcha')
    || url.includes('/sysconfig/public')
}

async function redirectToLogin(msg?: string) {
  if (isRedirectingToLogin) {
    return
  }
  isRedirectingToLogin = true

  const authStore = useAuthStore()
  authStore.clearSession()
  ElMessage.error(msg || '登录已失效，请重新登录')

  try {
    const { resetDynamicRoutes } = await import('@/router')
    resetDynamicRoutes()
  } catch { /* ignore */ }

  const hash = window.location.hash || ''
  if (!hash.startsWith('#/login')) {
    window.location.hash = '#/login'
  }
  window.location.reload()
}

/** 业务 code=401 或 HTTP 401：尝试刷新，失败则跳登录 */
async function handleUnauthorized(
  config: InternalAxiosRequestConfig & { _retry?: boolean },
  msg?: string,
): Promise<AxiosResponse> {
  if (!config || isAuthExemptUrl(config.url || '') || config._retry) {
    await redirectToLogin(msg)
    return Promise.reject(new Error(msg || '登录已失效'))
  }

  if (isRefreshing) {
    return new Promise((resolve, reject) => {
      refreshQueue.push({
        resolve: (token: string) => {
          config.headers.Authorization = `Bearer ${token}`
          resolve(instance(config))
        },
        reject,
      })
    })
  }

  config._retry = true
  isRefreshing = true

  const authStore = useAuthStore()
  try {
    const newToken = await authStore.refresh()
    config.headers.Authorization = `Bearer ${newToken}`
    refreshQueue.forEach(({ resolve }) => resolve(newToken))
    refreshQueue = []
    return instance(config)
  } catch (refreshError) {
    refreshQueue.forEach(({ reject }) => reject(refreshError))
    refreshQueue = []
    await redirectToLogin(msg || '登录已失效，请重新登录')
    return Promise.reject(refreshError)
  } finally {
    isRefreshing = false
  }
}

// 请求拦截器 — 自动附带 Access Token（刷新接口不带，避免过期 access 干扰）
instance.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const url = config.url || ''
    if (url.includes('/auth/refresh')) {
      return config
    }
    const authStore = useAuthStore()
    if (authStore.accessToken) {
      config.headers.Authorization = `Bearer ${authStore.accessToken}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

function rejectShown(message: string) {
  const err = new Error(message) as Error & { shown?: boolean }
  err.shown = true
  return Promise.reject(err)
}

// 响应拦截器 — 统一错误处理 + 401 自动刷新
instance.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data, config } = response

    // 后端统一 HTTP 200 + body.code；鉴权失败是 code=401
    if (data?.code === 401) {
      return handleUnauthorized(config as InternalAxiosRequestConfig & { _retry?: boolean }, data.msg)
    }

    // 业务层面的错误（code !== 200）
    if (data.code !== 200) {
      // 刷新接口失败：交给上层，由 handleUnauthorized 统一跳转，避免重复 toast
      if (isAuthExemptUrl(config.url || '') && (data.code === 1002 || data.code === 1003)) {
        return rejectShown(data.msg || '登录已失效')
      }
      if (!config.skipErrorToast) {
        ElMessage.error(data.msg || '请求失败')
      }
      return rejectShown(data.msg || '请求失败')
    }

    return response
  },
  async (error) => {
    const { config, response } = error
    const skipToast = !!config?.skipErrorToast

    // 网络错误
    if (!response) {
      if (!skipToast) ElMessage.error('网络连接失败，请检查网络')
      return rejectShown('网络连接失败，请检查网络')
    }

    const { status, data } = response
    const msg = data?.msg as string | undefined

    // HTTP 401（若网关/代理改写状态码）
    if (status === 401) {
      return handleUnauthorized(config, msg)
    }

    // 其他状态码
    const messageMap: Record<number, string> = {
      400: '请求参数错误',
      403: '没有操作权限',
      404: '请求的资源不存在',
      500: '服务器内部错误',
    }
    const text = msg || messageMap[status] || `请求失败 (${status})`
    if (!skipToast) ElMessage.error(text)
    return rejectShown(text)
  },
)

// 封装请求方法
const request = {
  get<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return instance.get(url, { params, ...config }).then((res) => res.data)
  },
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return instance.post(url, data, config).then((res) => res.data)
  },
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return instance.put(url, data, config).then((res) => res.data)
  },
  delete<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return instance.delete(url, { params, ...config }).then((res) => res.data)
  },
}

export default request
