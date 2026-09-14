<template>
  <div class="node-condition-setting">
    <el-form ref="formRef" :model="props.data.node" :rules="rules" label-width="auto" label-position="top">
      <el-form-item label="节点名称" prop="nodeName">
        <el-input v-model="props.data.node.nodeName" />
      </el-form-item>
      <el-form-item label="节点key" prop="nodeKey">
        <el-input v-model="props.data.node.nodeKey" />
      </el-form-item>

      <el-form-item label="条件分组" prop="conditionList" class="condition-form-item">
        <div class="alert-wrap">
          <el-alert title="满足以下条件时进入当前分支" type="info" show-icon :closable="false" />
        </div>
        <condition-groups
          v-model="props.data.node.conditionList"
          :process-form="props.flow?.processForm"
        />
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import conditionGroups from '../../common/condition/index.vue'

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

props.data.node = Object.assign({
  priorityLevel: 1,
  conditionList: [],
}, props.data.node)

if (!Array.isArray(props.data.node.conditionList)) {
  props.data.node.conditionList = []
}
</script>

<style lang="scss" scoped>
.node-condition-setting {
  width: 100%;

  .condition-form-item {
    :deep(.el-form-item__content) {
      display: block;
      width: 100%;
    }
  }

  .alert-wrap {
    padding-bottom: 10px;
    width: 100%;
  }
}
</style>
