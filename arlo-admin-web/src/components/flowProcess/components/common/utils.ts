/** 从表单定义中抽出可用于节点权限配置的字段列表（兼容 epic PageSchema / 旧 widgetList） */
function collectFormWidgets(processForm: any): any[] {
  if (!processForm) return []

  // 旧 VForm 占位结构
  if (Array.isArray(processForm.widgetList)) {
    return processForm.widgetList
  }

  // epic-designer PageSchema
  const list: any[] = []
  const walk = (nodes?: any[]) => {
    if (!Array.isArray(nodes)) return
    for (const node of nodes) {
      if (!node || typeof node !== 'object') continue
      if (node.input) list.push(node)
      if (Array.isArray(node.children)) walk(node.children)
      if (node.slots && typeof node.slots === 'object') {
        Object.values(node.slots).forEach((slot: any) => walk(slot))
      }
    }
  }
  walk(processForm.schemas)
  return list
}

/**
 * 按当前发起表单重建节点 formConfig：
 * - 新增字段写入默认权限
 * - 已有字段保留 opera，刷新 label
 * - 表单已删除的字段从权限表移除；绑定子表单时请用 getMergedFormConfig
 */
export function getFormConfig(
  processFormOrWidgetList: any,
  formConfig: any,
  defaultOpera: 0 | 1 | 2 = 1,
) {
  return getMergedFormConfig(
    [{ form: processFormOrWidgetList, defaultOpera }],
    formConfig,
  )
}

/**
 * 多张表单字段合并进权限表（主表 + 子表）。
 * 同 field 以先出现的为准；新字段用对应来源的 defaultOpera。
 */
export function getMergedFormConfig(
  sources: { form: any; defaultOpera: 0 | 1 | 2 }[],
  formConfig: any,
) {
  const prev = Array.isArray(formConfig) ? formConfig : []
  const fieldTypes = new Set([
    'input',
    'textarea',
    'number',
    'radio',
    'checkbox',
    'select',
    'time',
    'time-range',
    'date',
    'date-range',
    'daterange',
    'switch',
    'rate',
    'color',
    'slider',
    'picture-upload',
    'file-upload',
    'upload',
    'upload-file',
    'upload-image',
    'rich-editor',
    'cascader',
    'arlo-user',
    'arlo-role',
    'arlo-dept',
    'arlo-dict',
    'arlo-rich-editor',
  ])

  const prevMap = new Map(prev.map((it: any) => [String(it.id), it]))
  const next: any[] = []
  const seen = new Set<string>()

  for (const src of sources) {
    const widgetList = Array.isArray(src.form) ? src.form : collectFormWidgets(src.form)
    if (!Array.isArray(widgetList)) continue
    const defaultOpera = src.defaultOpera

    widgetList.forEach((item: any) => {
      const isField =
        item?.input === true || (item?.type && fieldTypes.has(item.type))
      if (!isField) return

      const id = String(item.id || item.field || '')
      if (!id || seen.has(id)) return
      seen.add(id)

      const label = item.label || item.options?.label || item.field || id
      const exists = prevMap.get(id)
      if (exists) {
        next.push({
          ...exists,
          id,
          label,
          opera: exists.opera == null ? defaultOpera : exists.opera,
        })
      } else {
        next.push({
          id,
          label,
          opera: defaultOpera,
        })
      }
    })
  }
  return next
}

