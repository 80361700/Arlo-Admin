import request from '../request'
import type { MenuTreeNode } from './system'

// ===================== 类型定义 =====================

export interface LoginParams {
  username: string
  password: string
  captchaId?: string
  captchaCode?: string
}

export interface LoginResult {
  accessToken: string
  refreshToken: string
}

export interface CaptchaResult {
  captchaId: string
  captcha: string // base64 图片
}

export interface UserInfo {
  id: number
  username: string
  nickname: string
  avatar: string
  email: string
  phone: string
  gender: number
  deptId: number
  deptName: string
  status: number
  remark: string
  roleNames: string[]
  postNames: string[]
  permissions: string[]
  dataScope: number
  deptIds: number[]
  mustChangePwd?: boolean
  pwdExpired?: boolean
}

export interface UpdateProfileParams {
  nickname: string
  gender: number
  phone: string
  email: string
  remark: string
  avatar?: string
}

export interface ChangePasswordParams {
  oldPassword: string
  newPassword: string
}

// ===================== 认证 API =====================

export function loginApi(params: LoginParams) {
  return request.post<LoginResult>('/v1/auth/login', params)
}

export function refreshTokenApi(refreshToken: string) {
  return request.post<LoginResult>('/v1/auth/refresh', { refreshToken })
}

export function logoutApi(refreshToken?: string) {
  return request.post('/v1/auth/logout', { refreshToken: refreshToken || '' })
}

export function getUserInfoApi() {
  return request.get<UserInfo>('/v1/auth/info')
}

export function getUserMenusApi() {
  return request.get<MenuTreeNode[]>('/v1/auth/menus')
}

export function updateProfileApi(data: UpdateProfileParams) {
  return request.put('/v1/auth/profile', data)
}

export function changePasswordApi(data: ChangePasswordParams) {
  return request.put('/v1/auth/password', data)
}

export function getCaptchaApi() {
  return request.get<CaptchaResult>('/v1/auth/captcha')
}
