<template>
  <div class="left-group">
    <div class="left-group-header">
      <div class="left-group-title">
        <span>流程分类</span>
        <div class="title-actions">
          <el-button
            v-permission="'flow:category:manage'"
            class="header-action-btn is-icon"
            size="small"
            title="新增分类"
            @click="addGroup()"
          >
            <el-icon :size="14"><Plus /></el-icon>
          </el-button>
          <el-button
            class="header-action-btn is-icon"
            size="small"
            title="刷新"
            @click="initList()"
          >
            <el-icon :size="14"><Refresh /></el-icon>
          </el-button>
        </div>
      </div>
      <div class="left-group-input">
        <el-input v-model="categoryName" placeholder="输入关键字进行过滤" clearable />
      </div>
    </div>
    <div class="group-content">
      <li :class="liIndex === '-1' ? 'li-color' : ''" @click="groupView({ id: '-1' })">
        <span>全部</span>
      </li>
      <li
        v-for="item in groupList"
        :key="item.id"
        :class="liIndex == item.id ? 'li-color' : ''"
        @click="groupView(item)"
      >
        <span>{{ item.name }}</span>
        <div class="oprate">
          <el-button
            v-permission="'flow:category:manage'"
            text
            type="primary"
            :icon="Edit"
            @click.stop="addGroup(item)"
          />
          <el-button
            v-permission="'flow:category:manage'"
            text
            type="danger"
            :icon="Delete"
            @click.stop="delGroup(item)"
          />
        </div>
      </li>
    </div>
  </div>

  <el-dialog
    v-model="groupModel"
    :title="groupData.id ? '编辑分类' : '添加分类'"
    width="500"
    :close-on-click-modal="false"
  >
    <el-form ref="modelRef" :model="groupData" :rules="groupRules" label-width="80px">
      <el-form-item label="分类名称" prop="name">
        <el-input v-model="groupData.name" maxlength="64" clearable />
      </el-form-item>
      <el-form-item label="备注" prop="remark">
        <el-input v-model="groupData.remark" type="textarea" maxlength="200" show-word-limit />
      </el-form-item>
      <el-form-item label="排序" prop="sort">
        <el-input-number v-model="groupData.sort" :min="0" :max="9999" controls-position="right" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="modelClose">取消</el-button>
      <el-button type="primary" @click="modelSubmit">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Edit, Delete, Plus, Refresh } from '@element-plus/icons-vue'
import {
  getFlowCategoryTree,
  createFlowCategory,
  updateFlowCategory,
  deleteFlowCategory,
} from '@/api'

const emits = defineEmits<{ event: [row: any] }>()

const categoryName = ref('')
const defaultGroupList = ref<any[]>([])
const groupList = ref<any[]>([])
const liIndex = ref('-1')

async function initList(row: any = {}) {
  const res = await getFlowCategoryTree()
  const list = (res.data || []).map((item: any) => ({
    id: item.categoryId,
    name: item.categoryName,
    remark: item.categoryRemark,
    sort: item.categorySort,
    processList: item.processList || [],
  }))
  defaultGroupList.value = list
  groupList.value = JSON.parse(JSON.stringify(list))
  if (row.id && row.id != '-1') {
    const hit = groupList.value.find((item: any) => item.id == row.id)
    groupView(hit || { id: '-1' })
  } else {
    groupView({ id: '-1' })
  }
}

initList()

watch(categoryName, (val) => {
  if (val) {
    groupList.value = defaultGroupList.value.filter((item: any) => item.name.indexOf(val) >= 0)
  } else {
    groupList.value = JSON.parse(JSON.stringify(defaultGroupList.value))
  }
})

function delGroup(row: any) {
  if (row.processList?.length) {
    ElMessage.warning('请先删除关联流程')
    return
  }
  ElMessageBox.confirm('确定将选择数据删除?', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await deleteFlowCategory(row.id)
      ElMessage.success('操作成功!')
      initList()
    })
    .catch(() => {})
}

function groupView(row: any) {
  liIndex.value = String(row.id)
  if (row.id == '-1') {
    const all = { id: '-1', processList: [] as any[] }
    defaultGroupList.value.forEach((item: any) => {
      if (item.processList) all.processList.push(...item.processList)
    })
    emits('event', all)
    return
  }
  emits('event', row)
}

const groupModel = ref(false)
const modelRef = ref<FormInstance>()
const groupData = ref<{ id: number | ''; name: string; remark: string; sort: number }>({
  id: '',
  name: '',
  remark: '',
  sort: 0,
})
const groupRules: FormRules = {
  name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
}

function addGroup(row: any = { id: '', name: '', remark: '', sort: 0 }) {
  groupData.value = {
    id: row.id || '',
    name: row.name || '',
    remark: row.remark || '',
    sort: Number(row.sort) || 0,
  }
  groupModel.value = true
}

async function modelSubmit() {
  await modelRef.value?.validate()
  if (groupData.value.id) {
    await updateFlowCategory({
      id: Number(groupData.value.id),
      name: groupData.value.name,
      remark: groupData.value.remark,
      sort: groupData.value.sort,
    })
    ElMessage.success('编辑成功!')
  } else {
    await createFlowCategory({
      name: groupData.value.name,
      remark: groupData.value.remark,
      sort: groupData.value.sort,
    })
    ElMessage.success('新增成功!')
  }
  groupModel.value = false
  initList()
}

function modelClose() {
  groupModel.value = false
}

defineExpose({ initList })
</script>

<style lang="scss" scoped>
.left-group {
  border-right: 1px solid var(--el-border-color-lighter);
  padding-right: 16px;
  height: calc(100% - 8px);
  overflow: auto;
  position: relative;
  background: var(--el-bg-color);
  &-header {
    background: var(--el-bg-color);
    position: sticky;
    top: 0;
    z-index: 2;
    padding-bottom: 16px;
  }
  &-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 10px;
    span {
      font-size: 16px;
      font-weight: 500;
    }
    .title-actions {
      display: flex;
      align-items: center;
      gap: 6px;
      .header-action-btn {
        margin: 0;
        height: 28px;
        padding: 0 10px;
        font-size: 13px;
        font-weight: 400;
        color: var(--el-text-color-regular);
        background: #fff;
        border: 1px solid var(--el-border-color);
        &:hover,
        &:focus {
          color: var(--el-color-primary);
          border-color: var(--el-color-primary-light-5);
          background: var(--el-color-primary-light-9);
        }
        &.is-icon {
          width: 28px;
          padding: 0;
          display: inline-flex;
          align-items: center;
          justify-content: center;
        }
      }
    }
  }
  &-input {
    display: flex;
  }
  .group-content {
    li {
      height: 34px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-radius: 4px;
      cursor: pointer;
      list-style: none;
      padding: 0 10px;
      box-sizing: border-box;
      span {
        display: inline-block;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        color: var(--el-text-color-primary);
        font-size: 14px;
      }
      .oprate {
        display: flex;
        align-items: center;
        .el-button {
          margin: 0;
          padding-right: 0;
        }
      }
    }
    .li-color {
      background: rgba(96, 98, 102, 0.05);
    }
  }
}
</style>
