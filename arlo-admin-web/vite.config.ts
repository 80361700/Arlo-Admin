import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import monacoEditorPlugin from 'vite-plugin-monaco-editor'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'src/types/auto-imports.d.ts',
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/types/components.d.ts',
    }),
    // epic-designer 源码编辑依赖 monaco；不同版本 default 导出方式不一致
    ((monacoEditorPlugin as any).default || monacoEditorPlugin)({
      languageWorkers: ['editorWorkerService', 'json'],
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      // 与 .env 中 VITE_API_BASE_URL=/api 对应；本地后端地址只在这里配
      '/api': {
        target: 'http://localhost:8090',
        changeOrigin: true,
        ws: true,
      },
    },
  },
})
