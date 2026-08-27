<template>
  <template v-if="item.type === 1">
    <!-- 目录 → el-sub-menu -->
    <el-sub-menu :index="resolvePath(item.path)">
      <template #title>
        <el-icon v-if="iconComp">
          <component :is="iconComp" />
        </el-icon>
        <span>{{ item.name }}</span>
      </template>
      <SidebarItem
        v-for="child in visibleChildren"
        :key="child.id"
        :item="child"
        :base-path="resolvePath(item.path)"
      />
    </el-sub-menu>
  </template>

  <template v-else-if="item.type === 2">
    <!-- 菜单 → el-menu-item -->
    <el-menu-item :index="resolvePath(item.path)">
      <el-icon v-if="iconComp">
        <component :is="iconComp" />
      </el-icon>
      <template #title>{{ item.name }}</template>
    </el-menu-item>
  </template>

  <!-- type=3 按钮不渲染 -->
</template>

<script setup lang="ts">
import { computed } from 'vue'
import * as Icons from '@element-plus/icons-vue'
import type { MenuTreeNode } from '@/api'

const props = defineProps<{
  item: MenuTreeNode
  basePath?: string
}>()

// 直接通过图标包把字符串名映射为组件引用，避免 <component :is="string"> 的运行时解析问题
const iconComp = computed(() => {
  if (!props.item.icon) return null
  return (Icons as Record<string, unknown>)[props.item.icon] || null
})

// 只看可见的子节点（排除 type=3 按钮和隐藏项）
const visibleChildren = computed(() =>
  (props.item.children || []).filter(c => c.type !== 3 && c.visible !== 0)
)

function resolvePath(childPath: string): string {
  if (!childPath) return ''
  // 已经是绝对路径就直接用
  if (childPath.startsWith('/')) return childPath
  // 拼接父路径
  const base = props.basePath || ''
  return base ? `${base}/${childPath}`.replace(/\/+/g, '/') : `/${childPath}`
}
</script>
