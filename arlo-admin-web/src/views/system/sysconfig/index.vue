<template>
  <div class="page-container">
    <ProTable
      ref="proTableRef"
      :data="tableData"
      :loading="loading"
      :total="tableData.length"
      :search-fields="searchFields"
      :show-pagination="false"
      :action-width="140"
      @search="handleSearch"
      @reset="handleReset"
    >
      <template #toolbar>
        <el-button v-permission="'sys:sysconfig:add'" type="primary" @click="handleAdd">新增配置</el-button>
      </template>

      <el-table-column prop="name" label="配置名称" min-width="120" />
      <el-table-column prop="key" label="配置键" min-width="140" show-overflow-tooltip />
      <el-table-column label="配置值" min-width="180">
        <template #default="{ row }">
          <template v-if="row.type === 3">
            <el-tag :type="isTruthy(row.value) ? 'success' : 'info'" size="small">
              {{ isTruthy(row.value) ? '开启' : '关闭' }}
            </el-tag>
          </template>
          <template v-else-if="row.type === 4">
            <AuthFileImage
              v-if="row.value"
              :file-ref="row.value"
              fit="cover"
              img-style="width: 40px; height: 40px; border-radius: 4px"
            />
            <span v-else style="color: #999">未设置</span>
          </template>
          <span v-else class="value-text">{{ row.value }}</span>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="typeMap[row.type]?.tag" size="small">{{ typeMap[row.type]?.label }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
      <el-table-column prop="createdAt" label="创建时间" width="160" />

      <template #actions="{ row }">
        <el-button v-permission="'sys:sysconfig:edit'" type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button v-permission="'sys:sysconfig:delete'" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <ProFormDialog
      v-model="dialogVisible"
      :model="form"
      :title="dialogTitle"
      :rules="formRules"
      @submit="handleSubmit"
    >
      <el-form-item label="配置名称" prop="name">
        <el-input v-model="form.name" placeholder="如：系统名称" maxlength="64" />
      </el-form-item>
      <el-form-item label="配置键" prop="key">
        <el-input v-model="form.key" placeholder="如：sys.name" maxlength="64" />
      </el-form-item>
      <el-form-item label="配置类型" prop="type">
        <el-select v-model="form.type" placeholder="选择类型" style="width: 100%" @change="onTypeChange">
          <el-option label="文本" :value="1" />
          <el-option label="JSON" :value="2" />
          <el-option label="开关" :value="3" />
          <el-option label="图片" :value="4" />
        </el-select>
      </el-form-item>
      <el-form-item label="配置值" prop="value">
        <div v-if="form.type === 3" class="switch-row">
          <el-switch
            v-model="form.value"
            active-value="true"
            inactive-value="false"
            active-text="开启"
            inactive-text="关闭"
          />
        </div>
        <el-input
          v-else-if="form.type === 2"
          v-model="form.value"
          type="textarea"
          :rows="10"
          placeholder='请输入 JSON，例如 {"key":"value"}'
          class="json-editor"
          spellcheck="false"
        />
        <div v-else-if="form.type === 4" class="image-value">
          <div v-if="form.value" class="image-preview" @click="pickerVisible = true">
            <AuthFileImage
              :file-ref="form.value"
              :preview="false"
              fit="cover"
              img-style="width: 80px; height: 80px; border-radius: 6px"
            />
            <span class="image-clear" title="清除" @click.stop="form.value = ''">
              <el-icon :size="12"><Close /></el-icon>
            </span>
          </div>
          <el-button v-else type="primary" plain @click="pickerVisible = true">选择图片</el-button>
        </div>
        <el-input v-else v-model="form.value" placeholder="配置值" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="200" />
      </el-form-item>
    </ProFormDialog>

    <FilePicker
      v-model="pickerVisible"
      title="选择图片"
      mode="single"
      :accept-types="['image']"
      :max-count="1"
      @confirm="onImagePicked"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import { Close } from '@element-plus/icons-vue'
import type { FormRules } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import FilePicker from '@/components/FilePicker.vue'
import AuthFileImage from '@/components/AuthFileImage.vue'
import {
  getConfigList,
  createConfig,
  updateConfig,
  deleteConfig,
  type ConfigItem,
  type CreateConfigParams,
  type UpdateConfigParams,
} from '@/api/modules/sysconfig'
import type { FileItem } from '@/api'
import { useAppStore } from '@/stores/app'
import { toFileRef } from '@/utils/fileUrl'

const appStore = useAppStore()
const proTableRef = ref()
const loading = ref(false)
const tableData = ref<ConfigItem[]>([])
const pickerVisible = ref(false)
const query = reactive<{ name?: string; key?: string; type?: number }>({})

const searchFields = [
  { prop: 'name', label: '配置名称', placeholder: '请输入配置名称' },
  { prop: 'key', label: '配置键', placeholder: '请输入配置键' },
  {
    prop: 'type',
    label: '类型',
    type: 'select' as const,
    options: [
      { label: '文本', value: 1 },
      { label: 'JSON', value: 2 },
      { label: '开关', value: 3 },
      { label: '图片', value: 4 },
    ],
  },
]

const typeMap: Record<number, { label: string; tag: 'success' | 'warning' | 'info' | undefined }> = {
  1: { label: '文本', tag: undefined },
  2: { label: 'JSON', tag: 'warning' },
  3: { label: '开关', tag: 'success' },
  4: { label: '图片', tag: 'info' },
}

function isTruthy(v: string) {
  return ['true', '1', 'on', 'yes', '开启', '开'].includes(String(v).trim().toLowerCase())
}

async function loadData(params: Record<string, any> = {}) {
  loading.value = true
  try {
    const res = await getConfigList(params)
    tableData.value = res.data || []
  } catch (err: any) {
    showRequestError(err, '加载失败')
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) {
  Object.assign(query, { name: undefined, key: undefined, type: undefined }, p)
  loadData(query)
}

function handleReset() {
  Object.assign(query, { name: undefined, key: undefined, type: undefined })
  loadData({})
}

// ========== 弹窗 ==========
const dialogVisible = ref(false)
const dialogTitle = ref('新增配置')
const isEdit = ref(false)
const form = reactive({ id: 0, name: '', key: '', value: '', type: 1 as number, remark: '' })

const formRules = computed<FormRules>(() => ({
  name: [{ required: true, message: '请输入配置名称', trigger: 'blur' }],
  key: [{ required: true, message: '请输入配置键', trigger: 'blur' }],
  type: [{ required: true, message: '请选择配置类型', trigger: 'change' }],
  value:
    form.type === 4 || form.type === 3
      ? []
      : form.type === 2
        ? [
            {
              validator: (_r, v, cb) => {
                if (!v || !String(v).trim()) {
                  cb()
                  return
                }
                try {
                  JSON.parse(String(v))
                  cb()
                } catch {
                  cb(new Error('JSON 格式不正确'))
                }
              },
              trigger: 'blur',
            },
          ]
        : [{ required: true, message: '请输入配置值', trigger: 'blur' }],
}))

const defaultForm = { id: 0, name: '', key: '', value: '', type: 1 as number, remark: '' }

function onTypeChange(type: number) {
  if (type === 3) form.value = 'false'
  else if (type === 2) form.value = '{\n  \n}'
  else form.value = ''
}

function onImagePicked(files: FileItem[]) {
  if (files?.[0]?.id) {
    form.value = toFileRef(files[0])
  }
}

function handleAdd() {
  dialogTitle.value = '新增配置'
  isEdit.value = false
  Object.assign(form, defaultForm)
  dialogVisible.value = true
}

function handleEdit(row: ConfigItem) {
  dialogTitle.value = '编辑配置'
  isEdit.value = true
  let value = row.value
  if (row.type === 3) {
    value = isTruthy(row.value) ? 'true' : 'false'
  }
  Object.assign(form, { id: row.id, name: row.name, key: row.key, value, type: row.type, remark: row.remark })
  dialogVisible.value = true
}

async function handleSubmit() {
  try {
    if (isEdit.value) {
      const params: UpdateConfigParams = {
        id: form.id,
        name: form.name,
        key: form.key,
        value: form.value,
        type: form.type,
        remark: form.remark,
      }
      await updateConfig(params)
      ElMessage.success('更新成功')
    } else {
      const params: CreateConfigParams = {
        name: form.name,
        key: form.key,
        value: form.value,
        type: form.type,
        remark: form.remark,
      }
      await createConfig(params)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    await loadData(query)
    await appStore.loadPublicConfig(true)
  } catch (err: any) {
    showRequestError(err, '操作失败')
  }
}

async function handleDelete(row: ConfigItem) {
  try {
    await ElMessageBox.confirm(`确认删除配置"${row.name}"？`, '删除确认', { type: 'warning' })
    await deleteConfig(row.id)
    ElMessage.success('删除成功')
    await loadData(query)
    await appStore.loadPublicConfig(true)
  } catch {
    // 取消
  }
}

onMounted(() => loadData({}))
</script>

<style scoped lang="scss">
.page-container { padding: 0; }
.value-text {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.switch-row {
  display: flex;
  align-items: center;
  min-height: 32px;
}
.json-editor :deep(textarea) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
}
.image-value {
  display: flex;
  align-items: center;
}
.image-preview {
  position: relative;
  cursor: pointer;
  display: inline-block;
}
.image-clear {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}
</style>
