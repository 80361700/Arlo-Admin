<template>
  <div class="sidebar-container">
    <div
      v-if="showLogo"
      class="sidebar-logo"
      title="首页"
      @click="goHome"
    >
      <img :src="logoDisplay" alt="logo" class="logo-img" />
      <span v-show="!appStore.sidebarCollapsed" class="logo-title">{{ appStore.systemName }}</span>
    </div>

    <el-menu
      :key="`${appStore.themeId}-${menuKey}`"
      :default-active="activeMenu"
      :collapse="appStore.sidebarCollapsed"
      :collapse-transition="false"
      menu-trigger="hover"
      router
      class="sidebar-menu"
    >
      <SidebarItem
        v-for="item in menuList"
        :key="item.id"
        :item="item"
      />
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { MenuTreeNode } from '@/api'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useAuthFileSrc } from '@/composables/useAuthFileSrc'
import SidebarItem from './SidebarItem.vue'

const props = withDefaults(defineProps<{
  /** 不传则使用完整菜单树（侧栏模式） */
  menus?: MenuTreeNode[]
  showLogo?: boolean
}>(), {
  showLogo: true,
})

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()

const menuList = computed(() => props.menus ?? authStore.menus)
const menuKey = computed(() => menuList.value.map((m) => m.id).join('-'))
const activeMenu = computed(() => route.path)

const logoSrc = useAuthFileSrc(() => appStore.systemLogo || '')
const logoDisplay = computed(() => logoSrc.value || '/vite.svg')

function goHome() {
  router.push('/dashboard')
}

onMounted(async () => {
  await appStore.loadPublicConfig()
  if (authStore.menus.length === 0) {
    try {
      await authStore.fetchMenus()
    } catch {
      /* 拦截器已提示 */
    }
  }
})
</script>

<style scoped lang="scss">
.sidebar-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  position: relative;
  z-index: 2;
  cursor: pointer;

  .logo-img {
    width: 28px;
    height: 28px;
    flex-shrink: 0;
  }

  .logo-title {
    margin-left: 10px;
    font-size: 16px;
    font-weight: 600;
    white-space: nowrap;
  }
}

.sidebar-menu {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}
</style>
