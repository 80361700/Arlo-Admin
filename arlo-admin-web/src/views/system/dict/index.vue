<template>
  <div class="page-container">
    <ProTable
      ref="tableRef"
      :data="typeList"
      :loading="typeLoading"
      :total="typeTotal"
      :search-fields="typeSearchFields"
      :show-index="false"
      @search="handleTypeSearch"
      @reset="handleTypeReset"
      @page-change="handleTypePageChange"
    >
      <template #toolbar>
        <el-button v-permission="'sys:dict:add'" type="primary" @click="handleAddType">新增类型</el-button>
      </template>

      <el-table-column prop="name" label="字典名称" min-width="140" />
      <el-table-column prop="code" label="字典编码" min-width="140" />
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />

      <template #actions="{ row }">
        <el-button v-permission="'sys:dict:edit'" type="primary" link size="small" @click="handleEditType(row)">编辑</el-button>
        <el-button v-permission="'sys:dict:config'" type="primary" link size="small" @click="handleTypeAction(row, 'config')">字典配置</el-button>
        <el-dropdown v-if="authStore.hasPermission('sys:dict:delete')" trigger="click" @command="(cmd: string) => handleTypeAction(row, cmd)">
          <el-button type="info" link size="small">
            更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="delete">删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
    </ProTable>

    <!-- 字典类型弹窗 -->
    <ProFormDialog
      ref="typeDialogRef"
      v-model="typeDialogVisible"
      :title="typeDialogTitle"
      :model="typeForm"
      :rules="typeFormRules"
      :submitting="typeSubmitting"
      @submit="handleTypeSubmit"
    >
      <template #default>
        <el-form-item label="字典名称" prop="name">
          <el-input v-model="typeForm.name" placeholder="请输入字典名称" maxlength="64" />
        </el-form-item>
        <el-form-item label="字典编码" prop="code">
          <el-input v-model="typeForm.code" placeholder="请输入字典编码" maxlength="64" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="typeForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="typeForm.remark" type="textarea" placeholder="请输入备注" :rows="2" />
        </el-form-item>
      </template>
    </ProFormDialog>

    <!-- 字典配置弹窗 -->
    <el-dialog
      v-model="configDialogVisible"
      :title="`${configType?.name || ''} 字典配置`"
      width="1000px"
      top="5vh"
      destroy-on-close
      @closed="configType = null"
    >
      <div class="config-toolbar">
        <el-button v-permission="'sys:dict:add'" type="primary" @click="handleAddData">新增数据</el-button>
        <div class="config-search">
          <el-input
            v-model="dataSearchLabel"
            placeholder="搜索字典标签"
            clearable
            size="default"
            style="width: 200px"
            @input="debounceSearchData"
          />
        </div>
      </div>

      <el-table :data="dataList" v-loading="dataLoading" border class="config-table">
        <el-table-column prop="label" label="字典标签" min-width="120" show-overflow-tooltip />
        <el-table-column prop="value" label="字典值" min-width="100" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="70" align="center" />
        <el-table-column label="默认" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.isDefault === 1 ? 'warning' : 'info'" size="small">
              {{ row.isDefault === 1 ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button v-permission="'sys:dict:edit'" type="primary" link size="small" @click="handleEditData(row as DictDataItem)">编辑</el-button>
              <el-button v-if="authStore.hasPermission('sys:dict:delete')" type="danger" link size="small" @click="handleDeleteData(row as DictDataItem)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="dataTotal > 0"
        v-model:current-page="dataQuery.page"
        v-model:page-size="dataQuery.pageSize"
        :total="dataTotal"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        class="config-pagination"
        @size-change="loadData"
        @current-change="loadData"
      />
    </el-dialog>

    <!-- 字典数据弹窗 -->
    <ProFormDialog
      ref="dataDialogRef"
      v-model="dataDialogVisible"
      :title="dataDialogTitle"
      :model="dataForm"
      :rules="dataFormRules"
      :submitting="dataSubmitting"
      @submit="handleDataSubmit"
    >
      <template #default>
        <el-form-item label="字典标签" prop="label">
          <el-input v-model="dataForm.label" placeholder="请输入字典标签" maxlength="64" />
        </el-form-item>
        <el-form-item label="字典值" prop="value">
          <el-input v-model="dataForm.value" placeholder="请输入字典值" maxlength="64" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="dataForm.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="是否默认">
          <el-radio-group v-model="dataForm.isDefault">
            <el-radio :value="0">否</el-radio>
            <el-radio :value="1">是</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="dataForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="dataForm.remark" type="textarea" placeholder="请输入备注" :rows="2" />
        </el-form-item>
      </template>
    </ProFormDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import { ArrowDown } from '@element-plus/icons-vue'
import type { FormRules } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import {
  getDictTypeList, getDictTypeDetail, createDictType, updateDictType, deleteDictType,
  getDictDataList, getDictDataDetail, createDictData, updateDictData, deleteDictData,
} from '@/api'
import type { DictTypeItem, DictTypeListParams, DictDataItem, DictDataListParams } from '@/api'
import { clearDictCache } from '@/utils/useDict'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

// ==================== 字典类型列表 ====================
const tableRef = ref()
const typeList = ref<DictTypeItem[]>([])
const typeLoading = ref(false)
const typeTotal = ref(0)
const typeQuery = reactive<DictTypeListParams>({ page: 1, pageSize: 10 })

const typeSearchFields = [
  { prop: 'name', label: '字典名称' },
  { prop: 'code', label: '字典编码' },
]

async function loadTypes() {
  typeLoading.value = true
  try {
    const res = await getDictTypeList(typeQuery)
    typeList.value = res.data.list || []
    typeTotal.value = res.data.total
  } finally {
    typeLoading.value = false
  }
}

function handleTypeSearch(p: any) { typeQuery.page = 1; Object.assign(typeQuery, p); loadTypes() }
function handleTypeReset() { typeQuery.page = 1; Object.assign(typeQuery, { page: 1, pageSize: 10, name: '', code: '' }); loadTypes() }
function handleTypePageChange(p: any) { Object.assign(typeQuery, p); loadTypes() }

// ==================== 字典类型 CRUD ====================
const typeDialogVisible = ref(false)
const typeDialogTitle = ref('新增类型')
const isEditType = ref(false)
const typeSubmitting = ref(false)
const typeDialogRef = ref()

const typeDefaultForm = { name: '', code: '', status: 1 as number, remark: '' }
const typeForm = reactive({ ...typeDefaultForm })

const typeFormRules: FormRules = {
  name: [{ required: true, message: '请输入字典名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入字典编码', trigger: 'blur' }],
}

function handleAddType() {
  isEditType.value = false
  typeDialogTitle.value = '新增字典类型'
  Object.assign(typeForm, { ...typeDefaultForm })
  typeDialogVisible.value = true
}

async function handleEditType(row: DictTypeItem) {
  isEditType.value = true
  typeDialogTitle.value = '编辑字典类型'
  const res = await getDictTypeDetail(row.id)
  const d = res.data
  Object.assign(typeForm, { id: d.id, name: d.name, code: d.code, status: d.status, remark: d.remark })
  typeDialogVisible.value = true
}

async function handleTypeSubmit() {
  typeSubmitting.value = true
  try {
    const p = { name: typeForm.name, code: typeForm.code, status: typeForm.status, remark: typeForm.remark }
    if (isEditType.value) {
      await updateDictType({ id: (typeForm as any).id, ...p })
      ElMessage.success('更新成功')
    } else {
      await createDictType(p)
      ElMessage.success('新增成功')
    }
    typeDialogVisible.value = false
    clearDictCache(typeForm.code)
    loadTypes()
  } catch (err: any) {
    showRequestError(err, '操作失败')
  } finally {
    typeSubmitting.value = false
  }
}

async function handleTypeAction(row: DictTypeItem, cmd: string) {
  if (cmd === 'config') {
    openConfigDialog(row)
  } else if (cmd === 'delete') {
    try {
      await ElMessageBox.confirm('确认删除该字典类型？删除后字典数据也会被清空。', '提示', { type: 'warning' })
      await deleteDictType(row.id)
      ElMessage.success('删除成功')
      clearDictCache(row.code)
      loadTypes()
    } catch (err: any) {
      showRequestError(err, '删除失败')
    }
  }
}

// ==================== 字典配置弹窗（字典数据管理） ====================
const configDialogVisible = ref(false)
const configType = ref<DictTypeItem | null>(null)

function openConfigDialog(row: DictTypeItem) {
  configType.value = row
  dataQuery.page = 1
  dataQuery.pageSize = 10
  dataQuery.dictTypeId = row.id
  dataSearchLabel.value = ''
  configDialogVisible.value = true
  loadData()
}

// ==================== 字典数据列表 ====================
const dataList = ref<DictDataItem[]>([])
const dataLoading = ref(false)
const dataTotal = ref(0)
const dataQuery = reactive<DictDataListParams>({ page: 1, pageSize: 10 })
const dataSearchLabel = ref('')

let dataSearchTimer: ReturnType<typeof setTimeout> | null = null
function debounceSearchData() {
  if (dataSearchTimer) clearTimeout(dataSearchTimer)
  dataSearchTimer = setTimeout(() => {
    dataQuery.page = 1
    dataQuery.label = dataSearchLabel.value.trim() || undefined
    loadData()
  }, 300)
}

async function loadData() {
  if (!dataQuery.dictTypeId) return
  dataLoading.value = true
  try {
    const res = await getDictDataList(dataQuery)
    dataList.value = res.data.list || []
    dataTotal.value = res.data.total
  } finally {
    dataLoading.value = false
  }
}

// ==================== 字典数据 CRUD ====================
const dataDialogVisible = ref(false)
const dataDialogTitle = ref('新增数据')
const isEditData = ref(false)
const dataSubmitting = ref(false)
const dataDialogRef = ref()

const dataDefaultForm = {
  label: '', value: '', sort: 0,
  isDefault: 0 as number, status: 1 as number, remark: '',
}
const dataForm = reactive({ ...dataDefaultForm })

const dataFormRules: FormRules = {
  label: [{ required: true, message: '请输入字典标签', trigger: 'blur' }],
  value: [{ required: true, message: '请输入字典值', trigger: 'blur' }],
}

function handleAddData() {
  isEditData.value = false
  dataDialogTitle.value = '新增字典数据'
  Object.assign(dataForm, { ...dataDefaultForm })
  dataDialogVisible.value = true
}

async function handleEditData(row: DictDataItem) {
  isEditData.value = true
  dataDialogTitle.value = '编辑字典数据'
  const res = await getDictDataDetail(row.id)
  const d = res.data
  Object.assign(dataForm, {
    id: d.id, label: d.label, value: d.value, sort: d.sort,
    isDefault: d.isDefault,
    status: d.status, remark: d.remark,
  })
  dataDialogVisible.value = true
}

async function handleDataSubmit() {
  dataSubmitting.value = true
  try {
    const p = {
      dictTypeId: configType.value!.id,
      label: dataForm.label, value: dataForm.value, sort: dataForm.sort,
      isDefault: dataForm.isDefault, status: dataForm.status, remark: dataForm.remark,
    }
    if (isEditData.value) {
      await updateDictData({ id: (dataForm as any).id, ...p })
      ElMessage.success('更新成功')
    } else {
      await createDictData(p)
      ElMessage.success('新增成功')
    }
    dataDialogVisible.value = false
    clearDictCache(configType.value?.code)
    loadData()
  } catch (err: any) {
    showRequestError(err, '操作失败')
  } finally {
    dataSubmitting.value = false
  }
}

async function handleDeleteData(row: DictDataItem) {
  try {
    await ElMessageBox.confirm('确认删除该字典数据？', '提示', { type: 'warning' })
    await deleteDictData(row.id)
    ElMessage.success('删除成功')
    clearDictCache(configType.value?.code)
    loadData()
  } catch (err: any) {
    showRequestError(err, '删除失败')
  }
}

onMounted(() => {
  loadTypes()
})
</script>

<style scoped lang="scss">
.config-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;

  .config-search {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.config-table {
  width: 100%;
}

.config-pagination {
  margin-top: 12px;
  justify-content: flex-end;
}
</style>
