<template>
  <UploadField
    v-model="model"
    kind="file"
    :multiple="multiple"
    :limit="limit"
    :disabled="disabled"
    @change="(v) => emit('change', v)"
  />
</template>

<script setup lang="ts">
import UploadField from './component.vue'

withDefaults(
  defineProps<{
    multiple?: boolean
    limit?: number
    disabled?: boolean
  }>(),
  {
    multiple: false,
    limit: 1,
  },
)

const model = defineModel<string>({ default: '' })
const emit = defineEmits<{ change: [value: string] }>()
</script>
