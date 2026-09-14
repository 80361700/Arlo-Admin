<template>
  <el-dialog
    :model-value="modelValue"
    title="评论"
    width="480px"
    destroy-on-close
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-position="top" @submit.prevent>
      <el-form-item label="@提醒">
        <el-select
          v-model="mentionIds"
          multiple
          filterable
          clearable
          placeholder="选择要 @ 的人（可选）"
          style="width: 100%"
          @change="onMentionChange"
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
      <el-form-item label="内容" required>
        <el-input
          v-model="content"
          type="textarea"
          :rows="4"
          maxlength="500"
          show-word-limit
          placeholder="请输入评论内容，可配合上方 @ 提醒相关人"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="loading" @click="onConfirm">发表</el-button>
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
  confirm: [payload: { content: string; mentionUserIds: number[] }]
}>()

const content = ref('')
const mentionIds = ref<number[]>([])
const options = ref<UserOption[]>([])
/** 已写入正文的 @姓名，避免重复追加 */
const insertedMentions = ref<Set<number>>(new Set())

function optionLabel(u: UserOption) {
  if (u.name && u.name !== u.username) return `${u.name}（${u.username}）`
  return u.name || u.username
}

function displayName(u: UserOption) {
  return u.name || u.username || String(u.id)
}

watch(
  () => props.modelValue,
  async (v) => {
    if (!v) return
    content.value = ''
    mentionIds.value = []
    insertedMentions.value = new Set()
    try {
      const res = await getAllUsers()
      options.value = res.data || []
    } catch {
      options.value = []
    }
  },
)

function onMentionChange(ids: number[]) {
  const map = new Map(options.value.map((u) => [u.id, u]))
  const next = new Set(ids)
  for (const id of ids) {
    if (insertedMentions.value.has(id)) continue
    const u = map.get(id)
    if (!u) continue
    const tag = `@${displayName(u)} `
    if (!content.value.includes(tag.trim())) {
      content.value = `${content.value}${content.value && !content.value.endsWith(' ') ? ' ' : ''}${tag}`
    }
    insertedMentions.value.add(id)
  }
  for (const id of [...insertedMentions.value]) {
    if (!next.has(id)) insertedMentions.value.delete(id)
  }
}

function onConfirm() {
  const text = content.value.trim()
  if (!text) {
    ElMessage.warning('请输入评论内容')
    return
  }
  emit('confirm', { content: text, mentionUserIds: [...mentionIds.value] })
}
</script>
