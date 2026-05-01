<template>
  <div class="add-channel-page">
    <PageHeader :title="isEdit ? '编辑通知渠道' : '添加通知渠道'" showBack backRoute="/settings/notification-channels" />

    <FormCard v-if="!pageLoading">
      <div class="form-field">
        <label class="field-label">渠道名称</label>
        <input v-model="form.name" placeholder="例如：开发团队钉钉群" class="field-input" />
      </div>

      <div class="form-field">
        <label class="field-label">渠道类型</label>
        <div class="type-selector">
          <button
            v-for="t in channelTypes"
            :key="t.value"
            class="type-btn"
            :class="{ active: form.type === t.value, disabled: !!editingId }"
            @click="selectType(t.value)"
          >
            <el-icon><component :is="t.icon" /></el-icon>
            {{ t.label }}
          </button>
        </div>
      </div>

      <div class="divider" v-if="form.type"></div>

      <template v-if="form.type === 'email'">
        <div class="config-header">
          <span class="config-title">邮件配置</span>
          <span class="type-badge badge-email">邮件</span>
        </div>
        <div class="form-row">
          <div class="form-field" style="flex:2">
            <label class="field-label">SMTP 服务器</label>
            <input v-model="configForm.smtp_host" placeholder="smtp.example.com" class="field-input" />
          </div>
          <div class="form-field" style="flex:1">
            <label class="field-label">端口</label>
            <input v-model="configForm.smtp_port" placeholder="587" class="field-input" />
          </div>
        </div>
        <div class="form-field">
          <label class="field-label">用户名</label>
          <input v-model="configForm.username" placeholder="发件邮箱账号" class="field-input" />
        </div>
        <div class="form-field">
          <label class="field-label">密码</label>
          <input v-model="configForm.password" type="password" placeholder="邮箱密码或授权码" class="field-input" />
        </div>
        <div class="form-field">
          <label class="field-label">发件人</label>
          <input v-model="configForm.from" placeholder="Git管理服务 &lt;noreply@example.com&gt;" class="field-input" />
        </div>
        <div class="form-field">
          <label class="field-label">收件人</label>
          <input v-model="configForm.to" placeholder="多个邮箱用逗号分隔" class="field-input" />
        </div>
      </template>

      <template v-if="form.type === 'dingtalk'">
        <div class="config-header">
          <span class="config-title">钉钉配置</span>
          <span class="type-badge badge-dingtalk">钉钉机器人</span>
        </div>
        <div class="form-field">
          <label class="field-label">Webhook URL</label>
          <input v-model="configForm.webhook_url" placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx" class="field-input" />
        </div>
        <div class="form-field">
          <label class="field-label">安全模式</label>
          <div class="mode-selector">
            <button class="mode-btn" :class="{ active: configForm.security_type === 'none' }" @click="configForm.security_type = 'none'">无</button>
            <button class="mode-btn" :class="{ active: configForm.security_type === 'sign' }" @click="configForm.security_type = 'sign'">签名</button>
            <button class="mode-btn" :class="{ active: configForm.security_type === 'keyword' }" @click="configForm.security_type = 'keyword'">关键字</button>
          </div>
        </div>
        <div class="form-field" v-if="configForm.security_type === 'sign'">
          <label class="field-label">签名密钥</label>
          <input v-model="configForm.secret" placeholder="SEC开头的密钥" class="field-input" />
        </div>
        <div class="form-field" v-if="configForm.security_type === 'keyword'">
          <label class="field-label">关键字</label>
          <input v-model="configForm.keywords" placeholder="消息中需要包含的关键字" class="field-input" />
        </div>
      </template>

      <template v-if="form.type === 'wechat'">
        <div class="config-header">
          <span class="config-title">企业微信配置</span>
          <span class="type-badge badge-wechat">企业微信</span>
        </div>
        <div class="form-field">
          <label class="field-label">Webhook URL</label>
          <input v-model="configForm.webhook_url" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx" class="field-input" />
        </div>
      </template>

      <template v-if="form.type === 'feishu'">
        <div class="config-header">
          <span class="config-title">飞书配置</span>
          <span class="type-badge badge-feishu">飞书机器人</span>
        </div>
        <div class="form-field">
          <label class="field-label">Webhook URL</label>
          <input v-model="configForm.webhook_url" placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx" class="field-input" />
        </div>
        <div class="form-field">
          <label class="field-label">安全模式</label>
          <div class="mode-selector">
            <button class="mode-btn" :class="{ active: configForm.security_type === 'none' }" @click="configForm.security_type = 'none'">无</button>
            <button class="mode-btn" :class="{ active: configForm.security_type === 'sign' }" @click="configForm.security_type = 'sign'">签名</button>
            <button class="mode-btn" :class="{ active: configForm.security_type === 'keyword' }" @click="configForm.security_type = 'keyword'">关键字</button>
          </div>
        </div>
        <div class="form-field" v-if="configForm.security_type === 'sign'">
          <label class="field-label">签名密钥</label>
          <input v-model="configForm.secret" placeholder="飞书签名密钥" class="field-input" />
        </div>
        <div class="form-field" v-if="configForm.security_type === 'keyword'">
          <label class="field-label">关键字</label>
          <input v-model="configForm.keywords" placeholder="消息中需要包含的关键字" class="field-input" />
        </div>
      </template>

      <template v-if="form.type === 'lanxin'">
        <div class="config-header">
          <span class="config-title">蓝信配置</span>
          <span class="type-badge badge-lanxin">蓝信机器人</span>
        </div>
        <div class="form-field">
          <label class="field-label">Webhook URL</label>
          <input v-model="configForm.webhook_url" placeholder="蓝信机器人 Webhook 地址" class="field-input" />
        </div>
        <div class="form-field">
          <label class="field-label">安全模式</label>
          <div class="mode-selector">
            <button class="mode-btn" :class="{ active: configForm.security_type === 'none' }" @click="configForm.security_type = 'none'">无</button>
            <button class="mode-btn" :class="{ active: configForm.security_type === 'sign' }" @click="configForm.security_type = 'sign'">签名</button>
            <button class="mode-btn" :class="{ active: configForm.security_type === 'keyword' }" @click="configForm.security_type = 'keyword'">关键字</button>
          </div>
        </div>
        <div class="form-field" v-if="configForm.security_type === 'sign'">
          <label class="field-label">签名密钥</label>
          <input v-model="configForm.sign" placeholder="签名密钥" class="field-input" />
        </div>
        <div class="form-field" v-if="configForm.security_type === 'keyword'">
          <label class="field-label">关键字</label>
          <input v-model="configForm.keywords" placeholder="消息中需要包含的关键字" class="field-input" />
        </div>
      </template>

      <template v-if="form.type === 'webhook'">
        <div class="config-header">
          <span class="config-title">Webhook 配置</span>
          <span class="type-badge badge-webhook">自定义 Webhook</span>
        </div>
        <div class="form-field">
          <label class="field-label">URL</label>
          <input v-model="configForm.url" placeholder="https://your-server.com/webhook" class="field-input" />
        </div>
        <div class="form-row">
          <div class="form-field" style="flex:1">
            <label class="field-label">请求方法</label>
            <select v-model="configForm.method" class="field-input">
              <option value="POST">POST</option>
              <option value="GET">GET</option>
            </select>
          </div>
          <div class="form-field" style="flex:1">
            <label class="field-label">Content-Type</label>
            <select v-model="configForm.content_type" class="field-input">
              <option value="application/json">application/json</option>
              <option value="application/x-www-form-urlencoded">application/x-www-form-urlencoded</option>
            </select>
          </div>
        </div>
      </template>

      <div class="divider" v-if="form.type"></div>

      <div class="config-header" v-if="form.type">
        <span class="config-title">触发时机</span>
        <div class="toggle-row">
          <button class="toggle-btn" :class="{ active: form.enabled }" @click="form.enabled = !form.enabled">
            <span class="toggle-dot" :class="{ right: form.enabled }"></span>
          </button>
          <span class="toggle-label">{{ form.enabled ? '启用' : '禁用' }}</span>
        </div>
      </div>

      <div class="trigger-grid" v-if="form.type">
        <div class="trigger-group">
          <div class="trigger-group-title">同步事件</div>
          <label class="checkbox-row" v-for="evt in syncEvents" :key="evt.key">
            <span class="checkbox-box" :class="{ checked: triggerEventsMap[evt.key], 'check-success': evt.key.includes('success'), 'check-danger': evt.key.includes('failure'), 'check-warning': evt.key.includes('conflict') }">
              <span v-if="triggerEventsMap[evt.key]" class="check-icon">&#10003;</span>
            </span>
            <span class="checkbox-label">{{ evt.label }}</span>
          </label>
        </div>
        <div class="trigger-group">
          <div class="trigger-group-title">Webhook 事件</div>
          <label class="checkbox-row" v-for="evt in webhookEvents" :key="evt.key">
            <span class="checkbox-box" :class="{ checked: triggerEventsMap[evt.key], 'check-danger': evt.key.includes('error') }">
              <span v-if="triggerEventsMap[evt.key]" class="check-icon">&#10003;</span>
            </span>
            <span class="checkbox-label">{{ evt.label }}</span>
          </label>
        </div>
        <div class="trigger-group">
          <div class="trigger-group-title">定时任务</div>
          <label class="checkbox-row" v-for="evt in cronEvents" :key="evt.key">
            <span class="checkbox-box" :class="{ checked: triggerEventsMap[evt.key] }">
              <span v-if="triggerEventsMap[evt.key]" class="check-icon">&#10003;</span>
            </span>
            <span class="checkbox-label">{{ evt.label }}</span>
          </label>
        </div>
        <div class="trigger-group">
          <div class="trigger-group-title">备份事件</div>
          <label class="checkbox-row" v-for="evt in backupEvents" :key="evt.key">
            <span class="checkbox-box" :class="{ checked: triggerEventsMap[evt.key], 'check-danger': evt.key.includes('failure') }">
              <span v-if="triggerEventsMap[evt.key]" class="check-icon">&#10003;</span>
            </span>
            <span class="checkbox-label">{{ evt.label }}</span>
          </label>
        </div>
      </div>

      <template #footer v-if="form.type">
        <ActionPill variant="outline" @click="router.push('/settings/notification-channels')">取消</ActionPill>
        <ActionPill variant="primary" :icon="Check" :disabled="saving" @click="handleSave">{{ saving ? '保存中...' : '保存渠道' }}</ActionPill>
      </template>
    </FormCard>

    <LoadingState v-else />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Check, ChatDotRound, Message, Link } from '@element-plus/icons-vue'
