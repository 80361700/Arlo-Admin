<template>
  <div class="page-container">
    <ProTable
      ref="proTableRef"
      :data="tableData"
      :loading="loading"
      :total="total"
      :search-fields="searchFields"
      :show-index="false"
      :action-width="240"
      @search="handleSearch"
      @reset="handleReset"
      @page-change="handlePageChange"
    >
      <template #toolbar>
        <el-button v-permission="'sys:file:upload'" type="primary" @click="uploadVisible = true">上传文件</el-button>
      </template>

      <el-table-column prop="id" label="ID" width="80" align="center" />
      <el-table-column label="预览" width="80" align="center">
        <template #default="{ row }">
          <AuthFileImage
            v-if="row.mimeType?.startsWith('image/')"
            :file-ref="row.accessKey"
            fit="cover"
            img-class="file-thumb-preview"
          >
            <template #error>
              <div class="file-thumb-fallback">
                <el-icon :size="20"><PictureFilled /></el-icon>
              </div>
            </template>
          </AuthFileImage>
          <div v-else class="file-thumb-fallback">
            <el-icon :size="20"><component :is="fileIcon(row.mimeType)" /></el-icon>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="文件名" min-width="180" show-overflow-tooltip />
      <el-table-column label="大小" width="100" align="right">
        <template #default="{ row }">
          {{ formatSize(row.size) }}
        </template>
      </el-table-column>
      <el-table-column prop="mimeType" label="MIME类型" width="180" show-overflow-tooltip />
      <el-table-column label="分类" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="categoryTag(row.category)" size="small">
            {{ categoryLabel(row.category) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="公开" width="100" align="center">
        <template #default="{ row }">
          <el-switch
            v-if="authStore.hasPermission('sys:file:upload')"
            :model-value="row.isPublic === 1"
            inline-prompt
            active-text="公"
            inactive-text="私"
            @change="(val: string | number | boolean) => handleTogglePublic(row as FileItem, Boolean(val))"
          />
          <el-tag v-else :type="row.isPublic === 1 ? 'success' : 'info'" size="small">
            {{ row.isPublic === 1 ? '公开' : '私有' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="md5" label="MD5" width="140" show-overflow-tooltip>
        <template #default="{ row }">
          <span style="font-family: monospace; font-size: 12px;">{{ row.md5 }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="uploader" label="上传者" width="100" />
      <el-table-column prop="createdAt" label="上传时间" width="170" />

      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="handleCopyUrl(row)">复制地址</el-button>
        <el-button type="primary" link size="small" @click="handleDownload(row)">下载</el-button>
        <el-button v-if="authStore.hasPermission('sys:file:delete')" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <!-- 上传弹窗 -->
    <el-dialog v-model="uploadVisible" title="上传文件" width="640px" :close-on-click-modal="false" @closed="resetUpload">
      <el-upload
        ref="uploadRef"
        class="upload-area"
        drag
        :auto-upload="false"
        multiple
        action=""
        :accept="UPLOAD_ACCEPT"
        :on-change="handleFileChange"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            单个不超过 50MB；允许图片/音视频/办公文档/压缩包，禁止可执行与脚本类文件
          </div>
        </template>
      </el-upload>

      <!-- 待上传文件列表 + 进度 -->
      <div v-if="uploadQueue.length > 0" class="upload-queue">
        <div
          v-for="item in uploadQueue"
          :key="item.uid"
          class="upload-queue-item"
        >
          <div class="queue-info">
            <span class="queue-name" :title="item.name">{{ item.name }}</span>
            <span class="queue-size">{{ formatSize(item.size) }}</span>
          </div>
          <el-progress
            v-if="item.status === 'uploading' || item.status === 'done' || item.status === 'error'"
            :percentage="item.percentage"
            :status="item.status === 'done' ? 'success' : item.status === 'error' ? 'exception' : ''"
            :show-text="item.status === 'uploading' || item.status === 'error'"
            :stroke-width="4"
            :indeterminate="item.status === 'uploading' && item.percentage === 0"
            style="flex: 1; margin-left: 12px;"
          />
          <el-icon
            v-if="item.status === 'pending'"
            class="queue-remove"
            @click="removeFile(item.uid)"
          >
            <CircleClose />
          </el-icon>
          <el-tag v-if="item.status === 'done'" type="success" size="small">成功</el-tag>
          <el-tag v-if="item.status === 'error'" type="danger" size="small">失败</el-tag>
        </div>
      </div>

      <template #footer>
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" :disabled="uploadQueue.length === 0" @click="handleUpload">
          上传{{ uploadQueue.length > 0 ? `（${uploadQueue.length} 个）` : '' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled, PictureFilled, VideoCamera, Document, Headset, CircleClose } from '@element-plus/icons-vue'
import type { UploadInstance, UploadFile, UploadRawFile } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import AuthFileImage from '@/components/AuthFileImage.vue'
import { useAuthStore } from '@/stores/auth'
import {
  getFileList, uploadFile, deleteFile, setFilePublic,
  type FileItem, type FileListQuery,
} from '@/api/modules/file'
import { getFileAccessUrl } from '@/utils/fileUrl'
import { downloadAuthFile } from '@/utils/authFileUrl'
import { UPLOAD_ACCEPT, isAllowedUploadFile } from '@/utils/uploadAllow'

// ==================== 查询 ====================
const authStore = useAuthStore()
const proTableRef = ref()
const tableData = ref<FileItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive<FileListQuery>({ page: 1, pageSize: 10 })

const searchFields = [
  { prop: 'name', label: '文件名' },
  {
    prop: 'category', label: '分类', type: 'select' as const,
    options: [
      { label: '图片', value: 'image' },
      { label: '视频', value: 'video' },
      { label: '音频', value: 'audio' },
      { label: '文档', value: 'document' },
      { label: '其他', value: 'other' },
    ],
  },
  {
    prop: 'isPublic', label: '公开', type: 'select' as const,
    options: [
      { label: '公开', value: 1 },
      { label: '私有', value: 0 },
    ],
  },
]

function categoryTag(c: string) {
  const m: Record<string, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    image: 'success', video: 'danger', audio: 'warning', document: '', other: 'info',
  }
  return m[c] || 'info'
}
function categoryLabel(c: string) {
  const m: Record<string, string> = {
    image: '图片', video: '视频', audio: '音频', document: '文档', other: '其他',
  }
  return m[c] || c
}
function formatSize(bytes: number) {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return (bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 1) + ' ' + units[i]
}

async function loadData(params: FileListQuery) {
  loading.value = true
  try {
    const res = await getFileList(params)
    tableData.value = res.data.list || []
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) {
  query.page = 1
  Object.assign(query, p)
  loadData(query)
}
function handleReset() {
  Object.assign(query, { page: 1, pageSize: query.pageSize || 10, name: undefined, category: undefined, isPublic: undefined })
  loadData(query)
}
function handlePageChange(p: any) {
  Object.assign(query, p)
  loadData(query)
}

// ==================== 上传 ====================
interface UploadQueueItem {
  uid: number
  name: string
  size: number
  raw: File
  status: 'pending' | 'uploading' | 'done' | 'error'
  percentage: number
  message: string
}

const uploadVisible = ref(false)
const uploading = ref(false)
const uploadRef = ref<UploadInstance>()
const uploadQueue = ref<UploadQueueItem[]>([])

function handleFileChange(file: UploadFile) {
  const raw = file.raw
  if (!raw) return

  // 去重：相同文件名 + 大小视为重复
  const exists = uploadQueue.value.some(
    item => item.name === raw.name && item.size === raw.size
  )
  if (exists) return

  if (!isAllowedUploadFile(raw)) {
    ElMessage.warning(`「${raw.name}」类型不允许上传`)
    return
  }

  // 50MB 单个限制
  if (raw.size > 50 * 1024 * 1024) {
    ElMessage.warning(`「${raw.name}」超过 50MB，已跳过`)
    return
  }

  uploadQueue.value.push({
    uid: file.uid,
    name: raw.name,
    size: raw.size,
    raw,
    status: 'pending',
    percentage: 0,
    message: '',
  })
}

function removeFile(uid: number) {
  const idx = uploadQueue.value.findIndex(item => item.uid === uid)
  if (idx !== -1) uploadQueue.value.splice(idx, 1)
}

function resetUpload() {
  uploadQueue.value = []
  uploadRef.value?.clearFiles()
}

async function handleUpload() {
  const pending = uploadQueue.value.filter(item => item.status === 'pending')
  if (pending.length === 0) return

  uploading.value = true
  let successCount = 0
  let errorCount = 0

  for (const item of pending) {
    item.status = 'uploading'
    item.percentage = 0

    try {
      await uploadFile(item.raw, {
        onUploadProgress: (progressEvent: any) => {
          const percent = progressEvent.total
            ? Math.round((progressEvent.loaded * 100) / progressEvent.total)
            : 0
          item.percentage = Math.max(percent, 1)
        },
      })
      item.status = 'done'
      item.percentage = 100
      successCount++
    } catch (err: any) {
      item.status = 'error'
      item.message = err.message || '上传失败'
      item.percentage = 100
      errorCount++
    }
  }

  uploading.value = false

  if (errorCount === 0) {
    ElMessage.success(`全部 ${successCount} 个文件上传成功`)
    uploadVisible.value = false
    loadData(query)
  } else {
    ElMessage.warning(`上传完成：${successCount} 成功，${errorCount} 失败`)
    uploadQueue.value = uploadQueue.value.filter(item => item.status !== 'done')
    if (successCount > 0) loadData(query)
  }
}

async function handleTogglePublic(row: FileItem, enabled: boolean) {
  const prev = row.isPublic
  row.isPublic = enabled ? 1 : 0
  try {
    await setFilePublic(row.id, enabled)
    ElMessage.success(enabled ? '已设为公开' : '已设为私有')
  } catch {
    row.isPublic = prev
  }
}

// ==================== 复制地址 / 下载 ====================
async function handleCopyUrl(row: FileItem) {
  const url = getFileAccessUrl(row.accessKey)
  try {
    await navigator.clipboard.writeText(url)
    ElMessage.success('地址已复制')
  } catch {
    // 降级：选区复制
    const input = document.createElement('input')
    input.value = url
    document.body.appendChild(input)
    input.select()
    try {
      document.execCommand('copy')
      ElMessage.success('地址已复制')
    } catch {
      ElMessage.error('复制失败，请手动复制：' + url)
    }
    document.body.removeChild(input)
  }
}

async function handleDownload(row: FileItem) {
  try {
    await downloadAuthFile(row.accessKey, row.name)
  } catch (e: any) {
    ElMessage.error(e?.message || '下载失败')
  }
}

// ==================== 删除 ====================
async function handleDelete(row: FileItem) {
  try {
    await ElMessageBox.confirm(`确认删除文件"${row.name}"？`, '删除确认', { type: 'warning' })
    await deleteFile(row.id)
    ElMessage.success('删除成功')
    loadData(query)
  } catch { /* 取消 */ }
}

function handleAction(row: FileItem, command: string) {
  if (command === 'delete') handleDelete(row)
}

function fileIcon(mimeType: string) {
  if (mimeType?.startsWith('video/')) return VideoCamera
  if (mimeType?.startsWith('audio/')) return Headset
  return Document
}

onMounted(() => loadData(query))
</script>

<style scoped lang="scss">
.upload-area {
  width: 100%;
}

.upload-queue {
  margin-top: 16px;
  max-height: 280px;
  overflow-y: auto;

  .upload-queue-item {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid #ebeef5;
    gap: 8px;

    &:last-child {
      border-bottom: none;
    }
  }

  .queue-info {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 200px;
    flex-shrink: 0;
  }

  .queue-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
    color: #303133;
  }

  .queue-size {
    font-size: 12px;
    color: #909399;
    flex-shrink: 0;
  }

  .queue-remove {
    cursor: pointer;
    color: #f56c6c;
    font-size: 16px;
    flex-shrink: 0;

    &:hover {
      color: #c45656;
    }
  }
}

.file-thumb-preview {
  width: 48px;
  height: 48px;
  border-radius: 4px;
  object-fit: cover;
  border: 1px solid #ebeef5;
  cursor: pointer;
}

.file-thumb-fallback {
  width: 48px;
  height: 48px;
  border-radius: 4px;
  background: #f0f2f5;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
}
</style>
