<template>
  <div class="platform-config-page">
    <div class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="$router.push('/settings')">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h2>平台配置</h2>
      </div>
      <div class="header-actions">
        <button class="action-pill action-pill--primary" @click="openAddDialog">
          <el-icon><Plus /></el-icon> 添加平台
        </button>
      </div>
    </div>

    <div v-if="loading" class="empty-card">
      <div class="loading-spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="providers.length === 0" class="empty-card">
      <div class="empty-icon">
        <el-icon :size="32"><Connection /></el-icon>
      </div>
      <div class="empty-text">暂无平台配置</div>
      <p class="empty-hint">添加 GitLab/GitHub/Gitea 平台集成，用于 CR 同步和 Webhook 事件</p>
      <button class="action-pill action-pill--primary" @click="openAddDialog">
        <el-icon><Plus /></el-icon> 添加平台
      </button>
    </div>

    <div v-else class="provider-list">
      <div v-for="p in providers" :key="p.id" class="provider-card">
        <div class="provider-header">
          <div class="provider-icon" :style="{ background: platformMeta(p.platform).iconBg }">
            <el-icon :size="20" :style="{ color: platformMeta(p.platform).iconColor }"><Connection /></el-icon>
          </div>
          <div class="provider-info">
            <h3>{{ p.name }}</h3>
            <span class="platform-type-badge" :style="{ background: platformMeta(p.platform).badgeBg, color: platformMeta(p.platform).iconColor }">{{ platformMeta(p.platform).label }}</span>
          </div>
          <div class="provider-actions">
            <button class="act-btn act-btn--outline" @click="handleTest(p)" :disabled="testingId === p.id">
              <el-icon><Connection /></el-icon> {{ testingId === p.id ? '测试中...' : '测试连接' }}
            </button>
            <button class="act-btn act-btn--outline" @click="openEditDialog(p)">
              <el-icon><Edit /></el-icon> 编辑
            </button>
            <button class="act-btn act-btn--danger" @click="handleDelete(p)">
              <el-icon><Delete /></el-icon> 删除
            </button>
          </div>
        </div>
        <div class="provider-meta">
          <div class="meta-item">
            <span class="meta-label">API 地址</span>
            <span class="meta-value mono">{{ p.base_url }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">Webhook</span>
            <span class="meta-value" :style="{ color: p.has_webhook_secret ? '#10B981' : '#F59E0B' }">{{ p.has_webhook_secret ? '已配置' : '未配置' }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">TLS 验证</span>
            <span class="meta-value" :style="{ color: p.skip_tls ? '#F59E0B' : '#10B981' }">{{ p.skip_tls ? '已跳过' : '已启用' }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">创建时间</span>
            <span class="meta-value">{{ formatDate(p.created_at) }}</span>
          </div>
        </div>
      </div>
    </div>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑平台' : '添加平台'" width="520px" destroy-on-close>
      <div class="dialog-form">
        <div class="form-field">
          <label class="field-label">名称</label>
          <input v-model="form.name" placeholder="例如: 公司 GitLab" class="field-input" />
        </div>
        <div class="form-field">
          <label class="field-label">平台类型</label>
          <div class="type-selector">
            <button v-for="t in platformTypes" :key="t.key" class="type-btn" :class="{ active: form.platform === t.key }" @click="form.platform = t.key">
              {{ t.label }}
            </button>
          </div>
        </div>
        <div class="form-field">
          <label class="field-label">API 地址</label>
          <input v-model="form.base_url" placeholder="https://gitlab.com/api/v4" class="field-input" />
        </div>
        <div class="form-field">
          <label class="field-label">凭证</label>
          <el-select v-model="form.credential_id" placeholder="选择凭证" style="width: 100%">
            <el-option v-for="c in credentials" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </div>
        <div class="form-field">
          <label class="field-label">Webhook Secret（可选）</label>
          <input v-model="form.webhook_secret" placeholder="用于验证 Webhook 请求" class="field-input" />
        </div>
        <div class="form-field">
          <label class="field-label">TLS 证书验证</label>
          <div class="toggle-row">
            <span class="toggle-hint">{{ form.skip_tls ? '已跳过证书验证（适用于自签名证书）' : '已启用证书验证（推荐）' }}</span>
            <button class="toggle-btn" :class="{ active: form.skip_tls }" @click="form.skip_tls = !form.skip_tls">
              <span class="toggle-dot" :class="{ right: form.skip_tls }"></span>
            </button>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">{{ editingId ? '保存' : '创建' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Plus, Connection, Edit, Delete } from '@element-plus/icons-vue'
import {
  createProvider, updateProvider, deleteProvider, testProvider,
} from '@/api/modules/provider'
import type { ProviderConfigDTO } from '@/api/modules/provider'
import { listCredentials } from '@/api/modules/credential'
import type { CredentialDTO } from '@/types/credential'
import { useProviderStore } from '@/stores/useProviderStore'

const providerStore = useProviderStore()

const loading = ref(false)
const saving = ref(false)
const testingId = ref<number | null>(null)
const providers = computed(() => providerStore.providers)
const credentials = ref<CredentialDTO[]>([])

const showDialog = ref(false)
const editingId = ref<number | null>(null)
const form = ref({
  name: '',
  platform: 'gitlab',
  base_url: '',
  credential_id: undefined as number | undefined,
  webhook_secret: '',
  skip_tls: false,
})

const platformTypes = [
  { key: 'gitlab', label: 'GitLab' },
  { key: 'github', label: 'GitHub' },
  { key: 'gitea', label: 'Gitea' },
]

const PLATFORM_META: Record<string, { label: string; iconBg: string; iconColor: string; badgeBg: string }> = {
  gitlab: { label: 'GitLab', iconBg: '#FFF4E6', iconColor: '#FC6D26', badgeBg: '#FFF4E6' },
  github: { label: 'GitHub', iconBg: '#F3F4F6', iconColor: '#24292F', badgeBg: '#F3F4F6' },
  gitea: { label: 'Gitea', iconBg: '#ECFDF5', iconColor: '#609926', badgeBg: '#ECFDF5' },
}

function platformMeta(p: string) {
  return PLATFORM_META[p] || { label: p, iconBg: '#F3F4F6', iconColor: '#6B7280', badgeBg: '#F3F4F6' }
}

function formatDate(t: string) {
  if (!t) return '-'
  const d = new Date(t)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

async function loadProviders() {
  loading.value = true
  try { await providerStore.fetchProviders(true) } catch { /* handled by store */ }
  finally { loading.value = false }
}

async function loadCredentials() {
  try { credentials.value = (await listCredentials()) || [] } catch { credentials.value = [] }
}

function openAddDialog() {
  editingId.value = null
  form.value = { name: '', platform: 'gitlab', base_url: '', credential_id: undefined, webhook_secret: '', skip_tls: false }
  showDialog.value = true
}

function openEditDialog(p: ProviderConfigDTO) {
  editingId.value = p.id
  form.value = { name: p.name, platform: p.platform, base_url: p.base_url, credential_id: p.credential_id, webhook_secret: '', skip_tls: p.skip_tls || false }
  showDialog.value = true
}

async function handleSave() {
  if (!form.value.name || !form.value.base_url) {
    ElMessage.warning('请填写名称和 API 地址')
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await updateProvider(editingId.value, {
        name: form.value.name,
        base_url: form.value.base_url,
        credential_id: form.value.credential_id,
        webhook_secret: form.value.webhook_secret || undefined,
        skip_tls: form.value.skip_tls,
      })
      ElMessage.success('保存成功')
    } else {
      await createProvider({
        name: form.value.name,
        platform: form.value.platform,
        base_url: form.value.base_url,
        credential_id: form.value.credential_id!,
        webhook_secret: form.value.webhook_secret || undefined,
        skip_tls: form.value.skip_tls,
      })
      ElMessage.success('创建成功')
    }
    showDialog.value = false
    providerStore.invalidate()
    loadProviders()
  } catch (e: any) {
    ElMessage.error((editingId.value ? '保存' : '创建') + '失败: ' + (e?.message || ''))
  } finally {
    saving.value = false
  }
}

async function handleDelete(p: ProviderConfigDTO) {
  try {
    await ElMessageBox.confirm(`确定删除平台「${p.name}」？`, '确认删除', { type: 'warning' })
    await deleteProvider(p.id)
    ElMessage.success('已删除')
    providerStore.invalidate()
    loadProviders()
  } catch {}
}

async function handleTest(p: ProviderConfigDTO) {
  testingId.value = p.id
  try {
    const res = await testProvider(p.id)
    if (res.connected) {
      ElMessage.success(`连接成功！平台: ${res.platform}，用户: ${res.user_name}`)
    } else {
      ElMessage.warning('连接失败: ' + (res.message || '未知错误'))
    }
  } catch (e: any) {
    ElMessage.error('测试失败: ' + (e?.message || ''))
  } finally {
    testingId.value = null
  }
}

onMounted(() => {
  loadProviders()
  loadCredentials()
})
</script>

<style scoped>
.platform-config-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  background: var(--bg-color-page, #fff);
  color: var(--text-color-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.back-btn:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.header-actions {
  display: flex;
  gap: 10px;
}

.action-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.action-pill--primary {
  background: var(--accent-primary, #6366F1);
  color: #fff;
}

.action-pill--primary:hover {
  background: #4F46E5;
}

.provider-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.provider-card {
  border-radius: 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.provider-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.provider-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  flex-shrink: 0;
}

.provider-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.provider-info h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.platform-type-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.provider-actions {
  display: flex;
  gap: 8px;
}

.act-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.act-btn--outline {
  border: 1px solid var(--border-color, #e5e7eb);
  background: transparent;
  color: var(--text-color-secondary);
}

.act-btn--outline:hover:not(:disabled) {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.act-btn--danger {
  border: 1px solid #FCA5A5;
  background: transparent;
  color: #EF4444;
}

.act-btn--danger:hover {
  background: #FEF2F2;
}

.act-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.provider-meta {
  display: flex;
  gap: 24px;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.meta-label {
  font-size: 11px;
  color: var(--text-color-placeholder);
}

.meta-value {
  font-size: 13px;
  color: var(--text-color-primary);
}

.meta-value.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}

.empty-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  gap: 12px;
  border-radius: 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
}

.empty-icon {
  color: var(--text-color-placeholder);
}

.empty-text {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.empty-hint {
  margin: 0;
  font-size: 13px;
  color: var(--text-color-secondary);
}

.dialog-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.field-label {
  font-size: 12px;
  color: var(--text-color-secondary);
  font-weight: 500;
}

.field-input {
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page, #fff);
  outline: none;
  transition: border-color 0.2s;
  width: 100%;
  box-sizing: border-box;
}

.field-input:focus {
  border-color: var(--accent-primary, #6366F1);
}

.type-selector {
  display: flex;
  gap: 8px;
}

.type-btn {
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.type-btn.active {
  background: var(--accent-primary, #6366F1);
  border-color: var(--accent-primary, #6366F1);
  color: #fff;
}

.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
}

.toggle-hint {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.toggle-btn {
  position: relative;
  width: 36px;
  height: 20px;
  border-radius: 10px;
  border: none;
  background: #E2E8F0;
  cursor: pointer;
  transition: background 0.2s;
  padding: 0;
  flex-shrink: 0;
}

.toggle-btn.active {
  background: var(--accent-primary, #6366F1);
}

.toggle-dot {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #fff;
  transition: left 0.2s;
}

.toggle-dot.right {
  left: 18px;
}
</style>
