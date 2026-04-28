<template>
  <div class="add-credential-page">
    <div class="title-row">
      <div class="title-left">
        <button class="back-btn" @click="router.push('/settings/credentials')">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h2 class="page-title">{{ isEdit ? '编辑凭证' : '添加凭证' }}</h2>
      </div>
    </div>

    <div class="form-card" v-if="!pageLoading">
      <div class="form-field">
        <label class="field-label">凭证名称</label>
        <input v-model="form.name" placeholder="例如: GitHub SSH Key" class="field-input" />
      </div>

      <div class="form-field">
        <label class="field-label">凭证类型</label>
        <div class="type-selector">
          <button
            v-for="t in credTypes"
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

      <div class="form-field">
        <label class="field-label">描述</label>
        <input v-model="form.description" placeholder="可选描述信息" class="field-input" />
      </div>

      <div class="divider"></div>

      <template v-if="form.type === 'ssh_key'">
        <div class="config-header">
          <span class="config-title">SSH 密钥配置</span>
          <span class="type-badge badge-ssh">SSH 密钥</span>
        </div>

        <div class="form-field">
          <label class="field-label">密钥来源</label>
          <div class="mode-selector">
            <button class="mode-btn" :class="{ active: sshSource === 'database' }" @click="sshSource = 'database'">数据库密钥</button>
            <button class="mode-btn" :class="{ active: sshSource === 'local' }" @click="sshSource = 'local'">本地文件</button>
          </div>
        </div>

        <div class="form-field" v-if="sshSource === 'database'">
          <label class="field-label">SSH 密钥</label>
          <select v-model="form.ssh_key_id" class="field-input">
            <option :value="undefined" disabled>选择数据库中的 SSH 密钥</option>
            <option v-for="key in sshKeys" :key="key.id" :value="key.id">{{ key.name }} ({{ sshKeyTypeLabel(key.key_type) }})</option>
          </select>
        </div>

        <template v-if="sshSource === 'local'">
          <div class="form-field">
            <label class="field-label">密钥路径</label>
            <input v-model="form.ssh_key_path" placeholder="例如: ~/.ssh/id_rsa" class="field-input" />
          </div>
          <div class="form-field">
            <label class="field-label">密码短语</label>
            <input v-model="form.secret" type="password" placeholder="可选 passphrase" class="field-input" />
          </div>
        </template>
      </template>

      <template v-if="form.type === 'http_basic' || form.type === 'http_token'">
        <div class="config-header">
          <span class="config-title">{{ form.type === 'http_basic' ? 'HTTP 账号密码配置' : 'HTTP Token 配置' }}</span>
          <span class="type-badge badge-http">HTTP</span>
        </div>

        <div class="form-field">
          <label class="field-label">用户名</label>
          <input v-model="form.username" :placeholder="form.type === 'http_token' ? '通常不需要，或填写用户名' : '用户名'" class="field-input" />
        </div>

        <div class="form-field">
          <label class="field-label">{{ form.type === 'http_token' ? 'Token' : '密码' }}</label>
          <input v-model="form.secret" type="password" :placeholder="form.type === 'http_token' ? 'Personal Access Token' : '密码'" class="field-input" />
        </div>
      </template>

      <div class="divider"></div>

      <div class="form-field">
        <label class="field-label">URL 匹配</label>
        <input v-model="form.url_pattern" placeholder="例如: *.github.com（可选，用于智能推荐）" class="field-input" />
        <div class="field-hint">当仓库 URL 匹配此模式时，系统将自动推荐该凭证。支持 * 通配符。</div>
      </div>

      <div class="form-actions">
        <button class="action-pill pill-outline" @click="router.push('/settings/credentials')">取消</button>
        <button class="action-pill pill-primary" @click="handleSave" :disabled="saving">
          <el-icon><Plus /></el-icon>
          {{ saving ? '保存中...' : (isEdit ? '保存凭证' : '创建凭证') }}
        </button>
      </div>
    </div>

    <div v-else class="loading-card">
      <div class="loading-spinner"></div>
      <span>加载中...</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Plus, Key, Lock, Ticket } from '@element-plus/icons-vue'
import { createCredential, updateCredential, getCredential } from '@/api/modules/credential'
import { listDBSSHKeys, type DBSSHKey } from '@/api/modules/sshkey'
import type { CredentialType } from '@/types/credential'

const router = useRouter()
const route = useRoute()

const editingId = computed(() => route.params.id ? Number(route.params.id) : null)
const isEdit = computed(() => !!editingId.value)
const pageLoading = ref(false)
const saving = ref(false)
const sshKeys = ref<DBSSHKey[]>([])
const sshSource = ref<'database' | 'local'>('database')

