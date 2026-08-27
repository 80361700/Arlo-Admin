<template>
  <el-dialog
    v-model="visible"
    :title="force ? (forceReason || '请修改密码') : '个人信息'"
    width="560px"
    :close-on-click-modal="false"
    :close-on-press-escape="!force"
    :show-close="!force"
    destroy-on-close
    @closed="handleClosed"
  >
    <el-alert
      v-if="force"
      :title="forceReason || '当前账号需要修改密码后才能继续使用'"
      type="warning"
      :closable="false"
      show-icon
      style="margin-bottom: 16px"
    />
    <el-tabs v-if="!force" v-model="activeTab">
      <el-tab-pane label="个人信息" name="info">
        <el-form
          ref="profileFormRef"
          :model="profileForm"
          :rules="profileRules"
          label-width="80px"
          v-loading="loading"
        >
          <el-form-item label="用户名">
            <el-input :model-value="authStore.userInfo?.username" disabled />
          </el-form-item>
          <el-form-item label="昵称" prop="nickname">
            <el-input v-model="profileForm.nickname" placeholder="请输入昵称" maxlength="32" />
          </el-form-item>
          <el-form-item label="头像">
            <div class="avatar-picker">
              <div class="avatar-wrap" @click="pickerVisible = true">
                <AuthAvatar :file-ref="profileForm.avatar" :size="64" shape="square" />
                <span
                  v-if="profileForm.avatar"
                  class="avatar-remove"
                  title="清除头像"
                  @click.stop="profileForm.avatar = ''"
                >
                  <el-icon :size="12"><Close /></el-icon>
                </span>
              </div>
            </div>
          </el-form-item>
          <el-form-item label="性别">
            <el-radio-group v-model="profileForm.gender">
              <el-radio v-for="opt in genderOptions" :key="String(opt.value)" :value="opt.value">
                {{ opt.label }}
              </el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="部门">
            <el-input :model-value="authStore.userInfo?.deptName || '-'" disabled />
          </el-form-item>
          <el-form-item label="手机号" prop="phone">
            <el-input v-model="profileForm.phone" placeholder="请输入手机号" maxlength="20" />
          </el-form-item>
          <el-form-item label="邮箱" prop="email">
            <el-input v-model="profileForm.email" placeholder="请输入邮箱" maxlength="64" />
          </el-form-item>
          <el-form-item label="状态">
            <el-tag :type="authStore.userInfo?.status === 1 ? 'success' : 'danger'" size="small">
              {{ statusLabel(authStore.userInfo?.status) }}
            </el-tag>
          </el-form-item>
          <el-form-item label="角色">
            <template v-if="(authStore.userInfo?.roleNames || []).length">
              <el-tag
                v-for="name in authStore.userInfo?.roleNames"
                :key="name"
                size="small"
                class="info-tag"
              >
                {{ name }}
              </el-tag>
            </template>
            <span v-else class="text-muted">-</span>
          </el-form-item>
          <el-form-item label="岗位">
            <template v-if="(authStore.userInfo?.postNames || []).length">
              <el-tag
                v-for="name in authStore.userInfo?.postNames"
                :key="name"
                type="info"
                size="small"
                class="info-tag"
              >
                {{ name }}
              </el-tag>
            </template>
            <span v-else class="text-muted">-</span>
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="profileForm.remark" type="textarea" placeholder="请输入备注" :rows="2" />
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="修改密码" name="password">
        <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="90px">
          <el-form-item label="原密码" prop="oldPassword">
            <el-input v-model="pwdForm.oldPassword" type="password" placeholder="请输入原密码" show-password maxlength="32" />
          </el-form-item>
          <el-form-item label="新密码" prop="newPassword">
            <el-input v-model="pwdForm.newPassword" type="password" placeholder="请输入新密码" show-password maxlength="32" />
          </el-form-item>
          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input v-model="pwdForm.confirmPassword" type="password" placeholder="请输入确认密码" show-password maxlength="32" />
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>

    <el-form v-else ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="90px">
      <el-form-item label="原密码" prop="oldPassword">
        <el-input v-model="pwdForm.oldPassword" type="password" placeholder="请输入原密码" show-password maxlength="32" />
      </el-form-item>
      <el-form-item label="新密码" prop="newPassword">
        <el-input v-model="pwdForm.newPassword" type="password" placeholder="请输入新密码" show-password maxlength="32" />
      </el-form-item>
      <el-form-item label="确认密码" prop="confirmPassword">
        <el-input v-model="pwdForm.confirmPassword" type="password" placeholder="请输入确认密码" show-password maxlength="32" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button v-if="!force" @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
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
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Close } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { updateProfileApi, changePasswordApi } from '@/api/modules/auth'
import type { FileItem } from '@/api/modules/file'
import FilePicker from '@/components/FilePicker.vue'
import AuthAvatar from '@/components/AuthAvatar.vue'
import { toFileRef } from '@/utils/fileUrl'
import { useDict, DictCode } from '@/utils/useDict'

