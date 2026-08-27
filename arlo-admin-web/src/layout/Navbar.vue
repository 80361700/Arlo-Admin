<template>
  <div class="navbar">
    <div class="navbar-left">
      <!-- 混合 / 顶栏：完整 Logo（点回首页） -->
      <div
        v-if="appStore.useHeaderBrand"
        class="nav-brand"
        :class="{
          'is-collapsed': appStore.isMixLayout && appStore.sidebarCollapsed,
          'is-plain': appStore.isTopbarLayout,
        }"
        title="首页"
        @click="goHome"
      >
        <img :src="logoDisplay" alt="logo" class="nav-brand-img" />
        <span
          v-show="!(appStore.isMixLayout && appStore.sidebarCollapsed)"
          class="nav-brand-title"
        >
          {{ appStore.systemName }}
        </span>
      </div>

      <el-icon
        v-if="!appStore.isTopbarLayout"
        class="collapse-btn"
        @click="appStore.toggleSidebar"
      >
        <Fold v-if="!appStore.sidebarCollapsed" />
        <Expand v-else />
      </el-icon>

      <!-- 侧栏：面包屑（开启页签时不显示，避免与 Tags 重复） -->
      <AppBreadcrumb v-if="appStore.isSideLayout && !appStore.tagsView" />

      <!-- 混合：顶栏一级菜单（仅 mix） -->
      <el-menu
        v-if="appStore.isMixLayout"
        :key="`nav-mix-${appStore.themeId}-${topMenus.map((m) => m.id).join('-')}`"
        mode="horizontal"
        :ellipsis="false"
        :default-active="mixActive"
        class="nav-top-menu"
        @select="onMixSelect"
      >
        <el-menu-item
          v-for="item in topMenus"
          :key="item.id"
          :index="String(item.id)"
        >
          <el-icon v-if="iconOf(item)">
            <component :is="iconOf(item)" />
          </el-icon>
          <span>{{ item.name }}</span>
        </el-menu-item>
      </el-menu>

      <!-- 顶栏：完整菜单树（仅 topbar） -->
      <el-menu
        v-if="appStore.isTopbarLayout"
        :key="`nav-topbar-${appStore.themeId}-${topMenus.map((m) => m.id).join('-')}`"
        mode="horizontal"
        :ellipsis="true"
        :default-active="route.path"
        router
        menu-trigger="hover"
        class="nav-top-menu nav-topbar-menu"
      >
        <SidebarItem
          v-for="item in topMenus"
          :key="item.id"
          :item="item"
        />
      </el-menu>
    </div>

    <div class="navbar-right">
      <div class="btn-icon">
        <el-popover
          placement="bottom"
          trigger="hover"
          :width="260"
          :show-after="120"
          :hide-after="200"
          :offset="8"
        >
          <template #reference>
            <div class="msg-bell">
              <el-icon :size="20"><Brush /></el-icon>
            </div>
          </template>
          <AppearanceDialog />
        </el-popover>

        <el-tooltip content="消息" placement="bottom">
          <div class="msg-bell" @click="goMessages">
            <el-badge :is-dot="messageStore.unreadCount > 0">
              <el-icon :size="20"><Bell /></el-icon>
            </el-badge>
          </div>
        </el-tooltip>
      </div>

      <el-dropdown trigger="hover" @command="handleCommand">
        <div class="user-info">
          <el-avatar :size="28" :src="avatarSrc || undefined" :icon="UserFilled" />
          <span class="username">{{ authStore.userInfo?.nickname || '用户' }}</span>
          <el-icon class="arrow"><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="home">首页</el-dropdown-item>
            <el-dropdown-item command="profile">个人信息</el-dropdown-item>
            <el-dropdown-item command="refreshPerm">刷新权限</el-dropdown-item>
            <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <ProfileDialog v-model="profileVisible" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as Icons from '@element-plus/icons-vue'
import { ArrowDown, UserFilled, Bell, Brush } from '@element-plus/icons-vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useMessageStore } from '@/stores/message'
import { resetDynamicRoutes } from '@/router'
import { useAuthFileSrc } from '@/composables/useAuthFileSrc'
import {
  getTopLevelMenus,
  resolveTopMenu,
  firstLeafPath,
} from '@/utils/navLayout'
import ProfileDialog from '@/components/ProfileDialog.vue'
import AppearanceDialog from '@/components/AppearanceDialog.vue'
import AppBreadcrumb from '@/components/AppBreadcrumb.vue'
import SidebarItem from './SidebarItem.vue'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const messageStore = useMessageStore()
const profileVisible = ref(false)

const avatarSrc = useAuthFileSrc(() => authStore.userInfo?.avatar || '')
const logoSrc = useAuthFileSrc(() => appStore.systemLogo || '')
const logoDisplay = computed(() => logoSrc.value || '/vite.svg')

const topMenus = computed(() => getTopLevelMenus(authStore.menus))
const mixActive = computed(() => {
  const hit =
    topMenus.value.find((m) => m.id === appStore.activeTopMenuId)
    || resolveTopMenu(authStore.menus, route.path)
  return hit ? String(hit.id) : ''
})

function iconOf(item: { icon?: string }) {
  if (!item.icon) return null
  return (Icons as Record<string, unknown>)[item.icon] || null
}

function onMixSelect(index: string) {
  const id = Number(index)
  const menu = topMenus.value.find((m) => m.id === id)
  if (!menu) return
  appStore.setActiveTopMenuId(id)
  const leaf = firstLeafPath(menu)
  if (leaf && leaf !== route.path) {
    router.push(leaf)
  }
}

