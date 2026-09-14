import type { ComponentConfigModel } from 'epic-designer'

/** 属性面板专用：字典类型下拉（不出现在组件库） */
const config: ComponentConfigModel = {
  component: () => import('./component.vue'),
  bindModel: 'modelValue',
  defaultSchema: {
    label: '字典编码',
    type: 'arlo-dict-code',
  },
  config: {},
}

export default config
