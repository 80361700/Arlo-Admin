<template>
  <el-select
    v-model="model"
    :placeholder="placeholder"
    :clearable="clearable"
    :disabled="disabled || !ready"
    :filterable="filterable"
    :multiple="multiple"
    teleported
    style="width: 100%"
    @change="(v: any) => emit('change', v)"
  >
    <el-option
      v-for="r in options"
      :key="r.id"
      :label="r.name"
      :value="r.id"
      :disabled="Number(r.status) === 0"
    />
  </el-select>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getAllRoles, type RoleItem } from '@/api'

withDefaults(
  defineProps<{
    placeholder?: string
    clearable?: boolean
    disabled?: boolean
    filterable?: boolean
    multiple?: boolean
  }>(),
  {
    placeholder: '请选择角色',
    clearable: true,
    filterable: true,
    multiple: false,
  },
)

const model = defineModel<number | number[] | undefined>()
const emit = defineEmits<{ change: [value: number | number[] | undefined] }>()

const options = ref<RoleItem[]>([])
const ready = ref(false)

onMounted(async () => {
  try {
    const res = await getAllRoles()
    options.value = res.data || []
  } catch {
    options.value = []
  } finally {
    ready.value = true
  }
})
</script>
