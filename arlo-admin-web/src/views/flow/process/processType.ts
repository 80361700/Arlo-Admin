/** 流程定义类型（与后端 process_type / 飞龙约定一致） */
export type ProcessType = 'main' | 'business' | 'child'

export type WizardStepKey = 'basic' | 'form' | 'process' | 'setting'

export interface WizardStep {
  key: WizardStepKey
  title: string
}

export const PROCESS_TYPE_META: Record<
  ProcessType,
  { label: string; createLabel: string; primary?: boolean }
> = {
  child: { label: '子流程', createLabel: '创建子流程' },
  business: { label: '业务流程', createLabel: '创建业务审批' },
  main: { label: '审批流程', createLabel: '创建审批', primary: true },
}

/** 创建入口顺序：子流程 / 业务审批 / 审批 */
export const PROCESS_CREATE_TYPES: ProcessType[] = ['child', 'business', 'main']

/**
 * 向导步骤：
 * - 审批流程：基础 → 表单 → 流程 → 扩展
 * - 业务审批：基础（含选择表单）→ 流程 → 扩展
 * - 子流程：基础 → 流程 → 扩展
 */
export function wizardStepsFor(type: string | undefined): WizardStep[] {
  const t = normalizeProcessType(type)
  if (t === 'child') {
    return [
      { key: 'basic', title: '基础信息' },
      { key: 'process', title: '流程设计' },
      { key: 'setting', title: '扩展设置' },
    ]
  }
  if (t === 'business') {
    return [
      { key: 'basic', title: '基础信息' },
      { key: 'process', title: '流程设计' },
      { key: 'setting', title: '扩展设置' },
    ]
  }
  return [
    { key: 'basic', title: '基础信息' },
    { key: 'form', title: '表单设计' },
    { key: 'process', title: '流程设计' },
    { key: 'setting', title: '扩展设置' },
  ]
}

export function normalizeProcessType(type: string | undefined | null): ProcessType {
  if (type === 'child' || type === 'business' || type === 'main') return type
  return 'main'
}

export function processTypeLabel(type: string | undefined | null): string {
  return PROCESS_TYPE_META[normalizeProcessType(type)].label
}
