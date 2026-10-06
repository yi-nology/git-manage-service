<template>
  <div class="audit-log-page">
    <PageHeader title="操作审计日志">
      <template #actions>
        <ActionPill variant="outline" :icon="RefreshRight" @click="loadLogs">刷新</ActionPill>
      </template>
    </PageHeader>

    <form class="filter-bar" @submit.prevent="loadLogs">
      <el-select
        v-model="filterAction"
        placeholder="全部操作"
        clearable
        class="filter-select"
        @change="loadLogs"
      >
        <el-option-group
          v-for="group in actionGroups"
          :key="group.label"
          :label="group.label"
        >
          <el-option
            v-for="item in group.items"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-option-group>
      </el-select>
      <input v-model="filterTarget" placeholder="目标对象" class="filter-input" />
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        value-format="YYYY-MM-DD"
        class="filter-date"
        @change="loadLogs"
      />
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
        <el-popover v-if="row.details && row.details !== '-'" trigger="click" :width="420" placement="top">
          <template #reference>
            <span class="detail-cell clickable">{{ truncateDetails(row.details) }}</span>
          </template>
          <pre class="details-json">{{ formatDetailsJSON(row.details) }}</pre>
        </el-popover>
        <span v-else class="detail-cell">-</span>
      </template>
      <template #cell-ip_address="{ row }">
        <span class="ip-cell">{{ row.ip_address || '-' }}</span>
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
      v-if="total_count > 0"
      :total="total_count"
      v-model:currentPage="currentPage"
      :page-size="page_size"
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
import {
  actionGroups, targetTypeMap, getActionLabel, truncateDetails,
  formatDetailsJSON, getActionClass, getStatusClass,
} from './auditMeta'

const repoNameMap = ref<Record<string, string>>({})

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
  { key: 'action', label: '操作', width: '150px' },
  { key: 'target', label: '目标', width: '160px' },
  { key: 'details', label: '详情', flex: 1 },
  { key: 'ip_address', label: 'IP 地址', width: '130px' },
  { key: 'status', label: '状态', width: '80px' },
]

const loading = ref(false)
const logs = ref<AuditLogDTO[]>([])
const total_count = ref(0)
const currentPage = ref(1)
const page_size = 20

const filterAction = ref('')
const filterTarget = ref('')
const dateRange = ref<string[] | null>(null)


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
    const params: Record<string, unknown> = {
      page: currentPage.value,
      page_size: page_size,
      action: filterAction.value || undefined,
      target: filterTarget.value || undefined,
    }
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res = await getAuditLogs(params)
    logs.value = res.items || []
    total_count.value = res.total || 0
  } catch {
    logs.value = []
    total_count.value = 0
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
  min-width: 180px;
}

.filter-input {
  padding: 6px 10px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  outline: none;
  width: 160px;
}

.filter-input:focus {
  border-color: var(--accent-primary);
}

.filter-date {
  width: 260px;
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

.detail-cell.clickable {
  cursor: pointer;
  text-decoration: underline dotted;
  text-underline-offset: 2px;
}

.detail-cell.clickable:hover {
  color: var(--accent-primary);
}

.ip-cell {
  font-size: 12px;
  color: var(--text-color-secondary);
  font-family: monospace;
}

.details-json {
  margin: 0;
  padding: 8px 12px;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 300px;
  overflow-y: auto;
  background: var(--bg-color-page);
  border-radius: 6px;
  color: var(--text-color-primary);
}
</style>
