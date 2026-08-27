<template>
  <el-breadcrumb class="app-breadcrumb" separator="/">
    <el-breadcrumb-item
      v-for="(item, index) in items"
      :key="`${item.title}-${index}`"
      :to="linkOf(item, index)"
    >
      {{ item.title }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup lang="ts">
import { useBreadcrumb, type BreadcrumbItem } from '@/composables/useBreadcrumb'

const { items } = useBreadcrumb()

function linkOf(item: BreadcrumbItem, index: number) {
  // 最后一级为当前页，不可点；无 path 的目录不可点
  if (index === items.value.length - 1) return undefined
  if (!item.path) return undefined
  return { path: item.path }
}
</script>

<style scoped lang="scss">
.app-breadcrumb {
  margin-left: 12px;
  line-height: 50px;
  font-size: 14px;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;

  :deep(.el-breadcrumb__inner) {
    color: #606266;
    font-weight: 400;
  }

  :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
    color: #303133;
    font-weight: 500;
  }

  :deep(.el-breadcrumb__inner.is-link) {
    color: #606266;

    &:hover {
      color: #409eff;
    }
  }
}
</style>
