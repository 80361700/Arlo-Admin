<template>
  <div class="appearance-panel">
    <section class="block">
      <h3 class="block-title">主题风格</h3>
      <div class="option-row">
        <el-tooltip
          v-for="t in themeOptions"
          :key="t.id"
          placement="bottom"
          :show-after="200"
          effect="light"
        >
          <template #content>
            <div class="tip-card">
              <div class="tip-title">{{ t.label }}</div>
              <div class="tip-desc">{{ t.desc }}</div>
            </div>
          </template>
          <button
            type="button"
            class="option-card"
            :class="{ 'is-active': appStore.themeId === t.id }"
            @click="appStore.setTheme(t.id)"
          >
            <div class="preview" :class="`preview-theme-${t.id}`">
              <div class="preview-side" />
              <div class="preview-main">
                <div class="preview-bar" />
                <div class="preview-body" />
              </div>
            </div>
          </button>
        </el-tooltip>
      </div>
    </section>

    <section class="block">
      <h3 class="block-title">导航布局</h3>
      <div class="option-row">
        <el-tooltip
          v-for="opt in layoutModeOptions"
          :key="opt.id"
          placement="bottom"
          :show-after="200"
          effect="light"
        >
          <template #content>
            <div class="tip-card">
              <div class="tip-title">{{ opt.label }}</div>
              <div class="tip-desc">{{ opt.desc }}</div>
            </div>
          </template>
          <button
            type="button"
            class="option-card"
            :class="{ 'is-active': appStore.layoutMode === opt.id }"
            @click="appStore.setLayoutMode(opt.id)"
          >
            <div class="preview" :class="`preview-layout-${opt.id}`">
              <template v-if="opt.id === 'side'">
                <div class="pl-aside" />
                <div class="pl-right">
                  <div class="pl-nav" />
                  <div class="pl-content" />
                </div>
              </template>
              <template v-else-if="opt.id === 'mix'">
                <div class="pl-top" />
                <div class="pl-bottom">
                  <div class="pl-aside-light" />
                  <div class="pl-content-light" />
                </div>
              </template>
              <template v-else>
                <div class="pl-top" />
                <div class="pl-content-full" />
              </template>
            </div>
          </button>
        </el-tooltip>
      </div>
    </section>

    <section class="block">
      <h3 class="block-title">界面选项</h3>
      <div class="option-switch">
        <span class="switch-label">展示页签</span>
        <el-switch
          :model-value="appStore.tagsView"
          @change="(v: string | number | boolean) => appStore.setTagsView(!!v)"
        />
      </div>
      <p class="option-hint">顶栏下方多页签；开启后按菜单「是否缓存」保留页面状态</p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore, layoutModeOptions } from '@/stores/app'

const appStore = useAppStore()

const themeDesc: Record<string, string> = {
  light: '浅色界面，适合日间使用',
  dark: '深色侧栏/顶栏，减少眩光',
}

const themeOptions = computed(() =>
  appStore.themeOptions.map((t) => ({
    ...t,
    desc: themeDesc[t.id] || t.label,
  })),
)
</script>

<style scoped lang="scss">
.appearance-panel {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 4px 2px 2px;
}

.block {
  padding-bottom: 16px;
  border-bottom: 1px solid #ebeef5;

  &:last-child {
    padding-bottom: 0;
    border-bottom: none;
  }
}

.block-title {
  margin: 0 0 10px;
  font-size: 13px;
  font-weight: 500;
  color: #606266;
}

.option-switch {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.switch-label {
  font-size: 13px;
  color: #303133;
}

.option-hint {
  margin: 8px 0 0;
  font-size: 12px;
  line-height: 1.4;
  color: #909399;
}

.option-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.option-card {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 32px;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  outline: none;
  vertical-align: top;

  &:hover .preview {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  &.is-active .preview {
    box-shadow: 0 0 0 2px #409eff;
  }
}

.preview {
  width: 40px;
  height: 32px;
  border-radius: 4px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  background: #fff;
  transition: box-shadow 0.15s ease;
}

.tip-card {
  .tip-title {
    margin: 0 0 4px;
    font-size: 13px;
    font-weight: 500;
    color: #303133;
    line-height: 1.3;
  }

  .tip-desc {
    margin: 0;
    font-size: 12px;
    color: #909399;
    line-height: 1.45;
    max-width: 180px;
  }
}

.preview-theme-light,
.preview-theme-dark {
  display: flex;
}

.preview-side {
  width: 34%;
  height: 100%;
  flex-shrink: 0;
}

.preview-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.preview-bar {
  height: 22%;
  flex-shrink: 0;
}

.preview-body {
  flex: 1;
}

.preview-theme-light {
  .preview-side {
    background: #304156;
  }

  .preview-bar {
    background: #fff;
  }

  .preview-body {
    background: #f0f2f5;
  }
}

.preview-theme-dark {
  .preview-side {
    background: #1f2d3d;
  }

  .preview-bar {
    background: #304156;
  }

  .preview-body {
    background: #e8eaed;
  }
}

.preview-layout-side {
  display: flex;

  .pl-aside {
    width: 34%;
    background: #304156;
    flex-shrink: 0;
  }

  .pl-right {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .pl-nav {
    height: 22%;
    background: #fff;
  }

  .pl-content {
    flex: 1;
    background: #f0f2f5;
  }
}

.preview-layout-mix {
  display: flex;
  flex-direction: column;

  .pl-top {
    height: 26%;
    background: #304156;
    flex-shrink: 0;
  }

  .pl-bottom {
    flex: 1;
    display: flex;
    min-height: 0;
  }

  .pl-aside-light {
    width: 28%;
    background: #e8eaed;
    flex-shrink: 0;
  }

  .pl-content-light {
    flex: 1;
    background: #fff;
  }
}

.preview-layout-topbar {
  display: flex;
  flex-direction: column;

  .pl-top {
    height: 26%;
    background: #304156;
    flex-shrink: 0;
  }

  .pl-content-full {
    flex: 1;
    background: #f0f2f5;
  }
}
</style>
