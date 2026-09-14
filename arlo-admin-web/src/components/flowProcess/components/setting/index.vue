<template>
  <el-drawer
    v-model="dialogShow"
    title="节点配置"
    direction="rtl"
    size="560px"
    append-to-body
    destroy-on-close
    class="flow-node-setting-drawer"
  >
    <component
      v-if="dialogShow && settingComp"
      ref="settingRef"
      :is="settingComp"
      :data="form"
      :flow="props.flow"
    />
    <el-empty v-else-if="dialogShow" description="暂无该节点配置面板" :image-size="72" />
    <template #footer>
      <div class="drawer-footer">
        <el-button type="primary" :icon="CirclePlus" @click="dialogSubmit">确定</el-button>
        <el-button :icon="CircleClose" @click="dialogClose">取消</el-button>
      </div>
    </template>
  </el-drawer>
</template>

<script lang="ts" setup>
import { computed, defineAsyncComponent, inject, ref } from 'vue'
import { CirclePlus, CircleClose } from '@element-plus/icons-vue'

const props = defineProps<{
  flow: any
}>()

const emits = defineEmits<{
  success: [value: any]
  close: []
}>()

const syncFlowProcessForm = inject<() => void>('syncFlowProcessForm', () => {})

const dialogShow = ref(false)
const dialogData = ref<any>({})
const settingRef = ref()
const form = ref<{ node: any; parent: any }>({
  node: {},
  parent: {},
})

/** Vite 需静态可分析；用 glob 收集已有面板，再按 key 取 */
const modules = import.meta.glob('./*/index.vue')

const componentCache = new Map<string, ReturnType<typeof defineAsyncComponent>>()

function resolveLoader(key: string) {
  const path = `./${key}/index.vue`
  return modules[path] as (() => Promise<any>) | undefined
}

function getAsyncComp(key: string) {
  if (componentCache.has(key)) return componentCache.get(key)!
  const loader = resolveLoader(key)
  if (!loader) return null
  const comp = defineAsyncComponent(loader)
  componentCache.set(key, comp)
  return comp
}

/**
 * 条件分支面板：node-4-3 / node-8-3 / node-9-3
 * 其它节点：只用 node-{type}（如审批在条件子树下父级是 3，不能拼成 node-3-1）
 */
function pickTemplateKey(node: any, parent: any): string | null {
  const type = node?.type
  if (type === undefined || type === null || type === '') return null
  const parentType = parent?.type
  if (parentType === 4 || parentType === 8 || parentType === 9) {
    const compound = `node-${parentType}-${type}`
    if (resolveLoader(compound)) return compound
  }
  const plain = `node-${type}`
  if (resolveLoader(plain)) return plain
  return null
}

const settingComp = computed(() => {
  if (!dialogShow.value) return null
  const key = pickTemplateKey(form.value.node, form.value.parent)
  if (!key) return null
  return getAsyncComp(key)
})

function init(obj: any = {}, parent: any = {}) {
  // 打开节点配置前，把表单设计器最新字段同步进 processForm / formConfig
  syncFlowProcessForm()
  form.value = {
    node: JSON.parse(JSON.stringify(obj)),
    parent,
  }
  dialogData.value = obj
  dialogShow.value = true
}

function dialogClose() {
  dialogShow.value = false
  emits('close')
}

function dialogSubmit() {
  settingRef.value?.getFormRef()?.validate((valid: boolean) => {
    if (valid) {
      dialogShow.value = false
      Object.assign(dialogData.value, form.value.node)
      emits('success', dialogData.value)
    }
  })
}

defineExpose({ init })
</script>

<style lang="scss">
.flow-node-setting-drawer {
  .el-drawer__header {
    margin-bottom: 0;
    padding: 16px 20px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }
  .el-drawer__title {
    font-size: 16px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    line-height: 24px;
  }
  .el-drawer__body {
    padding: 16px 20px;
  }
  .el-drawer__footer {
    padding: 12px 20px;
    border-top: 1px solid var(--el-border-color-lighter);
  }
}
</style>

<style lang="scss" scoped>
.drawer-footer {
  text-align: left;
  :deep(.el-button + .el-button) {
    margin-left: 0;
  }
  display: flex;
  gap: 10px;
}
</style>