const form = reactive({
  name: '',
  type: 'ssh_key' as CredentialType | '',
  description: '',
  ssh_key_id: undefined as number | undefined,
  ssh_key_path: '',
  username: '',
  secret: '',
  url_pattern: '',
})

const credTypes = [
  { value: 'ssh_key' as CredentialType, label: 'SSH 密钥', icon: Key },
  { value: 'http_basic' as CredentialType, label: 'HTTP 账号密码', icon: Lock },
  { value: 'http_token' as CredentialType, label: 'HTTP Token', icon: Ticket },
]

const SSH_KEY_TYPE_LABELS: Record<string, string> = {
  rsa: 'RSA', ed25519: 'Ed25519', ecdsa: 'ECDSA', dsa: 'DSA', unknown: '未知',
}

function sshKeyTypeLabel(t: string): string {
  return SSH_KEY_TYPE_LABELS[t?.toLowerCase()] ?? t?.toUpperCase() ?? ''
}

function selectType(type: CredentialType) {
  if (editingId.value) return
  form.type = type
  form.ssh_key_id = undefined
  form.ssh_key_path = ''
  form.username = ''
  form.secret = ''
}

async function handleSave() {
  if (!form.name) { ElMessage.warning('请输入凭证名称'); return }
  if (!form.type) { ElMessage.warning('请选择凭证类型'); return }

  saving.value = true
  try {
    const data: Record<string, any> = {
      name: form.name,
      type: form.type,
      description: form.description || undefined,
      url_pattern: form.url_pattern || undefined,
    }

    if (form.type === 'ssh_key') {
      if (sshSource.value === 'database') {
        data.ssh_key_id = form.ssh_key_id
        data.ssh_key_path = ''
      } else {
        data.ssh_key_path = form.ssh_key_path
        data.secret = form.secret || undefined
      }
    } else {
      data.username = form.username || undefined
      data.secret = form.secret || undefined
    }

    if (editingId.value) {
      await updateCredential(editingId.value, data)
      ElMessage.success('凭证已更新')
    } else {
      await createCredential(data as any)
      ElMessage.success('凭证已创建')
    }
    router.push('/settings/credentials')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    sshKeys.value = await listDBSSHKeys() || []
  } catch { /* ignore */ }

  if (editingId.value) {
    pageLoading.value = true
    try {
      const cred = await getCredential(editingId.value)
      if (!cred) { ElMessage.error('凭证不存在'); router.push('/settings/credentials'); return }
      form.name = cred.name
      form.type = cred.type
      form.description = cred.description || ''
      form.ssh_key_id = cred.ssh_key_id
      form.ssh_key_path = cred.ssh_key_path || ''
      form.username = cred.username || ''
      form.secret = ''
      form.url_pattern = cred.url_pattern || ''
      sshSource.value = cred.ssh_key_id ? 'database' : 'local'
    } catch (e: any) {
      ElMessage.error('加载凭证失败: ' + (e?.message || ''))
      router.push('/settings/credentials')
    } finally {
      pageLoading.value = false
    }
  }
})
</script>

<style scoped>
.add-credential-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-color-secondary);
  background: none;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 6px;
  padding: 6px 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.back-btn:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.form-card {
  border-radius: 12px;
  background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  padding: 24px;
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
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page, #fff);
  outline: none;
  width: 100%;
  box-sizing: border-box;
}

.field-input:focus {
  border-color: var(--accent-primary, #6366F1);
}

select.field-input {
  appearance: auto;
  cursor: pointer;
}

.field-hint {
  font-size: 12px;
  color: var(--text-color-secondary, #94A3B8);
  line-height: 1.5;
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
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.type-btn:hover:not(.disabled) {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.type-btn.active {
  background: var(--accent-primary, #6366F1);
  border-color: var(--accent-primary, #6366F1);
  color: #fff;
}

.type-btn.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.divider {
  height: 1px;
  background: var(--border-color, #e5e7eb);
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

.badge-ssh { background: #ECFDF5; color: #10B981; }
.badge-http { background: #EEF2FF; color: #6366F1; }

.mode-selector {
  display: flex;
  gap: 8px;
}

.mode-btn {
  padding: 8px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.mode-btn.active {
  background: var(--accent-primary, #6366F1);
  border-color: var(--accent-primary, #6366F1);
  color: #fff;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.action-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 500;
  padding: 10px 20px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
}

.action-pill:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pill-primary {
  background: var(--accent-primary, #6366F1);
  color: #fff;
}

.pill-primary:hover:not(:disabled) {
  opacity: 0.9;
}

.pill-outline {
  background: transparent;
  border: 1px solid var(--border-color, #e5e7eb);
  color: var(--text-color-primary);
}

.pill-outline:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
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
</style>
