<template>
  <div class="login-container">
    <div class="login-card">
      <div v-if="logoUrl" class="login-logo">
        <img :src="logoUrl" alt="logo" />
      </div>
      <h2 class="login-title">{{ appStore.systemName }}</h2>
      <p class="login-subtitle">后台管理系统</p>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        size="large"
        class="login-form"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item v-if="appStore.captchaEnabled" prop="captchaCode">
          <div class="captcha-row">
            <el-input
              v-model="form.captchaCode"
              placeholder="验证码"
              prefix-icon="Key"
              class="captcha-input"
              @keyup.enter="handleLogin"
            />
            <img
              :src="captchaImg"
              alt="验证码"
              class="captcha-img"
              title="点击刷新验证码"
              @click="refreshCaptcha"
            />
          </div>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            class="login-btn"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '登 录' }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import type { FormInstance, FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { getCaptchaApi } from '@/api'
import { resolveFileSrc } from '@/utils/fileUrl'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const captchaImg = ref('')
const captchaId = ref('')

const form = reactive({
  username: 'admin',
  password: 'admin123',
  captchaCode: '',
})

const rules = computed<FormRules>(() => {
  const base: FormRules = {
    username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
    password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  }
  if (appStore.captchaEnabled) {
    base.captchaCode = [{ required: true, message: '请输入验证码', trigger: 'blur' }]
  }
  return base
})

/** 登录页 Logo：仅公开文件可在未登录时展示 */
const logoUrl = computed(() => resolveFileSrc(appStore.systemLogo))

async function refreshCaptcha() {
  if (!appStore.captchaEnabled) return
  try {
    const res = await getCaptchaApi()
    captchaId.value = res.data.captchaId
    captchaImg.value = res.data.captcha
    form.captchaCode = ''
  } catch {
    // 静默失败，登录时校验
  }
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await authStore.login({
      username: form.username,
      password: form.password,
      captchaId: appStore.captchaEnabled ? captchaId.value : '',
      captchaCode: appStore.captchaEnabled ? form.captchaCode : '',
    })
    await authStore.fetchUserInfo()
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (err: any) {
    showRequestError(err, '登录失败')
    refreshCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await appStore.loadPublicConfig()
  document.title = appStore.systemName
  if (appStore.captchaEnabled) {
    refreshCaptcha()
  }
})
</script>

<style scoped lang="scss">
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}

.login-logo {
  display: flex;
  justify-content: center;
  margin-bottom: 12px;

  img {
    height: 48px;
    max-width: 160px;
    object-fit: contain;
  }
}

.login-title {
  text-align: center;
  font-size: 24px;
  color: #303133;
  margin-bottom: 8px;
}

.login-subtitle {
  text-align: center;
  font-size: 14px;
  color: #909399;
  margin-bottom: 32px;
}

.login-form {
  .login-btn {
    width: 100%;
  }
}

.captcha-row {
  display: flex;
  gap: 12px;
  align-items: center;

  .captcha-input {
    flex: 1;
  }

  .captcha-img {
    height: 40px;
    width: 120px;
    border-radius: 4px;
    cursor: pointer;
    border: 1px solid #dcdfe6;
    object-fit: contain;
  }
}
</style>
