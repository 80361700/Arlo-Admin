import type { FieldStates, PageSchema } from 'epic-designer'

/** epic formMode 下的默认根节点（与官方 useFormSchema 一致） */
export function createEmptyFormSchema(): PageSchema {
  return {
    schemas: [
      {
        id: 'root',
        label: '表单',
        type: 'form',
        props: {
          colon: true,
          labelAlign: 'right',
          labelCol: { span: 5 },
          labelLayout: 'fixed',
          labelPlacement: 'left',
          labelWidth: 100,
          layout: 'horizontal',
          name: 'default',
          wrapperCol: { span: 19 },
        },
        children: [],
      },
    ],
    script: `const { defineExpose, find } = epic;

function test (){
    console.log('test')
}

defineExpose({
 test
})`,
  }
}

export function isEpicPageSchema(value: unknown): value is PageSchema {
  return !!value && typeof value === 'object' && Array.isArray((value as PageSchema).schemas)
}

/** 是否含可展示表单字段（空根 children 视为无表单） */
export function pageSchemaHasFields(schema: PageSchema | null | undefined): boolean {
  if (!schema || !Array.isArray(schema.schemas) || !schema.schemas.length) return false
  const root = schema.schemas[0] as { children?: unknown[] }
  return Array.isArray(root?.children) && root.children.length > 0
}

/** 空数组 / 旧占位结构 → 合法 PageSchema，避免 epic 读 schemas[0] 崩溃 */
export function normalizeProcessForm(value: unknown): PageSchema {
  if (isEpicPageSchema(value) && value.schemas.length > 0) {
    const schema = JSON.parse(JSON.stringify(value)) as PageSchema
    normalizeFieldRules(schema)
    return schema
  }
  return createEmptyFormSchema()
}

/**
 * epic 默认必填规则常写成 type:string；日期区间/多选实际是数组，
 * 有值也会被判失败。按组件类型纠正 rule.type。
 */
export function normalizeFieldRules(schema: PageSchema) {
  const arrayTypes = new Set([
    'checkbox',
    'daterange',
    'date-range',
    'timerange',
    'time-range',
    'cascader',
    'upload-image',
    'upload-file',
    'picture-upload',
    'file-upload',
  ])
  const walk = (nodes?: any[]) => {
    if (!Array.isArray(nodes)) return
    for (const node of nodes) {
      if (!node || typeof node !== 'object') continue
      const typ = String(node.type || '')
      const propsType = String(node.props?.type || '')
      const isArrayValue =
        arrayTypes.has(typ) ||
        propsType === 'daterange' ||
        propsType === 'datetimerange' ||
        propsType === 'timerange' ||
        propsType === 'dates' ||
        (typ === 'date' && (propsType === 'daterange' || propsType === 'datetimerange')) ||
        (typ === 'select' && !!node.props?.multiple) ||
        (typ === 'arlo-user' && !!node.props?.multiple) ||
        (typ === 'arlo-role' && !!node.props?.multiple) ||
        (typ === 'arlo-dept' && !!node.props?.multiple) ||
        (typ === 'arlo-dict' && !!node.props?.multiple)

      if (isArrayValue && Array.isArray(node.rules)) {
        node.rules = node.rules.map((r: any) => {
          if (!r || typeof r !== 'object') return r
          if (r.type === 'string' || r.type === 'number') {
            return { ...r, type: 'array' }
          }
          if (r.required && !r.type) {
            return { ...r, type: 'array' }
          }
          return r
        })
      }
      if (Array.isArray(node.children)) walk(node.children)
      if (node.slots && typeof node.slots === 'object') {
        Object.values(node.slots).forEach((slot: any) => walk(slot))
      }
    }
  }
  walk(schema.schemas)
}

/**
 * 按节点 formConfig 生成 epic fieldStates。
 * opera: 0 只读 / 1 编辑 / 2 隐藏
 * writable=false（不可审批）时全部禁用，隐藏仍隐藏。
 *
 * 注意：只读必须用 DISABLED，不能用 READ。
 * epic 的 READ 只写 readonly，Element Plus 的 select/date/upload 等仍可交互。
 *
 * schema 用于：无 formConfig / 配置不全时，把父表单继承过来的字段一并禁用，
 * 避免 EBuilder 的 disabled 对 select 等组件失效导致「审批还能改单」。
 */
