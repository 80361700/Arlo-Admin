import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getPublicConfig } from '@/api/modules/sysconfig'
import { DEFAULT_THEME_ID, getTheme, listThemes, themeClassName } from '@/themes'

const THEME_KEY = 'arlo-theme'
const LAYOUT_KEY = 'arlo-layout-mode'
const TAGS_VIEW_KEY = 'arlo-tags-view'

/** side=侧栏 / mix=混合(顶栏一级+左侧二级) / topbar=纯顶栏 */
export type LayoutMode = 'side' | 'mix' | 'topbar'

export const layoutModeOptions: { id: LayoutMode; label: string; desc: string }[] = [
  { id: 'side', label: '侧栏模式', desc: '左侧展示完整菜单树' },
  { id: 'mix', label: '混合模式', desc: '顶栏一级，左侧二级；首页默认第一模块' },
  { id: 'topbar', label: '顶栏模式', desc: '菜单全部在顶部，无左侧栏' },
]

function readStoredThemeId(): string {
  const v = localStorage.getItem(THEME_KEY) || localStorage.getItem('arlo-sidebar-theme')
  if (v && listThemes().some((t) => t.id === v)) return v
  return DEFAULT_THEME_ID
}

function readStoredLayoutMode(): LayoutMode {
  const v = localStorage.getItem(LAYOUT_KEY)
  // 兼容旧键 top → mix
  if (v === 'top') {
    localStorage.setItem(LAYOUT_KEY, 'mix')
    return 'mix'
  }
  if (v === 'side' || v === 'mix' || v === 'topbar') return v
  return 'side'
}

function readStoredTagsView(): boolean {
  const v = localStorage.getItem(TAGS_VIEW_KEY)
  if (v === '0' || v === 'false') return false
  if (v === '1' || v === 'true') return true
  return true // 默认开启
}

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const themeId = ref(readStoredThemeId())
  const layoutMode = ref<LayoutMode>(readStoredLayoutMode())
  /** 顶栏下方是否展示页签 */
  const tagsView = ref(readStoredTagsView())
  /** 混合模式下当前选中的一级菜单 id */
  const activeTopMenuId = ref<number | null>(null)

  const systemName = ref('Arlo Admin')
  const captchaEnabled = ref(true)
  const systemLogo = ref('')
  const systemVersion = ref('1.0.0')
  const configLoaded = ref(false)

  const sidebarWidth = computed(() => (sidebarCollapsed.value ? '64px' : '220px'))
  const theme = computed(() => getTheme(themeId.value))
  const themeClass = computed(() => themeClassName(themeId.value))
  const themeOptions = computed(() => listThemes())

  const isSideLayout = computed(() => layoutMode.value === 'side')
  const isMixLayout = computed(() => layoutMode.value === 'mix')
  const isTopbarLayout = computed(() => layoutMode.value === 'topbar')
  /** 顶栏放 Logo（混合 / 纯顶栏） */
  const useHeaderBrand = computed(() => isMixLayout.value || isTopbarLayout.value)

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setTheme(id: string) {
    const next = getTheme(id)
    themeId.value = next.id
    localStorage.setItem(THEME_KEY, next.id)
  }

  function setLayoutMode(mode: LayoutMode) {
    layoutMode.value = mode
    localStorage.setItem(LAYOUT_KEY, mode)
  }

  function setTagsView(on: boolean) {
    tagsView.value = on
    localStorage.setItem(TAGS_VIEW_KEY, on ? '1' : '0')
  }

  function setActiveTopMenuId(id: number | null) {
    activeTopMenuId.value = id
  }

  function cycleTheme() {
    const list = listThemes()
    const idx = list.findIndex((t) => t.id === themeId.value)
    const next = list[(idx + 1) % list.length]
    setTheme(next.id)
  }

  async function loadPublicConfig(force = false) {
    if (configLoaded.value && !force) return
    try {
      const res = await getPublicConfig()
      const data = res.data
      if (data?.name) systemName.value = data.name
      captchaEnabled.value = !!data?.captcha
      systemLogo.value = data?.logo || ''
      if (data?.version) systemVersion.value = data.version
      configLoaded.value = true
      document.title = systemName.value
    } catch {
      // 公开配置失败时保留默认值
    }
  }

  return {
    sidebarCollapsed,
    themeId,
    layoutMode,
    tagsView,
    activeTopMenuId,
    theme,
    themeClass,
    themeOptions,
    isSideLayout,
    isMixLayout,
    isTopbarLayout,
    useHeaderBrand,
    sidebarWidth,
    systemName,
    captchaEnabled,
    systemLogo,
    systemVersion,
    configLoaded,
    toggleSidebar,
    setTheme,
    setLayoutMode,
    setTagsView,
    setActiveTopMenuId,
    cycleTheme,
    loadPublicConfig,
  }
})
