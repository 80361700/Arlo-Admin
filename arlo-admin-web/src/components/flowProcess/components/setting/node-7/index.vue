<template>
  <div>
    <el-alert
      type="info"
      :closable="false"
      show-icon
      title="到达节点后发起 HTTP 请求；失败将终止当前流程。未配置地址时跳过并继续。"
      style="margin-bottom: 12px"
    />
    <el-form ref="formRef" :model="props.data.node" :rules="rules" label-width="auto" label-position="top">
      <el-form-item label="节点名称" prop="nodeName">
        <el-input v-model="props.data.node.nodeName" />
      </el-form-item>
      <el-form-item label="节点key" prop="nodeKey">
        <el-input v-model="props.data.node.nodeKey" />
      </el-form-item>

      <el-form-item label="执行时机" prop="triggerType">
        <el-radio-group v-model="props.data.node.triggerType">
          <el-radio :value="1">立即执行</el-radio>
          <el-radio :value="2">延迟执行</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="Number(props.data.node.triggerType) === 2" label="延迟时长">
        <el-input v-model="number" clearable>
          <template #append>
            <el-select v-model="timer" placeholder="单位" style="width: 100px">
              <el-option label="天" value="d" />
              <el-option label="小时" value="h" />
              <el-option label="分钟" value="m" />
            </el-select>
          </template>
        </el-input>
        <div v-if="number" style="padding-top: 10px; width: 100%">
          <el-alert
            :title="`${number}${timer === 'd' ? '天' : timer === 'h' ? '小时' : '分钟'}后发起请求`"
            type="info"
            show-icon
            :closable="false"
          />
        </div>
      </el-form-item>

      <el-form-item label="请求地址" prop="extendConfig.trigger">
        <el-input
          v-model="props.data.node.extendConfig.trigger"
          clearable
          placeholder="https://example.com/webhook"
        />
      </el-form-item>

      <el-form-item label="请求方法">
        <el-select v-model="props.data.node.extendConfig.method" style="width: 100%">
          <el-option label="POST" value="POST" />
          <el-option label="GET" value="GET" />
          <el-option label="PUT" value="PUT" />
        </el-select>
      </el-form-item>

      <el-form-item label="附加参数（JSON，可选）">
        <el-input
          v-model="props.data.node.extendConfig.args"
          type="textarea"
          :rows="4"
          placeholder='例如 {"biz":"leave"}'
        />
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import config from '../../../config'

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
if (!props.data.node.extendConfig || typeof props.data.node.extendConfig !== 'object') {
  props.data.node.extendConfig = {}
}
if (props.data.node.triggerType == null) props.data.node.triggerType = 1
if (props.data.node.delayType == null) props.data.node.delayType = 1
if (!props.data.node.extendConfig.method) props.data.node.extendConfig.method = 'POST'
if (props.data.node.extendConfig.trigger == null) props.data.node.extendConfig.trigger = ''
if (props.data.node.extendConfig.args == null) props.data.node.extendConfig.args = ''

const number = ref('1')
const timer = ref('m')
if (props.data.node.extendConfig.time && typeof props.data.node.extendConfig.time === 'string') {
  const parts = String(props.data.node.extendConfig.time).split(':')
  if (parts.length >= 2 && ['d', 'h', 'm'].includes(parts[1])) {
    number.value = parts[0] || '1'
    timer.value = parts[1]
  }
}

watch([number, timer], () => {
  if (Number(props.data.node.triggerType) !== 2) return
  const n = String(number.value || '').trim() || '1'
  props.data.node.extendConfig.time = `${n}:${timer.value || 'm'}`
  props.data.node.delayType = 1
})

watch(
  () => props.data.node.triggerType,
  (v) => {
    if (Number(v) === 2) {
      const n = String(number.value || '').trim() || '1'
      props.data.node.extendConfig.time = `${n}:${timer.value || 'm'}`
      props.data.node.delayType = 1
    }
  },
)
</script>