import { createChannel, updateChannel, getChannel } from '@/api/modules/notification'
import PageHeader from '@/components/common/PageHeader.vue'
import FormCard from '@/components/common/FormCard.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import LoadingState from '@/components/common/LoadingState.vue'

const router = useRouter()
const route = useRoute()

const editingId = computed(() => route.params.id ? Number(route.params.id) : null)
const isEdit = computed(() => !!editingId.value)
const pageLoading = ref(false)

const form = reactive({
  name: '',
  type: '' as string,
  enabled: true,
  notify_on_success: false,
  notify_on_failure: true
})

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

const channelTypes = [
  { value: 'dingtalk', label: '钉钉', icon: ChatDotRound },
  { value: 'wechat', label: '企业微信', icon: Message },
  { value: 'feishu', label: '飞书', icon: ChatDotRound },
  { value: 'email', label: '邮件', icon: Message },
  { value: 'lanxin', label: '蓝信', icon: ChatDotRound },
  { value: 'webhook', label: 'Webhook', icon: Link },
]

const triggerEventsMap = reactive<Record<string, boolean>>({
  sync_success: false, sync_failure: true, sync_conflict: true,
  webhook_received: false, webhook_error: true,
  cron_triggered: false,
  backup_success: false, backup_failure: true
})

const syncEvents = [
  { key: 'sync_success', label: '同步成功' },
  { key: 'sync_failure', label: '同步失败' },
  { key: 'sync_conflict', label: '同步冲突' },
]
const webhookEvents = [
  { key: 'webhook_received', label: 'Webhook 接收' },
  { key: 'webhook_error', label: 'Webhook 错误' },
]
const cronEvents = [
  { key: 'cron_triggered', label: '定时任务触发' },
]
const backupEvents = [
  { key: 'backup_success', label: '备份成功' },
  { key: 'backup_failure', label: '备份失败' },
]

