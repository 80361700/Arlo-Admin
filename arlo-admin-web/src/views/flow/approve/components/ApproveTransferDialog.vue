<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="480px"
    destroy-on-close
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-position="top" @submit.prevent>
      <el-form-item
        v-if="pendingActors.length > 1"
        required
        label="转办对象"
      >
        <el-select v-model="fromActorId" placeholder="请选择要转办的办理人" style="width: 100%">
          <el-option
            v-for="a in pendingActors"
            :key="a.id"
            :label="a.actorName || `用户#${a.actorId}`"
            :value="a.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item required :label="title === '转办' ? '转办给' : '转交给'">
        <UserSelect v-model="toUserId" :placeholder="title === '转办' ? '请选择转办人' : '请选择转交人'" />
      </el-form-item>
      <el-form-item :label="title === '转办' ? '转办意见' : '转交意见'">
        <el-input
          v-model="opinion"
          type="textarea"
          :rows="3"
          maxlength="64"
          show-word-limit
          placeholder="选填"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="loading" @click="onConfirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import UserSelect from '@/components/UserSelect.vue'
import { getAllUsers } from '@/api'

export type TransferActorOption = {
  id: number
  actorId?: number
  actorName?: string
  actorState?: number
  actorType?: number
}

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    loading?: boolean
    title?: string
    /** 管理员转办：当前节点待办人（多人时需选择） */
    actors?: TransferActorOption[]
  }>(),
  { loading: false, title: '转交', actors: () => [] },
)

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  confirm: [payload: { toUserId: number; toName: string; opinion: string; fromActorId?: number }]
}>()

const toUserId = ref<number>()
const fromActorId = ref<number>()
const opinion = ref('')
const nameMap = ref<Record<number, string>>({})

const pendingActors = computed(() =>
  (props.actors || []).filter((a) => Number(a.actorState ?? 0) === 0 && a.id > 0),
)

watch(
  () => props.modelValue,
  async (v) => {
    if (!v) return
    toUserId.value = undefined
    opinion.value = ''
    const list = pendingActors.value
    fromActorId.value = list.length === 1 ? list[0].id : undefined
    try {
      const res = await getAllUsers()
      const map: Record<number, string> = {}
      for (const u of res.data || []) {
        map[u.id] = u.name || u.username
      }
      nameMap.value = map
    } catch {
      /* ignore */
    }
  },
)

function onConfirm() {
  if (pendingActors.value.length > 1 && !fromActorId.value) {
    ElMessage.warning('请选择要转办的办理人')
    return
  }
  if (!toUserId.value) {
    ElMessage.warning(props.title === '转办' ? '请选择转办人' : '请选择转交人')
    return
  }
  emit('confirm', {
    toUserId: toUserId.value,
    toName: nameMap.value[toUserId.value] || '',
    opinion: opinion.value.trim(),
    fromActorId: fromActorId.value || undefined,
  })
}
</script>
