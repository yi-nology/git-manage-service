<template>
  <div class="sync-config">
    <PageHeader title="高级配置">
      <template #actions>
        <el-button type="primary" @click="saveConfig" :icon="Check">保存配置</el-button>
      </template>
    </PageHeader>

    <el-tabs v-model="activeTab" type="card">
      <el-tab-pane label="全局配置" name="global">
        <el-card class="config-card">
          <template #header>
            <div class="card-header">
              <span>同步引擎配置</span>
            </div>
          </template>
          <el-form :model="globalConfig" label-width="180px">
            <el-form-item label="最大并发同步数">
              <el-input-number v-model="globalConfig.maxConcurrentSyncs" :min="1" :max="10" />
              <div class="form-tip">同时执行的同步任务数量</div>
            </el-form-item>

            <el-form-item label="默认超时时间（秒）">
              <el-input-number v-model="globalConfig.defaultTimeout" :min="30" :max="3600" />
            </el-form-item>

            <el-form-item label="默认重试次数">
              <el-input-number v-model="globalConfig.defaultRetryCount" :min="0" :max="5" />
            </el-form-item>

            <el-form-item label="重试间隔（秒）">
              <el-input-number v-model="globalConfig.retryInterval" :min="5" :max="300" />
            </el-form-item>

            <el-form-item label="历史记录保留天数">
              <el-input-number v-model="globalConfig.historyRetentionDays" :min="1" :max="365" />
            </el-form-item>

            <el-form-item label="启用增量同步">
              <el-switch v-model="globalConfig.enableIncrementalSync" />
              <div class="form-tip">仅同步变更的分支，提升同步效率</div>
            </el-form-item>

            <el-form-item label="启用分布式锁">
              <el-switch v-model="globalConfig.enableDistributedLock" />
              <div class="form-tip">多实例部署时防止并发冲突</div>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="Webhook 配置" name="webhook">
        <el-card class="config-card">
          <template #header>
            <div class="card-header">
              <span>Webhook 服务配置</span>
            </div>
          </template>
          <el-form :model="webhookConfig" label-width="180px">
            <el-form-item label="启用 Webhook">
              <el-switch v-model="webhookConfig.enabled" />
            </el-form-item>

            <el-form-item label="Webhook URL 前缀">
              <el-input v-model="webhookConfig.urlPrefix" placeholder="https://your-domain.com/api/webhook" />
            </el-form-item>

            <el-form-item label="签名密钥">
              <el-input v-model="webhookConfig.secretKey" type="password" show-password placeholder="用于验证 Webhook 签名" />
            </el-form-item>

            <el-form-item label="请求超时（秒）">
              <el-input-number v-model="webhookConfig.requestTimeout" :min="5" :max="60" />
            </el-form-item>

            <el-form-item label="允许的 IP 白名单">
              <el-input v-model="webhookConfig.ipWhitelist" type="textarea" :rows="3" placeholder="每行一个 IP，支持 CIDR 格式" />
              <div class="form-tip">留空表示允许所有 IP</div>
            </el-form-item>

            <el-form-item label="启用事件日志">
              <el-switch v-model="webhookConfig.enableEventLog" />
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="通知配置" name="notification">
        <el-card class="config-card">
          <template #header>
            <div class="card-header">
              <span>通知渠道配置</span>
            </div>
          </template>
          <el-form :model="notificationConfig" label-width="180px">
            <el-form-item label="启用邮件通知">
              <el-switch v-model="notificationConfig.emailEnabled" />
            </el-form-item>

            <el-form-item v-if="notificationConfig.emailEnabled" label="邮件服务器">
              <el-input v-model="notificationConfig.smtpHost" placeholder="smtp.example.com" />
            </el-form-item>

            <el-form-item v-if="notificationConfig.emailEnabled" label="SMTP 端口">
              <el-input-number v-model="notificationConfig.smtpPort" :min="1" :max="65535" />
            </el-form-item>

            <el-form-item v-if="notificationConfig.emailEnabled" label="收件人">
              <el-input v-model="notificationConfig.emailRecipients" type="textarea" :rows="2" placeholder="每行一个邮箱地址" />
            </el-form-item>

            <el-form-item label="启用 Slack 通知">
              <el-switch v-model="notificationConfig.slackEnabled" />
            </el-form-item>

            <el-form-item v-if="notificationConfig.slackEnabled" label="Slack Webhook">
              <el-input v-model="notificationConfig.slackWebhook" placeholder="https://hooks.slack.com/services/..." />
            </el-form-item>

            <el-form-item label="通知触发条件">
              <el-checkbox-group v-model="notificationConfig.notifyOn">
                <el-checkbox label="success">同步成功</el-checkbox>
                <el-checkbox label="failed">同步失败</el-checkbox>
                <el-checkbox label="timeout">同步超时</el-checkbox>
                <el-checkbox label="manual">手动触发</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { Check } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'

const activeTab = ref('global')

const globalConfig = reactive({
  maxConcurrentSyncs: 3,
  defaultTimeout: 300,
  defaultRetryCount: 2,
  retryInterval: 30,
  historyRetentionDays: 30,
  enableIncrementalSync: true,
  enableDistributedLock: true,
})

const webhookConfig = reactive({
  enabled: true,
  urlPrefix: 'https://your-domain.com/api/webhook',
  secretKey: 'your-secret-key-here',
  requestTimeout: 10,
  ipWhitelist: '',
  enableEventLog: true,
})

const notificationConfig = reactive({
  emailEnabled: false,
  smtpHost: 'smtp.example.com',
  smtpPort: 587,
  emailRecipients: '',
  slackEnabled: false,
  slackWebhook: '',
  notifyOn: ['failed', 'timeout'],
})

function saveConfig() {
  console.log('Save config:', {
    globalConfig,
    webhookConfig,
    notificationConfig,
  })
}
</script>

<style scoped lang="scss">
.sync-config {
  .config-card {
    max-width: 900px;

    .card-header {
      font-weight: 600;
    }

    .form-tip {
      font-size: 12px;
      color: var(--text-color-placeholder);
      margin-top: 4px;
    }
  }
}
</style>
