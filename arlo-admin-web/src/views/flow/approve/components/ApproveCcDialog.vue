<template>
  <el-dialog
    :model-value="modelValue"
    title="抄送"
    width="480px"
    destroy-on-close
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-position="top" @submit.prevent>
      <el-form-item required label="抄送给">
        <el-select
          v-model="userIds"
          multiple
          filterable
          clearable
          placeholder="请选择抄送人"
          style="width: 100%"
        >
          <el-option
            v-for="u in options"
            :key="u.id"
            :label="optionLabel(u)"
            :value="u.id"
            :disabled="u.status !== 1"
          />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="loading" @click="onConfirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getAllUsers, type UserOption } from '@/api'

const props = withDefaults(
  defineProps<{ modelValue: boolean; loading?: boolean }>(),
  { loading: false },
)

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  confirm: [payload: { users: { id: number; name: string }[] }]
}>()

const userIds = ref<number[]>([])
const options = ref<UserOption[]>([])

function optionLabel(u: UserOption) {
  if (u.name && u.name !== u.username) return `${u.name}（${u.username}）`
  return u.name || u.username
}

watch(
  () => props.modelValue,
  async (v) => {
    if (!v) return
    userIds.value = []
    try {
      const res = await getAllUsers()
      options.value = res.data || []
    } catch {
      options.value = []
    }
  },
)

function onConfirm() {
  if (!userIds.value.length) {
    ElMessage.warning('请选择抄送人')
    return
  }
  const map = new Map(options.value.map((u) => [u.id, u]))
  const users = userIds.value.map((id) => {
    const u = map.get(id)
    return { id, name: u?.name || u?.username || '' }
  })
  emit('confirm', { users })
}
</script>
