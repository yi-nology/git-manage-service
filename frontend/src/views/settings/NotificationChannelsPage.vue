<template>
  <div class="notification-page">
    <PageHeader title="通知渠道管理" subtitle="配置同步任务的消息通知渠道">
      <template #actions>
        <ActionPill variant="primary" :icon="Plus" @click="router.push('/settings/notification-channels/add')">
          添加渠道
        </ActionPill>
        <ActionPill variant="outline" :icon="Refresh" :disabled="loading" @click="loadChannels">
          刷新
        </ActionPill>
      </template>
    </PageHeader>

    <DataTable :columns="columns" :data="channels" row-key="id" :loading="loading">
      <template #empty>
        <EmptyState title="暂无通知渠道" description="点击上方按钮添加第一个通知渠道" />
      </template>
      <template #cell-name="{ row }">
        <span class="name-cell">{{ row.name }}</span>
      </template>
      <template #cell-type="{ row }">
        <span class="type-tag" :class="'type-' + row.type">{{ typeLabels[row.type] || row.type }}</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :variant="row.enabled ? 'success' : 'default'" :text="row.enabled ? '启用' : '禁用'" :showDot="false" />
      </template>
      <template #cell-trigger_events="{ row }">
        <div class="trigger-cell">
          <template v-if="parseTriggerEvents(row.trigger_events).length > 0">
            <span
              v-for="event in parseTriggerEvents(row.trigger_events)"
              :key="event"
              class="event-tag"
              :class="triggerEventClass(event)"
            >{{ triggerEventLabels[event] || event }}</span>
          </template>
          <template v-else>
            <span v-if="row.notify_on_success" class="event-tag tag-success">成功</span>
            <span v-if="row.notify_on_failure" class="event-tag tag-danger">失败</span>
          </template>
        </div>
      </template>
      <template #cell-updated_at="{ row }">
        <span class="time-cell">{{ formatDate(row.updated_at) }}</span>
      </template>
      <template #cell-actions="{ row }">
        <div class="action-btns">
          <ActionPill variant="green" small :icon="Promotion" :disabled="testingId === row.id" @click="handleTest(row.id)" />
          <ActionPill variant="primary" small :icon="Edit" @click="router.push(`/settings/notification-channels/${row.id}/edit`)" />
          <el-popconfirm title="确定删除此渠道?" @confirm="handleDelete(row.id)">
            <template #reference>
              <ActionPill variant="danger" small :icon="Delete" />
            </template>
          </el-popconfirm>
        </div>
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Refresh, Promotion, Edit, Delete } from '@element-plus/icons-vue'
import { listChannels, deleteChannel, testChannel } from '@/api/modules/notification'
import type { NotificationChannel } from '@/api/modules/notification'
import PageHeader from '@/components/common/PageHeader.vue'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'

const router = useRouter()

const loading = ref(false)
const channels = ref<NotificationChannel[]>([])
const testingId = ref<number | null>(null)

const columns: TableColumn[] = [
  { key: 'name', label: '名称', width: '160px' },
  { key: 'type', label: '类型', width: '120px' },
  { key: 'status', label: '状态', width: '100px' },
  { key: 'trigger_events', label: '触发时机', flex: 1 },
  { key: 'updated_at', label: '更新时间', width: '160px' },
  { key: 'actions', label: '操作', width: '140px' },
]

const typeLabels: Record<string, string> = {
  email: '邮件',
  dingtalk: '钉钉',
  wechat: '企业微信',
  lanxin: '蓝信',
  feishu: '飞书',
  webhook: 'Webhook'
}

const triggerEventLabels: Record<string, string> = {
  sync_success: '同步成功',
  sync_failure: '同步失败',
  sync_conflict: '同步冲突',
  webhook_received: 'Webhook 接收',
  webhook_error: 'Webhook 错误',
  cron_triggered: '定时任务触发',
  backup_success: '备份成功',
  backup_failure: '备份失败'
}

function triggerEventClass(event: string): string {
  if (event.includes('success')) return 'tag-success'
  if (event.includes('failure') || event.includes('error') || event.includes('conflict')) return 'tag-danger'
  if (event.includes('received') || event.includes('triggered')) return 'tag-warning'
  return 'tag-info'
}

function parseTriggerEvents(raw: string): string[] {
  if (!raw) return []
  try {
    const arr = JSON.parse(raw)
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

onMounted(() => {
  loadChannels()
})

async function loadChannels() {
  loading.value = true
  try {
    channels.value = await listChannels() || []
  } catch {
    ElMessage.error('加载通知渠道失败')
    channels.value = []
  } finally {
    loading.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await deleteChannel(id)
    ElMessage.success('渠道已删除')
    await loadChannels()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('删除失败: ' + (err.message || '未知错误'))
  }
}

async function handleTest(id: number) {
  testingId.value = id
  try {
    const result = await testChannel(id, '这是一条测试消息 - Git管理服务')
    if (result.success) {
      ElMessage.success('测试消息发送成功')
    } else {
      ElMessage.error('测试失败: ' + (result.error || '未知错误'))
    }
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('测试失败: ' + (err.message || '未知错误'))
  } finally {
    testingId.value = null
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}
</script>

<style scoped>
.notification-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.name-cell {
  font-weight: 500;
  color: var(--text-color-primary);
}

.time-cell {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
}

.type-tag.type-email { background: var(--accent-bg); color: #6366F1; }
.type-tag.type-dingtalk { background: #FFFBEB; color: #F59E0B; }
.type-tag.type-wechat { background: #ECFDF5; color: #10B981; }
.type-tag.type-lanxin { background: #FEF2F2; color: #EF4444; }
.type-tag.type-feishu { background: var(--accent-bg); color: #6366F1; }
.type-tag.type-webhook { background: var(--accent-bg); color: var(--text-color-secondary); }

.trigger-cell {
  white-space: normal;
}

.event-tag {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  margin-right: 4px;
  margin-bottom: 2px;
}

.event-tag.tag-success { background: #ECFDF5; color: #10B981; }
.event-tag.tag-danger { background: #FEF2F2; color: #EF4444; }
.event-tag.tag-warning { background: #FFFBEB; color: #F59E0B; }
.event-tag.tag-info { background: var(--accent-bg); color: #6366F1; }

.action-btns {
  display: flex;
  gap: 4px;
}
</style>
