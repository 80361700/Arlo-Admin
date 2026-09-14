<template>
  <div class="page-container">
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
        :model="searchForm"
        class="filter-form"
        label-position="right"
        @submit.prevent="handleSearch"
      >
        <div class="filter-grid">
          <el-form-item label="部门名称：" class="filter-item">
            <el-input
              v-model="searchForm.name"
              placeholder="请输入部门名称"
              clearable
              style="width: 160px"
            />
          </el-form-item>
          <el-form-item label="部门编码：" class="filter-item">
            <el-input
              v-model="searchForm.code"
              placeholder="请输入部门编码"
              clearable
              style="width: 160px"
            />
          </el-form-item>
          <el-form-item label="状态：" class="filter-item">
            <el-select
              v-model="searchForm.status"
              placeholder="请选择"
              clearable
              style="width: 160px"
            >
              <el-option label="启用" :value="1" />
              <el-option label="禁用" :value="0" />
            </el-select>
          </el-form-item>
          <div class="filter-actions">
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </div>
        </div>
      </el-form>
    </div>

    <div class="toolbar">
      <el-button v-permission="'sys:dept:add'" type="primary" @click="handleAdd()">新增部门</el-button>
    </div>

    <el-table
      :key="isSearching ? 'search' : 'tree'"
      :data="displayData"
      row-key="id"
      border
      v-loading="loading"
      default-expand-all
    >
      <el-table-column prop="name" label="部门名称" min-width="180" />
      <el-table-column prop="code" label="部门编码" min-width="120" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.code || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="80" align="center" />
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="leader" label="负责人" min-width="100">
        <template #default="{ row }">
          {{ row.leader || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="phone" label="联系电话" min-width="130">
        <template #default="{ row }">
          {{ row.phone || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.email || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.remark || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right" align="center">
        <template #default="{ row }">
          <div class="action-buttons">
            <el-button v-permission="'sys:dept:add'" type="primary" link size="small" @click="handleAdd(row as DeptTreeNode)">新增子级</el-button>
            <el-button v-permission="'sys:dept:edit'" type="primary" link size="small" @click="handleEdit(row as DeptTreeNode)">编辑</el-button>
            <el-dropdown v-if="authStore.hasPermission('sys:dept:delete')" trigger="click" @command="(cmd: string) => handleAction(row as DeptTreeNode, cmd)">
              <el-button type="info" link size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="delete">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <ProFormDialog
      ref="formDialogRef"
      v-model="dialogVisible"
      :title="dialogTitle"
      :model="form"
      :rules="formRules"
      :submitting="submitting"
      @submit="handleSubmit"
    >
      <template #default>
        <el-form-item label="上级部门">
          <el-tree-select
            v-model="parentSelect"
            :data="parentOptions"
            node-key="id"
            :props="{ label: 'name', children: 'children', disabled: 'disabled' }"
            check-strictly
            filterable
            clearable
            placeholder="空表示顶级部门"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入部门名称" maxlength="64" />
        </el-form-item>
        <el-form-item label="部门编码" prop="code">
          <el-input v-model="form.code" placeholder="请输入部门编码" maxlength="64" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="负责人">
          <UserSelect v-model="leaderSelect" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="请输入备注" maxlength="255" show-word-limit />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </template>
    </ProFormDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import { ArrowDown, ArrowUp, Filter } from '@element-plus/icons-vue'
import type { FormRules } from 'element-plus'
import ProFormDialog from '@/components/ProFormDialog.vue'
import UserSelect from '@/components/UserSelect.vue'
import { getDeptTree, createDept, updateDept, deleteDept } from '@/api'
import type { DeptTreeNode } from '@/api'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const loading = ref(false)
const filterCollapsed = ref(false)
/** 完整部门树（表单上级选择用） */
const fullTree = ref<DeptTreeNode[]>([])

interface DeptSearchQuery {
  name?: string
  code?: string
  status?: number
}

const searchForm = reactive<DeptSearchQuery>({
  name: undefined,
  code: undefined,
  status: undefined,
})
const activeQuery = reactive<DeptSearchQuery>({
  name: undefined,
  code: undefined,
  status: undefined,
})

const isSearching = computed(() => {
  const name = activeQuery.name?.trim()
  const code = activeQuery.code?.trim()
  return !!(name || code || activeQuery.status === 0 || activeQuery.status === 1)
})

function includesIgnoreCase(haystack: string, needle: string) {
  return haystack.toLowerCase().includes(needle.toLowerCase())
}

function matchNode(node: DeptTreeNode): boolean {
  const name = activeQuery.name?.trim()
  if (name && !includesIgnoreCase(node.name || '', name)) return false
  const code = activeQuery.code?.trim()
  if (code && !includesIgnoreCase(node.code || '', code)) return false
  if (activeQuery.status === 0 || activeQuery.status === 1) {
    if (node.status !== activeQuery.status) return false
  }
  return true
}

/** 保留命中节点及其祖先，子树仍保持树形结构 */
function filterDeptTree(nodes: DeptTreeNode[]): DeptTreeNode[] {
  const result: DeptTreeNode[] = []
  for (const node of nodes) {
    const children = filterDeptTree(node.children || [])
    if (matchNode(node) || children.length > 0) {
      result.push({ ...node, children })
    }
  }
  return result
}

const displayData = computed(() => {
  if (!isSearching.value) return fullTree.value
  return filterDeptTree(fullTree.value)
})

function applySearchParams(p: DeptSearchQuery) {
  activeQuery.name = p.name?.trim() || undefined
  activeQuery.code = p.code?.trim() || undefined
  activeQuery.status = p.status === 0 || p.status === 1 ? p.status : undefined
}

function handleSearch() {
  applySearchParams(searchForm)
}

function handleReset() {
  searchForm.name = undefined
  searchForm.code = undefined
  searchForm.status = undefined
  applySearchParams({})
}

async function loadData() {
  loading.value = true
  try {
    const res = await getDeptTree()
    fullTree.value = res.data || []
  } finally {
    loading.value = false
  }
}

const dialogVisible = ref(false)
const dialogTitle = ref('新增部门')
const isEdit = ref(false)
const submitting = ref(false)
const formDialogRef = ref()
const editingId = ref<number | undefined>()

interface ParentOption {
  id: number
  name: string
  disabled?: boolean
  children?: ParentOption[]
}

function collectSelfAndDescendantIds(node: DeptTreeNode | undefined, set: Set<number>) {
  if (!node) return
  set.add(node.id)
  for (const child of node.children || []) {
    collectSelfAndDescendantIds(child, set)
  }
}

function findNode(nodes: DeptTreeNode[], id: number): DeptTreeNode | undefined {
  for (const n of nodes) {
    if (n.id === id) return n
    const found = findNode(n.children || [], id)
    if (found) return found
  }
  return undefined
}

function mapParentOptions(nodes: DeptTreeNode[], excluded: Set<number>): ParentOption[] {
  return nodes.map((n) => ({
    id: n.id,
    name: n.name,
    disabled: excluded.has(n.id),
    children: mapParentOptions(n.children || [], excluded),
  }))
}

const parentOptions = computed<ParentOption[]>(() => {
  const excluded = new Set<number>()
  if (editingId.value) {
    collectSelfAndDescendantIds(findNode(fullTree.value, editingId.value), excluded)
  }
  return mapParentOptions(fullTree.value, excluded)
})

const parentSelect = computed({
  get: () => (form.parentId && form.parentId > 0 ? form.parentId : undefined),
  set: (val: number | undefined | null) => {
    form.parentId = val && val > 0 ? val : 0
  },
})

const leaderSelect = computed({
  get: () => (form.leaderId && form.leaderId > 0 ? form.leaderId : undefined),
  set: (val: number | undefined | null) => {
    form.leaderId = val && val > 0 ? val : 0
  },
})

const defaultForm = {
  parentId: 0 as number, name: '', code: '', sort: 0, leaderId: 0 as number,
  remark: '', status: 1 as number,
}
const form = reactive({ ...defaultForm })

const formRules: FormRules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
}

function handleAdd(parent?: DeptTreeNode) {
  isEdit.value = false
  editingId.value = undefined
  dialogTitle.value = parent ? '新增子部门' : '新增部门'
  Object.assign(form, { ...defaultForm, parentId: parent ? parent.id : 0 })
  dialogVisible.value = true
}

function handleEdit(row: DeptTreeNode) {
  isEdit.value = true
  editingId.value = row.id
  dialogTitle.value = '编辑部门'
  Object.assign(form, {
    id: row.id, parentId: row.parentId, name: row.name, code: row.code || '',
    sort: row.sort, leaderId: row.leaderId || 0, remark: row.remark || '', status: row.status,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  submitting.value = true
  try {
    const payload = {
      parentId: form.parentId, name: form.name, code: form.code, sort: form.sort,
      leaderId: form.leaderId, remark: form.remark, status: form.status,
    }
    if (isEdit.value) {
      await updateDept({ id: (form as any).id, ...payload })
      ElMessage.success('更新成功')
    } else {
      await createDept(payload)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (err: any) {
    showRequestError(err, '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await deleteDept(id)
    ElMessage.success('删除成功')
    loadData()
  } catch (err: any) {
    showRequestError(err, '删除失败')
  }
}

function handleAction(row: DeptTreeNode, command: string) {
  if (command === 'delete') {
    ElMessageBox.confirm('确认删除该部门及其子部门？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    }).then(() => {
      handleDelete(row.id)
    }).catch(() => {})
  }
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.filter-panel {
  margin-bottom: 12px;
  padding: 6px 16px 0;
  background: var(--el-fill-color-light);
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
    color: #409eff;
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

.filter-form {
  padding-bottom: 0;
}

.filter-grid {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0 8px;
}

.filter-item {
  margin: 0 16px 12px 0;
  width: auto;

  :deep(.el-form-item__label) {
    color: #606266;
    font-weight: 400;
    padding-right: 0;
  }

  :deep(.el-form-item__content) {
    flex: none;
  }
}

.filter-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 12px 0;

  :deep(.el-button + .el-button) {
    margin-left: 0;
  }
}

.toolbar { margin-bottom: 12px; }
.action-buttons {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;

  :deep(.el-button + .el-button) {
    margin-left: 0;
  }
}
</style>
