import request from '../request'
import type { PageData } from '../request'

export interface MemberItem {
  id: number
  phone: string
  nickname: string
  avatar: string
  gender: number
  source: string
  status: number
  lastLogin: string
  createdAt: string
}

export interface MemberDetail extends MemberItem {
  openid: string
  unionid: string
  mpOpenid: string
  updatedAt: string
}

export interface MemberListParams {
  page?: number
  pageSize?: number
  phone?: string
  nickname?: string
  source?: string
  status?: number
}

export interface CreateMemberParams {
  phone: string
  password?: string
  nickname: string
  avatar: string
  gender: number
  source: string
  status: number
}

export interface UpdateMemberParams {
  id: number
  nickname: string
  avatar: string
  gender: number
  source: string
  status: number
}

export interface UpdateMemberPasswordParams {
  id: number
  password: string
}

export function getMemberList(params: MemberListParams) {
  return request.get<PageData<MemberItem>>('/v1/system/member/list', params)
}

export function getMemberDetail(id: number) {
  return request.get<MemberDetail>(`/v1/system/member/${id}`)
}

export function createMember(data: CreateMemberParams) {
  return request.post('/v1/system/member', data)
}

export function updateMember(data: UpdateMemberParams) {
  return request.put('/v1/system/member', data)
}

export function updateMemberPassword(data: UpdateMemberPasswordParams) {
  return request.put('/v1/system/member/password', data)
}

export function updateMemberStatus(id: number, status: number) {
  return request.put(`/v1/system/member/${id}/status`, { status })
}

export function deleteMember(id: number) {
  return request.delete(`/v1/system/member/${id}`)
}
