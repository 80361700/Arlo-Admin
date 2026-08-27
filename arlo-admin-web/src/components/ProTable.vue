<template>
  <div class="pro-table">
    <!-- 筛选栏 -->
    <div v-if="searchFields.length" class="filter-panel">
      <div class="filter-header" @click="filterCollapsed = !filterCollapsed">
        <el-icon class="filter-icon"><Filter /></el-icon>
        <span>筛选</span>
        <el-icon class="filter-arrow" :class="{ 'is-collapsed': filterCollapsed }">
          <ArrowUp />
        </el-icon>
      </div>

      <el-form
        v-show="!filterCollapsed"
        :model="searchForm"
        class="filter-form"
        label-position="right"
        @submit.prevent="handleSearch"
      >
        <div class="filter-grid">
          <el-form-item
            v-for="field in searchFields"
            :key="field.prop"
            :label="field.label + '：'"
            class="filter-item"
          >
            <slot
              :name="`search-${field.prop}`"
              :value="searchForm[field.prop]"
              :update="(v: any) => searchForm[field.prop] = v"
            >
              <template v-if="field.type === 'select'">
                <el-select
                  v-model="searchForm[field.prop]"
                  :placeholder="field.placeholder || '请选择'"
                  clearable
                  style="width: 160px"
                >
                  <el-option
                    v-for="opt in field.options || []"
                    :key="opt.value"
                    :label="opt.label"
                    :value="opt.value"
                  />
                </el-select>
              </template>
              <template v-else-if="field.type === 'date'">
                <el-date-picker
                  v-model="searchForm[field.prop]"
                  type="date"
                  :placeholder="field.placeholder || '请选择日期'"
                  clearable
                  style="width: 160px"
                />
              </template>
              <template v-else-if="field.type === 'datetimerange'">
                <el-date-picker
                  v-model="searchForm[field.prop]"
                  type="datetimerange"
                  range-separator="—"
                  start-placeholder="开始日期"
                  end-placeholder="结束日期"
                  value-format="YYYY-MM-DD HH:mm:ss"
                  clearable
                  style="width: 360px"
                />
              </template>
              <template v-else>
                <el-input
                  v-model="searchForm[field.prop]"
                  :placeholder="field.placeholder || '请输入'"
                  clearable
                  style="width: 160px"
                />
              </template>
            </slot>
          </el-form-item>

          <div class="filter-actions">
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </div>
        </div>
      </el-form>
    </div>

    <!-- 工具栏 -->
    <div v-if="$slots.toolbar" class="toolbar">
      <slot name="toolbar" />
    </div>

    <!-- 表格 -->
    <el-table
      ref="tableRef"
      :data="data"
      :border="border"
      :stripe="stripe"
      v-loading="loading"
      v-bind="$attrs"
      @selection-change="handleSelectionChange"
    >
      <el-table-column
        v-if="selection"
        type="selection"
        width="50"
        fixed="left"
      />
      <el-table-column
        v-if="showIndex"
        type="index"
        label="序号"
        width="60"
        :index="indexMethod"
      />
      <slot />
      <el-table-column
        v-if="$slots.actions"
        label="操作"
        :width="actionWidth"
        fixed="right"
        align="center"
      >
        <template #default="scope">
          <div class="action-buttons">
            <slot name="actions" :row="(scope.row as T)" :index="scope.$index" />
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div v-if="showPagination" class="pagination-wrap">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="currentPageSize"
        :page-sizes="pageSizes"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts" generic="T extends Record<string, any> = Record<string, any>">
import { ref, reactive, watch } from 'vue'
import { Filter, ArrowUp } from '@element-plus/icons-vue'

export interface SearchField {
  prop: string
  label: string
  type?: 'input' | 'select' | 'date' | 'datetimerange'
  placeholder?: string
  options?: { label: string; value: any }[]
}

export interface PaginationParams {
  page: number
  pageSize: number
}

