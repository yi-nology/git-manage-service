<template>
  <template v-if="channelType === 'email'">
    <el-divider content-position="left">邮件配置</el-divider>
    <el-form-item label="SMTP 服务器">
      <el-row :gutter="12">
        <el-col :span="16">
          <el-input v-model="configForm.smtp_host" placeholder="smtp.example.com" />
        </el-col>
        <el-col :span="8">
          <el-input v-model="configForm.smtp_port" placeholder="端口 587" />
        </el-col>
      </el-row>
    </el-form-item>
    <el-form-item label="用户名">
      <el-input v-model="configForm.username" placeholder="发件邮箱账号" />
    </el-form-item>
    <el-form-item label="密码">
      <el-input v-model="configForm.password" type="password" show-password placeholder="邮箱密码或授权码" />
    </el-form-item>
    <el-form-item label="发件人">
      <el-input v-model="configForm.from" placeholder="Git管理服务 <noreply@example.com>" />
    </el-form-item>
    <el-form-item label="收件人">
      <el-input v-model="configForm.to" placeholder="多个邮箱用逗号分隔" />
    </el-form-item>
  </template>

  <template v-if="channelType === 'dingtalk'">
    <el-divider content-position="left">钉钉配置</el-divider>
    <el-form-item label="Webhook URL">
      <el-input v-model="configForm.webhook_url" placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx" />
    </el-form-item>
    <el-form-item label="安全模式">
      <el-radio-group v-model="configForm.security_type">
        <el-radio value="none">无</el-radio>
        <el-radio value="sign">签名</el-radio>
        <el-radio value="keyword">关键字</el-radio>
      </el-radio-group>
    </el-form-item>
    <el-form-item v-if="configForm.security_type === 'sign'" label="签名密钥">
      <el-input v-model="configForm.secret" placeholder="SEC开头的密钥" />
    </el-form-item>
    <el-form-item v-if="configForm.security_type === 'keyword'" label="关键字">
      <el-input v-model="configForm.keywords" placeholder="消息中需要包含的关键字" />
    </el-form-item>
  </template>

  <template v-if="channelType === 'wechat'">
    <el-divider content-position="left">企业微信配置</el-divider>
    <el-form-item label="Webhook URL">
      <el-input v-model="configForm.webhook_url" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx" />
    </el-form-item>
  </template>

  <template v-if="channelType === 'lanxin'">
    <el-divider content-position="left">蓝信配置</el-divider>
    <el-form-item label="Webhook URL">
      <el-input v-model="configForm.webhook_url" placeholder="蓝信机器人 Webhook 地址" />
    </el-form-item>
    <el-form-item label="安全模式">
      <el-radio-group v-model="configForm.security_type">
        <el-radio value="none">无</el-radio>
        <el-radio value="sign">签名</el-radio>
        <el-radio value="keyword">关键字</el-radio>
      </el-radio-group>
    </el-form-item>
    <el-form-item v-if="configForm.security_type === 'sign'" label="签名密钥">
      <el-input v-model="configForm.sign" placeholder="签名密钥" />
    </el-form-item>
    <el-form-item v-if="configForm.security_type === 'keyword'" label="关键字">
      <el-input v-model="configForm.keywords" placeholder="消息中需要包含的关键字" />
    </el-form-item>
  </template>

  <template v-if="channelType === 'feishu'">
    <el-divider content-position="left">飞书配置</el-divider>
    <el-form-item label="Webhook URL">
      <el-input v-model="configForm.webhook_url" placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx" />
    </el-form-item>
    <el-form-item label="安全模式">
      <el-radio-group v-model="configForm.security_type">
        <el-radio value="none">无</el-radio>
        <el-radio value="sign">签名</el-radio>
        <el-radio value="keyword">关键字</el-radio>
      </el-radio-group>
    </el-form-item>
    <el-form-item v-if="configForm.security_type === 'sign'" label="签名密钥">
      <el-input v-model="configForm.secret" placeholder="飞书签名密钥" />
    </el-form-item>
    <el-form-item v-if="configForm.security_type === 'keyword'" label="关键字">
      <el-input v-model="configForm.keywords" placeholder="消息中需要包含的关键字" />
    </el-form-item>
  </template>

  <template v-if="channelType === 'webhook'">
    <el-divider content-position="left">Webhook 配置</el-divider>
    <el-form-item label="URL">
      <el-input v-model="configForm.url" placeholder="https://your-server.com/webhook" />
    </el-form-item>
    <el-form-item label="请求方法">
      <el-select v-model="configForm.method" style="width: 100%">
        <el-option label="POST" value="POST" />
        <el-option label="GET" value="GET" />
      </el-select>
    </el-form-item>
    <el-form-item label="Content-Type">
      <el-select v-model="configForm.content_type" style="width: 100%">
        <el-option label="application/json" value="application/json" />
        <el-option label="application/x-www-form-urlencoded" value="application/x-www-form-urlencoded" />
      </el-select>
    </el-form-item>
  </template>
</template>

<script setup lang="ts">
defineProps<{
  channelType: string
  configForm: Record<string, any>
}>()
</script>
