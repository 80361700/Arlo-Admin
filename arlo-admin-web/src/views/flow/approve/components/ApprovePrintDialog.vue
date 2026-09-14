<template>
  <el-dialog
    v-model="visible"
    title="打印预览"
    width="720px"
    destroy-on-close
    :close-on-click-modal="false"
    class="approve-print-dialog"
    append-to-body
  >
    <div id="approve-print-area" class="print-area">
      <h1 class="print-title">{{ processName || '审批单' }}</h1>

      <table class="print-meta-table">
        <tbody>
          <tr>
            <th>审批编号</th>
            <td>{{ instanceId || '—' }}</td>
            <th>提交时间</th>
            <td>{{ createdAt || '—' }}</td>
          </tr>
        </tbody>
      </table>

      <table v-if="printRows.length" class="print-form-table">
        <tbody>
          <tr v-for="(row, idx) in printRows" :key="`${row.label}-${idx}`">
            <th>{{ row.label }}</th>
            <td>
              <div v-if="row.kind === 'images'" class="print-media">
                <template v-if="row.imageSrcs?.length">
                  <img
                    v-for="(src, i) in row.imageSrcs"
                    :key="`${src}-${i}`"
                    :src="src"
                    class="print-img"
                    alt=""
                  />
                </template>
                <span v-else-if="imagesLoading" class="print-muted">图片加载中…</span>
                <span v-else class="print-muted">{{ row.value || '—' }}</span>
              </div>
              <div v-else-if="row.kind === 'files'" class="print-files">
                {{ row.value || '—' }}
              </div>
              <template v-else>{{ row.value }}</template>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="print-empty">暂无表单内容</div>

      <table class="print-meta-table print-footer-table">
        <tbody>
          <tr>
            <th>打印人</th>
            <td>{{ printerName }}</td>
            <th>打印时间</th>
            <td>{{ printTime || '—' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="printing || imagesLoading" @click="doPrint">
        打印
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import printJS from 'print-js'
import type { FieldStates, PageSchema } from 'epic-designer'
import { useAuthStore } from '@/stores/auth'
import { pageSchemaHasFields } from '@/views/flow/process/components/formSchema'
import { acquireAuthFileUrl } from '@/utils/authFileUrl'
import { parseAccessKey } from '@/utils/fileUrl'

type PrintRow = {
  label: string
  value: string
  kind?: 'text' | 'images' | 'files'
  /** 原始文件引用，用于异步换图 */
  fileRefs?: string[]
  imageSrcs?: string[]
}

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    processName?: string
    instanceId?: number | string
    createdAt?: string
    formReady?: boolean
    formRenderType?: string
    formComponent?: string
    formData?: Record<string, any>
    pageSchema?: PageSchema | null
    fieldStates?: FieldStates
    renderKey?: number | string
  }>(),
  {
    processName: '',
    instanceId: '',
    createdAt: '',
    formReady: false,
    formRenderType: 'designer',
    formComponent: '',
    formData: () => ({}),
    pageSchema: null,
    fieldStates: () => [],
    renderKey: 0,
  },
)

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
}>()

const authStore = useAuthStore()
const printing = ref(false)
const printTime = ref('')
const imagesLoading = ref(false)
/** accessKey → blob/url，会话内给预览用 */
const imageUrlMap = ref<Record<string, string>>({})

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const printerName = computed(
  () => authStore.userInfo?.name || authStore.userInfo?.username || '—',
)

const hiddenFields = computed(() => {
  const set = new Set<string>()
  for (const s of props.fieldStates || []) {
    if (s?.state === 'HIDE' && s.field) set.add(String(s.field))
  }
  return set
})

const baseRows = computed(() => {
  if (!props.formReady) return [] as PrintRow[]
  const data = props.formData || {}
  const hidden = hiddenFields.value

  if (pageSchemaHasFields(props.pageSchema)) {
    const rows: PrintRow[] = []
    walkSchema(props.pageSchema?.schemas as any[], (node) => {
      const id = String(node.id || '')
      const field = String(node.field || node.id || '')
      if (!field || field === 'root') return
      if (hidden.has(field) || (id && hidden.has(id))) return
      if (!(node.input === true || node.field)) return
      if (isLayoutType(node.type)) return
      const label = String(node.label || node.props?.label || field).replace(/[:：]\s*$/, '')
      const raw = data[field] !== undefined ? data[field] : id ? data[id] : undefined
      rows.push(buildRow(label || field, raw, node))
    })
    if (rows.length) return rows
  }

  return Object.keys(data)
    .filter((k) => !k.startsWith('_') && !hidden.has(k))
    .map((k) => buildRow(k, data[k], null))
})

const printRows = computed(() =>
  baseRows.value.map((row) => {
    if (row.kind !== 'images' || !row.fileRefs?.length) return row
    const imageSrcs = row.fileRefs.map((k) => imageUrlMap.value[k]).filter(Boolean)
    return { ...row, imageSrcs }
  }),
)

