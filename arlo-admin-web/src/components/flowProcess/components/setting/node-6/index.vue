<template>
  <div>
    <el-form ref="formRef" :model="props.data.node" :rules="rules" label-width="auto" label-position="top">
      <el-form-item label="节点名称" prop="nodeName">
        <el-input v-model="props.data.node.nodeName"></el-input>
      </el-form-item>
      <el-form-item label="节点key" prop="nodeKey">
        <el-input v-model="props.data.node.nodeKey"></el-input>
      </el-form-item>

      <el-form-item label="延迟等待" prop="delayType">
        <el-radio-group v-model="props.data.node.delayType">
          <el-radio :value="1">固定时长</el-radio>
          <el-radio :value="2">自动计算</el-radio>
        </el-radio-group>

        <div v-if="props.data.node.delayType == 1" style="width: 100%; padding-top: 10px;">
          <el-input v-model="number" clearable>
            <template #append>
              <el-select v-model="timer" placeholder="单位" style="width: 100px">
                <el-option label="天" value="d" />
                <el-option label="小时" value="h" />
                <el-option label="分钟" value="m" />
              </el-select>
            </template>
          </el-input>
          <div v-if="number" style="padding-top: 10px; width: 100%;">
            <el-alert :title="`${number}${timer == 'd' ? '天' : (timer == 'h' ? '小时' : '分钟')}后进入下一步`" type="info" show-icon :closable="false" />
          </div>
        </div>

        <div v-else style="width: 100%; padding-top: 10px;">
          <el-time-picker v-model="props.data.node.extendConfig.time" placeholder="选择时间" value-format="HH:mm:ss" style="width: 100%" />
          <div v-if="props.data.node.extendConfig.time" style="padding-top: 10px; width: 100%;">
            <el-alert :title="`到达每天 ${props.data.node.extendConfig.time}（已过则次日）后进入下一步`" type="info" show-icon :closable="false" />
          </div>
        </div>
      </el-form-item>
      
    </el-form>
  </div>
</template>

<script setup lang="ts">

import { ref, defineProps, defineExpose, watch } from "vue"
import config from "../../../config"

const props = defineProps<{ 
  data: any, 
  flow: any,
}>()

const rules = ref({
  nodeName: [
    { required: true, message: '请输入节点名称', trigger: 'blur' },
  ],
  nodeKey: [
    { required: true, message: '请输入节点key', trigger: 'blur' },
  ],
})

const formRef = ref()
const getFormRef = () => {
  return formRef.value
}

defineExpose({
  getFormRef,
})

const number = ref("")
const timer = ref("m")

watch(() => props.data.node.delayType, (n , o) => {
  if (n == 1) {
    props.data.node.extendConfig.time = "1:m"
    number.value = '1'
    timer.value = 'm'
  }
  else {
    props.data.node.extendConfig.time = ""
  }
})
watch(() => number.value, (n, o) => {
  if (props.data.node.delayType == 1 && n) {
    props.data.node.extendConfig.time = `${n}:${timer.value}`
  }
})
watch(() => timer.value, (n, o) => {
  if (props.data.node.delayType == 1) {
    props.data.node.extendConfig.time = `${number.value}:${n}`
  }
})


props.data.node = Object.assign(JSON.parse(JSON.stringify(config.nodes.find((item: any) => item.type == props.data.node.type)?.config)), props.data.node)
if (props.data.node.delayType == 1 && props.data.node.extendConfig.time) {
  let strs = props.data.node.extendConfig.time.split(':')
  number.value = strs[0]
  timer.value = strs[1]
}
console.log("node-setting", props.data.node)


</script>

<style lang="scss" scoped>

.tags {
  width: 100%;
  padding-top: 10px;
  :deep(.el-tag) {
    margin-right: 10px;
  }
  .tips {
    font-size: 14px;
    color: #999;
  }
}
.table {
  width: 100%;
  :deep(.el-radio) {
    margin-right: 15px;
  }
}
.selectBtn {
  display: flex;
  .tips {
    font-size: 14px;
    padding-left: 10px;
    color: #999;
  }
}


</style>