const visible = defineModel<boolean>({ default: false })
const props = withDefaults(defineProps<{ force?: boolean }>(), { force: false })

const authStore = useAuthStore()
const { options: genderOptions } = useDict(DictCode.Gender)
const { getLabel: statusLabel } = useDict(DictCode.UserStatus)
const activeTab = ref('info')
const loading = ref(false)
const submitting = ref(false)
const profileFormRef = ref<FormInstance>()
const pwdFormRef = ref<FormInstance>()

const forceReason = computed(() => {
  if (authStore.userInfo?.mustChangePwd) return '管理员已重置密码，请先修改后再继续使用'
  if (authStore.userInfo?.pwdExpired) return '密码已过期，请先修改后再继续使用'
  return ''
})

const pickerVisible = ref(false)

const profileForm = reactive({
  nickname: '',
  avatar: '',
  gender: 0,
  phone: '',
  email: '',
  remark: '',
})

function onAvatarPicked(files: FileItem[]) {
  if (!files.length) return
  profileForm.avatar = toFileRef(files[0])
}

const profileRules: FormRules = {
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  phone: [{
    validator: (_rule, value, callback) => {
      if (!value) return callback()
      if (!/^1\d{10}$/.test(value)) return callback(new Error('手机号格式不正确'))
      callback()
    },
    trigger: 'blur',
  }],
  email: [{
    validator: (_rule, value, callback) => {
      if (!value) return callback()
      if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) return callback(new Error('邮箱格式不正确'))
      callback()
    },
    trigger: 'blur',
  }],
}

const pwdForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const pwdRules: FormRules = {
  oldPassword: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 32, message: '密码长度为 6-32 位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请输入确认密码', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value !== pwdForm.newPassword) callback(new Error('两次输入的密码不一致'))
        else callback()
      },
      trigger: 'blur',
    },
  ],
}

function fillProfileForm() {
  const info = authStore.userInfo
  if (!info) return
  profileForm.nickname = info.nickname || ''
  profileForm.avatar = info.avatar || ''
  profileForm.gender = info.gender ?? 0
  profileForm.phone = info.phone || ''
  profileForm.email = info.email || ''
  profileForm.remark = info.remark || ''
}

function resetPwdForm() {
  pwdForm.oldPassword = ''
  pwdForm.newPassword = ''
  pwdForm.confirmPassword = ''
  pwdFormRef.value?.clearValidate()
}

watch(visible, async (val) => {
  if (!val) return
  activeTab.value = props.force ? 'password' : 'info'
  resetPwdForm()
  if (props.force) return
  loading.value = true
  try {
    await authStore.fetchUserInfo()
    fillProfileForm()
  } finally {
    loading.value = false
  }
})

async function handleSubmit() {
  if (!props.force && activeTab.value === 'info') {
    const valid = await profileFormRef.value?.validate().catch(() => false)
    if (!valid) return
    submitting.value = true
    try {
      await updateProfileApi({
        nickname: profileForm.nickname,
        avatar: profileForm.avatar,
        gender: profileForm.gender,
        phone: profileForm.phone,
        email: profileForm.email,
        remark: profileForm.remark,
      })
      ElMessage.success('保存成功')
      await authStore.fetchUserInfo()
      visible.value = false
    } catch {
      /* interceptor */
    } finally {
      submitting.value = false
    }
    return
  }

  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    await changePasswordApi({
      oldPassword: pwdForm.oldPassword,
      newPassword: pwdForm.newPassword,
    })
    ElMessage.success('密码修改成功，请重新登录')
    visible.value = false
    await authStore.logout()
    window.location.hash = '#/login'
    window.location.reload()
  } catch {
    /* interceptor */
  } finally {
    submitting.value = false
  }
}

function handleClosed() {
  activeTab.value = 'info'
  resetPwdForm()
  profileFormRef.value?.clearValidate()
}
</script>

<style scoped lang="scss">
.info-tag {
  margin-right: 6px;
}
.text-muted {
  color: #909399;
  font-size: 13px;
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
}
</style>
