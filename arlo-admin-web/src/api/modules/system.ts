import request from '../request'

// ===================== 通用 =====================
export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

// ===================== 用户管理 =====================
export interface UserItem {
  id: number
  username: string
  nickname: string
  avatar: string
  email: string
  phone: string
  gender: number // 0=未知 1=男 2=女
  deptId: number
  deptName: string
  status: number // 0=禁用 1=启用
  remark: string
  roleIds: number[]
  roleNames: string[]
  postIds: number[]
  postNames: string[]
  createdAt: string
}

export interface UserListParams {
  page: number
  pageSize: number
  username?: string
  nickname?: string
  phone?: string
  status?: number
  deptId?: number
}

export interface CreateUserParams {
  username: string
  password: string
  nickname: string
  avatar?: string
  email?: string
  phone?: string
  gender?: number
  deptId?: number
  status?: number
  remark?: string
  roleIds?: number[]
  postIds?: number[]
}

export interface UpdateUserParams {
  id: number
  nickname: string
  avatar?: string
  email?: string
  phone?: string
  gender?: number
  deptId?: number
  status?: number
  remark?: string
  roleIds?: number[]
  postIds?: number[]
}

export interface UpdateUserPasswordParams {
  id: number
  password: string
}

export function getUserList(params: UserListParams) {
  return request.get<PageResult<UserItem>>('/v1/system/user/list', params)
}

export function getUserDetail(id: number) {
  return request.get<UserItem>(`/v1/system/user/${id}`)
}

export function createUser(params: CreateUserParams) {
  return request.post('/v1/system/user', params)
}

export function updateUser(params: UpdateUserParams) {
  return request.put('/v1/system/user', params)
}

export function deleteUser(id: number) {
  return request.delete(`/v1/system/user/${id}`)
}

export function updateUserPassword(params: UpdateUserPasswordParams) {
  return request.put('/v1/system/user/password', params)
}

export function unlockUser(id: number) {
  return request.put(`/v1/system/user/${id}/unlock`)
}

export function exportUsers(params?: Partial<UserListParams>) {
  return import('@/utils/download').then(({ downloadFile }) =>
    downloadFile('/v1/system/user/export', params, 'users.xlsx'),
  )
}

export function downloadUserImportTemplate() {
  return import('@/utils/download').then(({ downloadFile }) =>
    downloadFile('/v1/system/user/import/template', undefined, 'user_import_template.xlsx'),
  )
}

export function importUsers(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post<{ success: number; errors: string[] }>('/v1/system/user/import', formData, {
    headers: { 'Content-Type': undefined },
  })
}

// ===================== 角色管理 =====================
export interface RoleItem {
  id: number
  name: string
  code: string
  sort: number
  status: number
  remark: string
  dataScope: number // 1=全部 2=自定义 3=本部门及以下 4=本部门 5=仅本人
  deptIds: number[]
  createdAt: string
}

export interface RoleListParams {
  page: number
  pageSize: number
  name?: string
  code?: string
  status?: number
}

export interface CreateRoleParams {
  name: string
  code: string
  sort?: number
  status?: number
  remark?: string
  dataScope?: number
  deptIds?: number[]
}

export interface UpdateRoleParams {
  id: number
  name: string
  code: string
  sort?: number
  status?: number
  remark?: string
  dataScope?: number
  deptIds?: number[]
}

export interface AssignRoleMenusParams {
  roleId: number
  menuIds: number[]
}

export function getRoleList(params: RoleListParams) {
  return request.get<PageResult<RoleItem>>('/v1/system/role/list', params)
}

export function getAllRoles() {
  return request.get<RoleItem[]>('/v1/system/role/all')
}

export function getRoleDetail(id: number) {
  return request.get<RoleItem>(`/v1/system/role/${id}`)
}

export function createRole(params: CreateRoleParams) {
  return request.post('/v1/system/role', params)
}

export function updateRole(params: UpdateRoleParams) {
  return request.put('/v1/system/role', params)
}

export function deleteRole(id: number) {
  return request.delete(`/v1/system/role/${id}`)
}

export function getRoleMenus(id: number) {
  return request.get<number[]>(`/v1/system/role/${id}/menus`)
}

export function assignRoleMenus(params: AssignRoleMenusParams) {
  return request.post('/v1/system/role/assignMenus', params)
}

// ===================== 部门管理 =====================
export interface DeptTreeNode {
  id: number
  parentId: number
  name: string
  sort: number
  leader: string
  phone: string
  email: string
  status: number
  children: DeptTreeNode[]
}

export interface CreateDeptParams {
  parentId?: number
  name: string
  sort?: number
  leader?: string
  phone?: string
  email?: string
  status?: number
}

export interface UpdateDeptParams {
  id: number
  parentId?: number
  name: string
  sort?: number
  leader?: string
  phone?: string
  email?: string
  status?: number
}

export function getDeptTree() {
  return request.get<DeptTreeNode[]>('/v1/system/dept/tree')
}

export function createDept(params: CreateDeptParams) {
  return request.post('/v1/system/dept', params)
}

export function updateDept(params: UpdateDeptParams) {
  return request.put('/v1/system/dept', params)
}

