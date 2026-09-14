<template>
  <el-popover placement="right-start" :width="260" trigger="click" ref="nodeRef">
    <template #reference>
      <el-button icon="Plus" type="primary" circle />
    </template>
    <div class="node-btns">
      <div v-for="item in nodes" :key="item.type" class="node-btns-item" @click="handleNode(item)">
        <div class="icon">
          <el-icon :color="item.color" size="18">
            <component :is="item.icon"/>
          </el-icon>
        </div>
        <span>{{ item.name }}</span>
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { ref, defineEmits } from "vue"
import config from "../../config"

const emit = defineEmits<{
  (e: "onEvent", value: any): void,
}>()

const nodes = config.nodes
const nodeRef = ref()
const handleNode = (e: any) => {
  nodeRef.value.hide()
  emit('onEvent', e)
}
</script>

<style lang="scss" scoped>
.node-btns {
  display: flex;
  flex-wrap: wrap;
  &-item {
    width: 48px;
    margin: 10px 15px;
    display: flex;
    justify-content: center;
    flex-direction: column;
    align-items: center;
    cursor: pointer;
    &:hover {
      .icon {
        border: 1px solid #409eff;
        background-color: #409eff;
        i {
          color: white;
        }
      }
    }
    .icon {
      width: 40px;
      height: 40px;
      border: 1px solid #e4e7ed;
      border-radius: 20px;
      box-sizing: border-box;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    span {
      font-size: 12px;
      color: #606266;
      padding-top: 5px;
    }
  }
}
</style>