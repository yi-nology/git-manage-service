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

      <ChannelFormFields :type="form.type" :config="configForm" />

      <div class="divider" v-if="form.type"></div>

      <ChannelTriggerEvents
        v-if="form.type"
        v-model:enabled="form.enabled"
        :events-map="triggerEventsMap"
      />

      <template #footer v-if="form.type">
        <ActionPill variant="outline" @click="router.push(ROUTES.NotificationChannels)">取消</ActionPill>
        <ActionPill variant="primary" :icon="Check" :disabled="saving" @click="handleSave">{{ saving ? '保存中...' : '保存渠道' }}</ActionPill>
      </template>
    </FormCard>

    <LoadingState v-else />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ROUTES } from '@/router/paths'
import { toastApiError } from '@/composables/useNotification'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Check, ChatDotRound, Message, Link } from '@element-plus/icons-vue'
import { createChannel, updateChannel, getChannel } from '@/api/modules/notification'
import PageHeader from '@/components/common/PageHeader.vue'
import FormCard from '@/components/common/FormCard.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import ChannelFormFields from '@/components/settings/ChannelFormFields.vue'
import ChannelTriggerEvents from '@/components/settings/ChannelTriggerEvents.vue'

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
      if (!channel) { ElMessage.error('渠道不存在'); router.push(ROUTES.NotificationChannels); return }
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
      toastApiError(e, '加载渠道失败: ', '')
      router.push(ROUTES.NotificationChannels)
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
    router.push(ROUTES.NotificationChannels)
  } catch (e: any) {
    toastApiError(e, '保存失败: ', '未知错误')
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
</style>
