<template>
  <div class="form-list">
    <ProTable
      :data="data"
      :loading="loading"
      :total="data.length"
      :search-fields="searchFields"
      :show-index="false"
      :show-pagination="false"
      :selection="true"
      :action-width="180"
      @search="handleSearch"
      @reset="handleReset"
      @selection-change="(rows: any[]) => (selected = rows)"
    >
      <template #toolbar>
        <el-button v-permission="'flow:form:add'" type="primary" :icon="Plus" @click="openEdit()">新增</el-button>
        <el-button
          v-permission="'flow:form:delete'"
          type="danger"
          plain
          :disabled="!selected.length"
          @click="handleBatchDelete"
        >
          批量删除
        </el-button>
      </template>

      <el-table-column prop="name" label="模板名称" min-width="160" show-overflow-tooltip />
      <el-table-column prop="code" label="编码" min-width="120" show-overflow-tooltip />
      <el-table-column label="类型" width="110" align="center">
        <template #default="{ row }">
          <el-tag v-if="Number(row.formType) === 2" size="small" type="warning" effect="plain">系统表单</el-tag>
          <el-tag v-else size="small" effect="plain">设计表单</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="110" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.bound" type="primary" effect="plain">已绑定</el-tag>
          <el-switch
            v-else
            v-permission="'flow:form:edit'"
            :model-value="row.status === 1"
            @change="(v: string | number | boolean) => handleState(row, !!v)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ row.remark || '-' }}</template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="80" align="center" />

      <template #actions="{ row }">
        <el-button
          v-if="Number(row.formType) !== 2"
          v-permission="'flow:form:edit'"
          type="primary"
          link
          size="small"
          @click="handleDesign(row)"
        >
          设计
        </el-button>
        <el-button
          v-permission="'flow:form:edit'"
          type="primary"
          link
          size="small"
          @click="openEdit(row)"
        >
          编辑
        </el-button>
        <el-button
          v-permission="'flow:form:delete'"
          type="danger"
          link
          size="small"
          :disabled="row.bound"
          @click="handleDelete(row)"
        >
          删除
        </el-button>
      </template>
    </ProTable>

    <el-dialog
      v-model="editVisible"
      :title="editForm.formId ? '编辑表单模板' : '新增表单模板'"
      width="520"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form ref="editRef" :model="editForm" :rules="editRules" label-width="80px">
        <el-form-item label="表单分类" prop="categoryId">
          <el-select v-model="editForm.categoryId" placeholder="请选择" style="width: 100%">
            <el-option v-for="c in categoryOptions" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="editForm.name" maxlength="128" clearable />
        </el-form-item>
        <el-form-item label="模板编码" prop="code">
          <el-input v-model="editForm.code" maxlength="64" clearable :disabled="!!editForm.formId" />
        </el-form-item>
        <el-form-item label="类型" prop="formType">
          <el-radio-group v-model="editForm.formType" :disabled="!!editForm.formId">
            <el-radio :value="1">设计表单</el-radio>
            <el-radio :value="2">系统表单</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="editForm.formType === 2" label="组件路径" prop="pcUrl">
          <el-input
            v-model="editForm.pcUrl"
            maxlength="255"
            clearable
            placeholder="相对 views，如 business/demo/form"
          />
          <div class="field-tip">对应前端文件 src/views/&lt;路径&gt;.vue，需实现 validate / getData / setData</div>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <div class="status-row">
            <span>禁用</span>
            <el-switch v-model="editForm.status" :active-value="1" :inactive-value="0" />
            <span>正常</span>
          </div>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="editForm.sort" :min="0" :max="9999" controls-position="right" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="editForm.remark" type="textarea" maxlength="255" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitEdit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import {
  deleteFlowForm,
  getFlowFormCategoryOptions,
  saveFlowForm,
  updateFlowFormState,
  type FlowFormCategoryOption,
} from '@/api'

const emits = defineEmits<{ event: [data: any] }>()

const loading = ref(false)
const defaultData = ref<any>({ formList: [] })
const data = ref<any[]>([])
const selected = ref<any[]>([])
const filters = ref<{ name?: string; status?: number | '' }>({})

const searchFields = computed(() => [
  { prop: 'name', label: '模板名称', placeholder: '请输入' },
  {
    prop: 'status',
    label: '状态',
    type: 'select' as const,
    options: [
      { label: '正常', value: 1 },
      { label: '禁用', value: 0 },
    ],
  },
])

