import type { ComponentConfigModel } from 'epic-designer'

const config: ComponentConfigModel = {
  component: () => import('./component.vue'),
  bindModel: 'modelValue',
  groupName: '预设',
  icon: 'icon--epic--dialogs-outline-rounded',
  sort: 105,
  defaultSchema: {
    label: '角色',
    type: 'arlo-role',
    field: 'roleId',
    input: true,
    props: {
      placeholder: '请选择角色',
      clearable: true,
      filterable: true,
      multiple: false,
    },
  },
  config: {
    attribute: [
      { field: 'field', label: '数据字段', type: 'EpField' },
      { field: 'label', label: '标题', type: 'input' },
      { field: 'props.placeholder', label: '占位内容', type: 'input' },
      { field: 'props.multiple', label: '多选', type: 'switch' },
      { field: 'props.clearable', label: '可清空', type: 'switch' },
      { field: 'props.disabled', label: '禁用', type: 'switch' },
      { field: 'props.filterable', label: '可搜索', type: 'switch' },
    ],
    event: [{ type: 'change', description: '值改变时触发' }],
  },
}

export default config
