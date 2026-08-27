/** 全局主题清单（外观全在 styles/themes/_*.scss，class 约定为 theme-{id}） */

export interface ThemeOption {
  id: string
  label: string
}

/**
 * 扩展主题：加 styles/themes/_xxx.scss，并在此登记。
 * 根节点 class 自动为 `theme-${id}`。
 */
export const themes: ThemeOption[] = [
  { id: 'light', label: '浅色' },
  { id: 'dark', label: '深色' },
]

export const DEFAULT_THEME_ID = 'light'

export function getTheme(id: string): ThemeOption {
  return themes.find((t) => t.id === id) ?? themes[0]
}

export function listThemes(): ThemeOption[] {
  return themes
}

export function themeClassName(id: string): string {
  return `theme-${getTheme(id).id}`
}
