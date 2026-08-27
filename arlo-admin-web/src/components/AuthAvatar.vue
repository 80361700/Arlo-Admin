<template>
  <el-avatar
    :size="size"
    :src="url || undefined"
    :icon="icon"
    :shape="shape"
  />
</template>

<script setup lang="ts">
import { useAuthFileSrc } from '@/composables/useAuthFileSrc'
import { UserFilled } from '@element-plus/icons-vue'
import type { Component } from 'vue'
import { toRef } from 'vue'

const props = withDefaults(
  defineProps<{
    fileRef?: string | null
    size?: number
    shape?: 'circle' | 'square'
    icon?: Component
  }>(),
  {
    fileRef: '',
    size: 32,
    shape: 'circle',
    icon: UserFilled,
  },
)

const url = useAuthFileSrc(toRef(props, 'fileRef'))
</script>
