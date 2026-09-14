<template>
  <el-select
    v-model="model"
    :placeholder="placeholder"
    :clearable="clearable"
    :disabled="disabled || loading"
    :multiple="multiple"
    filterable
    style="width: 100%"
    @change="(v: any) => emit('change', v)"
  >
    <el-option
      v-for="opt in options"
      :key="String(opt.value)"
      :label="opt.label"
      :value="opt.value"
    />
  </el-select>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { getDictByCode } from '@/api/modules/system'
import type { DictOption } from '@/utils/useDict'

const props = withDefaults(
  defineProps<{
    dictCode?: string
    placeholder?: string
    clearable?: boolean
    disabled?: boolean
    multiple?: boolean
  }>(),
  {
    dictCode: '',
    placeholder: '请选择',
    clearable: true,
    multiple: false,
  },
)

const model = defineModel<string | number | Array<string | number> | undefined>()
const emit = defineEmits<{
  change: [value: string | number | Array<string | number> | undefined]
}>()

const options = ref<DictOption[]>([])
const loading = ref(false)

function parseValue(raw: string): string | number {
  if (/^-?\d+$/.test(raw)) return Number(raw)
  return raw
}

async function loadDict(code: string) {
  if (!code) {
    options.value = []
    return
  }
  loading.value = true
  try {
    const res = await getDictByCode(code)
    options.value = (res.data || []).map((d) => ({
      label: d.label,
      value: parseValue(d.value),
    }))
  } catch {
    options.value = []
  } finally {
    loading.value = false
  }
}

watch(
  () => props.dictCode,
  (code) => {
    void loadDict(code || '')
  },
  { immediate: true },
)
</script>