const SKIP_TYPES = new Set([
  'form',
  'page',
  'div',
  'row',
  'col',
  'card',
  'tabs',
  'tab-pane',
  'collapse',
  'collapse-item',
  'divider',
  'alert',
  'button',
  'space',
  'grid',
  'html',
])

function isLayoutType(type: unknown) {
  return SKIP_TYPES.has(String(type || '').toLowerCase())
}

function isImageField(node: any): boolean {
  const typ = String(node?.type || '').toLowerCase()
  return (
    typ === 'upload-image' ||
    typ === 'picture-upload' ||
    typ.includes('upload-image') ||
    (typ.includes('image') && typ.includes('upload'))
  )
}

function isFileField(node: any): boolean {
  const typ = String(node?.type || '').toLowerCase()
  return typ === 'upload-file' || typ === 'file-upload' || typ.includes('upload-file')
}

function walkSchema(nodes: any[] | undefined, visit: (node: any) => void) {
  if (!Array.isArray(nodes)) return
  for (const node of nodes) {
    if (!node || typeof node !== 'object') continue
    visit(node)
    if (Array.isArray(node.children)) walkSchema(node.children, visit)
    if (node.slots && typeof node.slots === 'object') {
      Object.values(node.slots).forEach((slot: any) => walkSchema(slot, visit))
    }
  }
}

function findOptionLabel(options: any[], value: unknown): string | undefined {
  if (!Array.isArray(options) || !options.length) return undefined
  const hit = options.find((o) => String(o?.value) === String(value))
  return hit ? String(hit.label ?? hit.text ?? hit.name ?? value) : undefined
}

function looksLikeDate(v: unknown) {
  return typeof v === 'string' && /^\d{4}-\d{2}-\d{2}/.test(v)
}

function parseFileRefs(raw: unknown): string[] {
  if (raw == null || raw === '') return []
  if (Array.isArray(raw)) {
    return raw.flatMap((x) => {
      if (typeof x === 'string') return parseFileRefs(x)
      if (x && typeof x === 'object') {
        const o = x as Record<string, any>
        const key = o.accessKey || o.url || o.id || ''
        return key ? [String(key)] : []
      }
      return []
    })
  }
  const s = String(raw).trim()
  if (!s) return []
  if (s.startsWith('[') || s.startsWith('{')) {
    try {
      return parseFileRefs(JSON.parse(s))
    } catch {
      /* ignore */
    }
  }
  return s
    .split(',')
    .map((x) => x.trim())
    .filter(Boolean)
}

function shortRef(ref: string) {
  const key = parseAccessKey(ref) || ref
  return key.length > 12 ? `${key.slice(0, 8)}…` : key
}

function buildRow(label: string, raw: unknown, node: any): PrintRow {
  if (isImageField(node)) {
    const refs = parseFileRefs(raw)
    return {
      label,
      kind: 'images',
      fileRefs: refs,
      value: refs.length ? refs.map(shortRef).join('、') : '—',
      imageSrcs: [],
    }
  }
  if (isFileField(node)) {
    const refs = parseFileRefs(raw)
    return {
      label,
      kind: 'files',
      fileRefs: refs,
      value: refs.length ? refs.map(shortRef).join('、') : '—',
    }
  }
  return {
    label,
    kind: 'text',
    value: formatPrintValue(raw, node),
  }
}

function formatPrintValue(raw: unknown, node: any): string {
  if (raw === null || raw === undefined || raw === '') return '—'

  const typ = String(node?.type || '').toLowerCase()
  const props = node?.props || {}
  const options = props.options || props.optionItems || []

  if (typeof raw === 'boolean') return raw ? '是' : '否'

  if (Array.isArray(raw)) {
    if (!raw.length) return '—'
    if (raw.length === 2 && looksLikeDate(raw[0]) && looksLikeDate(raw[1])) {
      return `${raw[0]} 至 ${raw[1]}`
    }
    if (raw.every((x) => x && typeof x === 'object')) {
      const names = raw
        .map((x: any) => x.name || x.fileName || x.label || x.title || x.url || '')
        .filter(Boolean)
      return names.length ? names.join('、') : '—'
    }
    return raw
      .map((v) => findOptionLabel(options, v) || String(v ?? ''))
      .filter((s) => s !== '')
      .join('、') || '—'
  }

  if (typeof raw === 'object') {
    const o = raw as Record<string, any>
    return String(o.name || o.fileName || o.label || o.title || o.url || JSON.stringify(raw))
  }

  const mapped = findOptionLabel(options, raw)
  if (mapped) return mapped

  if (typ.includes('rich') || typ.includes('editor') || typ.includes('html')) {
    return String(raw)
      .replace(/<[^>]+>/g, ' ')
      .replace(/\s+/g, ' ')
      .trim() || '—'
  }

  return String(raw)
}

