<template>
  <DeptTree
    v-model="model"
    :placeholder="placeholder"
    :clearable="clearable"
    :disabled="disabled"
    :filterable="filterable"
    :multiple="multiple"
    :check-strictly="checkStrictly"
    style="width: 100%"
    @change="(v) => emit('change', v)"
  />
</template>

<script setup lang="ts">
import DeptTree from '@/components/DeptTree.vue'

withDefaults(
  defineProps<{
    placeholder?: string
    clearable?: boolean
    disabled?: boolean
    filterable?: boolean
    multiple?: boolean
    checkStrictly?: boolean
  }>(),
  {
    placeholder: '请选择部门',
    clearable: true,
    filterable: true,
    multiple: false,
    checkStrictly: true,
  },
)

const model = defineModel<number | number[] | undefined>()
const emit = defineEmits<{ change: [value: number | number[] | undefined] }>()
</script>
