import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import 'epic-designer/dist/style.css'
import { setupElementPlus } from '@epic-designer/element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'
import { setupPermissionDirective } from './utils/directives/permission'
import { setupDesignerExtensions } from './components/designer-extensions'
import { resolveApiUrl } from './api/request'
import './styles/index.scss'

const uploadUrl = resolveApiUrl('/v1/file/upload')
setupElementPlus(undefined, {
  uploadFile: uploadUrl,
  uploadImage: uploadUrl,
})
setupDesignerExtensions()

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })
setupPermissionDirective(app)

// 全局注册所有 Element Plus 图标组件
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.mount('#app')
