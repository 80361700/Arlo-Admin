<template>
  <div class="page-container launch-page" v-loading="loading">
    <div class="launch-row">
      <aside class="launch-left">
        <div class="left-group">
          <div class="left-group-header">
            <div class="left-group-title">
              <span>流程分类</span>
              <div class="title-actions">
                <el-button class="header-action-btn is-icon" size="small" title="刷新" @click="load">
                  <el-icon :size="14"><Refresh /></el-icon>
                </el-button>
              </div>
            </div>
          </div>
          <ul class="group-list">
            <li
              v-for="c in categories"
              :key="String(c.id)"
              :class="{ active: activeCategoryId === c.id }"
              @click="activeCategoryId = c.id"
            >
              <span class="name">{{ c.name }}</span>
              <span class="count">{{ c.count }}</span>
            </li>
          </ul>
        </div>
      </aside>

      <section class="launch-right">
        <div class="panel-head">
          <div class="panel-title">{{ activeCategoryName }}</div>
          <div class="panel-sub">共 {{ filteredList.length }} 个可发起流程</div>
        </div>

        <div class="filter-panel">
          <div class="filter-header" @click="filterCollapsed = !filterCollapsed">
            <el-icon class="filter-icon"><Filter /></el-icon>
            <span>筛选</span>
            <el-icon class="filter-arrow" :class="{ 'is-collapsed': filterCollapsed }">
              <ArrowUp />
            </el-icon>
          </div>
          <el-form
            v-show="!filterCollapsed"
            class="filter-form"
            label-position="right"
            @submit.prevent="handleSearch"
          >
            <div class="filter-grid">
              <el-form-item label="流程名称：" class="filter-item">
                <el-input
                  v-model="processName"
                  placeholder="请输入流程名称"
                  clearable
                  style="width: 200px"
                  @keyup.enter="handleSearch"
                />
              </el-form-item>
              <el-form-item label="流程编号：" class="filter-item">
                <el-input
                  v-model="processKey"
                  placeholder="请输入流程编号"
                  clearable
                  style="width: 200px"
                  @keyup.enter="handleSearch"
                />
              </el-form-item>
              <div class="filter-actions">
                <el-button type="primary" @click="handleSearch">查询</el-button>
                <el-button @click="handleReset">重置</el-button>
              </div>
            </div>
          </el-form>
        </div>

        <div v-if="filteredList.length" class="card-grid">
          <div
            v-for="row in filteredList"
            :key="row.processId"
            class="process-card"
            @click="openLaunch(row)"
          >
            <div class="card-icon" :style="{ backgroundColor: row.processBgcolor || '#1e90ff' }">
              <el-icon v-if="row.processIcon" :size="22" color="#fff">
                <component :is="row.processIcon" />
              </el-icon>
              <span v-else class="icon-fallback">流</span>
            </div>
            <div class="card-text">
              <div class="card-name" :title="row.processName">{{ row.processName }}</div>
              <div class="card-desc" :title="row.remark || row.processKey">
                {{ row.remark || row.processKey || '暂无说明' }}
              </div>
            </div>
          </div>
        </div>
        <el-empty v-else description="暂无匹配的流程" :image-size="72" />
      </section>
    </div>

    <el-drawer v-model="drawer" :title="`发起 · ${current?.processName || ''}`" size="640px" destroy-on-close>
      <div v-loading="formLoading" class="launch-drawer-body">
        <CustomFlowFormHost
          v-if="formRenderType === 'vue' && formComponent"
          ref="customFormRef"
          :component-path="formComponent"
        />
        <EBuilder
          v-if="pageSchema"
          ref="builderRef"
          :page-schema="pageSchema"
          :field-states="fieldStates"
        />
        <el-empty
          v-else-if="!modelContent && !(formRenderType === 'vue' && formComponent)"
          description="暂无表单，可直接提交"
        />
        <LaunchFlowPreview v-if="modelContent" ref="previewRef" :model-content="modelContent" />
      </div>
      <template #footer>
        <el-button @click="drawer = false">取消</el-button>
        <el-button :loading="submitting" @click="submitLaunch(true)">暂存</el-button>
        <el-button type="primary" :loading="submitting" @click="submitLaunch(false)">提交</el-button>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowUp, Filter, Refresh } from '@element-plus/icons-vue'