function goHome() {
  router.push('/dashboard')
}

function goMessages() {
  router.push('/message/my')
}

onMounted(() => {
  // 强制改密期间不拉未读数，避免无权限接口干扰改密流程
  const forcePwd = !!(authStore.userInfo?.mustChangePwd || authStore.userInfo?.pwdExpired)
  if (authStore.accessToken && !forcePwd) {
    messageStore.startPolling()
  }
})

onBeforeUnmount(() => {
  messageStore.stopPolling()
})

async function handleCommand(cmd: string) {
  if (cmd === 'home') {
    goHome()
  } else if (cmd === 'profile') {
    profileVisible.value = true
  } else if (cmd === 'refreshPerm') {
    try {
      const path = router.currentRoute.value.fullPath
      await authStore.refreshPermissions()
      ElMessage.success('权限已刷新')
      const resolved = router.resolve(path)
      const stillValid = resolved.matched.some(
        (m) => m.name && m.name !== 'Layout' && m.name !== 'CatchAll',
      )
      if (!stillValid && path !== '/dashboard' && !path.startsWith('/dashboard')) {
        router.replace('/dashboard')
      } else {
        router.replace(path)
      }
    } catch {
      /* 拦截器已提示 */
    }
  } else if (cmd === 'logout') {
    messageStore.clear()
    await authStore.logout()
    resetDynamicRoutes()
    window.location.hash = '#/login'
    window.location.reload()
  }
}
</script>

<style scoped lang="scss">
.navbar {
  height: 100%;
  display: flex;
  align-items: stretch;
  justify-content: space-between;
}

.navbar-left {
  display: flex;
  align-items: center;
  min-width: 0;
  flex: 1;
  margin-right: 16px;

  .nav-brand {
    display: flex;
    align-items: center;
    justify-content: center;
    align-self: stretch;
    flex-shrink: 0;
    width: 220px;
    height: auto;
    margin: 0;
    cursor: pointer;
    overflow: hidden;
    box-sizing: border-box;
    transition: width 0.28s ease;
    background: var(--arlo-nav-brand-bg, transparent);
    position: relative;
    z-index: 2;
    margin-bottom: -1px;
    border-bottom: 1px solid var(--arlo-nav-brand-line, #e6e6e6);
    box-shadow: var(--arlo-nav-brand-shadow, -2px 1px 4px rgba(0, 0, 0, 0.12));

    &.is-collapsed {
      width: 64px;
    }

    /* 纯顶栏：跟白顶栏一体，无侧栏连续列样式 */
    &.is-plain {
      margin-bottom: 0;
      border-bottom: none;
      box-shadow: none;
      background: transparent;
    }

    .nav-brand-img {
      width: 28px;
      height: 28px;
      flex-shrink: 0;
    }

    .nav-brand-title {
      margin-left: 10px;
      font-size: 16px;
      font-weight: 600;
      color: var(--arlo-nav-brand-title, #303133);
      white-space: nowrap;
    }

    &.is-plain .nav-brand-title {
      color: var(--arlo-nav-brand-title, #303133);
    }
  }

  .collapse-btn {
    font-size: 18px;
    cursor: pointer;
    color: #666;
    flex-shrink: 0;

    &:hover {
      color: #409eff;
    }
  }

  .nav-brand + .collapse-btn {
    margin-left: 16px;
    margin-right: 8px;
  }
}

.nav-top-menu {
  flex: 1;
  min-width: 0;
  height: 50px;
  border-bottom: none !important;
  background: transparent !important;

  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    height: 50px;
    line-height: 50px;
    padding: 0 12px !important;
    margin: 0 !important;
    border-bottom: none !important;
    color: #303133;
    background: transparent !important;
  }

  :deep(.el-menu-item.is-active),
  :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
    color: #409eff !important;
    border-bottom: none !important;
    background: transparent !important;
  }

  :deep(.el-menu-item:hover),
  :deep(.el-menu-item:focus),
  :deep(.el-menu-item.is-active:hover),
  :deep(.el-sub-menu__title:hover) {
    background: transparent !important;
    border-bottom: none !important;
    color: #409eff !important;
  }
}

/* 仅纯顶栏菜单跟主题变量走，避免影响混合模式一级菜单 */
.nav-topbar-menu {
  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    color: var(--arlo-topbar-nav-text, #303133);
  }

  :deep(.el-menu-item.is-active),
  :deep(.el-sub-menu.is-active > .el-sub-menu__title),
  :deep(.el-menu-item:hover),
  :deep(.el-menu-item:focus),
  :deep(.el-menu-item.is-active:hover),
  :deep(.el-sub-menu__title:hover) {
    color: var(--arlo-topbar-nav-active, #409eff) !important;
  }

  :deep(.el-sub-menu .el-menu) {
    min-width: 160px;
  }
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
  align-self: center;

  :deep(.el-badge) {
    display: flex;
  }

  .btn-icon {
    display: flex;
    gap: 12px;
    padding-right: 10px;
  }

  .msg-bell {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 32px;
    cursor: pointer;
    color: #666;
    outline: none;

    &:hover {
      color: #409eff;
    }

    :deep(.el-badge__content) {
      transform: translateY(-2px) translateX(8px);
    }
  }

  .user-info {
    display: flex;
    align-items: center;
    cursor: pointer;
    gap: 8px;
    outline: none;

    .username {
      font-size: 14px;
      color: #333;
    }

    .arrow {
      font-size: 12px;
      color: #909399;
    }
  }
}
</style>
