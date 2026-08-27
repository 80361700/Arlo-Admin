import { ref, computed, type Ref, type ComputedRef } from 'vue'
import { getDictByCode, type DictDataItem } from '@/api/modules/system'

/** 常用字典编码（与 sys_dict_type.code 一致） */
export const DictCode = {
  UserStatus: 'sys_user_status',
  Gender: 'sys_gender',
  NoticeType: 'sys_notice_type',
  NoticeLevel: 'sys_notice_level',
  NoticeStatus: 'sys_notice_status',
  MessageType: 'sys_message_type',
  MemberSource: 'sys_member_source',
  DataScope: 'sys_data_scope',
} as const

export type DictOption = {
  label: string
  value: string | number
}

type CacheEntry = {
  items: DictDataItem[]
  promise?: Promise<DictDataItem[]>
}

const cache = new Map<string, CacheEntry>()

function parseValue(raw: string): string | number {
  if (/^-?\d+$/.test(raw)) return Number(raw)
  return raw
}

async function fetchDict(code: string, force = false): Promise<DictDataItem[]> {
  if (!force) {
    const hit = cache.get(code)
    if (hit?.items.length) return hit.items
    if (hit?.promise) return hit.promise
  }

  const promise = getDictByCode(code)
    .then((res) => {
      const items = res.data || []
      cache.set(code, { items })
      return items
    })
    .catch((err) => {
      cache.delete(code)
      throw err
    })

  cache.set(code, { items: cache.get(code)?.items || [], promise })
  return promise
}

export type UseDictResult = {
  options: Ref<DictOption[]>
  labelMap: ComputedRef<Record<string | number, string>>
  loading: Ref<boolean>
  getLabel: (value: string | number | null | undefined, fallback?: string) => string
  reload: () => Promise<void>
}

/** 加载字典并转为下拉/标签映射；带内存缓存，同 code 不重复请求 */
export function useDict(code: string): UseDictResult {
  const items = ref<DictDataItem[]>(cache.get(code)?.items || [])
  const loading = ref(false)

  const options = computed<DictOption[]>(() =>
    items.value.map((d) => ({
      label: d.label,
      value: parseValue(d.value),
    })),
  )

  const labelMap = computed(() => {
    const map: Record<string | number, string> = {}
    for (const d of items.value) {
      const v = parseValue(d.value)
      map[v] = d.label
      map[String(v)] = d.label
    }
    return map
  })

  function getLabel(value: string | number | null | undefined, fallback = '-') {
    if (value === null || value === undefined || value === '') return fallback
    return labelMap.value[value] ?? labelMap.value[String(value)] ?? fallback
  }

  async function reload() {
    loading.value = true
    try {
      items.value = await fetchDict(code, true)
    } finally {
      loading.value = false
    }
  }

  async function load() {
    if (items.value.length) return
    loading.value = true
    try {
      items.value = await fetchDict(code)
    } finally {
      loading.value = false
    }
  }

  void load()

  return { options, labelMap, loading, getLabel, reload }
}

/** 清除字典缓存（字典管理变更后可调用） */
export function clearDictCache(code?: string) {
  if (code) cache.delete(code)
  else cache.clear()
}
