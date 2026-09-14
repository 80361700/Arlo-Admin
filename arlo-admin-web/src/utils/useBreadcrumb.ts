import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { MenuTreeNode } from '@/api'

export type BreadcrumbItem = {
  title: string
  /** 可跳转路径；目录或当前页可不设 */
  path?: string
}

function normalizePath(path: string): string {
  if (!path) return ''
  const p = path.startsWith('/') ? path : `/${path}`
  return p.replace(/\/+/g, '/').replace(/\/$/, '') || '/'
}

function resolveMenuFullPath(menu: MenuTreeNode, parentPath: string): string {
  if (!menu.path) return ''
  if (menu.path.startsWith('/')) return normalizePath(menu.path)
  const base = parentPath ? parentPath : ''
  return normalizePath(`${base}/${menu.path}`)
}

/** 在菜单树中查找当前路由的面包屑链（不含「首页」） */
function findCrumbs(
  menus: MenuTreeNode[],
  targetPath: string,
  parentPath = '',
  parents: BreadcrumbItem[] = [],
): BreadcrumbItem[] | null {
  const target = normalizePath(targetPath)

  for (const menu of menus) {
    if (menu.visible === 0) continue

    if (menu.type === 1) {
      // 目录：参与展示，一般不可点
      const dirPath = menu.path ? resolveMenuFullPath(menu, parentPath) : parentPath
      const nextParents = [...parents, { title: menu.name }]
      const found = findCrumbs(menu.children || [], target, dirPath || parentPath, nextParents)
      if (found) return found
      continue
    }

    if (menu.type === 2) {
      const fullPath = resolveMenuFullPath(menu, parentPath)
      if (fullPath && fullPath === target) {
        return [...parents, { title: menu.name, path: fullPath }]
      }
      if (menu.children?.length) {
        const found = findCrumbs(menu.children, target, fullPath || parentPath, [
          ...parents,
          { title: menu.name, path: fullPath || undefined },
        ])
        if (found) return found
      }
    }
  }
  return null
}

export function useBreadcrumb() {
  const route = useRoute()
  const authStore = useAuthStore()

  const items = computed<BreadcrumbItem[]>(() => {
    const path = normalizePath(route.path)
    const home: BreadcrumbItem = { title: '首页', path: '/dashboard' }

    if (path === '/dashboard' || path === '/') {
      return [{ title: '首页' }]
    }

    const matched = findCrumbs(authStore.menus || [], path)
    if (matched?.length) {
      return [home, ...matched]
    }

    // 无菜单匹配时用路由 meta.title 兜底
    const title = (route.meta?.title as string) || path
    return [home, { title }]
  })

  return { items }
}