const allTriggerEventKeys = Object.keys(triggerEventsMap)
const saving = ref(false)

function selectType(type: string) {
  if (editingId.value) return
  form.type = type
}

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

onMounted(async () => {
  if (editingId.value) {
    pageLoading.value = true
    try {
      const channel = await getChannel(editingId.value)
      if (!channel) { ElMessage.error('渠道不存在'); router.push('/settings/notification-channels'); return }
      form.name = channel.name
      form.type = channel.type
      form.enabled = channel.enabled
      form.notify_on_success = channel.notify_on_success
      form.notify_on_failure = channel.notify_on_failure

      const events = parseTriggerEvents(channel.trigger_events)
      allTriggerEventKeys.forEach(k => { triggerEventsMap[k] = false })
      if (events.length > 0) {
        events.forEach(e => { if (e in triggerEventsMap) triggerEventsMap[e] = true })
      }

      const config = JSON.parse(channel.config || '{}')
      Object.keys(config).forEach(key => {
        if (key in configForm) (configForm as Record<string, string>)[key] = config[key]
      })
      if (!config.security_type) {
        if ((form.type === 'dingtalk' || form.type === 'feishu') && config.secret) configForm.security_type = 'sign'
        else if (form.type === 'lanxin' && config.sign) configForm.security_type = 'sign'
        else configForm.security_type = 'none'
      }
    } catch (e: any) {
      ElMessage.error('加载渠道失败: ' + (e?.message || ''))
      router.push('/settings/notification-channels')
    } finally {
      pageLoading.value = false
    }
  }
})

