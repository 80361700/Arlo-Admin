import { ref, watch, type Ref } from 'vue'
import { acquireAuthFileUrl } from '@/utils/authFileUrl'

/** 将文件引用转为可展示的 blob/外链地址（带登录态拉取） */
export function useAuthFileSrc(
  source: Ref<string | null | undefined> | (() => string | null | undefined),
): Ref<string> {
  const url = ref('')
  const getter = typeof source === 'function' ? source : () => source.value

  watch(
    getter,
    async (v) => {
      const raw = String(v ?? '').trim()
      if (!raw) {
        url.value = ''
        return
      }
      try {
        url.value = await acquireAuthFileUrl(raw)
      } catch {
        url.value = ''
      }
    },
    { immediate: true },
  )

  return url
}
