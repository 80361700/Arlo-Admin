<template>
  <el-select
    v-model="selectedValue"
    :placeholder="placeholder"
    :clearable="clearable"
    :disabled="disabled || !ready"
    :filterable="filterable"
    :filter-method="filterMethod"
    style="width: 100%"
    @change="handleChange"
  >
    <el-option
      v-for="u in filteredOptions"
      :key="u.id"
      :label="optionLabel(u)"
      :value="u.id"
      :disabled="u.status !== 1"
    />
  </el-select>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { getAllUsers } from '@/api'
import type { UserOption } from '@/api'

const props = withDefaults(
  defineProps<{
    placeholder?: string
    clearable?: boolean
    disabled?: boolean
    filterable?: boolean
    /** 不可选的用户 id（如不能委托给自己） */
    excludeIds?: number[]
  }>(),
  {
    placeholder: '请选择负责人',
    clearable: true,
    filterable: true,
    excludeIds: () => [],
  },
)

const selectedValue = defineModel<number | undefined>()

const emit = defineEmits<{
  change: [value: number | undefined, name?: string]
}>()

const allOptions = ref<UserOption[]>([])
const query = ref('')
const ready = ref(false)

const availableOptions = computed(() => {
  const exclude = new Set((props.excludeIds || []).filter((id) => id > 0))
  if (!exclude.size) return allOptions.value
  return allOptions.value.filter((u) => !exclude.has(u.id))
})

const filteredOptions = computed(() => {
  const q = query.value.trim().toLowerCase()
  const base = availableOptions.value
  if (!q) return base
  return base.filter((u) => {
    return (
      (u.name || '').toLowerCase().includes(q) ||
      (u.username || '').toLowerCase().includes(q)
    )
  })
})

function optionLabel(u: UserOption) {
  if (u.name && u.name !== u.username) {
    return `${u.name}（${u.username}）`
  }
  return u.name || u.username
}

function filterMethod(q: string) {
  query.value = q
}

function handleChange(val: number | undefined) {
  const u = allOptions.value.find((x) => x.id === val)
  emit('change', val, u ? u.name || u.username || '' : undefined)
}

watch(
  () => props.excludeIds,
  () => {
    if (selectedValue.value && (props.excludeIds || []).includes(selectedValue.value)) {
      selectedValue.value = undefined
      emit('change', undefined, undefined)
    }
  },
)

onMounted(async () => {
  try {
    const res = await getAllUsers()
    allOptions.value = res.data || []
  } catch {
    // 静默失败
  } finally {
    ready.value = true
  }
})
</script>
