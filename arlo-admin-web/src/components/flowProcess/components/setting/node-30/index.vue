<template>
  <div>
    <el-form ref="formRef" :model="props.data.node" :rules="rules" label-width="auto" label-position="top">
      <el-form-item label="节点名称" prop="nodeName">
        <el-input v-model="props.data.node.nodeName"></el-input>
      </el-form-item>
      <el-form-item label="节点key" prop="nodeKey">
        <el-input v-model="props.data.node.nodeKey"></el-input>
      </el-form-item>
      
    </el-form>
  </div>
</template>

<script setup lang="ts">

import { ref, defineProps, defineExpose } from "vue"
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


props.data.node = Object.assign(JSON.parse(JSON.stringify(config.nodes.find((item: any) => item.type == props.data.node.type)?.config)), props.data.node)
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