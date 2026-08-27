import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { loginApi, refreshTokenApi, logoutApi, getUserInfoApi, getUserMenusApi } from '@/api'
import type { UserInfo, MenuTreeNode } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  // ================== 状态 ==================
  const accessToken = ref<string>('')
  const refreshToken = ref<string>('')
  const userInfo = ref<UserInfo | null>(null)
  const menus = ref<MenuTreeNode[]>([])

  // ================== 计算属性 ==================
  const isLoggedIn = computed(() => !!accessToken.value)
  const roles = computed(() => userInfo.value?.roleNames || [])
  const permissions = computed(() => userInfo.value?.permissions || [])

  // ================== 方法 ==================
  /** 登录 */
  async function login(params: { username: string; password: string; captchaId: string; captchaCode: string }) {
    const res = await loginApi(params)
    accessToken.value = res.data.accessToken
    refreshToken.value = res.data.refreshToken
    // 持久化 token
    localStorage.setItem('accessToken', res.data.accessToken)
    localStorage.setItem('refreshToken', res.data.refreshToken)
  }

  /** 刷新 token（后端只签发新 access；refresh 沿用本地已有值） */
  async function refresh() {
    const stored = localStorage.getItem('refreshToken')
    if (!stored) {
      throw new Error('no refresh token')
    }
    const res = await refreshTokenApi(stored)
    accessToken.value = res.data.accessToken
    localStorage.setItem('accessToken', res.data.accessToken)
    // 若响应带回新的 refreshToken 则轮换，否则保留原 refresh
    if (res.data.refreshToken) {
      refreshToken.value = res.data.refreshToken
      localStorage.setItem('refreshToken', res.data.refreshToken)
    }
    return res.data.accessToken
  }

  /** 获取用户信息 */
  async function fetchUserInfo() {
    const res = await getUserInfoApi()
    userInfo.value = res.data
  }

  /** 获取当前角色菜单（侧边栏） */
  async function fetchMenus() {
    const res = await getUserMenusApi()
    menus.value = res.data || []
    return menus.value
  }

  /**
   * 会话内热刷新：重新拉权限 + 菜单 + 动态路由
   * 分配角色菜单后调用，无需重新登录
   */
  async function refreshPermissions() {
    const routerMod = await import('@/router')
    await fetchUserInfo()
    const menuTree = await routerMod.reloadDynamicRoutes()
    menus.value = menuTree
    const { useTagsStore } = await import('@/stores/tags')
    useTagsStore().pruneInvalid(routerMod.default)
  }

  /** 从 localStorage 恢复 session */
  function restoreSession() {
    const at = localStorage.getItem('accessToken')
    const rt = localStorage.getItem('refreshToken')
    if (at && rt) {
      accessToken.value = at
      refreshToken.value = rt
    }
  }

  /** 清空本地登录态（不调接口） */
  function clearSession() {
    accessToken.value = ''
    refreshToken.value = ''
    userInfo.value = null
    menus.value = []
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
    import('@/utils/authFileUrl').then(({ clearAuthFileCache }) => clearAuthFileCache())
    import('@/stores/tags').then(({ useTagsStore }) => useTagsStore().reset())
  }

  /**
   * 登出：先通知后端把 token 写入 Redis 黑名单，再清本地
   * 接口失败仍清本地，避免卡在已坏会话
   */
  async function logout() {
    const rt = refreshToken.value || localStorage.getItem('refreshToken') || ''
    try {
      if (accessToken.value || localStorage.getItem('accessToken')) {
        await logoutApi(rt)
      }
    } catch {
      /* 仍继续清本地 */
    }
    clearSession()
  }

  /** 检查是否有某权限（严格按 auth/info 返回的 permissions 判断） */
  function hasPermission(perm: string): boolean {
    return permissions.value.includes(perm)
  }

  return {
    accessToken,
    refreshToken,
    userInfo,
    menus,
    isLoggedIn,
    roles,
    permissions,
    login,
    refresh,
    fetchUserInfo,
    fetchMenus,
    refreshPermissions,
    restoreSession,
    clearSession,
    logout,
    hasPermission,
  }
})