import { EBuilder, type FieldStates, type PageSchema } from 'epic-designer'
import LaunchFlowPreview from '../components/LaunchFlowPreview.vue'
import CustomFlowFormHost from '../components/CustomFlowFormHost.vue'
import {
  getFlowCategoryOptions,
  getLaunchProcessForm,
  getLaunchProcessList,
  launchFlowProcess,
  type LaunchProcessItem,
  type LaunchSelectionNode,
} from '@/api'
import {
  buildFieldStates,
  normalizeProcessForm,
  readBuilderFormData,
} from '@/views/flow/process/components/formSchema'

const loading = ref(false)
const all = ref<LaunchProcessItem[]>([])
const categoryOptions = ref<{ id: number; name: string }[]>([])
const processName = ref('')
const processKey = ref('')
const appliedProcessName = ref('')
const appliedProcessKey = ref('')
const filterCollapsed = ref(false)
const activeCategoryId = ref<number | string>('all')

const drawer = ref(false)
const formLoading = ref(false)
const submitting = ref(false)
const current = ref<LaunchProcessItem | null>(null)
const pageSchema = ref<PageSchema | null>(null)
const fieldStates = ref<FieldStates>([])
const modelContent = ref<Record<string, any> | null>(null)
const builderRef = ref<InstanceType<typeof EBuilder>>()
const customFormRef = ref<InstanceType<typeof CustomFlowFormHost>>()
const previewRef = ref<InstanceType<typeof LaunchFlowPreview>>()
const selectionNodes = ref<LaunchSelectionNode[]>([])
const formRenderType = ref('designer')
const formComponent = ref('')

/** 发起人节点 formConfig；历史全只读视为未配置，按可编辑处理 */
function startFormConfigOf(mc: Record<string, any> | null | undefined) {
  const root = (mc?.nodeConfig || mc) as any
  const raw = root?.extendConfig?.formConfig
  if (!Array.isArray(raw) || !raw.length) return []
  if (raw.every((c: any) => Number(c.opera) === 0)) return []
  return raw
}

const searchedList = computed(() => {
  const nameKw = appliedProcessName.value.trim().toLowerCase()
  const keyKw = appliedProcessKey.value.trim().toLowerCase()
  if (!nameKw && !keyKw) return all.value
  return all.value.filter((p) => {
    if (nameKw && !(p.processName || '').toLowerCase().includes(nameKw)) return false
    if (keyKw && !(p.processKey || '').toLowerCase().includes(keyKw)) return false
    return true
  })
})

const categories = computed(() => {
  const countMap = new Map<number, number>()
  for (const p of searchedList.value) {
    const id = Number(p.categoryId) || 0
    countMap.set(id, (countMap.get(id) || 0) + 1)
  }

  const list = categoryOptions.value.map((c) => ({
    id: c.id,
    name: c.name,
    count: countMap.get(c.id) || 0,
  }))

  const uncategorized = countMap.get(0) || 0
  if (uncategorized > 0 && !list.some((c) => Number(c.id) === 0)) {
    list.push({ id: 0, name: '未分类', count: uncategorized })
  }

  return [{ id: 'all' as const, name: '全部', count: searchedList.value.length }, ...list]
})

const activeCategoryName = computed(() => {
  const hit = categories.value.find((c) => c.id === activeCategoryId.value)
  return hit?.name || '全部'
})

const filteredList = computed(() => {
  if (activeCategoryId.value === 'all') return searchedList.value
  return searchedList.value.filter((p) => Number(p.categoryId) === Number(activeCategoryId.value))
})

