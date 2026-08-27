<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="900px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:modelValue', $event)"
    @close="handleClose"
  >
    <!-- 类型 Tab -->
    <el-tabs v-model="activeTab" class="picker-tabs" @tab-change="onTabChange">
      <el-tab-pane label="图片" name="image" v-if="showTab('image')" />
      <el-tab-pane label="视频" name="video" v-if="showTab('video')" />
      <el-tab-pane label="文件" name="file" v-if="showTab('file')" />
    </el-tabs>

    <!-- 搜索工具栏 -->
    <div class="picker-toolbar">
      <el-select
        v-model="query.category"
        placeholder="素材分类"
        style="width: 140px"
        clearable
        @change="handleSearch"
      >
        <el-option
          v-for="opt in categoryOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <el-select
        v-model="query.isPublic"
        placeholder="公开状态"
        style="width: 120px"
        clearable
        @change="handleSearch"
      >
        <el-option label="公开" :value="1" />
        <el-option label="私有" :value="0" />
      </el-select>
      <el-input
        v-model="query.name"
        placeholder="请输入素材名称"
        clearable
        style="width: 200px"
        @keyup.enter="handleSearch"
        @clear="handleSearch"
      />
      <el-button type="primary" @click="handleSearch">
        <el-icon><Search /></el-icon>
        <span>搜索</span>
      </el-button>
      <el-button @click="handleReset">
        <el-icon><RefreshRight /></el-icon>
        <span>清空</span>
      </el-button>
      <div style="flex: 1"></div>
      <el-button v-permission="'sys:file:upload'" type="primary" @click="handleUploadClick">
        <el-icon><Upload /></el-icon>
        <span>上传素材</span>
      </el-button>
      <input
        ref="uploadInputRef"
        type="file"
        :accept="uploadAccept"
        multiple
        hidden
        @change="handleUploadFile"
      />
    </div>

    <!-- 文件列表 -->
    <el-table
      :data="fileList"
      height="340"
      highlight-current-row
      @row-click="handleRowClick"
      :row-class-name="rowClassName"
    >
      <el-table-column v-if="mode === 'multiple'" width="55" align="center">
        <template #header>
          <el-checkbox
            :model-value="isAllSelected"
            :indeterminate="isIndeterminate"
            @change="(val: any) => toggleSelectAll(val as boolean)"
          />
        </template>
        <template #default="{ row }">
          <el-checkbox
            :model-value="selectedMap.has(row.id)"
            @click.stop
            @change="(val: any) => toggleRow(row as FileItem, val as boolean)"
          />
        </template>
      </el-table-column>
      <el-table-column v-else width="45" align="center">
        <template #default="{ row }">
          <el-radio
            v-model="singleSelectedId"
            :label="row.id"
            style="margin-right: 0"
            @click.stop
          />
        </template>
      </el-table-column>

      <el-table-column label="素材名称" min-width="200">
        <template #default="{ row }">
          <div class="file-name-cell">
            <AuthFileImage
              v-if="row.mimeType?.startsWith('image/')"
              :file-ref="row.accessKey"
              fit="cover"
              img-class="file-thumb"
            >
              <template #error>
                <div class="file-thumb-placeholder">
                  <el-icon :size="20" color="#fff"><PictureFilled /></el-icon>
                </div>
              </template>
            </AuthFileImage>
            <div v-else class="file-thumb-placeholder">
              <el-icon :size="20" color="#fff">
                <component :is="fileIcon(row.mimeType)" />
              </el-icon>
            </div>
            <span class="file-name-text" :title="row.name">{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="素材分类" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="categoryTag(row.category)" size="small">
            {{ categoryLabel(row.category) }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="公开状态" width="90" align="center">
        <template #default="{ row }">
          <el-tag :type="row.isPublic === 1 ? 'success' : 'info'" size="small">
            {{ row.isPublic === 1 ? '公开' : '私有' }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="素材大小" width="100" align="center">
        <template #default="{ row }">
          <span class="size-tag">{{ formatSize(row.size) }}</span>
        </template>
      </el-table-column>

      <el-table-column prop="createdAt" label="上传时间" width="160" align="center" />

      <el-table-column label="操作" width="80" align="center" v-if="mode === 'single'">
        <template #default="{ row }">
          <el-button v-if="authStore.hasPermission('sys:file:delete')" type="danger" link size="small" @click.stop="handleDelete(row as FileItem)">
            <el-icon><Delete /></el-icon>
            <span>删除</span>
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="picker-pagination">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        small
        @change="handlePageChange"
      />
    </div>

    <!-- 底部已选提示 -->
    <div v-if="mode === 'multiple' && selectedFiles.length > 0" class="selected-bar">
      <span>已选 {{ selectedFiles.length }} 项</span>
      <span v-if="maxCount > 0" class="max-hint">（最多 {{ maxCount }} 项）</span>
      <el-button link size="small" @click="clearSelection">清空已选</el-button>
    </div>

    <!-- 底部按钮 -->
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel">取消</el-button>
        <el-button type="primary" @click="handleConfirm" :disabled="!canConfirm">
          确定{{ mode === 'multiple' && selectedFiles.length > 0 ? `（${selectedFiles.length}）` : '' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import { Search, RefreshRight, Upload, Delete, PictureFilled, VideoCamera, Document, Headset } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { getFileList, uploadFile, deleteFile, type FileItem } from '@/api/modules/file'
import AuthFileImage from '@/components/AuthFileImage.vue'
import { UPLOAD_ACCEPT, isAllowedUploadFile } from '@/utils/uploadAllow'

// ==================== Props / Emits ====================
const props = withDefaults(defineProps<{
  modelValue: boolean
  title?: string
  mode?: 'single' | 'multiple'
  acceptTypes?: ('image' | 'video' | 'file')[]
  maxCount?: number
}>(), {
  title: '选择素材',
  mode: 'single',
  acceptTypes: () => ['image', 'video', 'file'],
  maxCount: 0, // 0 表示不限制
})

const emit = defineEmits<{
  (e: 'update:modelValue', val: boolean): void
  (e: 'confirm', files: FileItem[]): void
  (e: 'cancel'): void
}>()

// ==================== State ====================
const authStore = useAuthStore()
const uploadInputRef = ref<HTMLInputElement>()

const activeTab = ref<'image' | 'video' | 'file'>('image')
const fileList = ref<FileItem[]>([])
const total = ref(0)

const query = reactive({
  page: 1,
  pageSize: 10,
  name: '',
  category: '',
  isPublic: undefined as number | undefined,
})

// 单选
const singleSelectedId = ref<number>(0)

// 多选
const selectedMap = ref<Map<number, FileItem>>(new Map())
const selectedFiles = computed(() => Array.from(selectedMap.value.values()))
const isAllSelected = computed(() => {
  if (fileList.value.length === 0) return false
  return fileList.value.every(row => selectedMap.value.has(row.id))
})
const isIndeterminate = computed(() => {
  const count = fileList.value.filter(row => selectedMap.value.has(row.id)).length
  return count > 0 && count < fileList.value.length
})

const canConfirm = computed(() => {
  if (props.mode === 'single') return singleSelectedId.value > 0
  return selectedMap.value.size > 0
})

// 分类下拉选项根据 acceptTypes 过滤
const categoryOptions = computed(() => {
  const options: { label: string; value: string }[] = [{ label: '全部', value: '' }]
  const seen = new Set<string>()
  for (const type of props.acceptTypes) {
    if (type === 'image' && !seen.has('image')) {
      seen.add('image')
      options.push({ label: '图片', value: 'image' })
    }
    if (type === 'video' && !seen.has('video')) {
      seen.add('video')
      options.push({ label: '视频', value: 'video' })
    }
    if (type === 'file') {
      for (const cat of ['audio', 'document', 'other'] as const) {
        if (!seen.has(cat)) {
          seen.add(cat)
          options.push({ label: categoryLabel(cat), value: cat })
        }
      }
    }
  }
  return options
})

// 上传文件类型
const uploadAccept = computed(() => {
  switch (activeTab.value) {
    case 'image': return 'image/*,.jpg,.jpeg,.png,.gif,.webp,.bmp,.ico,.svg'
    case 'video': return 'video/*,.mp4,.webm,.mov,.avi'
    case 'file': return UPLOAD_ACCEPT
    default: return UPLOAD_ACCEPT
  }
})

// ==================== Watch ====================
watch(() => props.modelValue, (visible) => {
  if (visible) {
    // 初始化默认tab
    const types = props.acceptTypes
    if (types.includes('image')) activeTab.value = 'image'
    else if (types.includes('video')) activeTab.value = 'video'
    else if (types.includes('file')) activeTab.value = 'file'

    // 重置查询
    query.page = 1
    query.pageSize = 10
    query.name = ''
    query.category = ''
    query.isPublic = undefined

    // 重置选择
    singleSelectedId.value = 0
    selectedMap.value.clear()

    loadFiles()
  }
})

// ==================== Tab / Search ====================
function showTab(type: 'image' | 'video' | 'file') {
  return props.acceptTypes.includes(type)
}

function onTabChange() {
  query.page = 1
  query.name = ''
  query.category = ''
  query.isPublic = undefined
  loadFiles()
}

function handleSearch() {
  query.page = 1
  loadFiles()
}

function handleReset() {
  query.name = ''
  query.category = ''
  query.isPublic = undefined
  query.page = 1
  handleSearch()
}

function handlePageChange() {
  loadFiles()
}

async function loadFiles() {
  try {
    const category = activeTab.value === 'file' ? '' : activeTab.value
    const res = await getFileList({
      page: query.page,
      pageSize: query.pageSize,
      name: query.name || undefined,
      category: query.category || category || undefined,
      isPublic: query.isPublic,
    })
    fileList.value = res.data.list || []
    total.value = res.data.total
  } catch (err: any) {
    showRequestError(err, '加载文件失败')
  }
}

// ==================== Selection ====================
function handleRowClick(row: any) {
  const file = row as FileItem
  if (props.mode === 'single') {
    singleSelectedId.value = file.id
  } else {
    toggleRow(file, !selectedMap.value.has(file.id))
  }
}

function toggleRow(row: FileItem, checked: boolean) {
  if (checked) {
    if (props.maxCount > 0 && selectedMap.value.size >= props.maxCount) {
      ElMessage.warning(`最多只能选择 ${props.maxCount} 个文件`)
      return
    }
    selectedMap.value.set(row.id, row)
  } else {
    selectedMap.value.delete(row.id)
  }
}

function toggleSelectAll(checked: boolean) {
  if (checked) {
    for (const row of fileList.value) {
      if (props.maxCount > 0 && selectedMap.value.size >= props.maxCount) {
        ElMessage.warning(`最多只能选择 ${props.maxCount} 个文件，已自动选满`)
        break
      }
      selectedMap.value.set(row.id, row)
    }
  } else {
    for (const row of fileList.value) {
      selectedMap.value.delete(row.id)
    }
  }
}

function rowClassName({ row }: { row: any }) {
  const file = row as FileItem
  if (props.mode === 'single') {
    return singleSelectedId.value === file.id ? 'selected-row' : ''
  }
  return selectedMap.value.has(file.id) ? 'selected-row' : ''
}

function clearSelection() {
  selectedMap.value.clear()
}

// ==================== Upload ====================
function handleUploadClick() {
  uploadInputRef.value?.click()
}

async function handleUploadFile(e: Event) {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files || files.length === 0) return

  // 检查数量限制
  if (props.maxCount > 0 && files.length > props.maxCount) {
    ElMessage.warning(`一次最多上传 ${props.maxCount} 个文件`)
    input.value = ''
    return
  }

  for (const file of files) {
    if (!isAllowedUploadFile(file)) {
      ElMessage.warning(`「${file.name}」类型不允许上传`)
      continue
    }
    try {
      const res = await uploadFile(file, { public: true })
      // 自动选中新上传的文件
      if (props.mode === 'single') {
        singleSelectedId.value = res.data.id
      } else {
        selectedMap.value.set(res.data.id, res.data)
      }
    } catch (err: any) {
      ElMessage.error(file.name + ': ' + (err.message || '上传失败'))
    }
  }

  input.value = ''
  loadFiles()
}

// ==================== Delete ====================
async function handleDelete(row: FileItem) {
  try {
    await ElMessageBox.confirm('确定删除该素材吗？', '提示', { type: 'warning' })
    await deleteFile(row.id)
    ElMessage.success('删除成功')
    if (singleSelectedId.value === row.id) {
      singleSelectedId.value = 0
    }
    selectedMap.value.delete(row.id)
    loadFiles()
  } catch (err: any) {
    if (err !== 'cancel') {
      showRequestError(err, '删除失败')
    }
  }
}

// ==================== Confirm / Cancel ====================
function handleConfirm() {
  if (props.mode === 'single') {
    const file = fileList.value.find(f => f.id === singleSelectedId.value)
    if (file) {
      emit('confirm', [file])
    }
  } else {
    emit('confirm', Array.from(selectedMap.value.values()))
  }
  emit('update:modelValue', false)
}

function handleCancel() {
  emit('cancel')
  emit('update:modelValue', false)
}

function handleClose() {
  emit('cancel')
}

// ==================== Tools ====================
function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function categoryLabel(cat: string): string {
  const map: Record<string, string> = {
    image: '图片', video: '视频', audio: '音频',
    document: '文档', other: '其他',
  }
  return map[cat] || cat
}

function categoryTag(cat: string): any {
  const map: Record<string, any> = {
    image: 'success', video: 'warning', audio: 'info',
    document: '', other: 'info',
  }
  return map[cat] || ''
}

function fileIcon(mimeType: string) {
  if (mimeType?.startsWith('video/')) return VideoCamera
  if (mimeType?.startsWith('audio/')) return Headset
  return Document
}
</script>

<style scoped lang="scss">
.picker-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 12px;
  }
  :deep(.el-tabs__item) {
    font-size: 14px;
  }
}

.picker-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  padding: 10px 12px;
  background: #f5f7fa;
  border-radius: 4px;

  .el-button span {
    margin-left: 4px;
  }
}

.picker-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}

.selected-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
  padding: 8px 12px;
  background: #ecf5ff;
  border-radius: 4px;
  font-size: 13px;
  color: #409eff;

  .max-hint {
    color: #909399;
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

// 表格行选中样式
:deep(.selected-row) {
  background-color: #ecf5ff !important;
}

// 文件名称单元格
  .file-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;

  .file-thumb {
    width: 60px;
    height: 60px;
    object-fit: cover;
    border-radius: 4px;
    border: 1px solid #ebeef5;
    flex-shrink: 0;
    background: #f5f7fa;
    cursor: pointer;
  }

  .file-thumb-placeholder {
    width: 60px;
    height: 60px;
    border-radius: 4px;
    background: #c0c4cc;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .file-name-text {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
    color: #303133;
  }
}

// 大小标签
.size-tag {
  color: #409eff;
  background: #ecf5ff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

// radio 单选居中
:deep(.el-radio__label) {
  display: none;
}
</style>