const editVisible = ref(false)
const saving = ref(false)
const editRef = ref<FormInstance>()
const categoryOptions = ref<FlowFormCategoryOption[]>([])
const editForm = reactive({
  formId: undefined as number | undefined,
  categoryId: undefined as number | undefined,
  name: '',
  code: '',
  formType: 1 as number,
  pcUrl: '',
  status: 1 as number,
  sort: 0,
  remark: '',
})
const editRules: FormRules = {
  categoryId: [{ required: true, message: '请选择分类', trigger: 'change' }],
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入模板编码', trigger: 'blur' }],
  pcUrl: [
    {
      validator: (_r, v, cb) => {
        if (editForm.formType === 2 && !String(v || '').trim()) {
          cb(new Error('请填写组件路径'))
          return
        }
        cb()
      },
      trigger: 'blur',
    },
  ],
  sort: [{ required: true, message: '请输入排序', trigger: 'change' }],
}

function applyFilter() {
  let list = defaultData.value.formList || []
  const name = (filters.value.name || '').trim()
  if (name) list = list.filter((item: any) => String(item.name).includes(name))
  if (filters.value.status === 0 || filters.value.status === 1) {
    list = list.filter((item: any) => item.status === filters.value.status)
  }
  data.value = list
  selected.value = []
}

function handleSearch(p: Record<string, any>) {
  filters.value = {
    name: p.name,
    status: p.status,
  }
  applyFilter()
}

function handleReset() {
  filters.value = {}
  applyFilter()
}

async function openEdit(row: any = {}) {
  const res = await getFlowFormCategoryOptions()
  categoryOptions.value = res.data || []
  editForm.formId = row.formId
  editForm.categoryId =
    row.categoryId ||
    (defaultData.value?.id && defaultData.value.id !== '-1'
      ? Number(defaultData.value.id)
      : categoryOptions.value[0]?.id)
  editForm.name = row.name || ''
  editForm.code = row.code || ''
  editForm.formType = Number(row.formType) === 2 ? 2 : 1
  editForm.pcUrl = row.pcUrl || ''
  editForm.status = row.status === 0 ? 0 : 1
  editForm.sort = Number(row.sort) || 0
  editForm.remark = row.remark || ''
  editVisible.value = true
}

async function submitEdit() {
  await editRef.value?.validate()
  saving.value = true
  try {
    await saveFlowForm({
      formId: editForm.formId,
      categoryId: editForm.categoryId,
      name: editForm.name,
      code: editForm.code,
      formType: editForm.formType,
      pcUrl: editForm.formType === 2 ? editForm.pcUrl : '',
      status: editForm.status,
      sort: editForm.sort,
      remark: editForm.remark,
    })
    ElMessage.success('保存成功')
    editVisible.value = false
    emits('event', { type: 'refresh', row: defaultData.value })
  } finally {
    saving.value = false
  }
}

function handleDesign(row: any) {
  emits('event', { type: 'design', row })
}

async function handleState(row: any, enabled: boolean) {
  const status = enabled ? 1 : 0
  await updateFlowFormState(row.formId, status)
  ElMessage.success('操作成功')
  emits('event', { type: 'refresh', row: defaultData.value })
}

function handleDelete(row: any) {
  if (row.bound) {
    ElMessage.warning('表单已被流程引用，无法删除')
    return
  }
  ElMessageBox.confirm('确定将选择数据删除?', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await deleteFlowForm(row.formId)
      ElMessage.success('操作成功!')
      emits('event', { type: 'refresh', row: defaultData.value })
    })
    .catch(() => {})
}

function handleBatchDelete() {
  const rows = selected.value.filter((r) => !r.bound)
  if (!rows.length) {
    ElMessage.warning('请选择可删除的表单')
    return
  }
  ElMessageBox.confirm(`确定删除选中的 ${rows.length} 条数据?`, {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      for (const row of rows) {
        await deleteFlowForm(row.formId)
      }
      ElMessage.success('操作成功!')
      emits('event', { type: 'refresh', row: defaultData.value })
    })
    .catch(() => {})
}

function reload(row: any) {
  defaultData.value = row || { formList: [] }
  defaultData.value.formList = defaultData.value.formList || []
  applyFilter()
}

defineExpose({ reload })
</script>

<style lang="scss" scoped>
.form-list {
  height: 100%;
  overflow: auto;
}
.status-row {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--el-text-color-regular);
}
.field-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
}
</style>
