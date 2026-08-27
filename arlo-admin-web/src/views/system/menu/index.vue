<template>
  <div class="page-container">
    <div class="toolbar">
      <el-button v-permission="'sys:menu:add'" type="primary" @click="handleAdd()">新增菜单</el-button>
    </div>

    <el-table
      ref="tableRef"
      :data="tableRoots"
      row-key="id"
      border
      lazy
      :load="loadTreeChildren"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      v-loading="loading"
    >
      <el-table-column prop="name" label="菜单名称" min-width="220" />
      <el-table-column label="图标" width="70" align="center">
        <template #default="{ row }">
          <el-icon v-if="row.icon" :size="18"><component :is="row.icon" /></el-icon>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="typeTag[row.type] || undefined" size="small">{{ typeMap[row.type] }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="permission" label="权限标识" min-width="180" show-overflow-tooltip />
      <el-table-column prop="path" label="路由路径" min-width="160" show-overflow-tooltip />
      <el-table-column prop="component" label="组件路径" min-width="180" show-overflow-tooltip />
      <el-table-column prop="sort" label="排序" width="80" align="center" />
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right" align="center">
        <template #default="{ row }">
          <div class="action-buttons">
            <el-button v-permission="'sys:menu:add'" type="primary" link size="small" @click="handleAdd(row as MenuTreeNode)">新增子级</el-button>
            <el-button v-permission="'sys:menu:edit'" type="primary" link size="small" @click="handleEdit(row as MenuTreeNode)">编辑</el-button>
            <el-dropdown v-if="authStore.hasPermission('sys:menu:delete')" trigger="click" @command="(cmd: string) => handleAction(row as MenuTreeNode, cmd)">
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

    <!-- 新增/编辑弹窗 -->
    <ProFormDialog
      ref="formDialogRef"
      v-model="dialogVisible"
      :title="dialogTitle"
      :model="form"
      :rules="formRules"
      width="600px"
      :submitting="submitting"
      @submit="handleSubmit"
    >
      <template #default>
        <el-form-item label="上级菜单">
          <el-tree-select
            v-model="parentSelect"
            :data="parentOptions"
            node-key="id"
            :props="{ label: 'name', children: 'children', disabled: 'disabled' }"
            check-strictly
            filterable
            clearable
            placeholder="空表示顶级菜单"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="菜单类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio :value="1">目录</el-radio>
            <el-radio :value="2">菜单</el-radio>
            <el-radio :value="3">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="菜单名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入菜单名称" maxlength="32" />
        </el-form-item>
        <el-form-item label="图标" v-if="form.type !== 3">
          <IconPicker v-model="form.icon" placeholder="点击选择图标" />
        </el-form-item>
        <el-form-item label="路由路径" v-if="form.type !== 3">
          <el-input v-model="form.path" placeholder="/system/user" maxlength="128" />
        </el-form-item>
        <el-form-item label="组件路径" v-if="form.type === 2">
          <el-input v-model="form.component" placeholder="system/user/index" maxlength="128" />
        </el-form-item>
        <el-form-item label="权限标识" prop="permission" v-if="form.type !== 1">
          <el-input v-model="form.permission" placeholder="sys:user:list" maxlength="100" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="是否可见" v-if="form.type !== 3">
          <el-radio-group v-model="form.visible">
            <el-radio :value="1">显示</el-radio>
            <el-radio :value="0">隐藏</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="是否缓存" v-if="form.type === 2">
          <el-radio-group v-model="form.keepAlive">
            <el-radio :value="1">是</el-radio>
            <el-radio :value="0">否</el-radio>
          </el-radio-group>
        </el-form-item>
      </template>
    </ProFormDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import { ArrowDown } from '@element-plus/icons-vue'
import type { FormRules, TableInstance } from 'element-plus'
import ProFormDialog from '@/components/ProFormDialog.vue'
import IconPicker from '@/components/IconPicker.vue'
import { getMenuTree, createMenu, updateMenu, deleteMenu } from '@/api'
import type { MenuTreeNode } from '@/api'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const typeMap: Record<number, string> = { 1: '目录', 2: '菜单', 3: '按钮' }
const typeTag: Record<number, 'success' | 'warning' | 'info' | 'primary' | 'danger'> = {
  2: 'success',
  3: 'warning',
}

const loading = ref(false)
const tableRef = ref<TableInstance>()
/** 完整菜单树（表单上级选择 / lazy 子节点来源） */
const fullTree = ref<MenuTreeNode[]>([])
/** 表格根节点（去掉 children，配合 lazy） */
const tableRoots = ref<LazyMenuRow[]>([])

type LazyMenuRow = Omit<MenuTreeNode, 'children'> & { hasChildren?: boolean }

function toLazyRow(node: MenuTreeNode): LazyMenuRow {
  const hasKids = (node.children?.length ?? 0) > 0
  const { children: _c, ...rest } = node
  return { ...rest, hasChildren: hasKids }
}

function loadTreeChildren(
  row: LazyMenuRow,
  _treeNode: unknown,
  resolve: (rows: LazyMenuRow[]) => void,
) {
  const node = findNode(fullTree.value, row.id)
  resolve((node?.children || []).map(toLazyRow))
}

/** 刷新已展开分支的 lazy 缓存，不重挂载表格，展开状态得以保留 */
function syncLazyLoadedChildren() {
  const table = tableRef.value as TableInstance & {
    store?: { states?: { lazyTreeNodeMap?: { value: Record<string, LazyMenuRow[]> } } }
    updateKeyChildren?: (key: string, data: LazyMenuRow[]) => void
  }
  const map = table?.store?.states?.lazyTreeNodeMap?.value
  if (!table?.updateKeyChildren || !map) return
  for (const key of Object.keys(map)) {
    const node = findNode(fullTree.value, Number(key))
    table.updateKeyChildren(key, (node?.children || []).map(toLazyRow))
  }
}

async function loadData() {
  loading.value = true
  try {
    const res = await getMenuTree()
    fullTree.value = res.data || []
    tableRoots.value = fullTree.value.map(toLazyRow)
    await nextTick()
    syncLazyLoadedChildren()
  } finally {
    loading.value = false
  }
}

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const dialogTitle = ref('新增菜单')
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

/** 收集某节点及其全部子孙 id（编辑时禁止选为自己的上级） */
function collectSelfAndDescendantIds(node: MenuTreeNode | undefined, set: Set<number>) {
  if (!node) return
  set.add(node.id)
  for (const child of node.children || []) {
    collectSelfAndDescendantIds(child, set)
  }
}

function findNode(nodes: MenuTreeNode[], id: number): MenuTreeNode | undefined {
  for (const n of nodes) {
    if (n.id === id) return n
    const found = findNode(n.children || [], id)
    if (found) return found
  }
  return undefined
}

function mapParentOptions(nodes: MenuTreeNode[], excluded: Set<number>): ParentOption[] {
  const result: ParentOption[] = []
  for (const n of nodes) {
    // 按钮一般不作为上级；目录/菜单可选
    if (n.type === 3) continue
    result.push({
      id: n.id,
      name: n.name,
      disabled: excluded.has(n.id),
      children: mapParentOptions(n.children || [], excluded),
    })
  }
  return result
}

const parentOptions = computed<ParentOption[]>(() => {
  const excluded = new Set<number>()
  if (editingId.value) {
    collectSelfAndDescendantIds(findNode(fullTree.value, editingId.value), excluded)
  }
  // 不用 id=0 节点：Element Plus tree-select 对 0 值回显不稳定，清空即表示顶级
  return mapParentOptions(fullTree.value, excluded)
})

/** 树选择绑定值：undefined=顶级，避免 0 回显问题 */
const parentSelect = computed({
  get: () => (form.parentId && form.parentId > 0 ? form.parentId : undefined),
  set: (val: number | undefined | null) => {
    form.parentId = val && val > 0 ? val : 0
  },
})

const defaultForm = {
  parentId: 0 as number, name: '', type: 2 as number, path: '', component: '',
  icon: '', sort: 0, permission: '', status: 1 as number, visible: 1 as number, keepAlive: 1 as number,
}
const form = reactive({ ...defaultForm })

const formRules: FormRules = {
  name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择菜单类型', trigger: 'change' }],
}

function handleAdd(parent?: MenuTreeNode) {
  isEdit.value = false
  editingId.value = undefined
  dialogTitle.value = parent ? `新增子菜单` : '新增菜单'
  Object.assign(form, {
    ...defaultForm,
    parentId: parent ? parent.id : 0,
    type: parent ? (parent.type === 1 ? 2 : 3) : 1,
  })
  dialogVisible.value = true
}

function handleEdit(row: MenuTreeNode) {
  isEdit.value = true
  editingId.value = row.id
  dialogTitle.value = '编辑菜单'
  Object.assign(form, {
    id: row.id,
    parentId: row.parentId ?? 0,
    name: row.name,
    type: row.type,
    path: row.path,
    component: row.component,
    icon: row.icon,
    sort: row.sort,
    permission: row.permission,
    status: row.status,
    visible: row.visible,
    keepAlive: row.keepAlive,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  submitting.value = true
  try {
    const parentId = form.parentId == null ? 0 : form.parentId
    const payload: any = {
      parentId,
      name: form.name,
      type: form.type,
      path: form.path,
      component: form.component,
      icon: form.icon,
      sort: form.sort,
      permission: form.permission,
      status: form.status,
      visible: form.visible,
      keepAlive: form.keepAlive,
    }
    if (isEdit.value) {
      payload.id = (form as any).id
      await updateMenu(payload)
      ElMessage.success('更新成功')
    } else {
      await createMenu(payload)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    await loadData()
  } catch (err: any) {
    showRequestError(err, '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await deleteMenu(id)
    ElMessage.success('删除成功')
    await loadData()
  } catch (err: any) {
    showRequestError(err, '删除失败')
  }
}

function handleAction(row: MenuTreeNode, command: string) {
  if (command === 'delete') {
    ElMessageBox.confirm('确认删除该菜单及其子菜单？', '提示', {
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
