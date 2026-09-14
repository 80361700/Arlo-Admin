<template>
  <el-dialog
    v-model="dialogShow"
    title="请选择组织成员"
    width="720px"
    destroy-on-close
    append-to-body
    :before-close="dialogClose"
  >
    <el-tabs v-model="activeTab">
      <el-tab-pane label="部门" name="dept" />
      <el-tab-pane label="角色" name="role" />
      <el-tab-pane label="用户" name="user" />
    </el-tabs>

    <div class="org-member-body">
      <div class="left">
        <el-input
          v-model="keyword"
          clearable
          placeholder="输入名称"
          :prefix-icon="Search"
          class="search"
        />
        <el-scrollbar height="360px">
          <el-tree
            v-if="activeTab === 'dept'"
            ref="deptTreeRef"
            :data="deptTree"
            node-key="id"
            show-checkbox
            default-expand-all
            :props="{ label: 'name', children: 'children' }"
            :filter-node-method="filterDept"
            @check="syncFromDeptTree"
          />
          <el-checkbox-group v-else-if="activeTab === 'role'" v-model="roleIds" class="check-list">
            <el-checkbox
              v-for="item in filteredRoles"
              :key="item.id"
              :value="String(item.id)"
            >
              {{ item.name }}
            </el-checkbox>
          </el-checkbox-group>
          <el-checkbox-group v-else v-model="userIds" class="check-list">
            <el-checkbox
              v-for="item in filteredUsers"
              :key="item.id"
              :value="String(item.id)"
              :disabled="Number(item.status) === 0"
            >
              {{ userLabel(item) }}
            </el-checkbox>
          </el-checkbox-group>
        </el-scrollbar>
      </div>
      <div class="right">
        <div class="right-title">已选（{{ selected.length }}）</div>
        <el-scrollbar height="400px">
          <div v-if="!selected.length" class="empty">暂无选择</div>
          <div v-for="item in selected" :key="`${item.kind}-${item.id}`" class="selected-item">
            <span>{{ item.name }}</span>
            <el-icon class="remove" @click="removeSelected(item)"><Close /></el-icon>
          </div>
        </el-scrollbar>
      </div>
    </div>

    <template #footer>
      <el-button @click="dialogClose">取消</el-button>
      <el-button type="primary" @click="dialogSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref, watch } from 'vue'
import { Close, Search } from '@element-plus/icons-vue'
import { getAllRoles, getAllUsers, getDeptTree } from '@/api'
import type { DeptTreeNode } from '@/api/modules/system'

export type OrgMemberKind = 'dept' | 'role' | 'user'
export type OrgMemberItem = { id: string | number; name: string; kind: OrgMemberKind }

const emits = defineEmits<{
  success: [list: OrgMemberItem[]]
  close: []
}>()

const dialogShow = ref(false)
const activeTab = ref<OrgMemberKind>('dept')
const keyword = ref('')
const selected = ref<OrgMemberItem[]>([])

const deptTree = ref<DeptTreeNode[]>([])
const deptTreeRef = ref()
const roles = ref<{ id: number; name: string }[]>([])
const users = ref<{ id: number; name: string; username: string; status: number }[]>([])
const roleIds = ref<string[]>([])
const userIds = ref<string[]>([])

const filteredRoles = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  if (!q) return roles.value
  return roles.value.filter((r) => (r.name || '').toLowerCase().includes(q))
})

const filteredUsers = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  if (!q) return users.value
  return users.value.filter((u) => userLabel(u).toLowerCase().includes(q))
})

function userLabel(u: { name?: string; username?: string }) {
  if (u.name && u.name !== u.username) return `${u.name}（${u.username}）`
  return u.name || u.username || ''
}

function filterDept(value: string, data: any) {
  if (!value) return true
  return String(data?.name || '').toLowerCase().includes(value.toLowerCase())
}

watch(keyword, (val) => {
  if (activeTab.value === 'dept') deptTreeRef.value?.filter(val)
})

watch(activeTab, () => {
  keyword.value = ''
  if (activeTab.value === 'dept') nextTick(() => syncDeptCheckedKeys())
})

