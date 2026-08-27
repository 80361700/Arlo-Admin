<template>
  <el-tree-select
    v-if="treeReady"
    v-model="selectedValue"
    :data="treeData"
    :props="treeProps"
    :placeholder="placeholder"
    :clearable="clearable"
    :disabled="disabled"
    :filterable="filterable"
    :multiple="multiple"
    :check-strictly="checkStrictly"
    v-bind="$attrs"
    @change="handleChange"
  />
  <el-select v-else :placeholder="placeholder" disabled style="width: 100%" />
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDeptTree } from '@/api'
import type { DeptTreeNode } from '@/api'

withDefaults(defineProps<{
  placeholder?: string
  clearable?: boolean
  disabled?: boolean
  filterable?: boolean
  multiple?: boolean
  /** 为 true 时可选任意节点（含父级）；false 时通常只能选叶子 */
  checkStrictly?: boolean
}>(), {
  placeholder: '请选择部门',
  clearable: true,
  filterable: true,
  // 用户/角色挂部门需要可选父级，默认放开父子关联限制
  checkStrictly: true,
})

const selectedValue = defineModel<number | number[] | undefined>()

const emit = defineEmits<{
  change: [value: number | number[] | undefined]
}>()

const treeData = ref<DeptTreeNode[]>([])
const treeReady = ref(false)

const treeProps = {
  label: 'name',
  value: 'id',
  children: 'children',
}

function handleChange(val: number | number[] | undefined) {
  emit('change', val)
}

onMounted(async () => {
  try {
    const res = await getDeptTree()
    treeData.value = res.data || []
  } catch {
    // 静默失败
  } finally {
    treeReady.value = true
  }
})
</script>
