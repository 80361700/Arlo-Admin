<template>
  <div class="page-container">
    <ProTable
      ref="proTableRef"
      :data="tableData"
      :loading="loading"
      :total="total"
      :search-fields="searchFields"
      :show-index="false"
      @search="handleSearch"
      @reset="handleReset"
      @page-change="handlePageChange"
    >
      <template #toolbar>
        <el-button v-permission="'sys:role:add'" type="primary" @click="handleAdd">新增角色</el-button>
      </template>

      <el-table-column prop="id" label="ID" width="80" align="center" />
      <el-table-column prop="name" label="角色名称" min-width="140" />
      <el-table-column prop="code" label="角色编码" min-width="140" />
      <el-table-column prop="sort" label="排序" width="80" align="center" />
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ statusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="数据范围" min-width="120">
        <template #default="{ row }">
          {{ resolveDataScopeLabel(row.dataScope) }}
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
      <el-table-column prop="createdAt" label="创建时间" min-width="170" />

      <template #actions="{ row }">
        <el-button v-permission="'sys:role:edit'" type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button v-permission="'sys:role:edit'" type="success" link size="small" @click="handleAssignMenus(row)">分配菜单</el-button>
        <el-dropdown
          v-if="authStore.hasPermission('sys:role:edit') || authStore.hasPermission('sys:role:delete')"
          trigger="click"
          @command="(cmd: string) => handleAction(row, cmd)"
        >
          <el-button type="info" link size="small">
            更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-if="authStore.hasPermission('sys:role:delete')" command="delete">删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
    </ProTable>

    <!-- 新增/编辑弹窗 -->
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
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" maxlength="32" />
        </el-form-item>
        <el-form-item label="角色编码" prop="code">
          <el-input v-model="form.code" placeholder="请输入角色编码" maxlength="32" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio v-for="opt in statusOptions" :key="String(opt.value)" :value="opt.value">
              {{ opt.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="数据范围">
          <el-select v-model="form.dataScope" style="width: 100%">
            <el-option
              v-for="opt in dataScopeOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
          <div class="form-tip">作用于用户、文件、日志、公告、消息发件等业务数据；不限制角色/部门/岗位/字典/菜单/会员等主数据</div>
        </el-form-item>
        <el-form-item label="指定部门" v-if="form.dataScope === 2">
          <DeptTree v-model="form.deptIds" multiple style="width: 100%" />
          <div class="form-tip">仅「自定义」时生效：勾选后该角色只能看到这些部门的数据</div>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" placeholder="请输入备注" :rows="2" />
        </el-form-item>
      </template>
    </ProFormDialog>

    <!-- 分配菜单弹窗 -->
    <el-dialog v-model="menuDialogVisible" title="分配菜单" width="500px" :close-on-click-modal="false">
      <el-tree
        ref="menuTreeRef"
        :data="menuTreeData"
        :props="{ label: 'name', children: 'children' }"
        node-key="id"
        show-checkbox
        default-expand-all
        :check-strictly="false"
      />
      <template #footer>
        <el-button @click="menuDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="menuSubmitting" @click="handleMenuSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import { ArrowDown } from '@element-plus/icons-vue'
import type { FormRules } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import DeptTree from '@/components/DeptTree.vue'
import {
  getRoleList, getRoleDetail, createRole, updateRole, deleteRole,
  getRoleMenus, assignRoleMenus, getMenuTree,
} from '@/api'
import type { RoleItem, RoleListParams, MenuTreeNode } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useDict, DictCode } from '@/utils/useDict'

const authStore = useAuthStore()
const { options: statusOptions, getLabel: statusLabel } = useDict(DictCode.UserStatus)

// ==================== 查询 ====================
const proTableRef = ref()
const tableData = ref<RoleItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive<RoleListParams>({ page: 1, pageSize: 10 })

const searchFields = computed(() => [
  { prop: 'name', label: '角色名称' },
  { prop: 'code', label: '角色编码' },
  { prop: 'status', label: '状态', type: 'select' as const, options: statusOptions.value },
])

const { options: dataScopeOptions, getLabel: dataScopeLabel } = useDict(DictCode.DataScope)

function resolveDataScopeLabel(v: number) {
  const fromDict = dataScopeLabel(v, '')
  if (fromDict) return fromDict
  const fallback: Record<number, string> = {
    1: '全部数据', 2: '自定义', 3: '本部门及以下', 4: '本部门', 5: '仅本人',
  }
  return fallback[v] || '-'
}

async function loadData(params: RoleListParams) {
  loading.value = true
  try {
    const res = await getRoleList(params)
    tableData.value = res.data.list || []
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) { query.page = 1; Object.assign(query, p); loadData(query) }
function handleReset() { loadData({ page: 1, pageSize: 10 }) }
function handlePageChange(p: any) { Object.assign(query, p); loadData(query) }

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const dialogTitle = ref('新增角色')
const isEdit = ref(false)
const submitting = ref(false)
const formDialogRef = ref()

const defaultForm = {
  name: '', code: '', sort: 0, status: 1 as number, remark: '',
  dataScope: 1 as number, deptIds: [] as number[],
}
const form = reactive({ ...defaultForm })

const formRules: FormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }],
}

