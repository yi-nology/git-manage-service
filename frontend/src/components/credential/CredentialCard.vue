<template>
  <div class="cred-card">
    <div class="card-top">
      <div class="card-title">
        <span class="type-badge" :class="'type-' + typeClass">{{ typeLabel }}</span>
        <span class="card-name">{{ credential.name }}</span>
      </div>
      <div class="card-actions">
        <button class="action-btn btn-edit" @click="$emit('edit', credential)">编辑</button>
        <el-popconfirm
          title="确定删除该凭证？"
          confirm-button-text="删除"
          cancel-button-text="取消"
          @confirm="$emit('delete', credential)"
        >
          <template #reference>
            <button class="action-btn btn-delete">删除</button>
          </template>
        </el-popconfirm>
      </div>
    </div>

    <div v-if="credential.url_pattern" class="card-url">
      <span class="url-icon">@</span>
      {{ credential.url_pattern }}
    </div>

    <div v-if="credential.description" class="card-desc">{{ credential.description }}</div>

    <div class="card-meta">
      <span class="meta-item" v-if="credential.ssh_key_name">
        <span class="meta-label">SSH 密钥:</span> {{ credential.ssh_key_name }}
        <span v-if="credential.ssh_key_type" class="meta-type" :class="'type-' + sshKeyClass(credential.ssh_key_type)">
          {{ sshKeyTypeLabel(credential.ssh_key_type) }}
        </span>
      </span>
      <span class="meta-item" v-if="credential.ssh_key_path">
        <span class="meta-label">路径:</span> {{ credential.ssh_key_path }}
      </span>
      <span class="meta-item" v-if="credential.username">
        <span class="meta-label">用户名:</span> {{ credential.username }}
      </span>
      <span class="meta-item">
        <span class="meta-label">密钥/密码:</span>
        <span class="secret-tag" :class="credential.has_secret ? 'tag-success' : 'tag-default'">
          {{ credential.has_secret ? '已配置' : '未配置' }}
        </span>
      </span>
      <span class="meta-item" v-if="credential.last_used_at">
        <span class="meta-label">最后使用:</span> {{ formatTime(credential.last_used_at) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CredentialDTO } from '@/types/credential'
import { computed } from 'vue'

const props = defineProps<{
  credential: CredentialDTO
}>()

defineEmits<{
  (e: 'edit', cred: CredentialDTO): void
  (e: 'delete', cred: CredentialDTO): void
}>()

const typeLabel = computed(() => {
  const map: Record<string, string> = {
    ssh_key: 'SSH',
    http_basic: 'HTTP',
    http_token: 'Token',
  }
  return map[props.credential.type] || props.credential.type
})

const typeClass = computed(() => {
  const map: Record<string, string> = {
    ssh_key: 'ssh',
    http_basic: 'http',
    http_token: 'token',
  }
  return map[props.credential.type] || 'default'
})

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString()
}

const SSH_KEY_TYPE_LABELS: Record<string, string> = {
  rsa: 'RSA', ed25519: 'Ed25519', ecdsa: 'ECDSA', dsa: 'DSA', unknown: '未知',
}

function sshKeyTypeLabel(t: string): string {
  return SSH_KEY_TYPE_LABELS[t?.toLowerCase()] ?? t?.toUpperCase() ?? ''
}

function sshKeyClass(t: string): string {
  const lower = t?.toLowerCase()
  if (lower === 'ed25519') return 'success'
  if (lower === 'rsa') return 'info'
  if (lower === 'ecdsa') return 'warning'
  return 'default'
}
</script>

<style scoped>
.cred-card {
  border-radius: 12px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.type-badge.type-ssh {
  background: #ECFDF5;
  color: #10B981;
}

.type-badge.type-http {
  background: #FFFBEB;
  color: #F59E0B;
}

.type-badge.type-token {
  background: var(--accent-bg);
  color: #6366F1;
}

.type-badge.type-default {
  background: var(--accent-bg);
  color: var(--text-color-secondary);
}

.card-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.card-actions {
  display: flex;
  gap: 4px;
}

.action-btn {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  border: 1px solid transparent;
  cursor: pointer;
  background: none;
  transition: all 0.2s;
}

.btn-edit {
  color: #6366F1;
  border-color: #6366F1;
}

.btn-edit:hover {
  background: var(--accent-bg);
}

.btn-delete {
  color: #EF4444;
  border-color: #EF4444;
}

.btn-delete:hover {
  background: #FEF2F2;
}

.card-url {
  font-size: 12px;
  color: var(--accent-primary);
  display: flex;
  align-items: center;
  gap: 4px;
}

.url-icon {
  font-weight: 600;
}

.card-desc {
  font-size: 12px;
  color: var(--text-color-secondary);
  line-height: 1.5;
}

.card-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.meta-item {
  font-size: 12px;
  color: var(--text-color-secondary);
  display: flex;
  align-items: center;
  gap: 4px;
}

.meta-label {
  font-weight: 500;
  color: var(--text-color-placeholder, #9ca3af);
}

.secret-tag {
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 11px;
}

.secret-tag.tag-success {
  background: #ECFDF5;
  color: #10B981;
}

.secret-tag.tag-default {
  background: var(--accent-bg);
  color: var(--text-color-secondary);
}

.meta-type {
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 11px;
}

.meta-type.type-success {
  background: #ECFDF5;
  color: #10B981;
}

.meta-type.type-info {
  background: var(--accent-bg);
  color: #6366F1;
}

.meta-type.type-warning {
  background: #FFFBEB;
  color: #F59E0B;
}

.meta-type.type-default {
  background: var(--accent-bg);
  color: var(--text-color-secondary);
}
</style>