watch(roleIds, () => {
  const map = new Map(roles.value.map((r) => [String(r.id), r.name]))
  selected.value = [
    ...selected.value.filter((s) => s.kind !== 'role'),
    ...roleIds.value
      .map((id) => ({ id, name: map.get(String(id)) || String(id), kind: 'role' as const }))
      .filter((s) => s.name),
  ]
})

watch(userIds, () => {
  const map = new Map(users.value.map((u) => [String(u.id), userLabel(u)]))
  selected.value = [
    ...selected.value.filter((s) => s.kind !== 'user'),
    ...userIds.value
      .map((id) => ({ id, name: map.get(String(id)) || String(id), kind: 'user' as const }))
      .filter((s) => s.name),
  ]
})

function findDeptName(nodes: DeptTreeNode[], id: string): string {
  for (const n of nodes) {
    if (String(n.id) === id) return n.name
    if (n.children?.length) {
      const found = findDeptName(n.children, id)
      if (found) return found
    }
  }
  return ''
}

function syncFromDeptTree() {
  const keys: string[] = (deptTreeRef.value?.getCheckedKeys?.(false) || []).map(String)
  selected.value = [
    ...selected.value.filter((s) => s.kind !== 'dept'),
    ...keys
      .map((id) => ({ id, name: findDeptName(deptTree.value, id) || id, kind: 'dept' as const }))
      .filter((s) => s.name),
  ]
}

function syncDeptCheckedKeys() {
  const keys = selected.value.filter((s) => s.kind === 'dept').map((s) => String(s.id))
  deptTreeRef.value?.setCheckedKeys?.(keys)
}

function removeSelected(item: OrgMemberItem) {
  selected.value = selected.value.filter((s) => !(s.kind === item.kind && String(s.id) === String(item.id)))
  if (item.kind === 'dept') syncDeptCheckedKeys()
  if (item.kind === 'role') roleIds.value = roleIds.value.filter((id) => String(id) !== String(item.id))
  if (item.kind === 'user') userIds.value = userIds.value.filter((id) => String(id) !== String(item.id))
}

async function init(obj: { selectData?: OrgMemberItem[] } = {}) {
  dialogShow.value = true
  activeTab.value = 'dept'
  keyword.value = ''
  selected.value = JSON.parse(JSON.stringify(obj.selectData || []))

  try {
    const [deptRes, roleRes, userRes] = await Promise.all([getDeptTree(), getAllRoles(), getAllUsers()])
    deptTree.value = deptRes.data || []
    roles.value = (roleRes.data || []).map((r: any) => ({ id: r.id, name: r.name || r.roleName || String(r.id) }))
    users.value = userRes.data || []
  } catch {
    deptTree.value = []
    roles.value = []
    users.value = []
  }

  roleIds.value = selected.value.filter((s) => s.kind === 'role').map((s) => String(s.id))
  userIds.value = selected.value.filter((s) => s.kind === 'user').map((s) => String(s.id))

  await nextTick()
  syncDeptCheckedKeys()
}

function dialogClose() {
  dialogShow.value = false
  emits('close')
}

function dialogSubmit() {
  dialogShow.value = false
  emits('success', selected.value.map((s) => ({ ...s })))
}

defineExpose({ init })
</script>

<style lang="scss" scoped>
.org-member-body {
  display: flex;
  gap: 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  overflow: hidden;
  .left,
  .right {
    flex: 1;
    padding: 12px;
    min-width: 0;
  }
  .left {
    border-right: 1px solid var(--el-border-color-lighter);
  }
  .search {
    margin-bottom: 10px;
  }
  .check-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    :deep(.el-checkbox) {
      margin-right: 0;
      height: auto;
      white-space: normal;
    }
  }
  .right-title {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    margin-bottom: 8px;
  }
  .empty {
    color: var(--el-text-color-placeholder);
    font-size: 13px;
    padding: 12px 0;
  }
  .selected-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 8px;
    margin-bottom: 6px;
    background: var(--el-fill-color-light);
    border-radius: 4px;
    font-size: 13px;
    .remove {
      cursor: pointer;
      color: var(--el-text-color-secondary);
      &:hover {
        color: var(--el-color-danger);
      }
    }
  }
}
</style>
