<template>
  <el-dialog
    v-model="dialogShow"
    title="选择人员"
    width="640px"
    destroy-on-close
    append-to-body
    :before-close="dialogClose"
  >
    <el-transfer
      v-model="selectionList"
      :data="data"
      :titles="['人员', '已选']"
      :props="{ key: 'key', label: 'label', disabled: 'disabled' }"
      filterable
      filter-placeholder="请输入搜索内容"
    />
    <template #footer>
      <div v-if="!dialogData.detail" class="dialog-footer">
        <el-button type="primary" @click="dialogSubmit">确定</el-button>
        <el-button @click="dialogClose">取消</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { getAllUsers } from '@/api'

const emits = defineEmits<{
  success: [list: { id: string | number; name: string }[]]
  close: []
}>()

interface Option {
  key: string
  label: string
  disabled: boolean
}

const data = ref<Option[]>([])
const selectionList = ref<string[]>([])
const dialogShow = ref(false)
const dialogData = ref<any>({})

async function init(obj: any = {}) {
  dialogData.value = obj
  dialogShow.value = true
  selectionList.value = (obj?.selectData || []).map((item: any) => String(item.id))
  const candidates = Array.isArray(obj?.candidates) ? obj.candidates : []
  if (candidates.length) {
    data.value = candidates.map((item: any) => ({
      key: String(item.id),
      label: item.name || String(item.id),
      disabled: false,
    }))
    return
  }
  try {
    const res = await getAllUsers()
    const list = res.data || []
    data.value = list.map((item: any) => ({
      key: String(item.id),
      label: item.name && item.name !== item.username ? `${item.name}（${item.username}）` : (item.name || item.username),
      disabled: Number(item.status) === 0,
    }))
  } catch {
    data.value = []
  }
}

function dialogClose() {
  dialogShow.value = false
  emits('close')
}

function dialogSubmit() {
  const list: { id: string | number; name: string }[] = []
  selectionList.value.forEach((item) => {
    const obj = data.value.find(i => i.key === String(item))
    if (obj) list.push({ id: obj.key, name: obj.label })
  })
  dialogShow.value = false
  emits('success', list)
}

defineExpose({ init })
</script>