/** 遍历流程树，按最新 processForm 刷新各节点 formConfig */
export function syncNodesFormConfig(nodeConfig: any, processForm: any) {
  if (!nodeConfig || typeof nodeConfig !== 'object') return

  const walk = (node: any) => {
    if (!node || typeof node !== 'object') return
    const type = Number(node.type)
    // 0 发起 / 1 审批：需要表单权限；子流程节点无人办理，不同步 formConfig
    if (type === 0 || type === 1) {
      if (!node.extendConfig || typeof node.extendConfig !== 'object') {
        node.extendConfig = {}
      }
      const defaultOpera = type === 0 ? 1 : 0
      const prev = Array.isArray(node.extendConfig.formConfig)
        ? node.extendConfig.formConfig
        : []
      const mainCfg = getFormConfig(processForm, prev, defaultOpera as 0 | 1 | 2)
      // 绑了子表单：保留主表里没有的权限行（来自子表），避免同步主表时冲掉
      if (String(node.actionUrl || '').trim()) {
        const mainIds = new Set(mainCfg.map((c: any) => String(c.id)))
        const extras = prev.filter((c: any) => c?.id && !mainIds.has(String(c.id)))
        node.extendConfig.formConfig = [...mainCfg, ...extras]
      } else {
        node.extendConfig.formConfig = mainCfg
      }
    } else if (type === 5) {
      // 清理历史误写在子流程节点上的表单权限 / 异步标记
      if (node.extendConfig && typeof node.extendConfig === 'object') {
        delete node.extendConfig.formConfig
        if (Object.keys(node.extendConfig).length === 0) delete node.extendConfig
      }
      delete node.callAsync
    }
    if (node.childNode) walk(node.childNode)
    const branches = [
      node.conditionNodes,
      node.parallelNodes,
      node.inclusiveNodes,
      node.routeNodes,
    ]
    for (const list of branches) {
      if (!Array.isArray(list)) continue
      for (const branch of list) walk(branch)
    }
  }

  walk(nodeConfig)
}

export function getFormFields(processForm: any): { id: string; label: string }[] {
  return getFormConfig(processForm, []).map((item: any) => ({
    id: String(item.id),
    label: String(item.label || item.id),
  }))
}

export function replaceSymbol(str: string) {
  const symbol: Record<string, string> = {
    '==': '等于',
    '!=': '不等于',
    '>': '大于',
    '>=': '大于等于',
    '<': '小于',
    '<=': '小于等于',
    include: '包含',
    notinclude: '不包含',
    belong: '属于',
    notbelong: '不属于',
  }
  return symbol[str] || str
}

/** 条件节点卡片展示文案 */
export function formatConditionText(it: any) {
  if (!it) return ''
  if (it.type === 'initiator') {
    const names = Array.isArray(it.valueList) && it.valueList.length
      ? it.valueList.map((v: any) => v.name).filter(Boolean).join('、')
      : (it.value || '')
    return `${it.label || '发起人'}${replaceSymbol(it.operator)}${names}`
  }
  return `${it.label || ''}${replaceSymbol(it.operator)}${it.value ?? ''}`
}

export function createFormCondition(field?: { id: string; label: string }) {
  return {
    type: 'form',
    label: field?.label || '',
    field: field?.id || '',
    operator: '==',
    value: '',
  }
}

export function createInitiatorCondition() {
  return {
    type: 'initiator',
    label: '发起人',
    field: 'initiator',
    operator: 'belong',
    value: '',
    valueList: [] as { id: string | number; name: string; kind: 'dept' | 'role' | 'user' }[],
  }
}

export function findNodesByType(node: any) {
  const result: { nodeName: string; nodeKey: string; type: number }[] = []

  function traverse(currentNode: any) {
    if (currentNode.type === 0 || currentNode.type === 1) {
      result.push({
        nodeName: currentNode.nodeName,
        nodeKey: currentNode.nodeKey,
        type: currentNode.type,
      })
    }

    if (currentNode.childNode) {
      traverse(currentNode.childNode)
    }

    if (currentNode.conditionNodes && Array.isArray(currentNode.conditionNodes)) {
      currentNode.conditionNodes.forEach((conditionNode: any) => {
        if (conditionNode.childNode) {
          traverse(conditionNode.childNode)
        }
      })
    }

    if (currentNode.parallelNodes && Array.isArray(currentNode.parallelNodes)) {
      currentNode.parallelNodes.forEach((parallelNode: any) => {
        if (parallelNode.childNode) {
          traverse(parallelNode.childNode)
        }
      })
    }

    if (currentNode.inclusiveNodes && Array.isArray(currentNode.inclusiveNodes)) {
      currentNode.inclusiveNodes.forEach((inclusiveNode: any) => {
        if (inclusiveNode.childNode) {
          traverse(inclusiveNode.childNode)
        }
      })
    }

    if (currentNode.routeNodes && Array.isArray(currentNode.routeNodes)) {
      currentNode.routeNodes.forEach((routeNode: any) => {
        if (routeNode.childNode) {
          traverse(routeNode.childNode)
        }
      })
    }
  }

  traverse(node)
  return result
}