function handleAdd() {
  isEdit.value = false
  dialogTitle.value = '新增角色'
  Object.assign(form, { ...defaultForm })
  dialogVisible.value = true
}

async function handleEdit(row: RoleItem) {
  isEdit.value = true
  dialogTitle.value = '编辑角色'
  try {
    const res = await getRoleDetail(row.id)
    const r = res.data
    Object.assign(form, {
      id: r.id,
      name: r.name,
      code: r.code,
      sort: r.sort,
      status: r.status,
      remark: r.remark,
      dataScope: r.dataScope,
      deptIds: r.deptIds || [],
    })
    dialogVisible.value = true
  } catch (err: any) {
    showRequestError(err, '获取角色信息失败')
  }
}

async function handleSubmit() {
  submitting.value = true
  try {
    const payload: any = {
      name: form.name, code: form.code, sort: form.sort,
      status: form.status, remark: form.remark,
      dataScope: form.dataScope,
      deptIds: form.dataScope === 2 ? form.deptIds : [],
    }
    if (isEdit.value) {
      payload.id = (form as any).id
      await updateRole(payload)
      ElMessage.success('更新成功')
    } else {
      await createRole(payload)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    loadData(query)
  } catch (err: any) {
    showRequestError(err, '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await deleteRole(id)
    ElMessage.success('删除成功')
    loadData(query)
  } catch (err: any) {
    showRequestError(err, '删除失败')
  }
}

function handleAction(row: RoleItem, command: string) {
  if (command === 'delete') {
    ElMessageBox.confirm('确认删除该角色？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    }).then(() => {
      handleDelete(row.id)
    }).catch(() => {})
  }
}

// ==================== 分配菜单 ====================
const menuDialogVisible = ref(false)
const menuSubmitting = ref(false)
const menuTreeData = ref<MenuTreeNode[]>([])
const menuTreeRef = ref()
let currentRoleId = 0

async function handleAssignMenus(row: RoleItem) {
  currentRoleId = row.id
  try {
    const [menuRes, checkedRes] = await Promise.all([getMenuTree(), getRoleMenus(row.id)])
    menuTreeData.value = menuRes.data || []
    // 先打开弹窗让 el-tree 挂载到 DOM，再设置选中
    menuDialogVisible.value = true
    await nextTick()
    // 只设置叶子节点（过滤掉父级目录），el-tree 会根据子节点自动计算父级的半选/全选状态
    const leafIds = getLeafMenuIds(menuTreeData.value)
    const checkedLeafIds = (checkedRes.data || []).filter(id => leafIds.includes(id))
    menuTreeRef.value?.setCheckedKeys(checkedLeafIds)
  } catch (err: any) {
    showRequestError(err, '获取菜单数据失败')
  }
}

// 从菜单树中提取所有叶子节点 ID（目录节点有 children，叶子节点没有）
function getLeafMenuIds(nodes: MenuTreeNode[]): number[] {
  const ids: number[] = []
  for (const node of nodes) {
    if (!node.children || node.children.length === 0) {
      ids.push(node.id)
    } else {
      ids.push(...getLeafMenuIds(node.children))
    }
  }
  return ids
}

async function handleMenuSubmit() {
  const checkedKeys = menuTreeRef.value?.getCheckedKeys() || []
  const halfCheckedKeys = menuTreeRef.value?.getHalfCheckedKeys() || []
  menuSubmitting.value = true
  try {
    await assignRoleMenus({ roleId: currentRoleId, menuIds: [...checkedKeys, ...halfCheckedKeys] })
    ElMessage.success('菜单分配成功')
    menuDialogVisible.value = false
    // 会话内热刷新：当前用户若受影响，侧边栏/按钮权限立即更新
    try {
      await authStore.refreshPermissions()
    } catch {
      ElMessage.warning('菜单已保存，请手动点击「刷新权限」或重新登录')
    }
  } catch (err: any) {
    showRequestError(err, '分配失败')
  } finally {
    menuSubmitting.value = false
  }
}

onMounted(() => {
  loadData(query)
})
</script>

<style scoped>
.form-tip {
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
}
</style>
