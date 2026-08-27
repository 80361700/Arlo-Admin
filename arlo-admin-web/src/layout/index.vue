<template>
  <el-container
    class="layout-container"
    :class="[appStore.themeClass, `layout-${appStore.layoutMode}`]"
  >
    <!-- 侧栏模式：左侧 Logo + 完整菜单 -->
    <el-aside
      v-if="appStore.isSideLayout"
      :width="appStore.sidebarWidth"
      class="layout-sidebar"
    >
      <Sidebar />
    </el-aside>

    <el-container direction="vertical" class="layout-header-shell">
      <el-header height="50px" class="layout-navbar">
        <Navbar />
      </el-header>

      <!-- 混合：页签仅在右侧列，且在滚动区外固定 -->
      <el-container v-if="appStore.isMixLayout" class="layout-body">
        <el-aside :width="appStore.sidebarWidth" class="layout-sidebar">
          <Sidebar :menus="sidebarMenus" :show-logo="false" />
        </el-aside>
        <div class="layout-content-col">
          <TagsView v-if="appStore.tagsView" />
          <el-main class="layout-main">
            <router-view v-slot="{ Component, route: r }">
              <transition name="fade" mode="out-in">
                <keep-alive v-if="appStore.tagsView" :include="tagsStore.cachedNames">
                  <component :is="Component" :key="String(r.name || r.path)" />
                </keep-alive>
                <component v-else :is="Component" :key="r.fullPath" />
              </transition>
            </router-view>
          </el-main>
        </div>
      </el-container>

      <!-- 侧栏 / 顶栏：页签在内容区顶，不随 main 滚动 -->
      <div v-else class="layout-content-col">
        <TagsView v-if="appStore.tagsView" />
        <el-main class="layout-main">
          <div v-if="appStore.isTopbarLayout && !isHomePage && !appStore.tagsView" class="layout-crumb">
            <AppBreadcrumb />
          </div>
          <router-view v-slot="{ Component, route: r }">
            <transition name="fade" mode="out-in">
              <keep-alive v-if="appStore.tagsView" :include="tagsStore.cachedNames">
                <component :is="Component" :key="String(r.name || r.path)" />
              </keep-alive>
              <component v-else :is="Component" :key="r.fullPath" />
            </transition>
          </router-view>
        </el-main>
      </div>
    </el-container>

    <ProfileDialog v-model="forcePwdVisible" force />
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useTagsStore } from '@/stores/tags'
import {
  resolveTopMenu,
  getSideMenusForTop,
} from '@/utils/navLayout'
import Sidebar from './Sidebar.vue'
import Navbar from './Navbar.vue'
import TagsView from './TagsView.vue'
import ProfileDialog from '@/components/ProfileDialog.vue'
import AppBreadcrumb from '@/components/AppBreadcrumb.vue'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const tagsStore = useTagsStore()
const forcePwdVisible = ref(false)

const isHomePage = computed(() => {
  const p = route.path.replace(/\/+$/, '') || '/'
  return p === '/' || p === '/dashboard'
})

const activeTopNode = computed(() => {
  if (!appStore.isMixLayout) return null
  return (
    authStore.menus.find((m) => m.id === appStore.activeTopMenuId)
    || resolveTopMenu(authStore.menus, route.path)
  )
})

const sidebarMenus = computed(() => {
  if (!appStore.isMixLayout) return undefined
  return getSideMenusForTop(activeTopNode.value)
})

watch(
  () => [authStore.userInfo?.mustChangePwd, authStore.userInfo?.pwdExpired, forcePwdVisible.value],
  () => {
    const need = !!(authStore.userInfo?.mustChangePwd || authStore.userInfo?.pwdExpired)
    if (need) forcePwdVisible.value = true
  },
  { immediate: true },
)

watch(
  () => [route.path, authStore.menus, appStore.layoutMode] as const,
  () => {
    if (!appStore.isMixLayout) return
    const hit = resolveTopMenu(authStore.menus, route.path)
    appStore.setActiveTopMenuId(hit?.id ?? null)
  },
  { immediate: true },
)

onMounted(async () => {
  await appStore.loadPublicConfig()
  if (authStore.menus.length === 0) {
    try {
      await authStore.fetchMenus()
    } catch {
      /* ignore */
    }
  }
})
</script>

<style scoped lang="scss">
.layout-container {
  height: 100vh;
  overflow: hidden;

  .layout-sidebar {
    overflow-y: auto;
    overflow-x: hidden;
    transition: width 0.28s ease, background-color 0.2s;
  }

  .layout-header-shell {
    display: flex;
    flex-direction: column;
    min-width: 0;
    flex: 1;
    min-height: 0;
    overflow: hidden;
    height: 100%;
  }

  .layout-body {
    display: flex;
    flex-direction: row;
    flex: 1;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
  }

  /* 内容列：Tags 固定，仅 layout-main 滚动 */
  .layout-content-col {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
  }

  .layout-navbar {
    background: #fff;
    border-bottom: 1px solid #e6e6e6;
    box-shadow: 2px 1px 4px rgba(0, 0, 0, 0.12);
    z-index: 10;
    padding: 0 16px;
    flex-shrink: 0;
    --el-header-padding: 0 16px;
    box-sizing: border-box;
  }

  &.layout-mix .layout-navbar,
  &.layout-topbar .layout-navbar {
    padding: 0 16px 0 0;
    --el-header-padding: 0 16px 0 0;
  }

  .layout-main {
    flex: 1;
    min-height: 0;
    min-width: 0;
    background: #f0f2f5;
    overflow-y: auto;
    padding: 16px;
  }

  .layout-crumb {
    margin: 0 0 12px;
    background: transparent;

    :deep(.app-breadcrumb) {
      margin-left: 0;
      line-height: 1.5;
      font-size: 13px;
    }
  }

  .page-container {
    padding: 16px;
    background-color: white;
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
