<template>
  <div class="create-approval-main">
    <div class="zoom-scale">
      <el-button icon="Plus" circle @click="scale < 1 ? (scale += 0.1) : 1.5" size="small" />
      <span>{{ Math.floor(scale * 100) }}%</span>
      <el-button icon="Minus" circle @click="scale > 0.2 ? (scale -= 0.1) : 0.5" size="small" />
    </div>

    <div class="canvas-scroll">
      <div class="sc-workflow-design">
        <div class="box-scale" :style="{ transform: `scale(${scale})` }">
          <node-list :parent="{}" :data="props.data.nodeConfig" @onEvent="nodeEvent" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import nodeList from './node.vue'

const props = defineProps<{
  data: any
}>()

const emit = defineEmits<{
  (e: 'onEvent', value: any, parent: any): void
}>()

const scale = ref(1)
const nodeEvent = (e: any, parent: any) => {
  emit('onEvent', e, parent)
}
</script>

<style lang="scss" scoped>
.create-approval-main {
  position: relative;
  flex: 1;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  background: #f6f8f9;
  border-radius: 4px;
}
.zoom-scale {
  position: absolute;
  top: 14px;
  right: 14px;
  z-index: 10;
  display: flex;
  align-items: center;
  background: rgba(255, 255, 255, 0.95);
  padding: 10px;
  border-radius: 4px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  span {
    padding: 0 10px;
    min-width: 40px;
    text-align: center;
  }
}
.canvas-scroll {
  height: 100%;
  overflow: auto;
}
.sc-workflow-design {
  min-height: 100%;
  background: #f6f8f9;
}
.box-scale {
  display: inline-block;
  position: relative;
  width: 100%;
  min-height: 100%;
  padding: 55px 100px;
  align-items: flex-start;
  justify-content: center;
  flex-wrap: wrap;
  min-width: -moz-min-content;
  min-width: min-content;
  transform-origin: 50% 0px 0px;
  background: #f6f8f9;
  box-sizing: border-box;
}
</style>