export function buildFieldStates(
  formConfig: { id?: string; field?: string; opera?: number | string }[] | undefined,
  writable: boolean,
  schema?: PageSchema | null,
): FieldStates {
  const list = formConfig || []
  const schemaIds = collectSchemaFieldIds(schema)

  if (!list.length) {
    if (writable) return []
    return schemaIds.map((field) => ({ field, state: 'DISABLED' as const }))
  }

  const operaOf = new Map<string, number>()
  for (const c of list) {
    const field = String(c.id || c.field || '')
    if (!field) continue
    operaOf.set(field, Number(c.opera))
  }

  const fields = new Set<string>([...operaOf.keys(), ...schemaIds])
  const states: FieldStates = []
  for (const field of fields) {
    // 未出现在 formConfig 中的 schema 字段：默认只读
    const opera = operaOf.has(field) ? (operaOf.get(field) as number) : 0
    if (opera === 2) {
      states.push({ field, state: 'HIDE' })
      continue
    }
    if (!writable || opera === 0) {
      states.push({ field, state: 'DISABLED' })
      continue
    }
    states.push({ field, state: 'WRITE' })
  }
  return states
}

/** 收集 epic schema 中可输入字段 id */
export function collectSchemaFieldIds(schema: PageSchema | null | undefined): string[] {
  const ids: string[] = []
  const walk = (nodes?: any[]) => {
    if (!Array.isArray(nodes)) return
    for (const node of nodes) {
      if (!node || typeof node !== 'object') continue
      const id = String(node.id || node.field || '')
      if (id && id !== 'root' && (node.input === true || node.field)) {
        ids.push(id)
      }
      if (Array.isArray(node.children)) walk(node.children)
      if (node.slots && typeof node.slots === 'object') {
        Object.values(node.slots).forEach((slot: any) => walk(slot))
      }
    }
  }
  walk(schema?.schemas as any[])
  return [...new Set(ids)]
}

/**
 * 办理态是否允许编辑表单：
 * - 发起人改单重提：可编辑
 * - 审批：仅当节点 formConfig 明确配置了可编辑字段（opera=1）
 * - 无 formConfig 时默认只读（避免子流程继承父表单后误开全表编辑）
 */
export function canEditApproveForm(
  formConfig: { opera?: number | string }[] | undefined,
  opts: { canConsent?: boolean; canResubmit?: boolean },
): boolean {
  if (opts.canResubmit) return true
  if (!opts.canConsent) return false
  return (formConfig || []).some((c) => Number(c.opera) === 1)
}

/**
 * @deprecated 优先用 buildFieldStates；保留给需要改 schema 的旧路径。
 * 必须写 props（不能写 componentProps）：epic migrate 会用 componentProps 整表覆盖 props。
 */
export function applyFormConfig(
  schema: PageSchema,
  formConfig: { id: string; opera: number }[],
  editable: boolean,
): PageSchema {
  const s = JSON.parse(JSON.stringify(schema || createEmptyFormSchema())) as PageSchema
  const map = new Map((formConfig || []).map((c) => [String(c.id), Number(c.opera)]))
  const walk = (nodes?: any[]) => {
    if (!Array.isArray(nodes)) return
    for (let i = nodes.length - 1; i >= 0; i--) {
      const node = nodes[i]
      if (!node || typeof node !== 'object') continue
      const id = String(node.id || node.field || '')
      if (id && map.has(id)) {
        const opera = map.get(id)!
        if (opera === 2) {
          nodes.splice(i, 1)
          continue
        }
        if (opera === 0 || !editable) {
          node.props = { ...(node.props || {}), disabled: true }
        }
      } else if (!editable && node.input) {
        node.props = { ...(node.props || {}), disabled: true }
      }
      if (Array.isArray(node.children)) walk(node.children)
      if (node.slots && typeof node.slots === 'object') {
        Object.values(node.slots).forEach((slot: any) => walk(slot))
      }
    }
  }
  walk(s.schemas)
  return s
}

/** EBuilder.getData 是异步的，发起/同意时必须 await，否则会把 Promise 序列化成 {} */
export async function readBuilderFormData(builder: any): Promise<Record<string, any>> {
  try {
    let raw = await builder?.getData?.()
    if (raw && typeof raw === 'object' && 'schemas' in raw) {
      raw = (raw as any).formData || (raw as any).data || {}
    }
    if (raw && typeof raw === 'object' && !Array.isArray(raw)) {
      return raw as Record<string, any>
    }
  } catch {
    /* ignore */
  }
  return {}
}

/**
 * epic setFormData 是合并写入：切换实例时若不先清掉，上一单字段会残留。
 * 先删掉 forms 再 setData，保证整表替换。
 */
export function writeBuilderFormData(builder: any, data: Record<string, any>) {
  const forms = builder?.pageManager?.forms
  if (forms && typeof forms === 'object') {
    for (const key of Object.keys(forms)) {
      delete forms[key]
    }
  }
  builder?.resetData?.()
  builder?.setData?.(data || {})
}