function parseTriggerEvents(raw: string): string[] {
  if (!raw) return []
  try { const arr = JSON.parse(raw); return Array.isArray(arr) ? arr : [] } catch { return [] }
}

function getTriggerEventsJson(): string {
  return JSON.stringify(allTriggerEventKeys.filter(k => triggerEventsMap[k]))
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
  if (!form.name) { ElMessage.warning('请输入渠道名称'); return }
  if (!form.type) { ElMessage.warning('请选择渠道类型'); return }

  saving.value = true
  try {
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
      event_templates_json: ''
    }
    if (editingId.value) {
      await updateChannel({ ...params, id: editingId.value })
      ElMessage.success('渠道更新成功')
    } else {
      await createChannel(params)
      ElMessage.success('渠道创建成功')
    }
    router.push('/settings/notification-channels')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || '未知错误'))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.add-channel-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.field-input {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  outline: none;
  width: 100%;
  box-sizing: border-box;
}

.field-input:focus {
  border-color: var(--accent-primary);
}

select.field-input {
  appearance: auto;
  cursor: pointer;
}

.type-selector {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.type-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.type-btn:hover:not(.disabled) {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.type-btn.active {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: #fff;
}

.type-btn.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.divider {
  height: 1px;
  background: var(--border-color);
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.config-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.type-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: normal;
}

.badge-dingtalk { background: #FFFBEB; color: #F59E0B; }
.badge-wechat { background: #ECFDF5; color: #10B981; }
.badge-feishu { background: var(--accent-bg); color: #6366F1; }
.badge-email { background: var(--accent-bg); color: #6366F1; }
.badge-lanxin { background: #FEF2F2; color: #EF4444; }
.badge-webhook { background: var(--accent-bg); color: var(--text-color-secondary); }

.form-row {
  display: flex;
  gap: 12px;
}

.mode-selector {
  display: flex;
  gap: 8px;
}

.mode-btn {
  padding: 8px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.mode-btn.active {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: #fff;
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: 8px;
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
  background: var(--accent-primary);
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

.trigger-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.trigger-group {
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.trigger-group-title {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-color-secondary);
  margin-bottom: 2px;
}

.checkbox-row {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-color-primary);
}

.checkbox-box {
  width: 14px;
  height: 14px;
  border-radius: 3px;
  border: 1.5px solid var(--border-color, #d1d5db);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.15s;
}

.checkbox-box.checked {
  border-color: var(--accent-primary);
  background: var(--accent-primary);
}

.checkbox-box.check-success.checked {
  border-color: #10B981;
  background: #10B981;
}

.checkbox-box.check-danger.checked {
  border-color: #EF4444;
  background: #EF4444;
}

.checkbox-box.check-warning.checked {
  border-color: #F59E0B;
  background: #F59E0B;
}

.check-icon {
  font-size: 10px;
  color: #fff;
  line-height: 1;
}

.checkbox-label {
  font-size: 13px;
}
</style>
