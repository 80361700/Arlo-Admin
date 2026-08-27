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
        <el-button v-permission="'sys:member:add'" type="primary" @click="handleAdd">新增会员</el-button>
      </template>

      <el-table-column prop="id" label="ID" width="80" align="center" />
      <el-table-column prop="phone" label="手机号" min-width="130" />
      <el-table-column prop="nickname" label="昵称" min-width="120" show-overflow-tooltip />
      <el-table-column label="头像" width="72" align="center">
        <template #default="{ row }">
          <AuthAvatar :file-ref="row.avatar" :size="32" />
        </template>
      </el-table-column>
      <el-table-column label="性别" width="80" align="center">
        <template #default="{ row }">
          {{ genderLabel(row.gender) }}
        </template>
      </el-table-column>
      <el-table-column label="来源" width="90" align="center">
        <template #default="{ row }">
          {{ sourceLabel(row.source) }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-switch
            v-if="authStore.hasPermission('sys:member:edit')"
            :model-value="row.status === 1"
            inline-prompt
            active-text="启"
            inactive-text="禁"
            @change="(val: string | number | boolean) => handleToggleStatus(row as MemberItem, Boolean(val))"
          />
          <el-tag v-else :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ statusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="lastLogin" label="最后登录" min-width="170">
        <template #default="{ row }">
          {{ row.lastLogin || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="注册时间" min-width="170" />

      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="handleDetail(row)">详情</el-button>
        <el-button
          v-if="authStore.hasPermission('sys:member:edit')"
          type="primary"
          link
          size="small"
          @click="handleEdit(row)"
        >
          编辑
        </el-button>
        <el-dropdown
          v-if="authStore.hasPermission('sys:member:edit') || authStore.hasPermission('sys:member:delete')"
          trigger="click"
          @command="(cmd: string) => handleAction(row, cmd)"
        >
          <el-button type="info" link size="small">
            更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-if="authStore.hasPermission('sys:member:edit')" command="resetPwd">
                重置密码
              </el-dropdown-item>
              <el-dropdown-item
                v-if="authStore.hasPermission('sys:member:delete')"
                command="delete"
                divided
              >
                删除
              </el-dropdown-item>
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
        <el-form-item label="手机号" prop="phone">
          <el-input
            v-model="form.phone"
            placeholder="请输入手机号"
            maxlength="11"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="留空则使用系统默认密码"
            show-password
            maxlength="32"
          />
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
        <el-form-item label="来源" prop="source">
          <el-select v-model="form.source" placeholder="请选择来源" style="width: 100%">
            <el-option
              v-for="opt in sourceOptions"
              :key="String(opt.value)"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio v-for="opt in statusOptions" :key="String(opt.value)" :value="opt.value">
              {{ opt.label }}
            </el-radio>
          </el-radio-group>
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

    <!-- 详情抽屉 -->
    <el-drawer v-model="detailVisible" title="会员详情" size="420px">
      <el-descriptions v-if="detail" :column="1" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ detail.phone }}</el-descriptions-item>
        <el-descriptions-item label="昵称">{{ detail.nickname || '-' }}</el-descriptions-item>
        <el-descriptions-item label="头像">
          <AuthAvatar :file-ref="detail.avatar" :size="40" />
        </el-descriptions-item>
        <el-descriptions-item label="性别">{{ genderLabel(detail.gender) }}</el-descriptions-item>
        <el-descriptions-item label="来源">{{ sourceLabel(detail.source) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detail.status === 1 ? 'success' : 'danger'" size="small">
            {{ statusLabel(detail.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="OpenID">{{ detail.openid || '-' }}</el-descriptions-item>
        <el-descriptions-item label="UnionID">{{ detail.unionid || '-' }}</el-descriptions-item>
        <el-descriptions-item label="公众号 OpenID">{{ detail.mpOpenid || '-' }}</el-descriptions-item>
        <el-descriptions-item label="最后登录">{{ detail.lastLogin || '-' }}</el-descriptions-item>
        <el-descriptions-item label="注册时间">{{ detail.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ detail.updatedAt }}</el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormRules } from 'element-plus'
import { ArrowDown, Close } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import FilePicker from '@/components/FilePicker.vue'
import AuthAvatar from '@/components/AuthAvatar.vue'
import {
  getMemberList,
  getMemberDetail,
  createMember,
  updateMember,
  updateMemberPassword,
  updateMemberStatus,
  deleteMember,
} from '@/api'
import type { MemberItem, MemberDetail, MemberListParams, FileItem } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { toFileRef } from '@/utils/fileUrl'
import { useDict, DictCode } from '@/utils/useDict'

const authStore = useAuthStore()
const { options: genderOptions, getLabel: genderLabel } = useDict(DictCode.Gender)
const { options: statusOptions, getLabel: statusLabel } = useDict(DictCode.UserStatus)
const { options: sourceOptions, getLabel: sourceLabel } = useDict(DictCode.MemberSource)

const proTableRef = ref()
const tableData = ref<MemberItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive<MemberListParams>({ page: 1, pageSize: 10 })

const searchFields = computed(() => [
  { prop: 'phone', label: '手机号' },
  { prop: 'nickname', label: '昵称' },
  {
    prop: 'source',
    label: '来源',
    type: 'select' as const,
    options: sourceOptions.value,
  },
  {
    prop: 'status',
    label: '状态',
    type: 'select' as const,
    options: statusOptions.value,
  },
])

async function loadData(params: MemberListParams) {
  loading.value = true
  try {
    const res = await getMemberList(params)
    tableData.value = res.data.list || []
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) {
  query.page = 1
  Object.assign(query, p)
  loadData(query)
}
function handleReset() {
  Object.assign(query, { page: 1, pageSize: 10, phone: undefined, nickname: undefined, source: undefined, status: undefined })
  loadData(query)
}
function handlePageChange(p: any) {
  Object.assign(query, p)
  loadData(query)
}

async function handleToggleStatus(row: MemberItem, enabled: boolean) {
  const next = enabled ? 1 : 0
  try {
    await updateMemberStatus(row.id, next)
    row.status = next
    ElMessage.success(next === 1 ? '已启用' : '已禁用')
  } catch {
    /* 拦截器已提示 */
  }
}

// ==================== 新增/编辑 ====================
const dialogVisible = ref(false)
const dialogTitle = ref('新增会员')
const isEdit = ref(false)
const submitting = ref(false)
const formDialogRef = ref()
const pickerVisible = ref(false)

const defaultForm = {
  id: 0,
  phone: '',
  password: '',
  nickname: '',
  avatar: '',
  gender: 0,
  source: 'h5',
  status: 1,
}
const form = reactive({ ...defaultForm })

const formRules: FormRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { len: 11, message: '手机号须为 11 位', trigger: 'blur' },
  ],
  password: [
    {
      validator: (_rule, value, callback) => {
        if (!value) return callback()
        if (value.length < 6 || value.length > 32) {
          return callback(new Error('密码长度 6-32 位'))
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  source: [{ required: true, message: '请选择来源', trigger: 'change' }],
}

function onAvatarPicked(files: FileItem[]) {
  if (!files.length) return
  form.avatar = toFileRef(files[0])
}

function handleAdd() {
  isEdit.value = false
  dialogTitle.value = '新增会员'
  Object.assign(form, defaultForm)
  dialogVisible.value = true
}

async function handleEdit(row: MemberItem) {
  isEdit.value = true
  dialogTitle.value = '编辑会员'
  try {
    const res = await getMemberDetail(row.id)
    const d = res.data
    Object.assign(form, {
      id: d.id,
      phone: d.phone,
      password: '',
      nickname: d.nickname,
      avatar: d.avatar,
      gender: d.gender,
      source: d.source || 'h5',
      status: d.status,
    })
    dialogVisible.value = true
  } catch {
    /* 拦截器已提示 */
  }
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateMember({
        id: form.id,
        nickname: form.nickname,
        avatar: form.avatar,
        gender: form.gender,
        source: form.source,
        status: form.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createMember({
        phone: form.phone,
        password: form.password,
        nickname: form.nickname,
        avatar: form.avatar,
        gender: form.gender,
        source: form.source,
        status: form.status,
      })
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    loadData(query)
  } catch {
    /* 拦截器已提示 */
  } finally {
    submitting.value = false
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

function handleResetPwd(row: MemberItem) {
  pwdForm.id = row.id
  pwdForm.password = ''
  pwdDialogVisible.value = true
}

function handleAction(row: MemberItem, command: string) {
  if (command === 'resetPwd') {
    handleResetPwd(row)
    return
  }
  if (command === 'delete') {
    ElMessageBox.confirm('确认删除该会员？', '提示', { type: 'warning' })
      .then(() => handleDelete(row.id))
      .catch(() => {})
  }
}

async function handlePwdSubmit() {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return
  pwdSubmitting.value = true
  try {
    await updateMemberPassword({ id: pwdForm.id, password: pwdForm.password })
    ElMessage.success('密码重置成功')
    pwdDialogVisible.value = false
  } catch {
    /* 拦截器已提示 */
  } finally {
    pwdSubmitting.value = false
  }
}

// ==================== 详情 ====================
const detailVisible = ref(false)
const detail = ref<MemberDetail | null>(null)

async function handleDetail(row: MemberItem) {
  try {
    const res = await getMemberDetail(row.id)
    detail.value = res.data
    detailVisible.value = true
  } catch {
    /* 拦截器已提示 */
  }
}

async function handleDelete(id: number) {
  try {
    await deleteMember(id)
    ElMessage.success('删除成功')
    loadData(query)
  } catch {
    /* 拦截器已提示 */
  }
}

onMounted(() => loadData(query))
</script>

<style scoped lang="scss">
.avatar-picker {
  display: flex;
  align-items: center;
}

.avatar-wrap {
  position: relative;
  width: 64px;
  height: 64px;
  cursor: pointer;
  border-radius: 50%;

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
    background: #e64545;
  }
}
</style>
