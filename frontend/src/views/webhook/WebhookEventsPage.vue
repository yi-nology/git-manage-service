<template>
  <div class="webhook-page-wrapper">
    <PageHeader :title="repoName || '仓库'" showBack :backRoute="`/local-repos/${repoKey}`">
      <template #title-suffix>
        <span v-if="currentVersion" class="version-tag">{{ currentVersion }}</span>
      </template>
      <template #actions>
        <ActionPill variant="outline" :icon="Refresh" @click="loadEvents">刷新</ActionPill>
      </template>
    </PageHeader>

    <div class="webhook-layout">
      <RepoSidebar :repo-key="repoKey" active-key="webhooks" />
      <div class="webhook-content">
        <div class="webhook-events-page">
          <StatsRow :stats="webhookStats" />

          <DataTable v-if="events.length > 0 || loading" :columns="eventColumns" :data="events" :loading="loading" row-key="id">
            <template #cell-event_type="{ row }">
              <span class="event-type">
                <el-icon :size="14" style="color:#6366F1"><Share /></el-icon>
                {{ row.event_type }}
              </span>
            </template>
            <template #cell-source="{ row }">
              <span class="platform-tag">
                <span class="platform-dot" :style="{ background: platformColor(row.source) }"></span>
                {{ row.source }}
              </span>
            </template>
            <template #cell-event_id="{ row }">
              <span class="mono" :title="row.event_id">{{ row.event_id.substring(0, 16) }}...</span>
            </template>
            <template #cell-status="{ row }">
              <StatusBadge :variant="whStatusVariant(row.status)" :text="statusLabel(row.status)" :showDot="false" />
            </template>
            <template #cell-created_at="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
            <template #row-actions="{ row }">
              <ActionPill v-if="row.status === 'failed'" variant="amber" small @click="handleRetry(row)">重试</ActionPill>
            </template>
          </DataTable>

          <EmptyState v-else-if="!loading" title="暂无事件" description="Webhook 事件将在此处显示" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, TrendCharts, CircleCheck, CircleClose, Share } from '@element-plus/icons-vue'
import { listWebhookEvents, retryWebhookEvent } from '@/api/modules/webhook-event'
import type { WebhookEventDTO } from '@/api/modules/webhook-event'
import RepoSidebar from '@/components/repo/RepoSidebar.vue'
import { getRepoDetail } from '@/api/modules/repo'
import { getCurrentVersion } from '@/api/modules/version'
import PageHeader from '@/components/common/PageHeader.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import StatsRow from '@/components/common/StatsRow.vue'

const route = useRoute()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const events = ref<WebhookEventDTO[]>([])
const total = ref(0)
const repoName = ref('')
const currentVersion = ref('')

const processedCount = computed(() => events.value.filter(e => e.status === 'processed').length)
const failedCount = computed(() => events.value.filter(e => e.status === 'failed').length)

const webhookStats = computed(() => [
  { label: '总事件', value: total.value, icon: TrendCharts, color: '#6366F1' },
  { label: '成功', value: processedCount.value, icon: CircleCheck, color: '#10B981' },
  { label: '失败', value: failedCount.value, icon: CircleClose, color: '#EF4444' },
])

const eventColumns: TableColumn[] = [
  { key: 'event_type', label: '事件类型', width: '140px' },
  { key: 'source', label: '来源', width: '80px' },
  { key: 'event_id', label: '事件 ID' },
  { key: 'actor_name', label: '触发者', width: '100px' },
  { key: 'status', label: '状态', width: '80px' },
  { key: 'created_at', label: '时间', width: '160px' },
]

const PLATFORM_COLORS: Record<string, string> = { gitlab: '#FC6D26', github: '#24292F', gitea: '#609926', gitee: '#C71D23', forgejo: '#F97316', tencent_code: '#1B5E20' }
function platformColor(s: string) { return PLATFORM_COLORS[s?.toLowerCase()] || '#6B7280' }
function statusLabel(s: string) {
  if (s === 'processed') return '已处理'
  if (s === 'received') return '待处理'
  if (s === 'failed') return '失败'
  return s
}
function whStatusVariant(s: string): 'success' | 'warning' | 'danger' | 'default' {
  if (s === 'processed') return 'success'
  if (s === 'received') return 'warning'
  if (s === 'failed') return 'danger'
  return 'default'
}
function formatTime(t: string) {
  if (!t) return '-'
  const d = new Date(t)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return `${d.getMonth() + 1}-${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

async function loadEvents() {
  loading.value = true
  try {
    const res = await listWebhookEvents({ page: 1, page_size: 50 })
    events.value = res?.items || []
    total.value = res?.total || 0
  } catch { events.value = [] }
  finally { loading.value = false }
}

async function handleRetry(row: WebhookEventDTO) {
  try {
    await retryWebhookEvent(row.id)
    ElMessage.success('已重试')
    loadEvents()
  } catch (e: any) {
    ElMessage.error('重试失败: ' + (e?.message || ''))
  }
}

onMounted(async () => {
  loadEvents()
  try {
    const r = await getRepoDetail(repoKey)
    repoName.value = r?.name || ''
  } catch {}
  try { currentVersion.value = (await getCurrentVersion(repoKey)) || '' } catch {}
})
</script>

<style scoped>
.webhook-page-wrapper {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.version-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 6px;
  background: var(--accent-bg);
  color: #6366F1;
  font-size: 12px;
  font-weight: 500;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}

.webhook-layout {
  display: flex;
  gap: 20px;
}

.webhook-content {
  flex: 1;
  min-height: calc(100vh - 180px);
}

.webhook-events-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.event-type { display: flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 500; color: var(--text-color-primary); }

.platform-tag { display: flex; align-items: center; gap: 4px; font-size: 11px; }
.platform-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }

.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}
</style>
