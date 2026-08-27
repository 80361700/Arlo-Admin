<template>
  <el-image
    v-if="url"
    :src="url"
    :preview-src-list="preview ? [url] : undefined"
    :fit="fit"
    :class="imgClass"
    :style="imgStyle"
    hide-on-click-modal
    preview-teleported
    :z-index="zIndex"
  >
    <template v-if="$slots.error" #error>
      <slot name="error" />
    </template>
  </el-image>
</template>

<script setup lang="ts">
import { useAuthFileSrc } from '@/composables/useAuthFileSrc'
import { toRef } from 'vue'

type ImageFit = '' | 'contain' | 'cover' | 'fill' | 'none' | 'scale-down'

const props = withDefaults(
  defineProps<{
    fileRef?: string | null
    fit?: ImageFit
    preview?: boolean
    imgClass?: string
    imgStyle?: string | Record<string, string>
    zIndex?: number
  }>(),
  {
    fileRef: '',
    fit: 'cover',
    preview: true,
    zIndex: 10000,
  },
)

const url = useAuthFileSrc(toRef(props, 'fileRef'))
</script>
