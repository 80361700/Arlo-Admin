<template>
  <div class="page-container">
    <div class="toolbar">
      <el-button v-permission="'sys:dept:add'" type="primary" @click="handleAdd()">新增部门</el-button>
    </div>

    <el-table
      :data="treeData"
      row-key="id"
      border
      v-loading="loading"
      default-expand-all
    >
      <el-table-column prop="name" label="部门名称" min-width="180" />
      <el-table-column prop="sort" label="排序" width="80" align="center" />
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="leader" label="负责人" min-width="100" />
      <el-table-column prop="phone" label="联系电话" min-width="130" />
      <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip />
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
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="form.leader" placeholder="请输入负责人" maxlength="32" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.phone" placeholder="请输入联系电话" maxlength="20" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="请输入邮箱" maxlength="64" />
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
import { ArrowDown } from '@element-plus/icons-vue'
import type { FormRules } from 'element-plus'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { getDeptTree, createDept, updateDept, deleteDept } from '@/api'
import type { DeptTreeNode } from '@/api'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const loading = ref(false)
const treeData = ref<DeptTreeNode[]>([])

async function loadData() {
  loading.value = true
  try {
    const res = await getDeptTree()
    treeData.value = res.data || []
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
    collectSelfAndDescendantIds(findNode(treeData.value, editingId.value), excluded)
  }
  return mapParentOptions(treeData.value, excluded)
})

const parentSelect = computed({
  get: () => (form.parentId && form.parentId > 0 ? form.parentId : undefined),
  set: (val: number | undefined | null) => {
    form.parentId = val && val > 0 ? val : 0
  },
})

const defaultForm = {
  parentId: 0 as number, name: '', sort: 0, leader: '', phone: '',
  email: '', status: 1 as number,
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
    id: row.id, parentId: row.parentId, name: row.name, sort: row.sort,
    leader: row.leader, phone: row.phone, email: row.email, status: row.status,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  submitting.value = true
  try {
    const payload = {
      parentId: form.parentId, name: form.name, sort: form.sort,
      leader: form.leader, phone: form.phone, email: form.email, status: form.status,
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