const props = withDefaults(defineProps<{
  data: T[]
  loading?: boolean
  total?: number
  searchFields?: SearchField[]
  selection?: boolean
  showIndex?: boolean
  showPagination?: boolean
  border?: boolean
  stripe?: boolean
  actionWidth?: number
  pageSizes?: number[]
  defaultPageSize?: number
}>(), {
  data: () => [],
  loading: false,
  total: 0,
  searchFields: () => [],
  selection: false,
  showIndex: true,
  showPagination: true,
  border: true,
  stripe: true,
  actionWidth: 180,
  pageSizes: () => [10, 20, 50, 100],
  defaultPageSize: 10,
})

const emit = defineEmits<{
  search: [params: Record<string, any>]
  reset: []
  pageChange: [params: PaginationParams & Record<string, any>]
  selectionChange: [rows: T[]]
}>()

defineSlots<{
  default?: () => any
  toolbar?: () => any
  actions?: (props: { row: T; index: number }) => any
  [name: `search-${string}`]: (props: { value: any; update: (v: any) => void }) => any
}>()

const filterCollapsed = ref(false)
const currentPage = ref(1)
const currentPageSize = ref(props.defaultPageSize)
const searchForm = reactive<Record<string, any>>({})

watch(() => props.searchFields, (fields) => {
  fields.forEach((f) => {
    if (!(f.prop in searchForm)) {
      searchForm[f.prop] = undefined
    }
  })
}, { immediate: true })

function indexMethod(index: number) {
  return (currentPage.value - 1) * currentPageSize.value + index + 1
}

function getSearchParams() {
  const params: Record<string, any> = {}
  props.searchFields.forEach((f) => {
    const v = searchForm[f.prop]
    if (v === undefined || v === null || v === '' || (Array.isArray(v) && v.length === 0)) {
      params[f.prop] = undefined
    } else {
      params[f.prop] = v
    }
  })
  return params
}

function handleSearch() {
  currentPage.value = 1
  emit('search', getSearchParams())
}

function handleReset() {
  props.searchFields.forEach((f) => {
    searchForm[f.prop] = undefined
  })
  currentPage.value = 1
  emit('reset')
}

function handlePageChange() {
  emit('pageChange', {
    page: currentPage.value,
    pageSize: currentPageSize.value,
    ...getSearchParams(),
  })
}

function handleSizeChange() {
  currentPage.value = 1
  emit('pageChange', {
    page: currentPage.value,
    pageSize: currentPageSize.value,
    ...getSearchParams(),
  })
}

function handleSelectionChange(rows: T[]) {
  emit('selectionChange', rows)
}

defineExpose({
  currentPage,
  currentPageSize,
  searchForm,
  getSearchParams,
  refresh() {
    handleSearch()
  },
  resetPage() {
    currentPage.value = 1
  },
})
</script>

<style scoped lang="scss">
.pro-table {
  .filter-panel {
    margin-bottom: 12px;
    padding: 6px 16px 0;
    background: var(--el-fill-color-light);
  }

  .filter-header {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    height: 28px;
    margin-bottom: 4px;
    font-size: 13px;
    color: #606266;
    cursor: pointer;
    user-select: none;

    &:hover {
      color: #409eff;
    }
  }

  .filter-icon {
    font-size: 14px;
  }

  .filter-arrow {
    font-size: 12px;
    transition: transform 0.2s ease;

    &.is-collapsed {
      transform: rotate(180deg);
    }
  }

  .filter-form {
    padding-bottom: 0;
  }

  .filter-grid {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0 8px;
  }

  .filter-item {
    margin: 0 16px 12px 0;
    width: auto;

    :deep(.el-form-item__label) {
      color: #606266;
      font-weight: 400;
      padding-right: 0;
    }

    :deep(.el-form-item__content) {
      flex: none;
    }
  }

  .filter-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 12px 0;

    :deep(.el-button + .el-button) {
      margin-left: 0;
    }
  }

  .toolbar {
    display: flex;
    gap: 0;
    margin-bottom: 12px;
  }

  .pagination-wrap {
    display: flex;
    justify-content: flex-end;
    border: 1px solid #ebeef5;
    border-top: none;
    padding: 12px;
  }

  .action-buttons {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;

    :deep(.el-button + .el-button) {
      margin-left: 0;
    }
  }
}
</style>
