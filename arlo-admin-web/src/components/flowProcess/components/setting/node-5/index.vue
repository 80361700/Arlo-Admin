<template>
  <div>
    <el-form ref="formRef" :model="props.data.node" :rules="rules" label-width="auto" label-position="top">
      <el-form-item label="节点名称" prop="nodeName">
        <el-input v-model="props.data.node.nodeName" />
      </el-form-item>
      <el-form-item label="节点key" prop="nodeKey">
        <el-input v-model="props.data.node.nodeKey" />
      </el-form-item>

      <el-form-item label="子流程配置" prop="subProcessValue">
        <el-select
          v-model="props.data.node.subProcessValue"
          placeholder="请选择已启用的子流程"
          clearable
          filterable
          style="width: 100%"
        >
          <el-option
            v-for="item in childData"
            :key="item.id"
            :label="`${item.name}（${item.key}）`"
            :value="String(item.id)"
          />
        </el-select>
        <div class="tips">
          仅展示已启用的其它子流程（不含当前流程）；父流程会同步等待子流程结束后继续
        </div>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import config from '../../../config'
import { getFlowProcessOptions } from '@/api'

const props = defineProps<{
  data: any
  flow: any
}>()

const rules = ref({
  nodeName: [{ required: true, message: '请输入节点名称', trigger: 'blur' }],
  nodeKey: [{ required: true, message: '请输入节点key', trigger: 'blur' }],
})

const formRef = ref()
const getFormRef = () => formRef.value
defineExpose({ getFormRef })

props.data.node = Object.assign(
  JSON.parse(JSON.stringify(config.nodes.find((item: any) => item.type == props.data.node.type)?.config)),
  props.data.node,
)
// 子流程节点不做人表单权限（无人工办理）；清理历史误写的 formConfig / callAsync
if (props.data.node.extendConfig) {
  delete props.data.node.extendConfig.formConfig
  if (Object.keys(props.data.node.extendConfig).length === 0) {
    delete props.data.node.extendConfig
  }
}
delete props.data.node.callAsync

const childData = ref<{ id: number; name: string; key: string }[]>([])

onMounted(async () => {
  const excludeId = props.flow?.processId ? Number(props.flow.processId) : undefined
  try {
    const res = await getFlowProcessOptions(excludeId)
    const list = res.data || []
    childData.value = excludeId
      ? list.filter((it) => Number(it.id) !== excludeId)
      : list
  } catch {
    childData.value = []
  }
  if (props.data.node.subProcessValue != null && props.data.node.subProcessValue !== '') {
    props.data.node.subProcessValue = String(props.data.node.subProcessValue)
  }
  if (excludeId && String(props.data.node.subProcessValue) === String(excludeId)) {
    props.data.node.subProcessValue = ''
    props.data.node.callProcess = ''
  }
})

watch(
  () => props.data.node.subProcessValue,
  () => {
    const obj = childData.value.find((it) => String(it.id) === String(props.data.node.subProcessValue))
    props.data.node.callProcess = obj ? `${obj.id}:${obj.name}` : ''
  },
)
</script>

<style lang="scss" scoped>
.tips {
  font-size: 13px;
  color: #999;
  line-height: 20px;
  margin-top: 6px;
}
</style>
