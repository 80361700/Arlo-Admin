<template>
  <div class="arlo-upload-field">
    <div v-if="keys.length" class="file-list">
      <div v-for="key in keys" :key="key" class="file-item">
        <AuthFileImage
          v-if="isImage"
          :file-ref="key"
          fit="cover"
          img-class="thumb"
        >
          <template #error>
            <div class="thumb placeholder">
              <el-icon :size="20"><PictureFilled /></el-icon>
            </div>
          </template>
        </AuthFileImage>
        <div v-else class="thumb placeholder">
          <el-icon :size="20"><Document /></el-icon>
        </div>
        <span class="name" :title="nameMap[key] || key">{{ nameMap[key] || shortKey(key) }}</span>
        <el-button
          v-if="!disabled"
          type="danger"
          link
          size="small"
          @click="remove(key)"
        >
          删除
        </el-button>
      </div>
    </div>

    <span v-else-if="disabled" class="empty-text">无</span>

    <el-button
      v-if="!disabled && canAdd"
      type="primary"
      plain
      @click="pickerVisible = true"
    >
      {{ selectButtonText }}
    </el-button>

    <FilePicker
      v-model="pickerVisible"
      :title="isImage ? '选择图片' : '选择文件'"
      :mode="isMultiple ? 'multiple' : 'single'"
      :accept-types="acceptTypes"
      :max-count="pickerMaxCount"
      @confirm="onConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Document, PictureFilled } from '@element-plus/icons-vue'
import FilePicker from '@/components/FilePicker.vue'
import AuthFileImage from '@/components/AuthFileImage.vue'
import type { FileItem } from '@/api'
import { toFileRef } from '@/utils/fileUrl'

const props = withDefaults(
  defineProps<{
    /** image | file */
    kind?: 'image' | 'file'
    multiple?: boolean | string | number
    /** 最多数量；多选时 0 表示不限制 */
    limit?: number | string
    disabled?: boolean
  }>(),
  {
    kind: 'file',
    multiple: false,
    limit: 1,
  },
)

const model = defineModel<string>({ default: '' })
const emit = defineEmits<{ change: [value: string] }>()

const pickerVisible = ref(false)
const nameMap = reactive<Record<string, string>>({})

const isImage = computed(() => props.kind === 'image')
const acceptTypes = computed((): ('image' | 'video' | 'file')[] =>
  isImage.value ? ['image'] : ['file'],
)

const isMultiple = computed(() => {
  const v = props.multiple
  return v === true || v === 'true' || v === 1 || v === '1'
})

const limitNum = computed(() => {
  const n = Number(props.limit)
  // 最小为 1
  if (!Number.isFinite(n) || n < 1) return 1
  return Math.floor(n)
})

const keys = computed(() =>
  String(model.value || '')
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean),
)

/** 传给 FilePicker：单选固定 1；多选按剩余名额 */
const pickerMaxCount = computed(() => {
  if (!isMultiple.value) return 1
  return Math.max(0, limitNum.value - keys.value.length)
})

const canAdd = computed(() => {
  if (!isMultiple.value) return keys.value.length === 0
  return keys.value.length < limitNum.value
})

const selectButtonText = computed(() => {
  if (!isMultiple.value) return isImage.value ? '选择图片' : '选择文件'
  return keys.value.length ? (isImage.value ? '继续添加图片' : '继续添加文件') : (isImage.value ? '选择图片' : '选择文件')
})

function shortKey(key: string) {
  return key.length > 10 ? `${key.slice(0, 8)}…` : key
}

function write(next: string[]) {
  const val = next.join(',')
  model.value = val
  emit('change', val)
}

function remove(key: string) {
  write(keys.value.filter((k) => k !== key))
}

function onConfirm(files: FileItem[]) {
  if (!files.length) return
  files.forEach((f) => {
    nameMap[f.accessKey] = f.name
  })
  if (isMultiple.value) {
    const merged = [...keys.value]
    for (const f of files) {
      const ref = toFileRef(f)
      if (!merged.includes(ref)) merged.push(ref)
      if (merged.length >= limitNum.value) break
    }
    write(merged.slice(0, limitNum.value))
  } else {
    write([toFileRef(files[0])])
  }
}

watch(
  () => [isMultiple.value, limitNum.value] as const,
  () => {
    if (keys.value.length > limitNum.value) {
      write(keys.value.slice(0, isMultiple.value ? limitNum.value : 1))
    }
    if (!isMultiple.value && keys.value.length > 1) {
      write(keys.value.slice(0, 1))
    }
  },
)
</script>

<style scoped lang="scss">
.arlo-upload-field {
  width: 100%;
}
.empty-text {
  color: var(--el-text-color-regular);
  font-size: var(--el-font-size-base);
  line-height: 32px;
}
.file-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 8px;
}
.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  background: var(--el-fill-color-blank);
}
.thumb {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  object-fit: cover;
  flex-shrink: 0;
}
.thumb.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--el-fill-color);
  color: var(--el-text-color-secondary);
}
.name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}
</style>
