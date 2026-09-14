<template>
  <div>
    <el-form ref="formRef" :model="props.data.node" :rules="rules" label-width="auto" label-position="top">
      <el-form-item label="节点名称" prop="nodeName">
        <el-input v-model="props.data.node.nodeName"></el-input>
      </el-form-item>
      <el-form-item label="节点key" prop="nodeKey">
        <el-input v-model="props.data.node.nodeKey"></el-input>
      </el-form-item>
      <el-form-item label="谁可以发起审批" prop="nodeAssigneeList">
        <div class="selectBtn">
          <el-button icon="Plus" round type="primary" @click="handleRole">选择角色</el-button>
        </div>
        
        <div v-if="props.data?.node?.nodeAssigneeList?.length" class="tags">
          <el-tag v-for="(item, index) in props.data?.node?.nodeAssigneeList" :key="index" closable @close="roleDelete(index)">{{item.name}}</el-tag>
        </div>
        <div v-else style="padding-top: 10px; width: 100%;">
          <el-alert title="不指定则默认所有人都可发起此审批" type="info" show-icon :closable="false" />
        </div>

      </el-form-item>
      <el-form-item label="表单权限">
        <div class="table">
          <el-table :data="props.data?.node?.extendConfig?.formConfig" border style="width: 100%">
            <el-table-column prop="label" label="表单字段" />
            <el-table-column prop="opera" label="操作" width="210">
              <template #default="scope">
                <el-radio-group v-model="scope.row.opera" size="small">
                  <el-radio :value="0">只读</el-radio>
                  <el-radio :value="1">编辑</el-radio>
                  <el-radio :value="2">隐藏</el-radio>
                </el-radio-group>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-form-item>
    </el-form>

    <role-select ref="roleRef" @success="roleSuccess"></role-select>
  </div>
</template>

<script setup lang="ts">

import { ref, defineProps, defineExpose, watch } from "vue"
import roleSelect from "../../common/role/index.vue"
import { getFormConfig } from "../../common/utils"

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


props.data.node = Object.assign({
  extendConfig: {
    formConfig: [],
  },
  nodeAssigneeList: [],
}, props.data.node)
let formConfig = getFormConfig(
  props.flow.processForm,
  props.data.node.extendConfig.formConfig,
  1,
)
// 发起人全「只读」基本不可用（填不了单）；视为历史默认误配，纠正为可编辑
if (formConfig.length && formConfig.every((c: any) => Number(c.opera) === 0)) {
  formConfig = formConfig.map((c: any) => ({ ...c, opera: 1 }))
}
props.data.node.extendConfig.formConfig = formConfig

watch(
  () => props.flow?.processForm,
  () => {
    if (!props.data?.node?.extendConfig) return
    let next = getFormConfig(
      props.flow.processForm,
      props.data.node.extendConfig.formConfig,
      1,
    )
    if (next.length && next.every((c: any) => Number(c.opera) === 0)) {
      next = next.map((c: any) => ({ ...c, opera: 1 }))
    }
    props.data.node.extendConfig.formConfig = next
  },
  { deep: true },
)

const roleRef = ref()
const handleRole = () => {
  roleRef.value.init({
    selectData: props.data.node.nodeAssigneeList
  })
}
const roleSuccess = (e: any) => {
  props.data.node.nodeAssigneeList = e
  console.log(e)
}
const roleDelete = (i: number) => {
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