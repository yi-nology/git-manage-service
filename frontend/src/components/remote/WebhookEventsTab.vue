<template>
  <div>
    <div class="content-header">
      <SectionTitle title="Webhook 事件" />
      <ActionPill variant="outline" :icon="Refresh" @click="loadWebhookEvents">刷新</ActionPill>
    </div>

    <DataTable :columns="whColumns" :data="webhookEvents" :loading="whLoading" row-key="id">
      <template #cell-event_type="{ row }">
        <span class="event-type-cell"><el-icon :size="14" style="color:#6366F1"><Share /></el-icon> {{ row.event_type }}</span>
      </template>
      <template #cell-event_id="{ row }">
        <span class="mono">{{ row.event_id?.substring(0, 16) }}...</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :variant="whStatusVariant(row.status)" :text="whStatusLabel(row.status)" :showDot="false" />
      </template>
      <template #cell-created_at="{ row }">
        {{ formatTime(row.created_at) }}
      </template>
      <template #empty>
        <EmptyState title="暂无 Webhook 事件" />
      </template>
      <template #row-actions="{ row }">
        <ActionPill v-if="row.status === 'failed'" variant="amber" small @click="handleRetryEvent(row)">重试</ActionPill>
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Share } from '@element-plus/icons-vue'
import { listWebhookEvents, retryWebhookEvent } from '@/api/modules/webhook-event'
import type { WebhookEventDTO } from '@/api/modules/webhook-event'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

defineProps<{
  active: boolean
}>()

const whLoading = ref(false)
const webhookEvents = ref<WebhookEventDTO[]>([])
const loaded = ref(false)

const whColumns: TableColumn[] = [
  { key: 'event_type', label: '事件类型', width: '120px' },
  { key: 'source', label: '来源', width: '80px' },
  { key: 'event_id', label: '事件 ID' },
  { key: 'status', label: '状态', width: '80px' },
  { key: 'created_at', label: '时间', width: '140px' },
]

function whStatusLabel(s: string) {
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
  return `${d.getMonth() + 1}-${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

async function loadWebhookEvents() {
  whLoading.value = true
  try {
    const res = await listWebhookEvents({ page: 1, page_size: 50 })
    webhookEvents.value = res?.items || []
  } catch { webhookEvents.value = [] }
  finally { whLoading.value = false }
}

async function handleRetryEvent(ev: WebhookEventDTO) {
  try {
    await retryWebhookEvent(ev.id)
    ElMessage.success('已重试')
    loadWebhookEvents()
  } catch (e: any) { ElMessage.error('重试失败: ' + (e?.message || '')) }
}

watch(() => props.active, (val) => {
  if (val && !loaded.value) {
    loadWebhookEvents()
    loaded.value = true
  }
}, { immediate: true })
</script>

<style scoped>
.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.event-type-cell {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}
</style>
