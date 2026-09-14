/** 审批列表高级筛选 */
export type ApproveAdvancedFilterValue = {
  createBy?: string
  instanceState?: number | undefined
  beginTime?: string
  endTime?: string
}

/** 审批列表默认时间范围（近 N 天），返回 YYYY-MM-DD */
export function defaultApproveDateRange(days = 90): [string, string] {
  const end = new Date()
  const start = new Date()
  start.setHours(0, 0, 0, 0)
  end.setHours(0, 0, 0, 0)
  start.setDate(start.getDate() - (days > 0 ? days : 90))
  return [fmtDay(start), fmtDay(end)]
}

export function defaultApproveAdvancedFilter(days = 90): ApproveAdvancedFilterValue {
  const [beginTime, endTime] = defaultApproveDateRange(days)
  return {
    createBy: '',
    instanceState: undefined,
    beginTime,
    endTime,
  }
}

export function emptyApproveAdvancedFilter(): ApproveAdvancedFilterValue {
  return {
    createBy: '',
    instanceState: undefined,
    beginTime: undefined,
    endTime: undefined,
  }
}

function fmtDay(d: Date) {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

/** 组装列表查询参数（服务端筛选；空值不传，避免误筛） */
export function approveFilterParams(filter?: ApproveAdvancedFilterValue | null) {
  const out: {
    createBy?: string
    instanceState?: number
    beginTime?: string
    endTime?: string
  } = {}
  const createBy = filter?.createBy?.trim()
  if (createBy) out.createBy = createBy
  if (filter?.instanceState !== undefined && filter?.instanceState !== null) {
    const n = Number(filter.instanceState)
    if (Number.isFinite(n)) out.instanceState = n
  }
  if (filter?.beginTime) out.beginTime = filter.beginTime
  if (filter?.endTime) out.endTime = filter.endTime
  return out
}
