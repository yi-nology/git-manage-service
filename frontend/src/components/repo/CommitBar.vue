<template>
  <div class="fe-commit-bar">
    <el-input
      :model-value="commitMsg"
      placeholder="提交信息 (Ctrl+Enter)..."
      size="small"
      class="fe-commit-input"
      @update:model-value="$emit('update:commitMsg', $event)"
      @keydown.ctrl.enter="$emit('commit')"
    />
    <el-button
      size="small"
      :loading="generatingMsg"
      @click="$emit('generateMsg')"
      title="AI 生成提交信息"
    >
      <el-icon><MagicStick /></el-icon>
    </el-button>
    <el-select
      :model-value="selectedAuthor"
      size="small"
      placeholder="选择作者"
      class="fe-author-select"
      clearable
      @update:model-value="$emit('update:selectedAuthor', $event)"
    >
      <el-option
        v-for="a in authors"
        :key="a.id"
        :label="a.canonicalName + ' <' + a.canonicalEmail + '>'"
        :value="a.id"
      />
    </el-select>
    <div class="fe-push-toggle">
      <span>推送</span>
      <el-switch
        :model-value="pushAfterCommit"
        size="small"
        @update:model-value="$emit('update:pushAfterCommit', $event)"
      />
    </div>
    <el-button
      type="primary"
      size="small"
      :loading="committing"
      :disabled="!commitMsg"
      @click="$emit('commit')"
    >
      <el-icon><Check /></el-icon> 提交
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { Check, MagicStick } from '@element-plus/icons-vue'
import type { AuthorIdentityDTO } from '@/api/modules/author'

defineProps<{
  commitMsg: string
  selectedAuthor: number | null
  pushAfterCommit: boolean
  authors: AuthorIdentityDTO[]
  committing: boolean
  generatingMsg: boolean
}>()

defineEmits<{
  'update:commitMsg': [value: string]
  'update:selectedAuthor': [value: number | null]
  'update:pushAfterCommit': [value: boolean]
  commit: []
  generateMsg: []
}>()
</script>

<style scoped>
.fe-commit-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-color-page);
  flex-shrink: 0;
}

.fe-commit-input {
  flex: 1;
}

.fe-author-select {
  width: 200px;
}

.fe-push-toggle {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-color-secondary);
  white-space: nowrap;
}
</style>
