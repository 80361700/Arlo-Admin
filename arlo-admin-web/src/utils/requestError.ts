import { ElMessage } from 'element-plus'

/** 请求层已弹过的错误（带 shown），页面 catch 里请用本方法，避免双提示 */
export function showRequestError(err: unknown, fallback = '操作失败') {
  if (err === 'cancel' || err === 'close') return
  const e = err as { shown?: boolean; message?: string } | null
  if (e?.shown) return
  ElMessage.error(e?.message || fallback)
}
