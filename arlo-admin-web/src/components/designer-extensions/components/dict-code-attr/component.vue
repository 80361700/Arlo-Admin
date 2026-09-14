<template>
  <el-select
    v-model="model"
    filterable
    clearable
    teleported
    placeholder="请选择字典"
    style="width: 100%"
  >
    <el-option
      v-for="item in options"
      :key="item.code"
      :label="`${item.name}（${item.code}）`"
      :value="item.code"
    />
  </el-select>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getAllDictTypes } from '@/api/modules/system'

const model = defineModel<string | undefined>()

const options = ref<{ id: number; name: string; code: string }[]>([])

onMounted(async () => {
  try {
    // 用仅需登录的 options 接口，避免表单设计依赖「字典管理」菜单权限
    const res = await getAllDictTypes()
    options.value = res.data || []
  } catch {
    options.value = []
  }
})
</script>
