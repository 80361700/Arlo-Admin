import { pluginManager } from 'epic-designer'
import { resolveApiUrl } from '@/api/request'
import { setupDesignerGlobals } from './globals'
import { ensurePresetGroupExpanded } from './ensurePresetExpanded'
import userSelect from './components/user-select'
import roleSelect from './components/role-select'
import deptSelect from './components/dept-select'
import dictSelect from './components/dict-select'
import dictCodeAttr from './components/dict-code-attr'
import richEditor from './components/rich-editor'
import uploadFile from './components/upload-field/file.index'
import uploadImage from './components/upload-field/image.index'

/**
 * Arlo 对 epic-designer 的扩展：
 * - 全局上传鉴权
 * - 预设：人员 / 角色 / 部门 / 字典
 * - 表单：富文本、上传文件/图片（FilePicker）
 */
export function setupDesignerExtensions(): void {
  ensurePresetGroupExpanded()
  setupDesignerGlobals()

  // form-mode 下根节点已是表单；传入 defaultSchema 时官方不会自动 hide，需手动隐藏
  pluginManager.component.hide('form')

  pluginManager.component.register(dictCodeAttr)
  pluginManager.component.register(userSelect)
  pluginManager.component.register(roleSelect)
  pluginManager.component.register(deptSelect)
  pluginManager.component.register(dictSelect)
  pluginManager.component.register(richEditor)
  // 覆盖内置上传组件
  pluginManager.component.register(uploadFile)
  pluginManager.component.register(uploadImage)

  // 兜底：若仍残留默认 action，改成系统上传地址
  const uploadUrl = resolveApiUrl('/v1/file/upload')
  for (const type of ['upload-file', 'upload-image'] as const) {
    const cfg = pluginManager.component.getConfigByType(type)
    if (cfg?.defaultSchema?.props?.action) {
      cfg.defaultSchema.props.action = uploadUrl
    }
  }
}
