<template>
  <div class="settings-page">
    <PageHeader title="系统设置" subtitle="管理应用配置和安全设置" />

    <div class="card-grid">
      <div v-for="card in cards" :key="card.route" class="settings-card" @click="$router.push(card.route)">
        <div class="card-icon" :class="'card-icon--' + card.color">
          <el-icon :size="20"><component :is="card.icon" /></el-icon>
        </div>
        <h3 class="card-title">{{ card.title }}</h3>
        <p class="card-desc">{{ card.desc }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Key, Lock, Bell, Connection, MagicStick, SetUp, Cpu, User } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'

const cards = [
  { title: 'LLM 配置', desc: '管理大模型提供商，用于代码审查、Spec 辅助等 AI 功能', icon: Cpu, color: 'violet', route: '/settings/llm' },
  { title: '代码审查', desc: '配置代码审查行为和审查规则', icon: MagicStick, color: 'purple', route: '/settings/code-review' },
  { title: 'SSH 密钥', desc: '管理 SSH 密钥，用于 Git 仓库认证', icon: Key, color: 'red', route: '/settings/ssh-keys' },
  { title: '凭证管理', desc: '管理平台访问凭证和 Token', icon: Lock, color: 'indigo', route: '/settings/credentials' },
  { title: '通知渠道', desc: '配置邮件、钉钉、微信等通知方式', icon: Bell, color: 'amber', route: '/settings/notification-channels' },
  { title: '平台配置', desc: '管理 GitLab/GitHub/Gitea 平台集成', icon: Connection, color: 'blue', route: '/settings/platforms' },
  { title: '分支规则', desc: '定义分支命名前缀和保护规则', icon: SetUp, color: 'teal', route: '/settings/branch-rules' },
  { title: 'Git 作者', desc: '管理 Git 提交身份和别名，修复历史提交作者', icon: User, color: 'green', route: '/settings/author' },
]
</script>

<style scoped>
.settings-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.settings-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 24px;
  background: var(--surface-card);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.settings-card:hover {
  background: var(--bg-color-page);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.08);
}

.card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
}

.card-icon--red { background: #FEF2F2; color: #F56C6C; }
.card-icon--indigo { background: var(--accent-bg); color: #6366F1; }
.card-icon--amber { background: #FFFBEB; color: #F59E0B; }
.card-icon--blue { background: #EFF6FF; color: #3B82F6; }
.card-icon--purple { background: #F3E8FF; color: #8B5CF6; }
.card-icon--violet { background: #EDE9FE; color: #7C3AED; }
.card-icon--teal { background: #F0FDFA; color: #14B8A6; }
.card-icon--green { background: #ECFDF5; color: #10B981; }

.card-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.card-desc {
  margin: 0;
  font-size: 13px;
  color: var(--text-color-secondary);
  line-height: 1.6;
}
</style>
