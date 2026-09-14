<template>
  <div>
    <el-form ref="formRef" :model="props.data.node" :rules="rules" label-width="auto" label-position="top">
      <el-form-item label="节点名称" prop="nodeName">
        <el-input v-model="props.data.node.nodeName"></el-input>
      </el-form-item>
      <el-form-item label="节点key" prop="nodeKey">
        <el-input v-model="props.data.node.nodeKey"></el-input>
      </el-form-item>
      <el-form-item label="选择要抄送的人员" prop="nodeAssigneeList">
        <div class="selectBtn">
          <el-button icon="Plus" round type="primary" @click="handleUser">选择人员</el-button>
        </div>
        
        <div v-if="props.data?.node?.nodeAssigneeList?.length" class="tags">
          <el-tag v-for="(item, index) in props.data?.node?.nodeAssigneeList" :key="index" closable @close="userDelete(index)">{{item.name}}</el-tag>
        </div>
      </el-form-item>

      <el-form-item label="抄送配置">
        <el-checkbox v-model="props.data.node.allowSelection" label="允许发起人自选抄送人"></el-checkbox>
        <el-checkbox v-model="props.data.node.remind" label="抄送提醒"></el-checkbox>
      </el-form-item>
      
    </el-form>

    <user-select ref="userRef" @success="userSuccess"></user-select>
  </div>
</template>

<script setup lang="ts">

import { ref, defineProps, defineExpose } from "vue"
import userSelect from "../../common/user/index.vue"
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


const userRef = ref()
const handleUser = () => {
  userRef.value.init({
    selectData: props.data.node.nodeAssigneeList
  })
}
const userSuccess = (e: any) => {
  props.data.node.nodeAssigneeList = e
  console.log(e)
}
const userDelete = (i: number) => {
  props.data.node.nodeAssigneeList.splice(i, 1)
}

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