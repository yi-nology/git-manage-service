<template>
  <div class="audit-log-page">
    <PageHeader title="操作审计日志">
      <template #actions>
        <ActionPill variant="outline" :icon="RefreshRight" @click="loadLogs">刷新</ActionPill>
      </template>
    </PageHeader>

    <form class="filter-bar" @submit.prevent="loadLogs">
      <select v-model="filterAction" class="filter-select" @change="loadLogs">
        <option value="">全部操作</option>
        <option v-for="(label, key) in actionLabelMap" :key="key" :value="key">{{ label }}</option>
      </select>
      <input v-model="filterTarget" placeholder="目标对象" class="filter-input" />
      <div class="filter-spacer"></div>
      <ActionPill variant="outline" :icon="Search">搜索</ActionPill>
    </form>

    <DataTable :columns="columns" :data="logs" :loading="loading">
      <template #cell-created_at="{ row }">
        <span class="time-cell">{{ formatDate(row.created_at) }}</span>
      </template>
      <template #cell-action="{ row }">
        <StatusBadge
          :variant="(getActionClass(row.action) as any)"
          :text="getActionLabel(row.action)"
          :show-dot="false"
        />
      </template>
      <template #cell-target="{ row }">
        <span class="repo-cell">{{ formatTarget(row.target) }}</span>
      </template>
      <template #cell-details="{ row }">
        <span class="detail-cell">{{ row.details || '-' }}</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge
          :variant="(getStatusClass(row.action) as any)"
          text="成功"
          :show-dot="false"
        />
      </template>
      <template #empty>
        <EmptyState title="暂无审计日志" />
      </template>
    </DataTable>

    <PaginationBar
      v-if="totalCount > 0"
      :total="totalCount"
      v-model:currentPage="currentPage"
      :page-size="pageSize"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { RefreshRight, Search } from '@element-plus/icons-vue'
import { getAuditLogs } from '@/api/modules/audit'
import { getRepoList } from '@/api/modules/repo'
import type { AuditLogDTO } from '@/types/stats'
import { formatDate } from '@/utils/format'
import PageHeader from '@/components/common/PageHeader.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'

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

const columns: TableColumn[] = [
  { key: 'created_at', label: '时间', width: '160px' },
  { key: 'action', label: '操作', width: '120px' },
  { key: 'target', label: '仓库', width: '160px' },
  { key: 'details', label: '详情', flex: 1 },
  { key: 'status', label: '状态', width: '100px' },
]

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

watch(currentPage, () => {
  loadLogs()
})

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
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 12px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
}

.filter-select {
  padding: 6px 10px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  outline: none;
  min-width: 140px;
}

.filter-select:focus {
  border-color: var(--accent-primary);
}

.filter-input {
  padding: 6px 10px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  outline: none;
  width: 180px;
}

.filter-input:focus {
  border-color: var(--accent-primary);
}

.filter-spacer {
  flex: 1;
}

.time-cell {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.repo-cell {
  font-size: 13px;
  color: var(--accent-primary);
}

.detail-cell {
  font-size: 12px;
  color: var(--text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
