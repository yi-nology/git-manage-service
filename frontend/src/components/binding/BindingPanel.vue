<template>
  <div class="binding-panel">
    <div class="panel-header">
      <span class="panel-title">远端关联</span>
      <el-button type="primary" size="small" @click="$emit('add')">+ 添加</el-button>
    </div>

    <div v-if="bindings.length === 0" class="empty-state">
      <el-text type="info">暂无远端关联，点击"添加"关联到远端平台</el-text>
    </div>

    <div v-for="b in bindings" :key="b.id" class="binding-card">
      <div class="binding-card-header">
        <el-tag :type="platformTagType(b.platform)" size="small" effect="dark">
          {{ platformLabel(b.platform) }}
        </el-tag>
        <span class="binding-repo-name">{{ b.platform_owner }}/{{ b.platform_repo }}</span>
        <el-tag v-if="b.is_primary" type="warning" size="small" effect="plain">主关联</el-tag>
      </div>
      <div class="binding-card-body">
        <span class="binding-meta">remote: {{ b.remote_name || '未绑定' }}</span>
        <span class="binding-meta">
          Webhook:
          <el-icon v-if="b.has_webhook" color="#67c23a"><SuccessFilled /></el-icon>
          <el-icon v-else color="#909399"><CircleCloseFilled /></el-icon>
        </span>
      </div>
      <div class="binding-card-actions">
        <el-button v-if="!b.is_primary" link type="primary" size="small" @click="$emit('set-primary', b.id)">
          设为主关联
        </el-button>
        <el-button v-if="!b.has_webhook" link type="primary" size="small" @click="$emit('register-webhook', b.id)">
          注册Webhook
        </el-button>
        <el-button v-else link type="warning" size="small" @click="$emit('delete-webhook', b.id)">
          删除Webhook
        </el-button>
        <el-button link type="danger" size="small" @click="handleDelete(b.id)">取消关联</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { SuccessFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import type { RepoProviderBindingDTO } from '@/types/binding'

defineProps<{
  bindings: RepoProviderBindingDTO[]
}>()

const emit = defineEmits<{
  add: []
  delete: [id: number]
  'set-primary': [id: number]
  'register-webhook': [id: number]
  'delete-webhook': [id: number]
}>()

function platformLabel(platform: string) {
  const map: Record<string, string> = { gitlab: 'GitLab', github: 'GitHub', gitea: 'Gitea', tencent_code: '腾讯工蜂' }
  return map[platform] || platform
}

function platformTagType(platform: string) {
  const map: Record<string, string> = { gitlab: 'danger', github: 'info', gitea: 'success', tencent_code: 'success' }
  return (map[platform] || 'info') as 'success' | 'warning' | 'danger' | 'info'
}

function handleDelete(id: number) {
  emit('delete', id)
}
</script>

<style scoped>
.binding-panel {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 16px;
}
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.panel-title {
  font-weight: 600;
  font-size: 14px;
}
.empty-state {
  text-align: center;
  padding: 20px 0;
  color: var(--el-text-color-secondary);
}
.binding-card {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 8px;
}
.binding-card:last-child {
  margin-bottom: 0;
}
.binding-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.binding-repo-name {
  font-weight: 500;
  font-size: 13px;
}
.binding-card-body {
  display: flex;
  gap: 16px;
  margin-bottom: 8px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.binding-meta {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.binding-card-actions {
  display: flex;
  gap: 4px;
}
</style>
