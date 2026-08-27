<template>
  <div class="rich-editor" :style="{ height: height + 'px' }">
    <Toolbar
      :editor="editorRef"
      :defaultConfig="toolbarConfig"
      mode="default"
      class="editor-toolbar"
    />
    <Editor
      v-model="valueHtml"
      :defaultConfig="editorConfig"
      mode="default"
      class="editor-content"
      @onCreated="handleCreated"
      @onChange="handleChange"
    />

    <!-- 素材选择弹窗 -->
    <FilePicker
      v-model="pickerVisible"
      :mode="pickerMode"
      :accept-types="pickerAcceptTypes"
      :max-count="pickerMaxCount"
      @confirm="onPickerConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, onBeforeUnmount, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import type { IDomEditor, IEditorConfig, IToolbarConfig } from '@wangeditor/editor'
import FilePicker from './FilePicker.vue'
import { uploadFile, type FileItem } from '@/api/modules/file'
import { getFileAccessUrl } from '@/utils/fileUrl'
import '@wangeditor/editor/dist/css/style.css'

// ==================== Props / Emits ====================
const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  height?: number
}>(), {
  placeholder: '请输入内容...',
  height: 400,
})

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
}>()

// ==================== Editor ====================
const editorRef = shallowRef<IDomEditor | undefined>(undefined)
const valueHtml = shallowRef(props.modelValue)

// 当前 pending 的 insertFn（由 customBrowseAndUpload 传入）
let pendingInsertFn: ((url: string, ...args: any[]) => void) | null = null
let pendingMediaType: 'image' | 'video' | null = null

const toolbarConfig: Partial<IToolbarConfig> = {
  // 网络图片等 Modal 挂到 body，避免被编辑器 overflow 裁切
  modalAppendToBody: true,
}

const editorConfig: Partial<IEditorConfig> = {
  placeholder: props.placeholder,
  maxLength: 5000,
  autoFocus: false,
  MENU_CONF: {
    // 图片：点击内置图片 icon → 弹 FilePicker（可多选）
    uploadImage: {
      async customBrowseAndUpload(insertFn: (url: string, alt: string, href: string) => void) {
        pendingInsertFn = insertFn
        pendingMediaType = 'image'
        pickerAcceptTypes.value = ['image']
        pickerMode.value = 'multiple'
        pickerMaxCount.value = 0
        pickerVisible.value = true
      },
      customUpload(file: File, insertFn: (url: string, alt: string, href: string) => void) {
        // 兜底：如果浏览器直接选了文件，也走上传
        uploadLocalFile(file, 'image', insertFn)
      },
    },
    // 视频：点击内置视频 icon → 弹 FilePicker（可多选）
    uploadVideo: {
      async customBrowseAndUpload(insertFn: (url: string, poster: string, name: string) => void) {
        pendingInsertFn = insertFn
        pendingMediaType = 'video'
        pickerAcceptTypes.value = ['video']
        pickerMode.value = 'multiple'
        pickerMaxCount.value = 0
        pickerVisible.value = true
      },
      customUpload(file: File, insertFn: (url: string, poster: string, name: string) => void) {
        uploadLocalFile(file, 'video', insertFn)
      },
    },
  },
}

/** 将 wangEditor Modal 相对当前公告/消息等 el-dialog 居中 */
function centerEditorModal($elem: { css: (s: Record<string, string | number>) => void; width: () => number; height: () => number }) {
  let host: HTMLElement | undefined
  for (const dialog of document.querySelectorAll<HTMLElement>('.el-overlay .el-dialog')) {
    const overlay = dialog.closest('.el-overlay') as HTMLElement | null
    if (!overlay || getComputedStyle(overlay).display === 'none') continue
    host = dialog
  }

  const mw = $elem.width()
  const mh = $elem.height()

  if (host) {
    const rect = host.getBoundingClientRect()
    $elem.css({
      position: 'fixed',
      left: `${Math.max(8, rect.left + (rect.width - mw) / 2)}px`,
      top: `${Math.max(8, rect.top + (rect.height - mh) / 2)}px`,
      marginLeft: '0',
      marginTop: '0',
      zIndex: 4000,
    })
    return
  }

  $elem.css({
    position: 'fixed',
    left: '50%',
    top: '50%',
    marginLeft: `${-mw / 2}px`,
    marginTop: `${-mh / 2}px`,
    zIndex: 4000,
  })
}

