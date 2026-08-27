import type { App, Directive, DirectiveBinding } from 'vue'
import { useAuthStore } from '@/stores/auth'

function checkPermission(el: HTMLElement, binding: DirectiveBinding<string | string[]>) {
  const { value } = binding
  if (!value) return

  const authStore = useAuthStore()
  const needed = Array.isArray(value) ? value : [value]
  const ok = needed.some((p) => authStore.hasPermission(p))

  if (!ok) {
    el.style.display = 'none'
    el.setAttribute('data-permission-hidden', '1')
  } else if (el.getAttribute('data-permission-hidden')) {
    el.style.display = ''
    el.removeAttribute('data-permission-hidden')
  }
}

/**
 * 按钮权限指令
 * 用法：v-permission="'sys:user:add'" 或 v-permission="['sys:user:add', 'sys:user:edit']"（满足其一即可）
 */
const permissionDirective: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    checkPermission(el, binding)
  },
  updated(el, binding) {
    checkPermission(el, binding)
  },
}

export function setupPermissionDirective(app: App) {
  app.directive('permission', permissionDirective)
}

export default permissionDirective
