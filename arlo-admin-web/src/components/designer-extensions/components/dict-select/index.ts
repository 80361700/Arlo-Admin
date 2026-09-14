import type { ComponentConfigModel } from 'epic-designer'

const config: ComponentConfigModel = {
  component: () => import('./component.vue'),
  bindModel: 'modelValue',
  groupName: '预设',
  icon: 'icon--epic--wysiwyg-rounded',
  sort: 120,
  defaultSchema: {
    label: '字典',
    type: 'arlo-dict',
    field: 'dictValue',
    input: true,
    props: {
      dictCode: 'sys_user_status',
      placeholder: '请选择',
      clearable: true,
      multiple: false,
    },
  },
  config: {
    attribute: [
      { field: 'field', label: '数据字段', type: 'EpField' },
      { field: 'label', label: '标题', type: 'input' },
      {
        field: 'props.dictCode',
        label: '字典编码',
        type: 'arlo-dict-code',
      },
      { field: 'props.placeholder', label: '占位内容', type: 'input' },
      { field: 'props.multiple', label: '多选', type: 'switch' },
      { field: 'props.clearable', label: '可清空', type: 'switch' },
      { field: 'props.disabled', label: '禁用', type: 'switch' },
    ],
    event: [{ type: 'change', description: '值改变时触发' }],
  },
}

export default config
