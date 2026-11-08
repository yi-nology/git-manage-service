<template>
  <template v-if="type === 'email'">
    <div class="config-header">
      <span class="config-title">邮件配置</span>
      <span class="type-badge badge-email">邮件</span>
    </div>
    <div class="form-row">
      <div class="form-field" style="flex:2">
        <label class="field-label">SMTP 服务器</label>
        <input v-model="config.smtp_host" placeholder="smtp.example.com" class="field-input" />
      </div>
      <div class="form-field" style="flex:1">
        <label class="field-label">端口</label>
        <input v-model="config.smtp_port" placeholder="587" class="field-input" />
      </div>
    </div>
    <div class="form-field">
      <label class="field-label">用户名</label>
      <input v-model="config.username" placeholder="发件邮箱账号" class="field-input" />
    </div>
    <div class="form-field">
      <label class="field-label">密码</label>
      <input v-model="config.password" type="password" placeholder="邮箱密码或授权码" class="field-input" />
    </div>
    <div class="form-field">
      <label class="field-label">发件人</label>
      <input v-model="config.from" placeholder="Git管理服务 &lt;noreply@example.com&gt;" class="field-input" />
    </div>
    <div class="form-field">
      <label class="field-label">收件人</label>
      <input v-model="config.to" placeholder="多个邮箱用逗号分隔" class="field-input" />
    </div>
  </template>

  <template v-if="type === 'dingtalk'">
    <div class="config-header">
      <span class="config-title">钉钉配置</span>
      <span class="type-badge badge-dingtalk">钉钉机器人</span>
    </div>
    <div class="form-field">
      <label class="field-label">Webhook URL</label>
      <input v-model="config.webhook_url" placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx" class="field-input" />
    </div>
    <div class="form-field">
      <label class="field-label">安全模式</label>
      <div class="mode-selector">
        <button class="mode-btn" :class="{ active: config.security_type === 'none' }" @click="config.security_type = 'none'">无</button>
        <button class="mode-btn" :class="{ active: config.security_type === 'sign' }" @click="config.security_type = 'sign'">签名</button>
        <button class="mode-btn" :class="{ active: config.security_type === 'keyword' }" @click="config.security_type = 'keyword'">关键字</button>
      </div>
    </div>
    <div class="form-field" v-if="config.security_type === 'sign'">
      <label class="field-label">签名密钥</label>
      <input v-model="config.secret" placeholder="SEC开头的密钥" class="field-input" />
    </div>
    <div class="form-field" v-if="config.security_type === 'keyword'">
      <label class="field-label">关键字</label>
      <input v-model="config.keywords" placeholder="消息中需要包含的关键字" class="field-input" />
    </div>
  </template>

  <template v-if="type === 'wechat'">
    <div class="config-header">
      <span class="config-title">企业微信配置</span>
      <span class="type-badge badge-wechat">企业微信</span>
    </div>
    <div class="form-field">
      <label class="field-label">Webhook URL</label>
      <input v-model="config.webhook_url" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx" class="field-input" />
    </div>
  </template>

  <template v-if="type === 'feishu'">
    <div class="config-header">
      <span class="config-title">飞书配置</span>
      <span class="type-badge badge-feishu">飞书机器人</span>
    </div>
    <div class="form-field">
      <label class="field-label">Webhook URL</label>
      <input v-model="config.webhook_url" placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx" class="field-input" />
    </div>
    <div class="form-field">
      <label class="field-label">安全模式</label>
      <div class="mode-selector">
        <button class="mode-btn" :class="{ active: config.security_type === 'none' }" @click="config.security_type = 'none'">无</button>
        <button class="mode-btn" :class="{ active: config.security_type === 'sign' }" @click="config.security_type = 'sign'">签名</button>
        <button class="mode-btn" :class="{ active: config.security_type === 'keyword' }" @click="config.security_type = 'keyword'">关键字</button>
      </div>
    </div>
    <div class="form-field" v-if="config.security_type === 'sign'">
      <label class="field-label">签名密钥</label>
      <input v-model="config.secret" placeholder="飞书签名密钥" class="field-input" />
    </div>
    <div class="form-field" v-if="config.security_type === 'keyword'">
      <label class="field-label">关键字</label>
      <input v-model="config.keywords" placeholder="消息中需要包含的关键字" class="field-input" />
    </div>
  </template>

  <template v-if="type === 'lanxin'">
    <div class="config-header">
      <span class="config-title">蓝信配置</span>
      <span class="type-badge badge-lanxin">蓝信机器人</span>
    </div>
    <div class="form-field">
      <label class="field-label">Webhook URL</label>
      <input v-model="config.webhook_url" placeholder="蓝信机器人 Webhook 地址" class="field-input" />
    </div>
    <div class="form-field">
      <label class="field-label">安全模式</label>
      <div class="mode-selector">
        <button class="mode-btn" :class="{ active: config.security_type === 'none' }" @click="config.security_type = 'none'">无</button>
        <button class="mode-btn" :class="{ active: config.security_type === 'sign' }" @click="config.security_type = 'sign'">签名</button>
        <button class="mode-btn" :class="{ active: config.security_type === 'keyword' }" @click="config.security_type = 'keyword'">关键字</button>
      </div>
    </div>
    <div class="form-field" v-if="config.security_type === 'sign'">
      <label class="field-label">签名密钥</label>
      <input v-model="config.sign" placeholder="签名密钥" class="field-input" />
    </div>
    <div class="form-field" v-if="config.security_type === 'keyword'">
      <label class="field-label">关键字</label>
      <input v-model="config.keywords" placeholder="消息中需要包含的关键字" class="field-input" />
    </div>
  </template>

  <template v-if="type === 'webhook'">
    <div class="config-header">
      <span class="config-title">Webhook 配置</span>
      <span class="type-badge badge-webhook">自定义 Webhook</span>
    </div>
    <div class="form-field">
      <label class="field-label">URL</label>
      <input v-model="config.url" placeholder="https://your-server.com/webhook" class="field-input" />
    </div>
    <div class="form-row">
      <div class="form-field" style="flex:1">
        <label class="field-label">请求方法</label>
        <select v-model="config.method" class="field-input">
          <option value="POST">POST</option>
          <option value="GET">GET</option>
        </select>
      </div>
      <div class="form-field" style="flex:1">
        <label class="field-label">Content-Type</label>
        <select v-model="config.content_type" class="field-input">
          <option value="application/json">application/json</option>
          <option value="application/x-www-form-urlencoded">application/x-www-form-urlencoded</option>
        </select>
      </div>
    </div>
  </template>
</template>

<script setup lang="ts">
// 通知渠道配置字段：按渠道类型渲染各字段（邮件/钉钉/企业微信/飞书/蓝信/Webhook）。
// config 为页面传入的响应式表单对象，字段直接回写（与 MirrorFormDialog 的 form prop 约定一致）。
defineProps<{
  type: string
  config: Record<string, string>
}>()
</script>

<style scoped>
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
</style>