function handleCreated(editor: IDomEditor) {
  editorRef.value = editor
  editor.on('modalOrPanelShow', (modalOrPanel: { type: string; $elem: any }) => {
    if (modalOrPanel.type !== 'modal') return
    // 等 DOM 完成布局后再量宽高
    requestAnimationFrame(() => centerEditorModal(modalOrPanel.$elem))
  })
}

function handleChange(editor: IDomEditor) {
  emit('update:modelValue', editor.getHtml())
}

watch(
  () => props.modelValue,
  (val) => {
    if (val !== valueHtml.value && editorRef.value) {
      valueHtml.value = val
    }
  },
)

onBeforeUnmount(() => {
  editorRef.value?.destroy()
})

/** 编辑器正文会落库，只插入不带 token 的固定公开地址 */
function getFileUrl(file: { accessKey: string }): string {
  return getFileAccessUrl(file.accessKey)
}

// ==================== FilePicker ====================
const pickerVisible = ref(false)
const pickerMode = ref<'single' | 'multiple'>('multiple')
const pickerAcceptTypes = ref<('image' | 'video' | 'file')[]>(['image'])
const pickerMaxCount = ref(0)

function onPickerConfirm(files: FileItem[]) {
  if (!pendingInsertFn || files.length === 0) return

  const publicFiles = files.filter((f) => f.isPublic === 1)
  const skipped = files.length - publicFiles.length
  if (publicFiles.length === 0) {
    ElMessage.warning('所选文件均为私有，写入正文后无法长期访问，请先在文件管理中设为公开')
    return
  }

  for (const file of publicFiles) {
    const url = getFileUrl(file)
    if (pendingMediaType === 'image') {
      ;(pendingInsertFn as (url: string, alt: string, href: string) => void)(url, file.name, url)
    } else if (pendingMediaType === 'video') {
      ;(pendingInsertFn as (url: string, poster: string, name: string) => void)(url, '', file.name)
    }
  }

  pendingInsertFn = null
  pendingMediaType = null
  if (skipped > 0) {
    ElMessage.warning(`已插入 ${publicFiles.length} 个，跳过 ${skipped} 个私有文件`)
  } else {
    ElMessage.success(`已插入 ${publicFiles.length} 个`)
  }
}

// ==================== 本地文件兜底上传（正文配图默认公开） ====================
async function uploadLocalFile(
  file: File,
  type: 'image' | 'video',
  insertFn: (...args: any[]) => void,
) {
  try {
    const res = await uploadFile(file, { public: true })
    const url = getFileUrl(res.data)
    if (type === 'image') {
      insertFn(url, res.data.name, url)
    } else {
      insertFn(url, '', res.data.name)
    }
  } catch (err: any) {
    showRequestError(err, '上传失败')
  }
}
</script>

<style scoped lang="scss">
.rich-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  :deep(.editor-toolbar) {
    border-bottom: 1px solid #dcdfe6;
    flex-shrink: 0;
  }

  :deep(.editor-content) {
    flex: 1;
    overflow: hidden;
  }

  :deep(.w-e-text-container) {
    background-color: transparent;
    height: 100% !important;
    overflow-y: auto !important;

    [data-slate-editor] {
      padding: 6px 10px;
    }
  }

  :deep(.w-e-bar-divider) {
    height: 28px;
    margin: 6px 5px;
  }
}
</style>
