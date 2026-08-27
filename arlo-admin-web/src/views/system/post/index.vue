<template>
  <div class="page-container">
    <ProTable
      ref="proTableRef"
      :data="tableData"
      :loading="loading"
      :total="total"
      :search-fields="searchFields"
      :show-index="false"
      @search="handleSearch"
      @reset="handleReset"
      @page-change="handlePageChange"
    >
      <template #toolbar>
        <el-button v-permission="'sys:post:add'" type="primary" @click="handleAdd">新增岗位</el-button>
      </template>

      <el-table-column prop="id" label="ID" width="80" align="center" />
      <el-table-column prop="code" label="岗位编码" min-width="140" />
      <el-table-column prop="name" label="岗位名称" min-width="140" />
      <el-table-column prop="sort" label="排序" width="80" align="center" />
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ statusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
      <el-table-column prop="createdAt" label="创建时间" min-width="170" />

      <template #actions="{ row }">
        <el-button v-permission="'sys:post:edit'" type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-popconfirm v-if="authStore.hasPermission('sys:post:delete')" title="确认删除该岗位？" @confirm="handleDelete(row.id)">
          <template #reference>
            <el-button type="danger" link size="small">删除</el-button>
          </template>
        </el-popconfirm>
      </template>
    </ProTable>

    <ProFormDialog
      ref="formDialogRef"
      v-model="dialogVisible"
      :title="dialogTitle"
      :model="form"
      :rules="formRules"
      :submitting="submitting"
      @submit="handleSubmit"
    >
      <template #default>
        <el-form-item label="岗位编码" prop="code">
          <el-input v-model="form.code" placeholder="请输入岗位编码" maxlength="32" />
        </el-form-item>
        <el-form-item label="岗位名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入岗位名称" maxlength="64" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio v-for="opt in statusOptions" :key="String(opt.value)" :value="opt.value">
              {{ opt.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" placeholder="请输入备注" :rows="2" />
        </el-form-item>
      </template>
    </ProFormDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import type { FormRules } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { getPostList, getPostDetail, createPost, updatePost, deletePost } from '@/api'
import type { PostItem, PostListParams } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useDict, DictCode } from '@/utils/useDict'

const authStore = useAuthStore()
const { options: statusOptions, getLabel: statusLabel } = useDict(DictCode.UserStatus)
const proTableRef = ref()
const tableData = ref<PostItem[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive<PostListParams>({ page: 1, pageSize: 10 })

const searchFields = computed(() => [
  { prop: 'code', label: '岗位编码' },
  { prop: 'name', label: '岗位名称' },
  { prop: 'status', label: '状态', type: 'select' as const, options: statusOptions.value },
])

async function loadData(params: PostListParams) {
  loading.value = true
  try {
    const res = await getPostList(params)
    tableData.value = res.data.list || []
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function handleSearch(p: any) { query.page = 1; Object.assign(query, p); loadData(query) }
function handleReset() { loadData({ page: 1, pageSize: 10 }) }
function handlePageChange(p: any) { Object.assign(query, p); loadData(query) }

const dialogVisible = ref(false)
const dialogTitle = ref('新增岗位')
const isEdit = ref(false)
const submitting = ref(false)
const formDialogRef = ref()

const defaultForm = { code: '', name: '', sort: 0, status: 1 as number, remark: '' }
const form = reactive({ ...defaultForm })

const formRules: FormRules = {
  code: [{ required: true, message: '请输入岗位编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入岗位名称', trigger: 'blur' }],
}

function handleAdd() {
  isEdit.value = false
  dialogTitle.value = '新增岗位'
  Object.assign(form, { ...defaultForm })
  dialogVisible.value = true
}

async function handleEdit(row: PostItem) {
  isEdit.value = true
  dialogTitle.value = '编辑岗位'
  const res = await getPostDetail(row.id)
  const p = res.data
  Object.assign(form, { id: p.id, code: p.code, name: p.name, sort: p.sort, status: p.status, remark: p.remark })
  dialogVisible.value = true
}

async function handleSubmit() {
  submitting.value = true
  try {
    const payload = { code: form.code, name: form.name, sort: form.sort, status: form.status, remark: form.remark }
    if (isEdit.value) {
      await updatePost({ id: (form as any).id, ...payload })
      ElMessage.success('更新成功')
    } else {
      await createPost(payload)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    loadData(query)
  } catch (err: any) {
    showRequestError(err, '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await deletePost(id)
    ElMessage.success('删除成功')
    loadData(query)
  } catch (err: any) {
    showRequestError(err, '删除失败')
  }
}

onMounted(() => loadData(query))
</script>
