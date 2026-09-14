<template>
  <div v-loading="loading" class="custom-flow-form-host">
    <component
      v-if="comp"
      :is="comp"
      ref="innerRef"
      :readonly="readonly"
      :form-data="formData"
    />
    <el-alert
      v-else-if="!loading && error"
      type="error"
      :closable="false"
      :title="error"
      show-icon
    />
  </div>
</template>

<script setup lang="ts">
import { defineAsyncComponent, onMounted, ref, shallowRef, watch, type Component } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 相对 src/views/，不含 .vue，如 business/demo/form */
    componentPath: string
    formData?: Record<string, any>
    readonly?: boolean
  }>(),
  {
    formData: () => ({}),
    readonly: false,
  },
)

const modules = import.meta.glob('@/views/**/*.vue') as Record<string, () => Promise<any>>

const loading = ref(false)
const error = ref('')
const comp = shallowRef<Component | null>(null)
const innerRef = ref<{
  validate?: () => boolean | Promise<boolean>
  getData?: () => Record<string, any>
  setData?: (data: Record<string, any>) => void
} | null>(null)

function resolveLoader(path: string): (() => Promise<any>) | null {
  const raw = String(path || '')
    .trim()
    .replace(/^\/+/, '')
    .replace(/^src\/views\//, '')
    .replace(/^views\//, '')
    .replace(/\.vue$/, '')
  if (!raw) return null
  const candidates = [
    `/src/views/${raw}.vue`,
    `@/views/${raw}.vue`,
  ]
  for (const key of candidates) {
    if (modules[key]) return modules[key]
  }
  // vite glob keys often look like /src/views/...
  for (const [k, loader] of Object.entries(modules)) {
    if (k.endsWith(`/views/${raw}.vue`) || k.endsWith(`${raw}.vue`)) {
      return loader
    }
  }
  return null
}

async function load() {
  loading.value = true
  error.value = ''
  comp.value = null
  const loader = resolveLoader(props.componentPath)
  if (!loader) {
    error.value = `未找到自定义表单组件：${props.componentPath || '(空)'}`
    loading.value = false
    return
  }
  comp.value = defineAsyncComponent({
    loader,
    delay: 0,
    onError() {
      error.value = `自定义表单加载失败：${props.componentPath}`
    },
  })
  loading.value = false
}

watch(
  () => props.componentPath,
  () => {
    load()
  },
)

watch(
  () => props.formData,
  (v) => {
    if (innerRef.value?.setData && v) {
      innerRef.value.setData(v)
    }
  },
  { deep: true },
)

onMounted(async () => {
  await load()
})

async function validate(): Promise<boolean> {
  const fn = innerRef.value?.validate
  if (!fn) return true
  return !!(await fn())
}

function getData(): Record<string, any> {
  return innerRef.value?.getData?.() || {}
}

function setData(data: Record<string, any>) {
  innerRef.value?.setData?.(data || {})
}

defineExpose({ validate, getData, setData })
</script>

<style scoped>
.custom-flow-form-host {
  min-height: 48px;
}
</style>
