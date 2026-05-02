<template>
  <div class="notification-manager">
    <div class="toolbar">
      <el-button type="primary" @click="openAddDialog">
        <el-icon><Plus /></el-icon> 添加渠道
      </el-button>
      <el-button @click="loadChannels" :loading="loading">
        <el-icon><Refresh /></el-icon> 刷新
      </el-button>
    </div>

    <el-table :data="channels" v-loading="loading" empty-text="暂无通知渠道">
      <el-table-column prop="name" label="名称" min-width="150" />
      <el-table-column prop="type" label="类型" width="120">
        <template #default="{ row }">
          <el-tag size="small">{{ typeLabels[row.type] || row.type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
            {{ row.enabled ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="触发时机" min-width="280">
        <template #default="{ row }">
          <template v-if="parseTriggerEvents(row.trigger_events).length > 0">
            <el-tag v-for="event in parseTriggerEvents(row.trigger_events)" :key="event" :type="triggerEventTagType(event)" size="small" class="mr-1 mb-1">
              {{ triggerEventLabels[event] || event }}
            </el-tag>
          </template>
          <template v-else>
            <el-tag v-if="row.notify_on_success" type="success" size="small" class="mr-1">成功</el-tag>
            <el-tag v-if="row.notify_on_failure" type="danger" size="small">失败</el-tag>
          </template>
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="160" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button-group size="small">
            <el-button @click="handleTest(row.id)" :loading="testingId === row.id" title="测试">
              <el-icon><Promotion /></el-icon>
            </el-button>
            <el-button @click="openEditDialog(row)" title="编辑">
              <el-icon><Edit /></el-icon>
            </el-button>
            <el-popconfirm title="确定删除此渠道?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button type="danger" title="删除">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </template>
            </el-popconfirm>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑通知渠道' : '添加通知渠道'" width="640px" destroy-on-close>
      <el-form :model="form" label-width="100px">
        <el-form-item label="渠道名称" required>
          <el-input v-model="form.name" placeholder="例如：开发团队钉钉群" />
        </el-form-item>
        <el-form-item label="渠道类型" required>
          <el-select v-model="form.type" style="width: 100%" :disabled="!!editingId">
            <el-option label="邮件 (Email)" value="email" />
            <el-option label="钉钉机器人" value="dingtalk" />
            <el-option label="企业微信机器人" value="wechat" />
            <el-option label="蓝信机器人" value="lanxin" />
            <el-option label="飞书机器人" value="feishu" />
            <el-option label="自定义 Webhook" value="webhook" />
          </el-select>
        </el-form-item>

        <ChannelConfigForms :channel-type="form.type" :config-form="configForm" />

        <el-divider content-position="left">通知选项</el-divider>
        <el-form-item label="启用渠道">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="触发时机">
          <div class="trigger-events-grid">
            <div class="trigger-group">
              <div class="trigger-group-title">同步事件</div>
              <el-checkbox v-model="triggerEventsMap.sync_success">同步成功</el-checkbox>
              <el-checkbox v-model="triggerEventsMap.sync_failure">同步失败</el-checkbox>
              <el-checkbox v-model="triggerEventsMap.sync_conflict">同步冲突</el-checkbox>
            </div>
            <div class="trigger-group">
              <div class="trigger-group-title">Webhook 事件</div>
              <el-checkbox v-model="triggerEventsMap.webhook_received">Webhook 接收</el-checkbox>
              <el-checkbox v-model="triggerEventsMap.webhook_error">Webhook 处理错误</el-checkbox>
            </div>
            <div class="trigger-group">
              <div class="trigger-group-title">定时任务</div>
              <el-checkbox v-model="triggerEventsMap.cron_triggered">定时任务触发</el-checkbox>
            </div>
            <div class="trigger-group">
              <div class="trigger-group-title">备份事件</div>
              <el-checkbox v-model="triggerEventsMap.backup_success">备份成功</el-checkbox>
              <el-checkbox v-model="triggerEventsMap.backup_failure">备份失败</el-checkbox>
            </div>
          </div>
        </el-form-item>

        <el-divider content-position="left">消息模板（按时机独立配置）</el-divider>

        <EventTemplateEditor
          ref="templateEditorRef"
          :enabled-events="enabledEvents"
          :trigger-event-labels="triggerEventLabels"
          :initial-templates-json="initialTemplatesJson"
        />
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Refresh, Promotion, Edit, Delete } from '@element-plus/icons-vue'
import { listChannels, createChannel, updateChannel, deleteChannel, testChannel } from '@/api/modules/notification'
import type { NotificationChannel } from '@/api/modules/notification'
import ChannelConfigForms from './ChannelConfigForms.vue'
import EventTemplateEditor from './EventTemplateEditor.vue'

const loading = ref(false)
const channels = ref<NotificationChannel[]>([])
const testingId = ref<number | null>(null)

const showDialog = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const initialTemplatesJson = ref('')

const form = reactive({
  name: '',
  type: '' as string,
  enabled: true,
  notify_on_success: false,
  notify_on_failure: true
})

const triggerEventsMap = reactive<Record<string, boolean>>({
  sync_success: false,
  sync_failure: true,
  sync_conflict: true,
  webhook_received: false,
  webhook_error: true,
  cron_triggered: false,
  backup_success: false,
  backup_failure: true
})

const allTriggerEventKeys = [
  'sync_success', 'sync_failure', 'sync_conflict',
  'webhook_received', 'webhook_error',
  'cron_triggered',
  'backup_success', 'backup_failure'
]

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

function triggerEventTagType(event: string): string {
  if (event.includes('success')) return 'success'
  if (event.includes('failure') || event.includes('error') || event.includes('conflict')) return 'danger'
  if (event.includes('received') || event.includes('triggered')) return 'warning'
  return 'info'
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

function getTriggerEventsJson(): string {
  const events = allTriggerEventKeys.filter(k => triggerEventsMap[k])
  return JSON.stringify(events)
}

function resetTriggerEvents() {
  allTriggerEventKeys.forEach(k => {
    triggerEventsMap[k] = false
  })
  triggerEventsMap.sync_failure = true
  triggerEventsMap.sync_conflict = true
  triggerEventsMap.webhook_error = true
  triggerEventsMap.backup_failure = true
}

function loadTriggerEvents(raw: string) {
  allTriggerEventKeys.forEach(k => {
    triggerEventsMap[k] = false
  })
  const events = parseTriggerEvents(raw)
  if (events.length > 0) {
    events.forEach(e => {
      if (e in triggerEventsMap) {
        triggerEventsMap[e] = true
      }
    })
  } else {
    triggerEventsMap.sync_failure = true
    triggerEventsMap.backup_failure = true
  }
}

const enabledEvents = computed(() => allTriggerEventKeys.filter(k => triggerEventsMap[k]))

const configForm = reactive({
  smtp_host: '',
  smtp_port: '',
  username: '',
  password: '',
  from: '',
  to: '',
  webhook_url: '',
  secret: '',
  sign: '',
  security_type: 'none',
  keywords: '',
  url: '',
  method: 'POST',
  content_type: 'application/json'
})

const typeLabels: Record<string, string> = {
  email: '邮件',
  dingtalk: '钉钉',
  wechat: '企业微信',
  lanxin: '蓝信',
  feishu: '飞书',
  webhook: 'Webhook'
}

const templateEditorRef = ref<InstanceType<typeof EventTemplateEditor> | null>(null)

onMounted(() => {
  loadChannels()
})

watch(() => form.type, () => {
  if (!editingId.value) {
    Object.keys(configForm).forEach(key => {
      (configForm as Record<string, string>)[key] = ''
    })
    configForm.method = 'POST'
    configForm.content_type = 'application/json'
    configForm.security_type = 'none'
  }
})

async function loadChannels() {
  loading.value = true
  try {
    channels.value = await listChannels()
  } catch {
    ElMessage.error('加载通知渠道失败')
  } finally {
    loading.value = false
  }
}

function openAddDialog() {
  editingId.value = null
  form.name = ''
  form.type = ''
  form.enabled = true
  form.notify_on_success = false
  form.notify_on_failure = true
  resetTriggerEvents()
  initialTemplatesJson.value = ''
  Object.keys(configForm).forEach(key => {
    (configForm as Record<string, string>)[key] = ''
  })
  configForm.method = 'POST'
  configForm.content_type = 'application/json'
  configForm.security_type = 'none'
  showDialog.value = true
}

async function openEditDialog(channel: NotificationChannel) {
  editingId.value = channel.id
  form.name = channel.name
  form.type = channel.type
  form.enabled = channel.enabled
  form.notify_on_success = channel.notify_on_success
  form.notify_on_failure = channel.notify_on_failure
  loadTriggerEvents(channel.trigger_events)
  initialTemplatesJson.value = channel.event_templates_json || ''

  try {
    const config = JSON.parse(channel.config || '{}')
    Object.keys(config).forEach(key => {
      if (key in configForm) {
        (configForm as Record<string, string>)[key] = config[key]
      }
    })
    if (!config.security_type) {
      if ((form.type === 'dingtalk' || form.type === 'feishu') && config.secret) {
        configForm.security_type = 'sign'
      } else if (form.type === 'lanxin' && config.sign) {
        configForm.security_type = 'sign'
      } else {
        configForm.security_type = 'none'
      }
    }
  } catch { /* ignore */ }

  showDialog.value = true
}

function getConfigJson(): string {
  const configKeys: Record<string, string[]> = {
    email: ['smtp_host', 'smtp_port', 'username', 'password', 'from', 'to'],
    dingtalk: ['webhook_url', 'security_type', 'secret', 'keywords'],
    wechat: ['webhook_url'],
    lanxin: ['webhook_url', 'security_type', 'sign', 'keywords'],
    feishu: ['webhook_url', 'security_type', 'secret', 'keywords'],
    webhook: ['url', 'method', 'content_type']
  }

  const keys = configKeys[form.type] || []
  const config: Record<string, string> = {}
  keys.forEach(key => {
    const value = (configForm as Record<string, string>)[key]
    if (value) config[key] = value
  })
  return JSON.stringify(config)
}

async function handleSave() {
  if (!form.name || !form.type) {
    ElMessage.warning('请填写名称和类型')
    return
  }
  saving.value = true
  try {
    const eventTemplatesJson = templateEditorRef.value?.buildJson() || ''
    const params = {
      name: form.name,
      type: form.type,
      config: getConfigJson(),
      enabled: form.enabled,
      notify_on_success: form.notify_on_success,
      notify_on_failure: form.notify_on_failure,
      trigger_events: getTriggerEventsJson(),
      title_template: '',
      content_template: '',
      event_templates_json: eventTemplatesJson
    }
    if (editingId.value) {
      await updateChannel({ ...params, id: editingId.value })
      ElMessage.success('渠道更新成功')
    } else {
      await createChannel(params)
      ElMessage.success('渠道创建成功')
    }
    showDialog.value = false
    await loadChannels()
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('保存失败: ' + (err.message || '未知错误'))
  } finally {
    saving.value = false
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
</script>

<style scoped>
.notification-manager {
  padding: 8px 0;
}
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.mr-1 {
  margin-right: 4px;
}
.mb-1 {
  margin-bottom: 4px;
}
.trigger-events-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  width: 100%;
}
.trigger-group {
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
}
.trigger-group-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
  font-weight: 500;
}
.trigger-group .el-checkbox {
  display: block;
  margin-left: 0;
  margin-bottom: 2px;
}
</style>
