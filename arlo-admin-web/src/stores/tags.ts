import { defineStore } from 'pinia'
import { computed, nextTick, ref } from 'vue'
import type { RouteLocationNormalizedLoaded, Router } from 'vue-router'

const TAGS_SESSION_KEY = 'arlo-visited-tags'

export interface TagItem {
  path: string
  fullPath: string
  title: string
  name: string
  affix?: boolean
  keepAlive?: boolean
}

/** 首页固定页签 */
export const HOME_TAG: TagItem = {
  path: '/dashboard',
  fullPath: '/dashboard',
  title: '首页',
  name: 'Dashboard',
  affix: true,
  keepAlive: true,
}

function readSessionTags(): TagItem[] {
  try {
    const raw = sessionStorage.getItem(TAGS_SESSION_KEY)
    if (!raw) return [HOME_TAG]
    const list = JSON.parse(raw) as TagItem[]
    if (!Array.isArray(list) || list.length === 0) return [HOME_TAG]
    const hasHome = list.some((t) => t.path === HOME_TAG.path)
    return hasHome ? list : [HOME_TAG, ...list]
  } catch {
    return [HOME_TAG]
  }
}

function writeSessionTags(list: TagItem[]) {
  sessionStorage.setItem(TAGS_SESSION_KEY, JSON.stringify(list))
}

export const useTagsStore = defineStore('tags', () => {
  const visited = ref<TagItem[]>(readSessionTags())
  /** 刷新时临时踢出 keep-alive，下一帧恢复 */
  const droppedCacheNames = ref<string[]>([])

  /** keep-alive include：当前仍打开且菜单允许缓存的路由 name */
  const cachedNames = computed(() =>
    visited.value
      .filter((t) => t.keepAlive && t.name && !droppedCacheNames.value.includes(t.name))
      .map((t) => t.name),
  )

  function persist() {
    writeSessionTags(visited.value)
  }

  function navigateToTag(path: string, router: Router) {
    const tag = visited.value.find((t) => t.path === path)
    router.push(tag?.fullPath || path)
  }

  function navigateIfCurrentRemoved(router: Router) {
    const cur = router.currentRoute.value.path
    if (!visited.value.some((t) => t.path === cur)) {
      const last = visited.value[visited.value.length - 1] || HOME_TAG
      router.push(last.fullPath || last.path)
    }
  }

  function addView(route: RouteLocationNormalizedLoaded) {
    const path = route.path
    if (!path || path === '/login' || path === '/404' || path === '/403') return
    // 只记录挂在 Layout 下的业务页
    const underLayout = route.matched.some((m) => m.name === 'Layout')
    if (!underLayout) return

    const name = typeof route.name === 'string' ? route.name : ''
    if (!name || name === 'Layout' || name === 'CatchAll') return

    const title = String(route.meta?.title || name)
    const keepAlive = route.meta?.keepAlive === true || route.meta?.affix === true
    const affix = route.meta?.affix === true || path === HOME_TAG.path

    const idx = visited.value.findIndex((t) => t.path === path)
    if (idx >= 0) {
      const prev = visited.value[idx]
      visited.value[idx] = {
        ...prev,
        fullPath: route.fullPath,
        title,
        name,
        keepAlive: affix ? true : keepAlive,
        affix: prev.affix || affix,
      }
    } else {
      visited.value.push({
        path,
        fullPath: route.fullPath,
        title,
        name,
        affix,
        keepAlive: affix ? true : keepAlive,
      })
    }
    persist()
  }

  /** 按当前路由表过滤失效页签（刷新后 / 权限变更后） */
  function pruneInvalid(router: Router) {
    visited.value = visited.value.filter((tag) => {
      if (tag.affix || tag.path === HOME_TAG.path) return true
      const resolved = router.resolve(tag.path)
      return resolved.matched.some(
        (m) => m.name && m.name !== 'Layout' && m.name !== 'CatchAll' && m.name !== 'NotFound',
      )
    })
    if (!visited.value.some((t) => t.path === HOME_TAG.path)) {
      visited.value.unshift({ ...HOME_TAG })
    }
    persist()
  }

  function closeTag(path: string, router: Router) {
    const idx = visited.value.findIndex((t) => t.path === path)
    if (idx < 0) return
    const tag = visited.value[idx]
    if (tag.affix) return

    const isActive = router.currentRoute.value.path === path
    visited.value.splice(idx, 1)
    persist()

    if (!isActive) return
    const next = visited.value[idx] || visited.value[idx - 1] || HOME_TAG
    router.push(next.fullPath || next.path)
  }

  function closeOthers(path: string, router?: Router) {
    visited.value = visited.value.filter((t) => t.affix || t.path === path)
    persist()
    if (router && router.currentRoute.value.path !== path) {
      navigateToTag(path, router)
    }
  }

  function closeLeft(path: string, router: Router) {
    const idx = visited.value.findIndex((t) => t.path === path)
    if (idx <= 0) return
    visited.value = visited.value.filter((t, i) => t.affix || i >= idx)
    persist()
    navigateIfCurrentRemoved(router)
  }

  function closeRight(path: string, router: Router) {
    const idx = visited.value.findIndex((t) => t.path === path)
    if (idx < 0 || idx >= visited.value.length - 1) return
    visited.value = visited.value.filter((t, i) => t.affix || i <= idx)
    persist()
    navigateIfCurrentRemoved(router)
  }

  function closeAll(router: Router) {
    visited.value = visited.value.filter((t) => t.affix)
    if (!visited.value.length) visited.value = [{ ...HOME_TAG }]
    persist()
    const cur = router.currentRoute.value.path
    if (!visited.value.some((t) => t.path === cur)) {
      router.push(HOME_TAG.path)
    }
  }

  function reset() {
    visited.value = [{ ...HOME_TAG }]
    droppedCacheNames.value = []
    sessionStorage.removeItem(TAGS_SESSION_KEY)
  }

  /** 刷新指定页签（踢出 keep-alive 后重新进入） */
  async function refreshTag(path: string, router: Router) {
    const tag = visited.value.find((t) => t.path === path)
    if (!tag?.name) return

    if (router.currentRoute.value.path !== path) {
      await router.push(tag.fullPath || tag.path)
    }

    if (!droppedCacheNames.value.includes(tag.name)) {
      droppedCacheNames.value.push(tag.name)
    }
    await nextTick()
    await router.replace(tag.fullPath || tag.path)
    await nextTick()
    droppedCacheNames.value = droppedCacheNames.value.filter((n) => n !== tag.name)
  }

  return {
    visited,
    cachedNames,
    addView,
    pruneInvalid,
    closeTag,
    closeOthers,
    closeLeft,
    closeRight,
    closeAll,
    refreshTag,
    reset,
  }
})