async function resolveImages() {
  const refs = new Set<string>()
  for (const row of baseRows.value) {
    if (row.kind === 'images') {
      for (const r of row.fileRefs || []) refs.add(r)
    }
  }
  if (!refs.size) {
    imageUrlMap.value = {}
    imagesLoading.value = false
    return
  }
  imagesLoading.value = true
  const next: Record<string, string> = { ...imageUrlMap.value }
  try {
    await Promise.all(
      [...refs].map(async (ref) => {
        if (next[ref]) return
        try {
          next[ref] = await acquireAuthFileUrl(ref)
        } catch {
          next[ref] = ''
        }
      }),
    )
    imageUrlMap.value = next
  } finally {
    imagesLoading.value = false
  }
}

watch(
  () => props.modelValue,
  async (open) => {
    if (open) {
      const d = new Date()
      const pad = (n: number) => String(n).padStart(2, '0')
      printTime.value = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
      await resolveImages()
    } else {
      imageUrlMap.value = {}
    }
  },
)

watch(
  () => [props.formData, props.pageSchema, props.formReady] as const,
  () => {
    if (visible.value) resolveImages()
  },
  { deep: true },
)

const PRINT_CSS = `
  @page { margin: 12mm 10mm; }
  body { font-family: "PingFang SC", "Microsoft YaHei", sans-serif; color: #303133; font-size: 14px; }
  .print-area { color: #303133; }
  .print-title {
    text-align: center;
    font-size: 22px;
    font-weight: 600;
    margin: 0 0 20px;
    letter-spacing: 1px;
  }
  .print-meta-table,
  .print-form-table {
    width: 100%;
    border-collapse: collapse;
    table-layout: fixed;
    margin: 0 0 16px;
  }
  .print-meta-table th,
  .print-meta-table td,
  .print-form-table th,
  .print-form-table td {
    border: 1px solid #d0d3d9;
    padding: 10px 12px;
    vertical-align: top;
    line-height: 1.6;
    word-break: break-word;
  }
  .print-meta-table th,
  .print-form-table th {
    width: 120px;
    background: #f7f8fa;
    font-weight: 500;
    text-align: right;
    color: #606266;
  }
  .print-meta-table th { width: 88px; }
  .print-footer-table { margin: 20px 0 0; }
  .print-empty {
    border: 1px solid #e4e7ed;
    padding: 24px;
    text-align: center;
    color: #909399;
    margin-bottom: 16px;
  }
  .print-media { display: flex; flex-wrap: wrap; gap: 8px; }
  .print-img {
    max-width: 180px;
    max-height: 140px;
    object-fit: contain;
    border: 1px solid #e4e7ed;
    border-radius: 4px;
    background: #fff;
  }
  .print-muted { color: #909399; }
`

async function waitImagesReady() {
  if (imagesLoading.value) {
    await new Promise<void>((resolve) => {
      const stop = watch(imagesLoading, (v) => {
        if (!v) {
          stop()
          resolve()
        }
      })
    })
  }
  const imgs = Array.from(document.querySelectorAll('#approve-print-area img.print-img')) as HTMLImageElement[]
  await Promise.all(
    imgs.map(
      (img) =>
        new Promise<void>((resolve) => {
          if (img.complete) {
            resolve()
            return
          }
          img.onload = () => resolve()
          img.onerror = () => resolve()
        }),
    ),
  )
}

async function doPrint() {
  printing.value = true
  try {
    await waitImagesReady()
    printJS({
      printable: 'approve-print-area',
      type: 'html',
      scanStyles: false,
      style: PRINT_CSS,
    })
  } finally {
    setTimeout(() => {
      printing.value = false
    }, 400)
  }
}
</script>

<style scoped lang="scss">
.print-area {
  color: #303133;
  font-size: 14px;
}
.print-title {
  text-align: center;
  font-size: 22px;
  font-weight: 600;
  margin: 0 0 20px;
  letter-spacing: 1px;
  color: #303133;
}
.print-meta-table,
.print-form-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  margin: 0 0 16px;
  th,
  td {
    border: 1px solid #d0d3d9;
    padding: 10px 12px;
    vertical-align: top;
    line-height: 1.6;
    word-break: break-word;
  }
  th {
    width: 120px;
    background: #f7f8fa;
    font-weight: 500;
    text-align: right;
    color: #606266;
  }
}
.print-meta-table th {
  width: 88px;
}
.print-footer-table {
  margin: 20px 0 0;
}
.print-empty {
  border: 1px solid #e4e7ed;
  padding: 24px;
  text-align: center;
  color: #909399;
  margin-bottom: 16px;
}
.print-media {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.print-img {
  max-width: 180px;
  max-height: 140px;
  object-fit: contain;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #fff;
}
.print-muted {
  color: #909399;
}
</style>