function handleSearch() {
  appliedProcessName.value = processName.value
  appliedProcessKey.value = processKey.value
  if (!categories.value.some((c) => c.id === activeCategoryId.value)) {
    activeCategoryId.value = 'all'
  }
}

function handleReset() {
  processName.value = ''
  processKey.value = ''
  appliedProcessName.value = ''
  appliedProcessKey.value = ''
  activeCategoryId.value = 'all'
}

async function load() {
  loading.value = true
  try {
    const [procRes, catRes] = await Promise.all([
      getLaunchProcessList(),
      getFlowCategoryOptions(),
    ])
    all.value = procRes.data || []
    categoryOptions.value = (catRes.data || []).map((c) => ({
      id: Number(c.id),
      name: c.name,
    }))
  } finally {
    loading.value = false
  }
}

async function openLaunch(row: LaunchProcessItem) {
  current.value = row
  drawer.value = true
  formLoading.value = true
  pageSchema.value = null
  fieldStates.value = []
  modelContent.value = null
  selectionNodes.value = []
  formRenderType.value = 'designer'
  formComponent.value = ''
  try {
    const res = await getLaunchProcessForm(row.processId)
    formRenderType.value = res.data?.formRenderType === 'vue' ? 'vue' : 'designer'
    formComponent.value = res.data?.formComponent || ''
    const schema = normalizeProcessForm(res.data?.processForm)
    pageSchema.value = schema?.schemas?.length ? schema : null
    const mc = res.data?.modelContent
    modelContent.value = mc && typeof mc === 'object' && Object.keys(mc).length ? mc : null
    fieldStates.value = buildFieldStates(startFormConfigOf(modelContent.value), true)
    selectionNodes.value = res.data?.selectionNodes || []
  } finally {
    formLoading.value = false
  }
}

async function submitLaunch(saveAsDraft = false) {
  if (!current.value) return
  const picks = previewRef.value?.getSelectionPicks?.() || { nodeAssignees: {}, nodeCcUsers: {} }
  if (!saveAsDraft) {
    for (const node of selectionNodes.value) {
      const bag = node.kind === 'cc' ? picks.nodeCcUsers : picks.nodeAssignees
      if (node.required && !(bag[node.nodeKey]?.length)) {
        ElMessage.warning(`请为「${node.nodeName}」选择${node.kind === 'cc' ? '抄送人' : '审批人'}`)
        return
      }
    }
    if (formRenderType.value === 'vue') {
      const ok = await customFormRef.value?.validate?.()
      if (!ok) {
        ElMessage.warning('请完善表单必填项')
        return
      }
    }
    if (builderRef.value) {
      try {
        await (builderRef.value as any).validate?.()
      } catch {
        ElMessage.warning('请完善表单必填项')
        return
      }
    }
  }
  submitting.value = true
  try {
    const formData: Record<string, any> = {}
    if (formRenderType.value === 'vue') {
      Object.assign(formData, customFormRef.value?.getData?.() || {})
    }
    if (builderRef.value) {
      Object.assign(formData, await readBuilderFormData(builderRef.value))
    }
    await launchFlowProcess({
      processId: current.value.processId,
      formData,
      nodeAssignees: picks.nodeAssignees,
      nodeCcUsers: picks.nodeCcUsers,
      saveAsDraft,
    })
    ElMessage.success(saveAsDraft ? '已暂存，可在「我的申请」继续发起' : '发起成功')
    drawer.value = false
  } finally {
    submitting.value = false
  }
}

onMounted(() => load())
</script>

<style lang="scss" scoped>
.launch-page {
  height: 100%;
  min-height: 520px;
  background: var(--el-bg-color);
}
.launch-row {
  display: flex;
  align-items: stretch;
  height: 100%;
  min-height: 520px;
  gap: 16px;
}
.launch-left {
  width: 380px;
  flex: 0 0 380px;
  min-width: 380px;
  max-width: 380px;
  height: 100%;
}
.launch-right {
  flex: 1;
  min-width: 0;
  height: 100%;
  overflow: auto;
}

