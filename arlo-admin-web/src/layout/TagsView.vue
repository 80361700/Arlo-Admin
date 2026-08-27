<template>
  <div v-if="tags.length" class="tags-view">
    <div class="tags-scroll">
      <el-dropdown
        v-for="tag in tags"
        :key="tag.path"
        trigger="contextmenu"
        placement="bottom-start"
        @command="(cmd: string) => onCommand(cmd, tag)"
      >
        <div
          class="tags-item"
          :class="{ 'is-active': tag.path === route.path }"
          @click="go(tag)"
          @click.middle="onClose(tag)"
        >
          <span class="tags-title">{{ tag.title }}</span>
          <span
            v-if="!tag.affix"
            class="tags-close"
            @click.stop="onClose(tag)"
          >
            <el-icon :size="12"><Close /></el-icon>
          </span>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="refresh">刷新</el-dropdown-item>
            <el-dropdown-item command="close" :disabled="tag.affix">关闭</el-dropdown-item>
            <el-dropdown-item command="closeOthers">关闭其它</el-dropdown-item>
            <el-dropdown-item command="closeLeft" :disabled="!canCloseLeft(tag)">
              关闭左侧
            </el-dropdown-item>
            <el-dropdown-item command="closeRight" :disabled="!canCloseRight(tag)">
              关闭右侧
            </el-dropdown-item>
            <el-dropdown-item command="closeAll">关闭全部</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Close } from '@element-plus/icons-vue'
import { useTagsStore, type TagItem } from '@/stores/tags'

const route = useRoute()
const router = useRouter()
const tagsStore = useTagsStore()

const tags = computed(() => tagsStore.visited)

function tagIndex(tag: TagItem) {
  return tagsStore.visited.findIndex((t) => t.path === tag.path)
}

function canCloseLeft(tag: TagItem) {
  const idx = tagIndex(tag)
  if (idx <= 0) return false
  return tagsStore.visited.slice(0, idx).some((t) => !t.affix)
}

function canCloseRight(tag: TagItem) {
  const idx = tagIndex(tag)
  if (idx < 0 || idx >= tagsStore.visited.length - 1) return false
  return tagsStore.visited.slice(idx + 1).some((t) => !t.affix)
}

function go(tag: TagItem) {
  if (tag.path === route.path) return
  router.push(tag.fullPath || tag.path)
}

function onClose(tag: TagItem) {
  if (tag.affix) return
  tagsStore.closeTag(tag.path, router)
}

function onCommand(cmd: string, tag: TagItem) {
  switch (cmd) {
    case 'refresh':
      tagsStore.refreshTag(tag.path, router)
      break
    case 'close':
      tagsStore.closeTag(tag.path, router)
      break
    case 'closeOthers':
      tagsStore.closeOthers(tag.path, router)
      break
    case 'closeLeft':
      tagsStore.closeLeft(tag.path, router)
      break
    case 'closeRight':
      tagsStore.closeRight(tag.path, router)
      break
    case 'closeAll':
      tagsStore.closeAll(router)
      break
  }
}
</script>

<style scoped lang="scss">
.tags-view {
  display: flex;
  align-items: center;
  height: 44px;
  padding: 0 16px;
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
  flex-shrink: 0;
}

.tags-scroll {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow-x: auto;
  overflow-y: hidden;
  white-space: nowrap;
  width: 100%;
  scrollbar-width: thin;

  &::-webkit-scrollbar {
    height: 4px;
  }

  :deep(.el-dropdown) {
    flex-shrink: 0;
  }
}

.tags-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 26px;
  padding: 0 8px;
  font-size: 12px;
  color: #606266;
  background: #f4f4f5;
  border: 1px solid #e4e7ed;
  border-radius: 3px;
  cursor: pointer;
  user-select: none;
  flex-shrink: 0;
  transition: color 0.15s, background 0.15s, border-color 0.15s;

  &:hover {
    color: var(--el-color-primary);
  }

  &.is-active {
    color: #fff;
    background: var(--el-color-primary);
    border-color: var(--el-color-primary);

    .tags-close:hover {
      background: rgba(255, 255, 255, 0.25);
    }
  }
}

.tags-title {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tags-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  margin-left: 2px;

  &:hover {
    background: rgba(0, 0, 0, 0.08);
  }
}
</style>
