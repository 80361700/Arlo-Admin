<template>
  <UserSelect
    v-model="model"
    :placeholder="placeholder"
    :clearable="clearable"
    :disabled="disabled"
    :filterable="filterable"
    @change="(v) => emit('change', v)"
  />
</template>

<script setup lang="ts">
import UserSelect from '@/components/UserSelect.vue'

withDefaults(
  defineProps<{
    placeholder?: string
    clearable?: boolean
    disabled?: boolean
    filterable?: boolean
  }>(),
  {
    placeholder: '请选择人员',
    clearable: true,
    filterable: true,
  },
)

const model = defineModel<number | undefined>()
const emit = defineEmits<{ change: [value: number | undefined] }>()
</script>
