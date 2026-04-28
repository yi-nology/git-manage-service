<template>
  <div class="webhook-page-wrapper">
    <div class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="$router.push(`/local-repos/${repoKey}`)">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h2>{{ repoName || '仓库' }}</h2>
        <span v-if="currentVersion" class="version-tag">{{ currentVersion }}</span>
      </div>
      <div class="header-actions">
        <button class="action-pill action-pill--outline" @click="loadEvents">
          <el-icon><Refresh /></el-icon> 刷新
        </button>
      </div>
    </div>

    <div class="webhook-layout">
      <RepoSidebar :repo-key="repoKey" active-key="webhooks" />
      <div class="webhook-content">
        <div class="webhook-events-page">
          <div class="stats-row">
          <div class="stat-card">
            <div class="stat-label-row">
              <el-icon :size="20" style="color:#6366F1"><TrendCharts /></el-icon>
              <span class="stat-label">总事件数</span>
            </div>
            <span class="stat-value">{{ total }}</span>
          </div>
          <div class="stat-card">
            <div class="stat-label-row">
              <el-icon :size="20" style="color:#10B981"><CircleCheck /></el-icon>
              <span class="stat-label">已处理</span>
            </div>
            <span class="stat-value" style="color:#10B981">{{ processedCount }}</span>
          </div>
          <div class="stat-card">
            <div class="stat-label-row">
              <el-icon :size="20" style="color:#F59E0B"><Warning /></el-icon>
              <span class="stat-label">待处理</span>
            </div>
            <span class="stat-value" style="color:#F59E0B">{{ pendingCount }}</span>
          </div>
          <div class="stat-card">
            <div class="stat-label-row">
              <el-icon :size="20" style="color:#EF4444"><CircleClose /></el-icon>
              <span class="stat-label">失败</span>
            </div>
            <span class="stat-value" style="color:#EF4444">{{ failedCount }}</span>
          </div>
        </div>

        <div class="table-card" v-if="events.length > 0">
          <div class="table-header">
            <span class="th" style="width:140px">事件类型</span>
            <span class="th" style="width:80px">来源</span>
            <span class="th" style="flex:1">事件 ID</span>
            <span class="th" style="width:100px">触发者</span>
            <span class="th" style="width:80px">状态</span>
            <span class="th" style="width:160px">时间</span>
            <span class="th" style="width:70px">操作</span>
          </div>
          <div v-for="row in events" :key="row.id" class="table-row">
            <span class="td" style="width:140px">
              <span class="event-type">
                <el-icon :size="14" style="color:#6366F1"><Share /></el-icon>
                {{ row.event_type }}
              </span>
            </span>
            <span class="td" style="width:80px">
              <span class="platform-tag">
                <span class="platform-dot" :style="{ background: platformColor(row.source) }"></span>
                {{ row.source }}
              </span>
            </span>
            <span class="td td-mono" style="flex:1" :title="row.event_id">{{ row.event_id.substring(0, 16) }}...</span>
            <span class="td" style="width:100px">{{ row.actor_name || '-' }}</span>
            <span class="td" style="width:80px">
              <span class="status-pill" :class="'status-' + row.status">{{ statusLabel(row.status) }}</span>
            </span>
            <span class="td td-time" style="width:160px">{{ formatTime(row.created_at) }}</span>
            <span class="td" style="width:70px">
              <button v-if="row.status === 'failed'" class="retry-btn" @click="handleRetry(row)">重试</button>
            </span>
          </div>
        </div>

        <div v-else-if="!loading" class="empty-state">
          <el-icon class="empty-icon"><Link /></el-icon>
          <h3>暂无事件</h3>
          <p>Webhook 事件将在此处显示</p>
        </div>
      </div>
    </div>
  </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Refresh, TrendCharts, CircleCheck, Warning, CircleClose, Share, Link } from '@element-plus/icons-vue'
import { listWebhookEvents, retryWebhookEvent } from '@/api/modules/webhook-event'
import type { WebhookEventDTO } from '@/api/modules/webhook-event'
import RepoSidebar from '@/components/repo/RepoSidebar.vue'
import { getRepoDetail } from '@/api/modules/repo'
import { getCurrentVersion } from '@/api/modules/version'

const route = useRoute()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const events = ref<WebhookEventDTO[]>([])
const total = ref(0)
const repoName = ref('')
const currentVersion = ref('')

const processedCount = computed(() => events.value.filter(e => e.status === 'processed').length)
const pendingCount = computed(() => events.value.filter(e => e.status === 'received').length)
const failedCount = computed(() => events.value.filter(e => e.status === 'failed').length)

const PLATFORM_COLORS: Record<string, string> = { gitlab: '#FC6D26', github: '#24292F', gitea: '#609926' }
function platformColor(s: string) { return PLATFORM_COLORS[s?.toLowerCase()] || '#6B7280' }
function statusLabel(s: string) {
  if (s === 'processed') return '已处理'
  if (s === 'received') return '待处理'
  if (s === 'failed') return '失败'
  return s
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

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  background: var(--bg-color-page, #fff);
  color: var(--text-color-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.back-btn:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.version-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 6px;
  background: #EEF2FF;
  color: #6366F1;
  font-size: 12px;
  font-weight: 500;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.action-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  color: var(--text-color-secondary);
}

.action-pill:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
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

.stats-row { display: flex; gap: 16px; }

.stat-card {
  flex: 1; border-radius: 12px; background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb); padding: 20px;
  display: flex; flex-direction: column; gap: 8px;
}
.stat-label-row { display: flex; align-items: center; gap: 8px; }
.stat-label { font-size: 12px; color: var(--text-color-secondary); }
.stat-value { font-size: 28px; font-weight: 700; }

.table-card {
  border-radius: 12px; border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff); overflow: hidden;
}

.table-header { display: flex; align-items: center; padding: 12px 20px; background: #EEF2FF; }
.th { font-size: 12px; font-weight: 600; color: var(--text-color-secondary); }

.table-row {
  display: flex; align-items: center; padding: 12px 20px;
  border-bottom: 1px solid var(--border-color, #e5e7eb); transition: background 0.15s;
}
.table-row:last-child { border-bottom: none; }
.table-row:hover { background: #F8FAFC; }

.td { font-size: 12px; color: var(--text-color-secondary); }
.td-mono { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; font-size: 12px; }
.td-time { color: var(--text-color-secondary, #94A3B8); font-size: 12px; }

.event-type { display: flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 500; color: var(--text-color-primary); }

.platform-tag { display: flex; align-items: center; gap: 4px; font-size: 11px; }
.platform-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }

.status-pill {
  display: inline-block; padding: 4px 8px; border-radius: 9999px;
  font-size: 11px; font-weight: 500; text-align: center; width: 56px;
}
.status-processed { background: #ECFDF5; color: #059669; }
.status-received { background: #FFFBEB; color: #D97706; }
.status-failed { background: #FEF2F2; color: #DC2626; }

.retry-btn {
  padding: 2px 8px; border-radius: 4px; border: 1px solid #F59E0B;
  background: transparent; font-size: 11px; color: #F59E0B; cursor: pointer; transition: all 0.2s;
}
.retry-btn:hover { background: #FFFBEB; }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 60px 0; gap: 12px; }
.empty-icon { font-size: 48px; color: var(--text-color-placeholder); }
.empty-state h3 { margin: 0; font-size: 16px; color: var(--text-color-primary); }
.empty-state p { margin: 0; font-size: 13px; color: var(--text-color-secondary); }
</style>
