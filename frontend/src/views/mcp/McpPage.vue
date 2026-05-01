<template>
  <div class="mcp-page">
    <PageHeader title="MCP 配置" subtitle="多模型协作平台连接配置" />

    <FormCard>
      <template #header>
        <span class="conn-title">连接参数</span>
        <StatusBadge :variant="mcpEnabled ? 'success' : 'default'" :text="mcpEnabled ? '已连接' : '未连接'" />
      </template>
      <div class="conn-row">
        <div class="conn-field">
          <label class="field-label">Host</label>
          <input v-model="mcpConfig.serviceUrl" placeholder="请输入 MCP 服务地址" class="field-input" />
        </div>
        <div class="conn-field">
          <label class="field-label">Port</label>
          <input v-model.number="mcpConfig.port" type="number" :min="1" :max="65535" class="field-input" />
        </div>
        <div class="conn-field">
          <label class="field-label">状态</label>
          <div class="toggle-row">
            <button class="toggle-btn" :class="{ active: mcpEnabled }" @click="mcpEnabled = !mcpEnabled">
              <span class="toggle-dot" :class="{ right: mcpEnabled }"></span>
            </button>
            <span class="toggle-label">{{ mcpEnabled ? '已开启' : '已关闭' }}</span>
          </div>
        </div>
      </div>
      <div class="conn-actions">
        <ActionPill variant="primary" @click="saveConfig">保存配置</ActionPill>
        <ActionPill variant="outline" @click="testConnection">测试连接</ActionPill>
      </div>
    </FormCard>

    <SectionTitle>可用工具 ({{ tools.length }})</SectionTitle>

    <div class="tools-grid">
      <div v-for="tool in tools" :key="tool.name" class="tool-card">
        <div class="tool-header">
          <span class="tool-icon" :style="{ background: toolMeta(tool.name).bg, color: toolMeta(tool.name).fg }">
            {{ toolMeta(tool.name).initial }}
          </span>
          <span class="tool-name">{{ tool.name }}</span>
        </div>
        <span class="tool-desc">{{ tool.desc }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import FormCard from '@/components/common/FormCard.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const mcpEnabled = ref(false)

const mcpConfig = reactive({
  serviceUrl: 'http://localhost:3002',
  apiKey: '',
  port: 3002
})

const TOOL_META: Record<string, { initial: string; bg: string; fg: string }> = {
  repo_list: { initial: 'R', bg: '#EEF2FF', fg: '#6366F1' },
  repo_info: { initial: 'R', bg: '#EEF2FF', fg: '#6366F1' },
  branch_list: { initial: 'B', bg: '#ECFDF5', fg: '#10B981' },
  git_commit_search: { initial: 'C', bg: '#FFF7ED', fg: '#F97316' },
  git_tag_create: { initial: 'T', bg: '#FDF2F8', fg: '#EC4899' },
  git_diff: { initial: 'D', bg: '#FEF3C7', fg: '#D97706' },
  git_push: { initial: 'P', bg: '#DBEAFE', fg: '#3B82F6' },
  git_pull: { initial: 'P', bg: '#DBEAFE', fg: '#3B82F6' },
  git_stash: { initial: 'S', bg: '#F3E8FF', fg: '#8B5CF6' },
  git_merge: { initial: 'M', bg: '#E0F2FE', fg: '#0EA5E9' },
  git_stats: { initial: 'G', bg: '#F0FDF4', fg: '#22C55E' },
  sync_task: { initial: 'S', bg: '#FEF2F2', fg: '#EF4444' },
  file_read: { initial: 'F', bg: '#F5F3FF', fg: '#7C3AED' },
}

function toolMeta(name: string) {
  return TOOL_META[name] || { initial: name[0]?.toUpperCase() || '?', bg: '#F3F4F6', fg: '#6B7280' }
}

const tools = [
  { name: 'repo_list', desc: '列出仓库' },
  { name: 'repo_info', desc: '获取仓库详细信息' },
  { name: 'branch_list', desc: '列出所有分支信息' },
  { name: 'git_commit_search', desc: '搜索仓库提交历史记录' },
  { name: 'git_tag_create', desc: '创建标签并可选推送' },
  { name: 'git_diff', desc: '查看文件或分支差异' },
  { name: 'git_push', desc: '推送分支到远程仓库' },
  { name: 'git_pull', desc: '拉取远程分支更新' },
  { name: 'git_stash', desc: '管理暂存区变更' },
  { name: 'git_merge', desc: '合并分支并处理冲突' },
  { name: 'git_stats', desc: '获取代码统计信息' },
  { name: 'sync_task', desc: '管理同步任务执行' },
  { name: 'file_read', desc: '读取仓库文件内容' },
]

const saveConfig = () => {
  console.log('保存配置:', mcpConfig)
}

const testConnection = () => {
  console.log('测试连接:', mcpConfig.serviceUrl)
}
</script>

<style scoped>
.mcp-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.conn-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.conn-row {
  display: flex;
  gap: 16px;
}

.conn-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.field-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-color-secondary);
}

.field-input {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  outline: none;
  width: 100%;
  box-sizing: border-box;
}

.field-input:focus {
  border-color: var(--accent-primary);
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 36px;
}

.toggle-btn {
  position: relative;
  width: 40px;
  height: 22px;
  border-radius: 11px;
  border: none;
  background: var(--border-color, #d1d5db);
  cursor: pointer;
  transition: background 0.2s;
  padding: 0;
}

.toggle-btn.active {
  background: var(--accent-primary);
}

.toggle-dot {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #fff;
  transition: left 0.2s;
}

.toggle-dot.right {
  left: 20px;
}

.toggle-label {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.conn-actions {
  display: flex;
  gap: 8px;
}

.tools-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 12px;
}

.tool-card {
  border-radius: 8px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.tool-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tool-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 700;
  flex-shrink: 0;
}

.tool-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--accent-primary);
}

.tool-desc {
  font-size: 11px;
  color: var(--text-color-secondary);
}
</style>
