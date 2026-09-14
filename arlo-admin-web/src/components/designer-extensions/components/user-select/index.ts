import type { ComponentConfigModel } from 'epic-designer'

const config: ComponentConfigModel = {
  component: () => import('./component.vue'),
  bindModel: 'modelValue',
  groupName: '预设',
  icon: 'icon--epic--select',
  sort: 100,
  defaultSchema: {
    label: '人员',
    type: 'arlo-user',
    field: 'userId',
    input: true,
    props: {
      placeholder: '请选择人员',
      clearable: true,
      filterable: true,
    },
  },
  config: {
    attribute: [
      { field: 'field', label: '数据字段', type: 'EpField' },
      { field: 'label', label: '标题', type: 'input' },
      { field: 'props.placeholder', label: '占位内容', type: 'input' },
      { field: 'props.clearable', label: '可清空', type: 'switch' },
      { field: 'props.disabled', label: '禁用', type: 'switch' },
      { field: 'props.filterable', label: '可搜索', type: 'switch' },
    ],
    event: [{ type: 'change', description: '值改变时触发' }],
  },
}

export default config
