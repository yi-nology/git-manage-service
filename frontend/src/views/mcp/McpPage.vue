<template>
  <div class="mcp-page">
    <div class="title-row">
      <div class="title-left">
        <h2 class="page-title">MCP 配置</h2>
        <p class="page-subtitle">多模型协作平台连接配置</p>
      </div>
    </div>

    <div class="conn-card">
      <div class="conn-header">
        <span class="conn-title">连接参数</span>
        <span class="conn-status" :class="mcpEnabled ? 'status-connected' : 'status-disconnected'">
          {{ mcpEnabled ? '已连接' : '未连接' }}
        </span>
      </div>
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
        <button class="action-primary" @click="saveConfig">保存配置</button>
        <button class="action-outline" @click="testConnection">测试连接</button>
      </div>
    </div>

    <div class="section-title">可用工具 ({{ tools.length }})</div>

    <div class="tools-grid">
      <div v-for="tool in tools" :key="tool.name" class="tool-card">
        <span class="tool-name">{{ tool.name }}</span>
        <span class="tool-desc">{{ tool.desc }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

const mcpEnabled = ref(false)

const mcpConfig = reactive({
  serviceUrl: 'http://localhost:3002',
  apiKey: '',
  port: 3002
})

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

.title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.page-subtitle {
  margin: 0;
  font-size: 13px;
  font-weight: normal;
  color: var(--text-color-secondary);
}

.conn-card {
  border-radius: 12px;
  background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.conn-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.conn-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.conn-status {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
}

.conn-status.status-connected {
  background: #ECFDF5;
  color: #10B981;
}

.conn-status.status-disconnected {
  background: var(--accent-bg, #EEF2FF);
  color: var(--text-color-secondary);
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
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page, #fff);
  outline: none;
  width: 100%;
  box-sizing: border-box;
}

.field-input:focus {
  border-color: var(--accent-primary, #6366F1);
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
  background: var(--accent-primary, #6366F1);
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

.action-primary {
  padding: 8px 16px;
  border-radius: 6px;
  border: none;
  background: var(--accent-primary, #6366F1);
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.action-primary:hover {
  opacity: 0.9;
}

.action-outline {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: none;
  color: var(--text-color-regular);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.action-outline:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.tools-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.tool-card {
  border-radius: 8px;
  background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.tool-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--accent-primary, #6366F1);
}

.tool-desc {
  font-size: 11px;
  color: var(--text-color-secondary);
}
</style>
