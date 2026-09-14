import type { ComponentConfigModel } from 'epic-designer'

const config: ComponentConfigModel = {
  component: () => import('./component.vue'),
  bindModel: 'modelValue',
  groupName: '表单',
  icon: 'icon--epic--wysiwyg-rounded',
  sort: 910,
  defaultSchema: {
    label: '富文本',
    type: 'arlo-rich-editor',
    field: 'content',
    input: true,
    props: {
      placeholder: '请输入内容...',
      height: 280,
    },
  },
  config: {
    attribute: [
      { field: 'field', label: '数据字段', type: 'EpField' },
      { field: 'label', label: '标题', type: 'input' },
      { field: 'props.placeholder', label: '占位内容', type: 'input' },
      { field: 'props.height', label: '高度', type: 'number' },
      { field: 'props.disabled', label: '禁用', type: 'switch' },
    ],
    event: [{ type: 'change', description: '内容变化时触发' }],
  },
}

export default config
