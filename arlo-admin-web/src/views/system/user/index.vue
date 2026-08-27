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
        <el-button v-permission="'sys:user:add'" type="primary" @click="handleAdd">新增用户</el-button>
        <el-button v-permission="'sys:user:export'" @click="handleExport">导出</el-button>
        <el-button v-permission="'sys:user:import'" @click="handleDownloadTemplate">下载模板</el-button>
        <el-upload
          v-permission="'sys:user:import'"
          :show-file-list="false"
          :http-request="handleImport"
          accept=".xlsx,.xls"
          style="display: inline-block; margin-left: 12px"
        >
          <el-button :loading="importing">导入</el-button>
        </el-upload>
      </template>

      <el-table-column prop="id" label="ID" width="80" align="center" />
      <el-table-column label="头像" width="72" align="center">
        <template #default="{ row }">
          <AuthAvatar :file-ref="row.avatar" :size="32" />
        </template>
      </el-table-column>
      <el-table-column prop="username" label="用户名" min-width="120" />
      <el-table-column prop="nickname" label="昵称" min-width="120" />
      <el-table-column label="性别" width="80" align="center">
        <template #default="{ row }">
          {{ genderLabel(row.gender) }}
        </template>
      </el-table-column>
      <el-table-column prop="deptName" label="部门" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.deptName || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="phone" label="手机号" min-width="140" />
      <el-table-column prop="email" label="邮箱" min-width="180" show-overflow-tooltip />
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ statusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" min-width="170" />

      <template #actions="{ row }">
        <el-button v-permission="'sys:user:edit'" type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button v-permission="'sys:user:edit'" type="warning" link size="small" @click="handleResetPwd(row)">重置密码</el-button>
        <el-dropdown
          v-if="authStore.hasPermission('sys:user:delete') || authStore.hasPermission('sys:user:unlock')"
          trigger="click"
          @command="(cmd: string) => handleAction(row, cmd)"
        >
          <el-button type="info" link size="small">
            更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-if="authStore.hasPermission('sys:user:unlock')" command="unlock">解锁</el-dropdown-item>
              <el-dropdown-item v-if="authStore.hasPermission('sys:user:delete')" command="delete" divided>删除</el-dropdown-item>
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
        <el-form-item label="用户名" prop="username" v-if="!isEdit">
          <el-input v-model="form.username" placeholder="请输入用户名" maxlength="32" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password maxlength="32" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="请输入昵称" maxlength="32" />
        </el-form-item>
        <el-form-item label="头像">
          <div class="avatar-picker">
            <div class="avatar-wrap" @click="pickerVisible = true">
              <AuthAvatar :file-ref="form.avatar" :size="64" shape="square" />
              <span
                v-if="form.avatar"
                class="avatar-remove"
                title="清除头像"
                @click.stop="form.avatar = ''"
              >
                <el-icon :size="12"><Close /></el-icon>
              </span>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="性别">
          <el-radio-group v-model="form.gender">
            <el-radio v-for="opt in genderOptions" :key="String(opt.value)" :value="opt.value">
              {{ opt.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="部门">
          <DeptTree v-model="form.deptId" style="width: 100%" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.phone" placeholder="请输入手机号" maxlength="20" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="请输入邮箱" maxlength="64" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio v-for="opt in statusOptions" :key="String(opt.value)" :value="opt.value">
              {{ opt.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="角色">
          <el-checkbox-group v-model="form.roleIds">
            <el-checkbox v-for="r in roleList" :key="r.id" :value="r.id">
              {{ r.name }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="岗位">
          <el-checkbox-group v-model="form.postIds">
            <el-checkbox v-for="p in postList" :key="p.id" :value="p.id">
              {{ p.name }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" placeholder="请输入备注" :rows="2" />
        </el-form-item>
      </template>
    </ProFormDialog>

    <!-- 重置密码弹窗 -->
    <el-dialog v-model="pwdDialogVisible" title="重置密码" width="420px" :close-on-click-modal="false">
      <el-form :model="pwdForm" :rules="pwdRules" ref="pwdFormRef" label-width="80px">
        <el-form-item label="新密码" prop="password">
          <el-input v-model="pwdForm.password" type="password" placeholder="请输入新密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="pwdSubmitting" @click="handlePwdSubmit">确定</el-button>
      </template>
    </el-dialog>

    <FilePicker
      v-model="pickerVisible"
      title="选择头像"
      mode="single"
      :accept-types="['image']"
      :max-count="1"
      @confirm="onAvatarPicked"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import { ArrowDown, Close } from '@element-plus/icons-vue'
import type { FormRules, UploadRequestOptions } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import DeptTree from '@/components/DeptTree.vue'
import FilePicker from '@/components/FilePicker.vue'
import AuthAvatar from '@/components/AuthAvatar.vue'
import {
  getUserList, getUserDetail, createUser, updateUser, deleteUser, updateUserPassword,
  unlockUser, exportUsers, downloadUserImportTemplate, importUsers,
  getAllRoles, getAllPosts,
} from '@/api'
import type { UserItem, UserListParams, RoleItem, PostItem, FileItem } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { toFileRef } from '@/utils/fileUrl'
import { useDict, DictCode } from '@/utils/useDict'

const authStore = useAuthStore()
const { options: genderOptions, getLabel: genderLabel } = useDict(DictCode.Gender)
const { options: statusOptions, getLabel: statusLabel } = useDict(DictCode.UserStatus)

// ==================== 查询 ====================
const proTableRef = ref()
const tableData = ref<UserItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive<UserListParams>({ page: 1, pageSize: 10 })

const searchFields = computed(() => [
  { prop: 'username', label: '用户名' },
  { prop: 'nickname', label: '昵称' },
  { prop: 'phone', label: '手机号' },
  {
    prop: 'status', label: '状态', type: 'select' as const,
    options: statusOptions.value,
  },
])

async function loadData(params: UserListParams) {
  loading.value = true
  try {
    const res = await getUserList(params)
    tableData.value = res.data.list || []
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) { query.page = 1; Object.assign(query, p); loadData(query) }
function handleReset() { loadData({ page: 1, pageSize: 10 }) }
function handlePageChange(p: any) { Object.assign(query, p); loadData(query) }

// ==================== 角色/岗位选项 ====================
const roleList = ref<RoleItem[]>([])
const postList = ref<PostItem[]>([])

async function loadOptions() {
  const [rRes, pRes] = await Promise.allSettled([getAllRoles(), getAllPosts()])
  if (rRes.status === 'fulfilled') roleList.value = rRes.value.data || []
  if (pRes.status === 'fulfilled') postList.value = pRes.value.data || []
}

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const dialogTitle = ref('新增用户')
const isEdit = ref(false)
const submitting = ref(false)
const formDialogRef = ref()

const defaultForm = {
  id: 0 as number,
  username: '', password: '', nickname: '', avatar: '', gender: 0 as number,
  deptId: undefined as number | undefined, phone: '', email: '',
  status: 1 as number, remark: '', roleIds: [] as number[], postIds: [] as number[],
}
const form = reactive({ ...defaultForm })
const pickerVisible = ref(false)

function onAvatarPicked(files: FileItem[]) {
  if (!files.length) return
  form.avatar = toFileRef(files[0])
}
const formRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
}

function handleAdd() {
  isEdit.value = false
  dialogTitle.value = '新增用户'
  Object.assign(form, defaultForm, { roleIds: [], postIds: [] })
  dialogVisible.value = true
}

async function handleEdit(row: UserItem) {
  isEdit.value = true
  dialogTitle.value = '编辑用户'
  try {
    const res = await getUserDetail(row.id)
    const u = res.data
    Object.assign(form, {
      id: u.id,
      username: u.username,
      password: '',
      nickname: u.nickname,
      avatar: u.avatar || '',
      gender: u.gender,
      deptId: u.deptId,
      phone: u.phone,
      email: u.email,
      status: u.status,
      remark: u.remark,
      roleIds: u.roleIds || [],
      postIds: u.postIds || [],
    })
    dialogVisible.value = true
  } catch (err: any) {
    showRequestError(err, '获取用户信息失败')
  }
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateUser({
        id: form.id,
        nickname: form.nickname,
        avatar: form.avatar,
        gender: form.gender,
        deptId: form.deptId,
        phone: form.phone,
        email: form.email,
        status: form.status,
        remark: form.remark,
        roleIds: form.roleIds,
        postIds: form.postIds,
      })
      ElMessage.success('更新成功')
    } else {
      await createUser({
        ...form,
        username: form.username.trim(),
        nickname: form.nickname.trim(),
        avatar: form.avatar,
      })
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
    await deleteUser(id)
    ElMessage.success('删除成功')
    loadData(query)
  } catch (err: any) {
    showRequestError(err, '删除失败')
  }
}

function handleAction(row: UserItem, command: string) {
  if (command === 'unlock') {
    handleUnlock(row)
    return
  }
  if (command === 'delete') {
    ElMessageBox.confirm('确认删除该用户？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    }).then(() => {
      handleDelete(row.id)
    }).catch(() => {})
  }
}

// ==================== 重置密码 ====================
const pwdDialogVisible = ref(false)
const pwdSubmitting = ref(false)
const pwdForm = reactive({ id: 0, password: '' })
const pwdFormRef = ref()
const pwdRules: FormRules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 32, message: '密码长度 6-32 位', trigger: 'blur' },
  ],
}

function handleResetPwd(row: UserItem) {
  pwdForm.id = row.id
  pwdForm.password = ''
  pwdDialogVisible.value = true
}

async function handlePwdSubmit() {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return
  pwdSubmitting.value = true
  try {
    await updateUserPassword({ id: pwdForm.id, password: pwdForm.password })
    ElMessage.success('密码重置成功')
    pwdDialogVisible.value = false
  } catch (err: any) {
    showRequestError(err, '重置失败')
  } finally {
    pwdSubmitting.value = false
  }
}

const importing = ref(false)

async function handleExport() {
  try {
    await exportUsers({
      username: query.username,
      nickname: query.nickname,
      phone: query.phone,
      status: query.status,
      deptId: query.deptId,
    })
    ElMessage.success('导出成功')
  } catch (err: any) {
    showRequestError(err, '导出失败')
  }
}

async function handleDownloadTemplate() {
  try {
    await downloadUserImportTemplate()
  } catch (err: any) {
    showRequestError(err, '下载模板失败')
  }
}

async function handleImport(opt: UploadRequestOptions) {
  importing.value = true
  try {
    const res = await importUsers(opt.file as File)
    const errs = res.data?.errors || []
    if (errs.length) {
      ElMessageBox.alert(errs.slice(0, 20).join('<br/>'), `导入完成：成功 ${res.data?.success || 0} 条`, {
        dangerouslyUseHTMLString: true,
        type: 'warning',
      })
    } else {
      ElMessage.success(`导入成功 ${res.data?.success || 0} 条`)
    }
    loadData(query)
    opt.onSuccess?.(res as any)
  } catch (err: any) {
    showRequestError(err, '导入失败')
    opt.onError?.(err)
  } finally {
    importing.value = false
  }
}

async function handleUnlock(row: UserItem) {
  try {
    await unlockUser(row.id)
    ElMessage.success(`已解锁用户 ${row.username}`)
  } catch (err: any) {
    showRequestError(err, '解锁失败')
  }
}

onMounted(() => {
  loadData(query)
  loadOptions()
})
</script>

<style scoped lang="scss">
.page-container {
  padding: 0;
}

.avatar-picker {
  display: flex;
  align-items: center;
}

.avatar-wrap {
  position: relative;
  width: 64px;
  height: 64px;
  cursor: pointer;
  border-radius: 6px;

  &:hover {
    opacity: 0.9;
  }

  :deep(.el-avatar) {
    display: flex;
    align-items: center;
    justify-content: center;

    .el-icon {
      display: block;
      margin: 0;
      font-size: 36px;
    }

    svg {
      display: block;
      margin: 0;
      width: 1em;
      height: 1em;
    }
  }
}

.avatar-remove {
  position: absolute;
  top: -4px;
  right: -4px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #f56c6c;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 0 0 1px #fff;
  z-index: 1;

  &:hover {
    background: #f78989;
  }
}
</style>
