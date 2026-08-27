import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useTagsStore } from '@/stores/tags'
import { getUserMenusApi } from '@/api'
import type { MenuTreeNode } from '@/api'

// 自动扫描 views 目录下所有 .vue 文件，菜单 component 字段直接对应文件路径
const views = import.meta.glob('/src/views/**/*.vue')

/** 给异步组件打上 name，供 keep-alive include 匹配 */
function namedView(loader: () => Promise<unknown>, name: string) {
  return () =>
    (loader() as Promise<{ default: Record<string, unknown> }>).then((mod) => {
      const comp = mod.default
      if (comp && typeof comp === 'object') {
        comp.name = name
        comp.__name = name
      }
      return mod
    })
}

// ================== 静态路由（无需从菜单加载） ==================
const staticRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录', noAuth: true },
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/layout/index.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: namedView(() => import('@/views/dashboard/index.vue'), 'Dashboard'),
        meta: { title: '首页', icon: 'HomeFilled', affix: true, keepAlive: true },
      },
    ],
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: { title: '无权限', noAuth: true },
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { title: '404', noAuth: true },
  },
]

const CATCH_ALL_NAME = 'CatchAll'

// 兜底路由（动态路由加载完成后再注册，避免刷新时抢先匹配）
const catchAllRoute: RouteRecordRaw = {
  path: '/:pathMatch(.*)*',
  name: CATCH_ALL_NAME,
  redirect: '/404',
}

const router = createRouter({
  history: createWebHashHistory(),
  routes: [...staticRoutes],
})

// ================== 从菜单树生成 Vue Router 路由 ==================
function menuTreeToRoutes(menus: MenuTreeNode[], parentPath: string = ''): RouteRecordRaw[] {
  const routes: RouteRecordRaw[] = []

  for (const menu of menus) {
    if (menu.visible === 0) continue

    if (menu.type === 1) {
      const childRoutes = menuTreeToRoutes(menu.children || [], menu.path || '')
      for (const child of childRoutes) {
        routes.push(child)
      }
    } else if (menu.type === 2 && menu.component) {
      const key = `/src/views/${menu.component.replace(/^\//, '').replace(/\.vue$/, '')}.vue`
      const importFn = views[key]
      if (!importFn) {
        console.warn(`[router] 组件未找到: "${menu.component}" → ${key}`)
        continue
      }

      const routePath = menu.path
        ? (menu.path.startsWith('/') ? menu.path : `${parentPath ? parentPath + '/' : '/'}${menu.path}`)
        : '/'

      const routeName = `MenuView_${menu.id}`

      routes.push({
        path: routePath,
        name: routeName,
        component: namedView(importFn as () => Promise<unknown>, routeName),
        meta: {
          title: menu.name,
          icon: menu.icon,
          menuId: menu.id,
          keepAlive: menu.keepAlive === 1,
        },
      })
    }
  }

  return routes
}

// ================== 动态路由状态 ==================
const WHITE_LIST = ['/login', '/404', '/403']

let routesLoaded = false
const dynamicRouteNames: string[] = []

/** 登出时重置动态路由标记 */
export function resetDynamicRoutes() {
  clearDynamicRoutes()
  routesLoaded = false
}

function clearDynamicRoutes() {
  for (const name of dynamicRouteNames.splice(0)) {
    if (router.hasRoute(name)) {
      router.removeRoute(name)
    }
  }
  if (router.hasRoute(CATCH_ALL_NAME)) {
    router.removeRoute(CATCH_ALL_NAME)
  }
}

/** 按当前角色菜单重新注册动态路由（会话内权限热刷新） */
export async function reloadDynamicRoutes(): Promise<MenuTreeNode[]> {
  clearDynamicRoutes()

  const res = await getUserMenusApi()
  const menuTree = res.data || []
  const dynamicRoutes = menuTreeToRoutes(menuTree)

  for (const route of dynamicRoutes) {
    router.addRoute('Layout', route)
    if (route.name) {
      dynamicRouteNames.push(String(route.name))
    }
  }
  router.addRoute(catchAllRoute)
  routesLoaded = true
  return menuTree
}

router.beforeEach(async (to, _from, next) => {
  const appStore = useAppStore()
  if (!appStore.configLoaded) {
    await appStore.loadPublicConfig()
  }
  const brand = appStore.systemName || 'Arlo Admin'
  document.title = `${to.meta.title || brand} - ${brand}`

  const authStore = useAuthStore()

  if (WHITE_LIST.includes(to.path) || to.meta.noAuth) {
    next()
    return
  }

  if (!authStore.accessToken) {
    authStore.restoreSession()
  }

  if (!authStore.accessToken) {
    next(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
    return
  }

  if (!authStore.userInfo) {
    try {
      await authStore.fetchUserInfo()
    } catch {
      authStore.clearSession()
      next(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
      return
    }
  }

  // 按当前用户角色菜单加载动态路由（不是全量 menu/tree）
  if (!routesLoaded) {
    try {
      const menuTree = await reloadDynamicRoutes()
      authStore.menus = menuTree
      useTagsStore().pruneInvalid(router)
      next({ ...to, replace: true })
      return
    } catch {
      routesLoaded = true
      next('/403')
      return
    }
  }

  next()
})

router.afterEach((to) => {
  useTagsStore().addView(to)
})

export default router
