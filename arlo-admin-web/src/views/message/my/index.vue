<template>
  <div class="page-container">
    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <!-- 我的消息 -->
      <el-tab-pane label="我的消息" name="received">
        <ProTable
          ref="receivedTableRef"
          :data="receivedTableData"
          :loading="receivedLoading"
          :total="receivedTotal"
          :search-fields="receivedSearchFields"
          :show-index="false"
          :action-width="200"
          @search="handleReceivedSearch"
          @reset="handleReceivedReset"
          @page-change="handleReceivedPageChange"
        >
          <template #toolbar>
            <el-button v-permission="['message:my:add', 'message:my:list']" type="primary" @click="handleSend">发送消息</el-button>
            <el-button
              v-if="activeTab === 'received' && unreadCount > 0"
              v-permission="['message:my:edit', 'message:my:list']"
              @click="handleMarkAllRead"
            >
              全部已读 ({{ unreadCount }})
            </el-button>
          </template>
          <el-table-column prop="id" label="ID" width="80" align="center" />
          <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip>
            <template #default="{ row }">
              <span :style="{ fontWeight: row.isRead === 0 ? 'bold' : 'normal' }">
                {{ row.title }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="类型" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="typeTag(row.type)" size="small">{{ typeLabel(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sender" label="发送者" width="100" />
          <el-table-column label="接收者" width="100">
            <template #default="{ row }">
              {{ row.receiverId === 0 ? '全部用户' : '指定用户' }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.isRead === 1 ? 'success' : 'warning'" size="small">
                {{ row.isRead === 1 ? '已读' : '未读' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="发送时间" width="170" />
          <template #actions="{ row }">
            <el-button
              v-permission="['message:my:view', 'message:my:edit', 'message:my:list']"
              type="primary"
              link
              size="small"
              @click="handleDetail(row)"
            >详情</el-button>
            <el-button
              v-if="row.isRead === 0"
              v-permission="['message:my:edit', 'message:my:list']"
              type="success"
              link
              size="small"
              @click="handleMarkRead(row)"
            >标为已读</el-button>
            <el-button
              v-if="authStore.hasPermission('message:my:delete')"
              type="danger"
              link
              size="small"
              @click="handleDelete(row)"
            >删除</el-button>
          </template>
        </ProTable>
      </el-tab-pane>

      <!-- 发送记录 -->
      <el-tab-pane label="发送记录" name="sent">
        <ProTable
          ref="sentTableRef"
          :data="sentTableData"
          :loading="sentLoading"
          :total="sentTotal"
          :search-fields="sentSearchFields"
          :show-index="false"
          :action-width="160"
          @search="handleSentSearch"
          @reset="handleSentReset"
          @page-change="handleSentPageChange"
        >
          <template #toolbar>
            <el-button v-permission="['message:my:add', 'message:my:list']" type="primary" @click="handleSend">发送消息</el-button>
          </template>
          <el-table-column prop="id" label="ID" width="80" align="center" />
          <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
          <el-table-column label="类型" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="typeTag(row.type)" size="small">{{ typeLabel(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="接收者" width="120">
            <template #default="{ row }">
              <span v-if="row.receiverId === 0">全部用户</span>
              <span v-else-if="(row.receiverCount || 0) > 1">指定用户 ({{ row.receiverCount }}人)</span>
              <span v-else>指定用户</span>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="发送时间" width="170" />
          <template #actions="{ row }">
            <el-button
              v-permission="['message:my:view', 'message:my:list']"
              type="primary"
              link
              size="small"
              @click="handleDetail(row)"
            >详情</el-button>
            <el-button
              v-if="authStore.hasPermission('message:my:delete')"
              type="danger"
              link
              size="small"
              @click="handleDelete(row)"
            >删除</el-button>
          </template>
        </ProTable>
      </el-tab-pane>
    </el-tabs>

    <!-- 发送消息弹窗 -->
    <ProFormDialog
      ref="formDialogRef"
      v-model="dialogVisible"
      title="发送消息"
      width="1000px"
      :model="form"
      :rules="formRules"
      :submitting="submitting"
      @submit="handleSendSubmit"
    >
      <el-form-item label="标题" prop="title">
        <el-input v-model="form.title" placeholder="请输入标题" maxlength="128" />
      </el-form-item>
      <el-form-item label="类型" prop="type">
        <el-select v-model="form.type" placeholder="请选择">
          <el-option
            v-for="opt in messageTypeOptions"
            :key="String(opt.value)"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="接收者" prop="receiverId">
        <el-radio-group v-model="form.receiverId" @change="onReceiverChange">
          <el-radio :value="0">全部用户</el-radio>
          <el-radio :value="1">指定用户</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="form.receiverId === 1" label="选择用户" prop="targetUserIds">
        <div class="user-select-area">
          <el-button type="primary" size="small" @click="openUserPicker">
            <el-icon><User /></el-icon>
            <span>选择用户</span>
          </el-button>
          <div v-if="selectedUsers.length > 0" class="selected-user-tags">
            <el-tag
              v-for="user in selectedUsers"
              :key="user.id"
              closable
              size="small"
              @close="removeSelectedUser(user.id)"
            >
              {{ user.nickname || user.username }}
            </el-tag>
          </div>
        </div>
      </el-form-item>
      <el-form-item label="内容" prop="content">
        <RichEditor v-model="form.content" placeholder="请输入消息内容" />
      </el-form-item>
    </ProFormDialog>

    <!-- 用户选择弹框 -->
    <el-dialog
      v-model="userPickerVisible"
      title="选择用户"
      width="720px"
      :close-on-click-modal="false"
    >
      <div class="user-picker-toolbar">
        <el-input
          v-model="userQuery.keyword"
          placeholder="搜索用户名/昵称"
          clearable
          style="width: 200px"
          @keyup.enter="loadUserPicker"
          @clear="loadUserPicker"
        />
        <el-button type="primary" @click="loadUserPicker">搜索</el-button>
        <el-button @click="userQuery.keyword = ''; loadUserPicker()">清空</el-button>
      </div>

      <el-table
        :data="userPickerList"
        height="320"
        @row-click="handleUserRowClick"
        :row-class-name="userRowClassName"
      >
        <el-table-column width="55" align="center">
          <template #header>
            <el-checkbox
              :model-value="isAllUserSelected"
              :indeterminate="isUserIndeterminate"
              @change="(val: any) => toggleSelectAllUsers(val as boolean)"
            />
          </template>
          <template #default="{ row }">
            <el-checkbox
              :model-value="userSelectedMap.has(row.id)"
              @click.stop
              @change="(val: any) => toggleUserRow(row as UserItem, val as boolean)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="nickname" label="昵称" width="120" />
        <el-table-column prop="phone" label="手机号" width="120" />
        <el-table-column prop="deptName" label="部门" min-width="120" />
      </el-table>

      <div class="user-picker-pagination">
        <el-pagination
          v-model:current-page="userQuery.page"
          v-model:page-size="userQuery.pageSize"
          :total="userTotal"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          small
          @change="loadUserPicker"
        />
      </div>

      <div v-if="userSelectedMap.size > 0" class="user-selected-bar">
        <span>已选 {{ userSelectedMap.size }} 人</span>
        <el-button link size="small" @click="clearUserSelection">清空已选</el-button>
      </div>

      <template #footer>
        <el-button @click="userPickerVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmUserPicker" :disabled="userSelectedMap.size === 0">
          确定 ({{ userSelectedMap.size }})
        </el-button>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="消息详情" width="620px" :close-on-click-modal="false">
      <template v-if="detailData">
        <div class="detail-header">
          <h3>{{ detailData.title }}</h3>
          <div class="detail-meta">
            <el-tag :type="typeTag(detailData.type)" size="small">{{ typeLabel(detailData.type) }}</el-tag>
            <span>发送者：{{ detailData.sender || '-' }}</span>
            <span>接收者：{{ detailData.receiverId === 0 ? '全部用户' : '指定用户' }}</span>
            <span>{{ detailData.createdAt }}</span>
          </div>
        </div>
        <div class="detail-body">
          <RichContent :html="detailData.content" />
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { showRequestError } from '@/utils/requestError'
import { User } from '@element-plus/icons-vue'
import type { FormRules } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import RichEditor from '@/components/RichEditor.vue'
import RichContent from '@/components/RichContent.vue'
import {
  getMessageList, sendMessage, markMessageRead, markAllMessageRead, deleteMessage,
  type MessageItem, type MessageListQuery, type SendMessageParams,
} from '@/api/modules/message'
import { getUserList, type UserItem } from '@/api/modules/system'
import { useAuthStore } from '@/stores/auth'
import { useMessageStore } from '@/stores/message'
import { useDict, DictCode } from '@/utils/useDict'
import { storeToRefs } from 'pinia'

const authStore = useAuthStore()
const messageStore = useMessageStore()
const { unreadCount } = storeToRefs(messageStore)
const { options: messageTypeOptions, getLabel: messageTypeLabel } = useDict(DictCode.MessageType)

// ==================== 公共辅助函数 ====================
function typeTag(t: number) {
  const m: Record<number, 'primary' | 'warning' | 'info'> = { 1: 'primary', 2: 'warning', 3: 'info' }
  return m[t] || 'info'
}
function typeLabel(t: number) {
  return messageTypeLabel(t)
}

// ==================== 我的消息状态 ====================
const receivedTableData = ref<MessageItem[]>([])
const receivedLoading = ref(false)
const receivedTotal = ref(0)
const receivedQuery = reactive<MessageListQuery>({ page: 1, pageSize: 10, direction: 1 })

async function loadReceivedData() {
  receivedLoading.value = true
  try {
    const res = await getMessageList(receivedQuery)
    receivedTableData.value = res.data.list || []
    receivedTotal.value = res.data.total
  } finally {
    receivedLoading.value = false
  }
}

function handleReceivedSearch(p: any) {
  receivedQuery.page = 1
  Object.assign(receivedQuery, p)
  loadReceivedData()
}

function handleReceivedReset() {
  receivedQuery.page = 1
  receivedQuery.isRead = undefined
  loadReceivedData()
}

function handleReceivedPageChange(p: any) {
  Object.assign(receivedQuery, p)
  loadReceivedData()
}

// ==================== 发送记录状态 ====================
const sentTableData = ref<MessageItem[]>([])
const sentLoading = ref(false)
const sentTotal = ref(0)
const sentQuery = reactive<MessageListQuery>({ page: 1, pageSize: 10, direction: 2 })

async function loadSentData() {
  sentLoading.value = true
  try {
    const res = await getMessageList(sentQuery)
    sentTableData.value = res.data.list || []
    sentTotal.value = res.data.total
  } finally {
    sentLoading.value = false
  }
}

function handleSentSearch(p: any) {
  sentQuery.page = 1
  Object.assign(sentQuery, p)
  loadSentData()
}

function handleSentReset() {
  sentQuery.page = 1
  loadSentData()
}

function handleSentPageChange(p: any) {
  Object.assign(sentQuery, p)
  loadSentData()
}

// ==================== Tabs ====================
const activeTab = ref<'received' | 'sent'>('received')

const receivedSearchFields = [
  {
    prop: 'isRead', label: '状态', type: 'select' as const,
    options: [
      { label: '未读', value: 0 },
      { label: '已读', value: 1 },
    ],
  },
]
const sentSearchFields: any[] = [] // 发送记录不需要搜索字段

function onTabChange() {
  if (activeTab.value === 'received') {
    loadReceivedData()
  } else {
    loadSentData()
  }
}

onMounted(() => {
  loadReceivedData()
  void messageStore.fetchUnreadCount()
})

// ==================== 发送消息 ====================
const dialogVisible = ref(false)
const submitting = ref(false)
const formDialogRef = ref()

const defaultForm = { title: '', content: '', type: 1, receiverId: 0, targetUserIds: [] as number[] }
const form = reactive({ ...defaultForm })

const selectedUsers = ref<UserItem[]>([])

const formRules: FormRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }],
  targetUserIds: [{ validator: (_rule: any, _value: any, callback: any) => {
    if (form.receiverId === 1 && form.targetUserIds.length === 0) {
      callback(new Error('请选择用户'))
    } else {
      callback()
    }
  }, trigger: 'change' }],
}

function handleSend() {
  Object.assign(form, defaultForm)
  selectedUsers.value = []
  dialogVisible.value = true
}

function onReceiverChange(val: string | number | boolean | undefined) {
  if (val === 0) {
    form.targetUserIds = []
    selectedUsers.value = []
  }
}

function removeSelectedUser(id: number) {
  form.targetUserIds = form.targetUserIds.filter(uid => uid !== id)
  selectedUsers.value = selectedUsers.value.filter(u => u.id !== id)
}

async function handleSendSubmit() {
  submitting.value = true
  try {
    const payload: SendMessageParams = {
      title: form.title,
      content: form.content,
      type: form.type,
    }
    if (form.receiverId === 1) {
      payload.receiverIds = form.targetUserIds
    }
    await sendMessage(payload)
    ElMessage.success('发送成功')
    dialogVisible.value = false
    loadSentData()
    activeTab.value = 'sent'
  } catch (err: any) {
    showRequestError(err, '发送失败')
  } finally {
    submitting.value = false
  }
}

// ==================== 用户选择弹框 ====================
const userPickerVisible = ref(false)
const userPickerList = ref<UserItem[]>([])
const userTotal = ref(0)
const userQuery = reactive({ page: 1, pageSize: 10, keyword: '' })
const userSelectedMap = ref<Map<number, UserItem>>(new Map())

const isAllUserSelected = computed(() => {
  if (userPickerList.value.length === 0) return false
  return userPickerList.value.every(row => userSelectedMap.value.has(row.id))
})
const isUserIndeterminate = computed(() => {
  const count = userPickerList.value.filter(row => userSelectedMap.value.has(row.id)).length
  return count > 0 && count < userPickerList.value.length
})

function openUserPicker() {
  userQuery.page = 1
  userQuery.pageSize = 10
  userQuery.keyword = ''
  userSelectedMap.value.clear()
  for (const user of selectedUsers.value) {
    userSelectedMap.value.set(user.id, user)
  }
  loadUserPicker()
  userPickerVisible.value = true
}

async function loadUserPicker() {
  try {
    const params: any = { page: userQuery.page, pageSize: userQuery.pageSize }
    if (userQuery.keyword) {
      params.username = userQuery.keyword
      params.nickname = userQuery.keyword
    }
    const res = await getUserList(params)
    userPickerList.value = res.data.list || []
    userTotal.value = res.data.total
  } catch (err: any) {
    showRequestError(err, '加载用户失败')
  }
}

function handleUserRowClick(row: any) {
  const user = row as UserItem
  toggleUserRow(user, !userSelectedMap.value.has(user.id))
}

function toggleUserRow(user: UserItem, checked: boolean) {
  if (checked) {
    userSelectedMap.value.set(user.id, user)
  } else {
    userSelectedMap.value.delete(user.id)
  }
}

function toggleSelectAllUsers(checked: boolean) {
  if (checked) {
    for (const row of userPickerList.value) {
      userSelectedMap.value.set(row.id, row)
    }
  } else {
    for (const row of userPickerList.value) {
      userSelectedMap.value.delete(row.id)
    }
  }
}

function userRowClassName({ row }: { row: any }) {
  const user = row as UserItem
  return userSelectedMap.value.has(user.id) ? 'selected-row' : ''
}

function clearUserSelection() {
  userSelectedMap.value.clear()
}

function confirmUserPicker() {
  const users = Array.from(userSelectedMap.value.values())
  selectedUsers.value = users
  form.targetUserIds = users.map(u => u.id)
  userPickerVisible.value = false
}

// ==================== 操作 ====================
async function handleMarkRead(row: MessageItem) {
  try {
    await markMessageRead(row.id)
    ElMessage.success('已标记为已读')
    loadReceivedData()
    void messageStore.fetchUnreadCount()
  } catch (err: any) {
    showRequestError(err, '操作失败')
  }
}

async function handleMarkAllRead() {
  try {
    await ElMessageBox.confirm('确认将所有消息标记为已读？', '提示', { type: 'warning' })
    await markAllMessageRead()
    ElMessage.success('操作成功')
    loadReceivedData()
    void messageStore.fetchUnreadCount()
  } catch { /* 取消 */ }
}

async function handleDelete(row: MessageItem) {
  const isSent = activeTab.value === 'sent'
  try {
    await ElMessageBox.confirm(
      isSent
        ? '确认从发送记录中移除？对方收件箱中的消息不会受影响。'
        : '确认从我的消息中删除？仅对自己不可见。',
      '删除确认',
      { type: 'warning' },
    )
    await deleteMessage(row.id, isSent ? 'sent' : 'received')
    ElMessage.success('删除成功')
    if (isSent) {
      loadSentData()
    } else {
      loadReceivedData()
      void messageStore.fetchUnreadCount()
    }
  } catch { /* 取消 */ }
}

// ==================== 详情 ====================
const detailVisible = ref(false)
const detailData = ref<MessageItem | null>(null)

function handleDetail(row: MessageItem) {
  detailData.value = row
  detailVisible.value = true
}
</script>

<style scoped lang="scss">

.detail-header {
  margin-bottom: 20px;
  h3 { margin: 0 0 12px; font-size: 18px; }
  .detail-meta {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 13px;
    color: #909399;
    span { line-height: 1; }
  }
}
.detail-body {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 4px;
}

// 用户选择区域
.user-select-area {
  display: flex;
  flex-direction: column;
  gap: 8px;

  .selected-user-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
}

// 用户选择弹框
.user-picker-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  padding: 10px 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.user-picker-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}

.user-selected-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
  padding: 8px 12px;
  background: #ecf5ff;
  border-radius: 4px;
  font-size: 13px;
  color: #409eff;
}

:deep(.selected-row) {
  background-color: #ecf5ff !important;
}
</style>
