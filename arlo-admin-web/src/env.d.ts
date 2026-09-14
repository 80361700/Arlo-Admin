/// <reference types="vite/client" />

declare module 'print-js' {
  interface PrintJSParams {
    printable: string | string[]
    type?: 'pdf' | 'html' | 'image' | 'json' | 'raw-html'
    scanStyles?: boolean
    targetStyles?: string | string[]
    style?: string
    documentTitle?: string
    [key: string]: any
  }
  function printJS(params: PrintJSParams | string): void
  export default printJS
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

export {}

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    icon?: string
    noAuth?: boolean
    affix?: boolean
    keepAlive?: boolean
    menuId?: number
  }
}
