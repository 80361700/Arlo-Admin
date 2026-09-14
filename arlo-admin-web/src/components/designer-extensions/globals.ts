import { pluginManager } from 'epic-designer'
import { resolveApiUrl } from '@/api/request'

/** 把 epic 内置上传接到 Arlo 文件接口，并带上登录态 */
export function setupDesignerGlobals(): void {
  const uploadUrl = resolveApiUrl('/v1/file/upload')
  pluginManager.global.uploadFile = uploadUrl
  pluginManager.global.uploadImage = uploadUrl
  pluginManager.global.axiosConfig = {
    headers: {
      get Authorization() {
        const token = localStorage.getItem('accessToken') || ''
        return token ? `Bearer ${token}` : ''
      },
    },
  }
}
