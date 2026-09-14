<template>
  <div class="node-condition-setting">
    <el-form ref="formRef" :model="props.data.node" :rules="rules" label-width="auto" label-position="top">
      <el-form-item label="节点名称" prop="nodeName">
        <el-input v-model="props.data.node.nodeName" />
      </el-form-item>
      <el-form-item label="节点key" prop="nodeKey">
        <el-input v-model="props.data.node.nodeKey" />
      </el-form-item>

      <el-form-item label="动态路由" prop="routeNodes" class="condition-form-item">
        <div
          v-for="(routeNode, routeIndex) in props.data.node.routeNodes"
          :key="routeIndex"
          class="routeNodes"
        >
          <div class="customForm">
            <div class="item">
              <span>路由名称</span>
              <el-input v-model="routeNode.nodeName" />
            </div>
            <div class="item">
              <span>路由节点</span>
              <el-select v-model="routeNode.nodeKey" placeholder="请选择流程节点">
                <el-option
                  v-for="value in routeNodes"
                  :key="value.nodeKey"
                  :label="value.nodeName"
                  :value="value.nodeKey"
                />
              </el-select>
            </div>
          </div>

          <div class="alert-wrap">
            <el-alert title="满足以下条件时进入当前分支" type="info" show-icon :closable="false" />
          </div>
          <condition-groups
            v-model="routeNode.conditionList"
            :process-form="props.flow?.processForm"
          >
            <template #extra>
              <el-button icon="Delete" type="danger" plain @click="deleteRouteNodes(routeIndex)">
                删除路由
              </el-button>
            </template>
          </condition-groups>
        </div>
        <div class="footer">
          <el-button icon="Plus" type="primary" plain @click="pushRouteNodes">添加路由分支</el-button>
        </div>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import config from '../../../config'
import { findNodesByType } from '../../common/utils'
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

props.data.node = Object.assign(
  JSON.parse(JSON.stringify(config.nodes.find((item: any) => item.type == props.data.node.type)?.config)),
  props.data.node,
)

if (!Array.isArray(props.data.node.routeNodes)) {
  props.data.node.routeNodes = []
}
props.data.node.routeNodes.forEach((route: any) => {
  if (!Array.isArray(route.conditionList)) route.conditionList = []
})

const routeNodes = ref(findNodesByType(props.flow.modelContent.nodeConfig))

const pushRouteNodes = () => {
  props.data.node.routeNodes.push({
    nodeName: `路由${props.data.node.routeNodes.length + 1}`,
    nodeKey: '',
    priorityLevel: 1,
    conditionList: [],
  })
  if (props.data.node.routeNodes.length == 1) {
    props.data.node.routeNodes[0].local_isEdit = false
  }
  updateLevel()
}

const deleteRouteNodes = (index: number) => {
  props.data.node.routeNodes.splice(index, 1)
  updateLevel()
}

const updateLevel = () => {
  props.data.node.routeNodes.forEach((item: any, index: number) => {
    item.priorityLevel = index + 1
  })
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

  .routeNodes {
    border: 1px solid #eee;
    padding: 10px;
    margin-bottom: 10px;
    width: 100%;
    box-sizing: border-box;
    position: relative;
  }

  .customForm {
    display: flex;
    justify-content: space-between;
    width: 100%;
    padding-bottom: 10px;
    .item {
      display: flex;
      flex-direction: column;
      width: 48%;
      span {
        width: 100px;
        text-align: left;
        color: #606266;
        margin-bottom: 4px;
      }
    }
  }

  .footer {
    margin-top: 4px;
  }
}
</style>
