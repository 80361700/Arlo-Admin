/** 确保左侧组件库「预设」分组默认展开 */
export function ensurePresetGroupExpanded(): void {
  const key = 'ep-component-view-keys'
  try {
    const raw = localStorage.getItem(key)
    if (!raw) return // 空 = expand all
    const keys = JSON.parse(raw)
    if (!Array.isArray(keys) || keys.length === 0) return
    if (!keys.includes('预设')) {
      keys.push('预设')
      localStorage.setItem(key, JSON.stringify(keys))
    }
  } catch {
    // ignore
  }
}
