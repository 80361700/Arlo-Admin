import type { ComponentConfigModel } from 'epic-designer'

const config: ComponentConfigModel = {
  component: () => import('./component.vue'),
  bindModel: 'modelValue',
  groupName: '预设',
  icon: 'icon--epic--list-alt-outline-rounded',
  sort: 110,
  defaultSchema: {
    label: '部门',
    type: 'arlo-dept',
    field: 'deptId',
    input: true,
    props: {
      placeholder: '请选择部门',
      clearable: true,
      filterable: true,
      checkStrictly: true,
      multiple: false,
    },
  },
  config: {
    attribute: [
      { field: 'field', label: '数据字段', type: 'EpField' },
      { field: 'label', label: '标题', type: 'input' },
      { field: 'props.placeholder', label: '占位内容', type: 'input' },
      { field: 'props.multiple', label: '多选', type: 'switch' },
      { field: 'props.checkStrictly', label: '任意节点可选', type: 'switch' },
      { field: 'props.clearable', label: '可清空', type: 'switch' },
      { field: 'props.disabled', label: '禁用', type: 'switch' },
    ],
    event: [{ type: 'change', description: '值改变时触发' }],
  },
}

export default config
