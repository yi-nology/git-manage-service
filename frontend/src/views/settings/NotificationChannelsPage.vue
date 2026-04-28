<template>
  <div class="notification-page">
    <div class="title-row">
      <div class="title-left">
        <h2 class="page-title">通知渠道管理</h2>
        <p class="page-subtitle">配置同步任务的消息通知渠道</p>
      </div>
      <div class="title-actions">
        <button class="add-btn" @click="router.push('/settings/notification-channels/add')">
          <el-icon><Plus /></el-icon>
          添加渠道
        </button>
        <button class="refresh-btn" @click="loadChannels" :disabled="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading-card">
      <div class="loading-spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="channels.length === 0" class="empty-card">
      <div class="empty-icon">
        <el-icon :size="32"><Bell /></el-icon>
      </div>
      <div class="empty-text">暂无通知渠道</div>
      <div class="empty-sub">点击上方按钮添加第一个通知渠道</div>
    </div>

    <div v-else class="table-card">
      <div class="table-header">
        <span class="th" style="width:160px">名称</span>
        <span class="th" style="width:120px">类型</span>
        <span class="th" style="width:100px">状态</span>
        <span class="th" style="flex:1">触发时机</span>
        <span class="th" style="width:160px">更新时间</span>
        <span class="th" style="width:140px">操作</span>
      </div>
      <div v-for="channel in channels" :key="channel.id" class="table-row">
        <span class="td name-cell" style="width:160px">{{ channel.name }}</span>
        <span class="td" style="width:120px">
          <span class="type-tag" :class="'type-' + channel.type">{{ typeLabels[channel.type] || channel.type }}</span>
        </span>
        <span class="td" style="width:100px">
          <span class="status-badge" :class="channel.enabled ? 'status-active' : 'status-inactive'">
            {{ channel.enabled ? '启用' : '禁用' }}
          </span>
        </span>
        <span class="td" style="flex:1">
          <template v-if="parseTriggerEvents(channel.trigger_events).length > 0">
            <span
              v-for="event in parseTriggerEvents(channel.trigger_events)"
              :key="event"
              class="event-tag"
              :class="triggerEventClass(event)"
            >{{ triggerEventLabels[event] || event }}</span>
          </template>
          <template v-else>
            <span v-if="channel.notify_on_success" class="event-tag tag-success">成功</span>
            <span v-if="channel.notify_on_failure" class="event-tag tag-danger">失败</span>
          </template>
        </span>
        <span class="td time-cell" style="width:160px">{{ formatDate(channel.updated_at) }}</span>
        <span class="td" style="width:140px">
          <div class="action-btns">
            <button class="action-btn btn-test" @click="handleTest(channel.id)" :disabled="testingId === channel.id" title="测试">
              <el-icon><Promotion /></el-icon>
            </button>
            <button class="action-btn btn-edit" @click="router.push(`/settings/notification-channels/${channel.id}/edit`)" title="编辑">
              <el-icon><Edit /></el-icon>
            </button>
            <el-popconfirm title="确定删除此渠道?" @confirm="handleDelete(channel.id)">
              <template #reference>
                <button class="action-btn btn-delete" title="删除">
                  <el-icon><Delete /></el-icon>
                </button>
              </template>
            </el-popconfirm>
          </div>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Refresh, Promotion, Edit, Delete, Bell } from '@element-plus/icons-vue'
import { listChannels, deleteChannel, testChannel } from '@/api/modules/notification'
import type { NotificationChannel } from '@/api/modules/notification'

const router = useRouter()

const loading = ref(false)
const channels = ref<NotificationChannel[]>([])
const testingId = ref<number | null>(null)

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

.title-actions {
  display: flex;
  gap: 8px;
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 8px;
  border: none;
  background: var(--accent-primary, #6366F1);
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.add-btn:hover {
  opacity: 0.9;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  color: var(--text-color-regular);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.empty-sub {
  font-size: 13px;
  color: var(--text-color-secondary);
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

.type-tag.type-email { background: #EEF2FF; color: #6366F1; }
.type-tag.type-dingtalk { background: #FFFBEB; color: #F59E0B; }
.type-tag.type-wechat { background: #ECFDF5; color: #10B981; }
.type-tag.type-lanxin { background: #FEF2F2; color: #EF4444; }
.type-tag.type-feishu { background: #EEF2FF; color: #6366F1; }
.type-tag.type-webhook { background: var(--accent-bg, #EEF2FF); color: var(--text-color-secondary); }

.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
}

.status-badge.status-active { background: #ECFDF5; color: #10B981; }
.status-badge.status-inactive { background: var(--accent-bg, #EEF2FF); color: var(--text-color-secondary); }

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
.event-tag.tag-info { background: #EEF2FF; color: #6366F1; }

.action-btns {
  display: flex;
  gap: 4px;
}

.action-btn {
  width: 28px;
  height: 28px;
  border-radius: 4px;
  border: 1px solid transparent;
  cursor: pointer;
  background: none;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.2s;
}

.btn-test { color: #10B981; border-color: #10B981; }
.btn-test:hover:not(:disabled) { background: #ECFDF5; }
.btn-test:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-edit { color: #6366F1; border-color: #6366F1; }
.btn-edit:hover { background: #EEF2FF; }

.btn-delete { color: #EF4444; border-color: #EF4444; }
.btn-delete:hover { background: #FEF2F2; }
</style>
