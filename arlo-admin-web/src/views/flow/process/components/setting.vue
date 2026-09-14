<template>
  <el-form :model="modelValue.flow.processSetting" label-width="160px" class="setting-form">
    <el-divider content-position="center">提交人权限</el-divider>
    <el-form-item label="撤销审批中的申请">
      <div class="item">
        <el-checkbox v-model="modelValue.flow.processSetting.allowRevocation" label="允许撤销审批中的申请" />
        <div class="tips">第一个审批节点通过后，提交人仍可撤销申请（配置前已发起的申请不生效）</div>
      </div>
    </el-form-item>

    <el-divider content-position="center">审批操作</el-divider>
    <el-form-item label="委托">
      <div class="item">
        <el-checkbox v-model="modelValue.flow.processSetting.allowDelegate" label="允许委托他人审批" />
        <div class="tips">关闭后本流程待办不会因个人委托进入受托人列表，受托人也无法代批办理</div>
      </div>
    </el-form-item>
    <el-form-item label="批量操作">
      <el-checkbox v-model="modelValue.flow.processSetting.allowBatchOperate" label="允许批量同意/拒绝" />
    </el-form-item>
    <el-form-item label="二次确认">
      <el-checkbox v-model="modelValue.flow.processSetting.secondOperatePrompt" label="关键操作二次确认提示" />
    </el-form-item>
    <el-form-item label="重复审批">
      <div class="item">
        <el-checkbox v-model="modelValue.flow.processSetting.repeatOperateSkip" label="同一人重复节点自动跳过" />
        <div class="tips">同一审批人在本流程中已同意过，再次出现为审批人时自动跳过（认领池角色节点除外）</div>
      </div>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
defineProps<{ modelValue: any }>()
</script>

<style lang="scss" scoped>
.setting-form {
  // max-width: 720px;
}
.tips {
  font-size: 13px;
  color: #999;
  line-height: 22px;
  margin-top: 4px;
}
.item {
  width: 100%;
}
</style>
