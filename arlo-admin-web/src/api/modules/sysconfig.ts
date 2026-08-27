import request from '../request'

// ===================== 系统配置 =====================
export interface ConfigItem {
  id: number
  name: string
  key: string
  value: string
  type: number // 1=文本 2=JSON 3=开关 4=图片
  remark: string
  createdAt: string
  updatedAt: string
}

export interface ConfigListQuery {
  name?: string
  key?: string
  type?: number
}

export interface CreateConfigParams {
  name: string
  key: string
  value: string
  type: number
  remark?: string
}

export interface UpdateConfigParams {
  id: number
  name: string
  key: string
  value: string
  type: number
  remark?: string
}

/** 公开配置（登录页免鉴权） */
export interface PublicConfig {
  name: string
  captcha: boolean
  logo: string
  version: string
}

const BASE = '/v1/sysconfig'

export function getConfigList(params?: ConfigListQuery) {
  return request.get<ConfigItem[]>(`${BASE}/list`, params)
}

/** 公开配置，无需登录 */
export function getPublicConfig() {
  return request.get<PublicConfig>(`${BASE}/public`)
}

export function createConfig(data: CreateConfigParams) {
  return request.post<ConfigItem>(BASE, data)
}

export function updateConfig(data: UpdateConfigParams) {
  return request.put<ConfigItem>(BASE, data)
}

export function deleteConfig(id: number) {
  return request.delete(`${BASE}/${id}`)
}