export function deleteDept(id: number) {
  return request.delete(`/v1/system/dept/${id}`)
}

// ===================== 菜单管理 =====================
export interface MenuTreeNode {
  id: number
  parentId: number
  name: string
  type: number // 1=目录 2=菜单 3=按钮
  path: string
  component: string
  icon: string
  sort: number
  permission: string
  status: number
  visible: number
  keepAlive: number
  children: MenuTreeNode[]
}

export interface CreateMenuParams {
  parentId?: number
  name: string
  type: number
  path?: string
  component?: string
  icon?: string
  sort?: number
  permission?: string
  status?: number
  visible?: number
  keepAlive?: number
}

export interface UpdateMenuParams {
  id: number
  parentId?: number
  name: string
  type: number
  path?: string
  component?: string
  icon?: string
  sort?: number
  permission?: string
  status?: number
  visible?: number
  keepAlive?: number
}

export function getMenuTree() {
  return request.get<MenuTreeNode[]>('/v1/system/menu/tree')
}

export function createMenu(params: CreateMenuParams) {
  return request.post('/v1/system/menu', params)
}

export function updateMenu(params: UpdateMenuParams) {
  return request.put('/v1/system/menu', params)
}

export function deleteMenu(id: number) {
  return request.delete(`/v1/system/menu/${id}`)
}

// ===================== 岗位管理 =====================
export interface PostItem {
  id: number
  code: string
  name: string
  sort: number
  status: number
  remark: string
  createdAt: string
}

export interface PostListParams {
  page: number
  pageSize: number
  code?: string
  name?: string
  status?: number
}

export interface CreatePostParams {
  code: string
  name: string
  sort?: number
  status?: number
  remark?: string
}

export interface UpdatePostParams {
  id: number
  code: string
  name: string
  sort?: number
  status?: number
  remark?: string
}

export function getPostList(params: PostListParams) {
  return request.get<PageResult<PostItem>>('/v1/system/post/list', params)
}

export function getAllPosts() {
  return request.get<PostItem[]>('/v1/system/post/all')
}

export function getPostDetail(id: number) {
  return request.get<PostItem>(`/v1/system/post/${id}`)
}

export function createPost(params: CreatePostParams) {
  return request.post('/v1/system/post', params)
}

export function updatePost(params: UpdatePostParams) {
  return request.put('/v1/system/post', params)
}

export function deletePost(id: number) {
  return request.delete(`/v1/system/post/${id}`)
}

// ===================== 字典管理 =====================
export interface DictTypeItem {
  id: number
  name: string
  code: string
  status: number
  remark: string
  createdAt: string
}

export interface DictTypeListParams {
  page: number
  pageSize: number
  name?: string
  code?: string
  status?: number
}

export interface CreateDictTypeParams {
  name: string
  code: string
  status?: number
  remark?: string
}

export interface UpdateDictTypeParams {
  id: number
  name: string
  code: string
  status?: number
  remark?: string
}

export interface DictDataItem {
  id: number
  dictTypeId: number
  label: string
  value: string
  sort: number
  isDefault: number
  status: number
  remark: string
  createdAt: string
}

export interface DictDataListParams {
  page: number
  pageSize: number
  dictTypeId?: number
  label?: string
  status?: number
}

export interface CreateDictDataParams {
  dictTypeId: number
  label: string
  value: string
  sort?: number
  isDefault?: number
  status?: number
  remark?: string
}

export interface UpdateDictDataParams {
  id: number
  dictTypeId: number
  label: string
  value: string
  sort?: number
  isDefault?: number
  status?: number
  remark?: string
}

// 字典类型
export function getDictTypeList(params: DictTypeListParams) {
  return request.get<PageResult<DictTypeItem>>('/v1/system/dict/type/list', params)
}

export function getDictTypeDetail(id: number) {
  return request.get<DictTypeItem>(`/v1/system/dict/type/${id}`)
}

export function createDictType(params: CreateDictTypeParams) {
  return request.post('/v1/system/dict/type', params)
}

export function updateDictType(params: UpdateDictTypeParams) {
  return request.put('/v1/system/dict/type', params)
}

export function deleteDictType(id: number) {
  return request.delete(`/v1/system/dict/type/${id}`)
}

// 字典数据
export function getDictDataList(params: DictDataListParams) {
  return request.get<PageResult<DictDataItem>>('/v1/system/dict/data/list', params)
}

export function getDictDataDetail(id: number) {
  return request.get<DictDataItem>(`/v1/system/dict/data/${id}`)
}

export function createDictData(params: CreateDictDataParams) {
  return request.post('/v1/system/dict/data', params)
}

export function updateDictData(params: UpdateDictDataParams) {
  return request.put('/v1/system/dict/data', params)
}

export function deleteDictData(id: number) {
  return request.delete(`/v1/system/dict/data/${id}`)
}

/** 按字典编码获取启用中的字典项（业务下拉，仅需登录） */
export function getDictByCode(code: string) {
  return request.get<DictDataItem[]>(`/v1/system/dict/data/code/${encodeURIComponent(code)}`)
}
