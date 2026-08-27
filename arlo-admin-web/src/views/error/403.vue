<template>
  <div class="error-page">
    <h1>403</h1>
    <p>没有访问权限</p>
    <div class="actions">
      <el-button type="primary" @click="$router.push('/dashboard')">返回首页</el-button>
      <el-button @click="handleRefresh">刷新权限</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

async function handleRefresh() {
  try {
    await authStore.refreshPermissions()
    ElMessage.success('权限已刷新')
    router.replace('/dashboard')
  } catch {
    /* 拦截器已提示 */
  }
}
</script>

<style scoped lang="scss">
.error-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  h1 {
    font-size: 80px;
    color: #dcdfe6;
  }
  p {
    font-size: 18px;
    color: #909399;
    margin: 16px 0 32px;
  }
  .actions {
    display: flex;
    gap: 12px;
  }
}
</style>
