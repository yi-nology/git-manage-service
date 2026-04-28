<template>
  <div class="audit-log-page">
    <div class="title-row">
      <div class="title-left">
        <h2 class="page-title">操作审计日志</h2>
      </div>
      <button class="refresh-btn" @click="loadLogs">
        <el-icon><RefreshRight /></el-icon>
        刷新
      </button>
    </div>

    <div class="filter-bar">
      <div class="filter-item">
        <select v-model="filterAction" class="filter-select" @change="loadLogs">
          <option value="">全部操作</option>
          <option v-for="(label, key) in actionLabelMap" :key="key" :value="key">{{ label }}</option>
        </select>
      </div>
      <div class="filter-item">
        <input v-model="filterTarget" placeholder="目标对象" class="filter-input" @keyup.enter="loadLogs" />
      </div>
      <div class="filter-spacer"></div>
      <button class="filter-search-btn" @click="loadLogs">
        <el-icon><Search /></el-icon>
        搜索
      </button>
    </div>

    <div v-if="loading" class="loading-card">
      <div class="loading-spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="logs.length === 0" class="empty-card">
      <div class="empty-icon">
        <el-icon :size="32"><Warning /></el-icon>
      </div>
      <div class="empty-text">暂无审计日志</div>
    </div>

    <div v-else class="table-card">
      <div class="table-header">
        <span class="th" style="width:160px">时间</span>
        <span class="th" style="width:120px">操作</span>
        <span class="th" style="width:160px">仓库</span>
        <span class="th" style="flex:1">详情</span>
        <span class="th" style="width:100px">状态</span>
      </div>
      <div v-for="(log, idx) in logs" :key="idx" class="table-row">
        <span class="td time-cell" style="width:160px">{{ formatDate(log.created_at) }}</span>
        <span class="td" style="width:120px">
          <span class="action-tag" :class="getActionClass(log.action)">{{ getActionLabel(log.action) }}</span>
        </span>
        <span class="td repo-cell" style="width:160px">{{ formatTarget(log.target) }}</span>
        <span class="td detail-cell" style="flex:1">{{ log.details || '-' }}</span>
        <span class="td" style="width:100px">
          <span class="status-tag" :class="getStatusClass(log.action)">成功</span>
        </span>
      </div>
    </div>

    <div v-if="totalCount > 0" class="pagination-bar">
      <span class="pagination-info">
        显示 {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, totalCount) }} 共 {{ totalCount }} 条
      </span>
      <div class="pagination-btns">
        <button class="page-btn" :disabled="currentPage <= 1" @click="currentPage--; loadLogs()">上一页</button>
        <span class="page-num">{{ currentPage }}</span>
        <button class="page-btn" :disabled="currentPage * pageSize >= totalCount" @click="currentPage++; loadLogs()">下一页</button>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Warning, RefreshRight, Search } from '@element-plus/icons-vue'
import { getAuditLogs } from '@/api/modules/audit'
import { getRepoList } from '@/api/modules/repo'
import type { AuditLogDTO } from '@/types/stats'
import { formatDate } from '@/utils/format'

const repoNameMap = ref<Record<string, string>>({})

const actionLabelMap: Record<string, string> = {
  CREATE: '创建',
  UPDATE: '更新',
  DELETE: '删除',
  FETCH_REPO: '拉取仓库',
  CREATE_BRANCH: '创建分支',
  DELETE_BRANCH: '删除分支',
  UPDATE_BRANCH: '更新分支',
  CHECKOUT_BRANCH: '切换分支',
  PUSH_BRANCH: '推送分支',
  PULL_BRANCH: '拉取分支',
  MERGE_CONFLICT: '合并冲突',
  MERGE_SUCCESS: '合并成功',
  CHERRY_PICK: 'Cherry-Pick',
  REBASE: '变基',
  REBASE_ABORT: '中止变基',
  REBASE_CONTINUE: '继续变基',
  SUBMODULE_ADD: '添加子模块',
  SUBMODULE_INIT: '初始化子模块',
  SUBMODULE_UPDATE: '更新子模块',
  SUBMODULE_SYNC: '同步子模块',
  SUBMODULE_REMOVE: '删除子模块',
  STASH_SAVE: '保存暂存',
  STASH_APPLY: '应用暂存',
  STASH_POP: '弹出暂存',
  STASH_DROP: '丢弃暂存',
  STASH_CLEAR: '清空暂存',
  SYNC: '同步',
  SYNC_ADHOC: '手动同步',
  SUBMIT_CHANGES: '提交变更',
  WEBHOOK_TRIGGER: 'Webhook 触发',
  WEBHOOK_TRIGGER_BY_TOKEN: 'Token 触发',
  NOTIFICATION_CHANNEL_CREATE: '创建通知渠道',
  NOTIFICATION_CHANNEL_UPDATE: '更新通知渠道',
  NOTIFICATION_CHANNEL_DELETE: '删除通知渠道',
}

