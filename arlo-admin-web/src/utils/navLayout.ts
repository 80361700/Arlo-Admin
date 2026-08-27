import type { MenuTreeNode } from '@/api'

/** 可见的非按钮子节点 */
export function visibleNavChildren(item: MenuTreeNode): MenuTreeNode[] {
  return (item.children || []).filter((c) => c.type !== 3 && c.visible !== 0)
}

/** 顶栏一级：根上 type=1 的目录，或根上直接挂的 type=2 页面 */
export function getTopLevelMenus(menus: MenuTreeNode[]): MenuTreeNode[] {
  return (menus || []).filter((m) => m.visible !== 0 && (m.type === 1 || m.type === 2))
}

function normalizePath(p: string): string {
  if (!p) return ''
  const s = p.startsWith('/') ? p : `/${p}`
  return s.replace(/\/+/g, '/').replace(/\/$/, '') || '/'
}

function pathMatches(menuPath: string, routePath: string): boolean {
  const a = normalizePath(menuPath)
  const b = normalizePath(routePath)
  if (!a || a === '/') return false
  return b === a || b.startsWith(`${a}/`)
}

/** 节点或其子孙是否匹配当前路由 */
export function menuContainsPath(node: MenuTreeNode, routePath: string): boolean {
  if (node.type === 2 && pathMatches(node.path, routePath)) return true
  return visibleNavChildren(node).some((c) => menuContainsPath(c, routePath))
}

/** 根据路由反推当前一级模块 */
export function findActiveTopMenu(menus: MenuTreeNode[], routePath: string): MenuTreeNode | null {
  const tops = getTopLevelMenus(menus)
  for (const m of tops) {
    if (menuContainsPath(m, routePath)) return m
  }
  return null
}

/** 混合模式默认一级：优先路由匹配，否则第一个带二级菜单的目录，再否则第一个 */
export function resolveTopMenu(menus: MenuTreeNode[], routePath: string): MenuTreeNode | null {
  const hit = findActiveTopMenu(menus, routePath)
  if (hit) return hit
  const tops = getTopLevelMenus(menus)
  if (!tops.length) return null
  return tops.find((m) => m.type === 1 && visibleNavChildren(m).length > 0) || tops[0]
}

/** 取模块下第一个可跳转的页面 path */
export function firstLeafPath(node: MenuTreeNode): string | null {
  if (node.type === 2 && node.path) return normalizePath(node.path)
  for (const c of visibleNavChildren(node)) {
    const p = firstLeafPath(c)
    if (p) return p
  }
  return null
}

/** 混合模式侧栏菜单：当前一级的子树（二级及以下） */
export function getSideMenusForTop(top: MenuTreeNode | null): MenuTreeNode[] {
  if (!top) return []
  if (top.type === 2) return [] // 一级本身就是页面，无需侧栏
  return visibleNavChildren(top)
}
