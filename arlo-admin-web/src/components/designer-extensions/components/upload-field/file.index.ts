import type { ComponentConfigModel } from 'epic-designer'

/** 覆盖 epic 内置 upload-file，改为 Arlo FilePicker */
const config: ComponentConfigModel = {
  component: () => import('./file.vue'),
  bindModel: 'modelValue',
  groupName: '表单',
  icon: 'icon--epic--upload-rounded',
  sort: 920,
  defaultSchema: {
    label: '上传文件',
    type: 'upload-file',
    field: 'uploadFile',
    input: true,
    props: {
      multiple: false,
      limit: 1,
    },
  },
  config: {
    attribute: [
      { field: 'field', label: '数据字段', type: 'EpField' },
      { field: 'label', label: '标题', type: 'input' },
      { field: 'props.multiple', label: '多选', type: 'switch' },
      {
        field: 'props.limit',
        label: '最大数量',
        type: 'number',
        props: { min: 1 },
      },
      { field: 'props.disabled', label: '禁用', type: 'switch' },
    ],
    event: [{ type: 'change', description: '值改变时触发' }],
  },
}

export default config