function getActionLabel(action: string): string {
  return actionLabelMap[action] || action
}

const targetTypeMap: Record<string, string> = {
  repo: '仓库',
  task: '同步任务',
  task_key: '同步任务',
  channel: '通知渠道',
}

function formatTarget(target: string): string {
  const sepIdx = target.indexOf(':')
  if (sepIdx === -1) return target
  const prefix = target.substring(0, sepIdx)
  const value = target.substring(sepIdx + 1)
  const label = targetTypeMap[prefix]
  if (!label) return target
  if (prefix === 'repo') {
    const name = repoNameMap.value[value]
    return name ? `${label}: ${name}` : `${label}: ${value}`
  }
  return `${label}: ${value}`
}

const loading = ref(false)
const logs = ref<AuditLogDTO[]>([])
const totalCount = ref(0)
const currentPage = ref(1)
const pageSize = 20

const filterAction = ref('')
const filterTarget = ref('')

function getActionClass(action: string): string {
  if (action.includes('DELETE') || action.includes('REMOVE') || action === 'MERGE_CONFLICT') return 'danger'
  if (action.includes('CREATE') || action === 'MERGE_SUCCESS' || action === 'SUBMODULE_ADD') return 'success'
  if (action.includes('UPDATE') || action.includes('PUSH') || action === 'SUBMIT_CHANGES') return 'warning'
  return 'info'
}

function getStatusClass(action: string): string {
  if (action === 'MERGE_CONFLICT') return 'danger'
  return 'success'
}

onMounted(async () => {
  try {
    const repos = await getRepoList() || []
    const map: Record<string, string> = {}
    for (const r of repos) {
      map[r.key] = r.name
    }
    repoNameMap.value = map
  } catch {
    // ignore
  }
  loadLogs()
})

async function loadLogs() {
  loading.value = true
  try {
    const res = await getAuditLogs({
      page: currentPage.value,
      page_size: pageSize,
      action: filterAction.value || undefined,
      target: filterTarget.value || undefined,
    })
    logs.value = res.items || []
    totalCount.value = res.total || 0
  } catch {
    logs.value = []
    totalCount.value = 0
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.audit-log-page {
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

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  color: var(--text-color-regular);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.refresh-btn:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 12px;
  background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
}

.filter-select {
  padding: 6px 10px;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page, #fff);
  outline: none;
  min-width: 140px;
}

.filter-select:focus {
  border-color: var(--accent-primary, #6366F1);
}

.filter-input {
  padding: 6px 10px;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page, #fff);
  outline: none;
  width: 180px;
}

.filter-input:focus {
  border-color: var(--accent-primary, #6366F1);
}

.filter-spacer {
  flex: 1;
}

.filter-search-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 10px;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 4px;
  background: none;
  color: var(--text-color-regular);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-search-btn:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.loading-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
  border-radius: 12px;
  background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  color: var(--text-color-secondary);
  font-size: 13px;
}

.loading-spinner {
  width: 24px;
  height: 24px;
  border: 3px solid var(--border-color, #e5e7eb);
  border-top-color: var(--accent-primary, #6366F1);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 48px 24px;
  border-radius: 12px;
  background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
}

.empty-icon {
  color: var(--text-color-placeholder, #9ca3af);
  margin-bottom: 4px;
}

.empty-text {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.table-card {
  border-radius: 12px;
  background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  overflow: hidden;
}

.table-header {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  background: var(--accent-bg, #EEF2FF);
}

.th {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color-secondary);
}

.table-row {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
}

.table-row:last-child {
  border-bottom: none;
}

.td {
  font-size: 13px;
  color: var(--text-color-regular);
}

.time-cell {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.repo-cell {
  font-size: 13px;
  color: var(--accent-primary, #6366F1);
}

.detail-cell {
  font-size: 12px;
  color: var(--text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
}

.action-tag.success {
  background: #ECFDF5;
  color: #10B981;
}

.action-tag.warning {
  background: #FFFBEB;
  color: #F59E0B;
}

.action-tag.danger {
  background: #FEF2F2;
  color: #EF4444;
}

.action-tag.info {
  background: #EEF2FF;
  color: #6366F1;
}

.status-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
}

.status-tag.success {
  background: #ECFDF5;
  color: #10B981;
}

.status-tag.danger {
  background: #FEF2F2;
  color: #EF4444;
}

.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination-info {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.pagination-btns {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn {
  padding: 6px 12px;
  border-radius: 6px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  color: var(--text-color-regular);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-num {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color-primary);
  min-width: 30px;
  text-align: center;
}

.detail-content {
  background: var(--bg-color);
  padding: 12px;
  border-radius: 8px;
  max-height: 500px;
  overflow: auto;
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 13px;
}
</style>