.left-group {
  border-right: 1px solid var(--el-border-color-lighter);
  padding-right: 16px;
  height: calc(100% - 8px);
  overflow: auto;
  position: relative;
  background: var(--el-bg-color);
  display: flex;
  flex-direction: column;
}
.left-group-header {
  background: var(--el-bg-color);
  position: sticky;
  top: 0;
  z-index: 2;
  padding-bottom: 12px;
}
.left-group-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
  span {
    font-size: 16px;
    font-weight: 500;
  }
  .title-actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .header-action-btn {
    margin: 0;
    height: 28px;
    padding: 0 10px;
    font-size: 13px;
    font-weight: 400;
    color: var(--el-text-color-regular);
    background: #fff;
    border: 1px solid var(--el-border-color);
    &:hover,
    &:focus {
      color: var(--el-color-primary);
      border-color: var(--el-color-primary-light-5);
      background: var(--el-color-primary-light-9);
    }
    &.is-icon {
      width: 28px;
      padding: 0;
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }
  }
}

.group-list {
  list-style: none;
  margin: 0;
  padding: 0 0 12px;
  overflow: auto;
  flex: 1;

  li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 10px 12px;
    margin-bottom: 2px;
    border-radius: 6px;
    cursor: pointer;
    color: var(--el-text-color-regular);
    transition: background 0.15s, color 0.15s;

    &:hover {
      background: var(--el-fill-color-light);
      color: var(--el-color-primary);
    }

    &.active {
      background: var(--el-color-primary-light-9);
      color: var(--el-color-primary);
      font-weight: 500;
    }

    .name {
      min-width: 0;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
      font-size: 14px;
    }
    .count {
      flex-shrink: 0;
      font-size: 12px;
      color: var(--el-text-color-secondary);
      background: var(--el-fill-color);
      border-radius: 10px;
      padding: 0 7px;
      line-height: 18px;
      min-width: 22px;
      text-align: center;
    }
    &.active .count {
      background: var(--el-color-primary-light-7);
      color: var(--el-color-primary);
    }
  }
}

.filter-panel {
  margin-bottom: 12px;
  padding: 6px 16px 0;
  background: var(--el-fill-color-light);
  border-radius: 4px;
}
.filter-header {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 28px;
  margin-bottom: 4px;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
  user-select: none;
  &:hover {
    color: var(--el-color-primary);
  }
}
.filter-icon {
  font-size: 14px;
}
.filter-arrow {
  font-size: 12px;
  transition: transform 0.2s ease;
  &.is-collapsed {
    transform: rotate(180deg);
  }
}
.filter-grid {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0 8px;
}
.filter-form {
  :deep(.el-form-item) {
    margin: 0 16px 12px 0;
  }
  :deep(.el-form-item__label) {
    color: #606266;
    font-weight: 400;
    padding-right: 0;
  }
}
.filter-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 12px;
  :deep(.el-button + .el-button) {
    margin-left: 0;
  }
}

.panel-head {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 16px;
}
.panel-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}
.panel-sub {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}
.process-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s, transform 0.15s;

  &:hover {
    border-color: var(--el-color-primary);
    box-shadow: 0 4px 12px rgba(64, 158, 255, 0.12);
    transform: translateY(-1px);
  }
}
.card-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.icon-fallback {
  color: #fff;
  font-size: 14px;
  font-weight: 600;
}
.card-text {
  min-width: 0;
  flex: 1;
}
.card-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.card-desc {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.launch-drawer-body {
  min-height: 240px;
}

@media (max-width: 1400px) {
  .card-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (max-width: 992px) {
  .launch-row {
    flex-direction: column;
  }
  .launch-left {
    width: 100%;
    flex: none;
    min-width: 0;
    max-width: none;
    height: auto;
  }
  .left-group {
    height: auto;
    max-height: 280px;
    padding-right: 0;
    border-right: 0;
    margin-bottom: 8px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }
  .card-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
