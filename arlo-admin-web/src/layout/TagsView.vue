<template>
  <div v-if="tags.length" class="tags-view">
    <button
      type="button"
      class="tags-nav"
      :disabled="!canScrollLeft"
      title="向左滚动"
      @click="scrollBy(-1)"
    >
      <el-icon :size="14"><DArrowLeft /></el-icon>
    </button>

    <div ref="scrollRef" class="tags-scroll" @scroll="updateScrollState">
      <el-dropdown
        v-for="tag in tags"
        :key="tag.path"
        trigger="contextmenu"
        placement="bottom-start"
        @command="(cmd: string) => onCommand(cmd, tag)"
      >
        <div
          class="tags-item"
          :class="{
            'is-active': tag.path === route.path,
            'is-home': isHome(tag),
          }"
          @click="go(tag)"
          @click.middle="onClose(tag)"
        >
          <el-icon v-if="isHome(tag)" class="tags-home-icon" :size="16">
            <HomeFilled />
          </el-icon>
          <template v-else>
            <span class="tags-title">{{ tag.title }}</span>
            <span
              v-if="!tag.affix"
              class="tags-close"
              @click.stop="onClose(tag)"
            >
              <el-icon :size="12"><Close /></el-icon>
            </span>
          </template>
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

    <button
      type="button"
      class="tags-nav"
      :disabled="!canScrollRight"
      title="向右滚动"
      @click="scrollBy(1)"
    >
      <el-icon :size="14"><DArrowRight /></el-icon>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Close, DArrowLeft, DArrowRight, HomeFilled } from '@element-plus/icons-vue'
import { useTagsStore, HOME_TAG, type TagItem } from '@/stores/tags'

const route = useRoute()
const router = useRouter()
const tagsStore = useTagsStore()

const tags = computed(() => tagsStore.visited)
const scrollRef = ref<HTMLElement | null>(null)
const canScrollLeft = ref(false)
const canScrollRight = ref(false)

function isHome(tag: TagItem) {
  return tag.path === HOME_TAG.path
}

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

function updateScrollState() {
  const el = scrollRef.value
  if (!el) {
    canScrollLeft.value = false
    canScrollRight.value = false
    return
  }
  const max = el.scrollWidth - el.clientWidth
  canScrollLeft.value = el.scrollLeft > 1
  canScrollRight.value = max > 1 && el.scrollLeft < max - 1
}

function scrollBy(dir: -1 | 1) {
  const el = scrollRef.value
  if (!el) return
  const step = Math.max(120, Math.floor(el.clientWidth * 0.6))
  el.scrollBy({ left: dir * step, behavior: 'smooth' })
}

function scrollActiveIntoView() {
  const el = scrollRef.value
  if (!el) return
  const active = el.querySelector('.tags-item.is-active') as HTMLElement | null
  active?.scrollIntoView({ inline: 'nearest', block: 'nearest', behavior: 'smooth' })
  nextTick(updateScrollState)
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

let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  updateScrollState()
  const el = scrollRef.value
  if (el && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => updateScrollState())
    resizeObserver.observe(el)
  }
  window.addEventListener('resize', updateScrollState)
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  window.removeEventListener('resize', updateScrollState)
})

watch(
  () => [tags.value.length, route.path] as const,
  async () => {
    await nextTick()
    scrollActiveIntoView()
    updateScrollState()
  },
)
</script>

<style scoped lang="scss">
.tags-view {
  display: flex;
  align-items: center;
  height: 40px;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  flex-shrink: 0;
  box-sizing: border-box;
}

.tags-nav {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 36px;
  height: 100%;
  padding: 0;
  border: none;
  background: transparent;
  color: #8c8c8c;
  cursor: pointer;
  transition: color 0.15s;

  &:first-child {
    border-right: 1px solid #e8e8e8;
  }

  &:hover:not(:disabled) {
    color: var(--el-color-primary);
  }

  &:disabled {
    color: #d9d9d9;
    cursor: not-allowed;
  }
}

.tags-scroll {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  /* 向下多出 1px，盖住栏底部分隔线；未选中透明，线仍可见 */
  height: calc(100% + 1px);
  margin-bottom: -1px;
  overflow-x: auto;
  overflow-y: hidden;
  white-space: nowrap;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }

  :deep(.el-dropdown) {
    height: 100%;
    flex-shrink: 0;
  }

  :deep(.el-tooltip__trigger) {
    display: flex !important;
    align-items: center;
    height: 100%;
  }
}

.tags-item {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 100%;
  padding: 0 14px;
  font-size: 13px;
  line-height: 1;
  color: #595959;
  background: transparent;
  border-right: 1px solid #e8e8e8;
  cursor: pointer;
  user-select: none;
  box-sizing: border-box;
  transition: color 0.15s, background 0.15s;

  &:hover {
    color: var(--el-color-primary);
  }

  &.is-home {
    padding: 0 12px;
  }

  &.is-active {
    color: var(--el-color-primary);
    background: #f0f2f5;

    .tags-close {
      color: var(--el-color-primary);
    }
  }
}

.tags-home-icon {
  color: inherit;
}

.tags-title {
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tags-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  color: #8c8c8c;
  border-radius: 2px;
  transition: color 0.15s, background 0.15s;

  &:hover {
    color: var(--el-color-primary);
    background: rgba(0, 0, 0, 0.06);
  }
}
</style>
