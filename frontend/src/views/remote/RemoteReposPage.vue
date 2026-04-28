<template>
  <div class="remote-repos-page">
    <div class="page-header">
      <div class="header-left">
        <h2>远端仓库管理</h2>
        <p class="page-subtitle">浏览和管理所有已关联平台的远端仓库</p>
      </div>
      <div class="header-actions">
        <button class="action-pill action-pill--outline" @click="loadData" :disabled="loading">
          <el-icon><Refresh /></el-icon> 刷新
        </button>
      </div>
    </div>

    <div v-if="providers.length === 0 && !loading" class="empty-card">
      <div class="empty-icon">
        <el-icon :size="32"><Connection /></el-icon>
      </div>
      <div class="empty-text">暂无平台配置</div>
      <p class="empty-hint">请先在「设置 → 平台配置」中添加 GitLab/GitHub/Gitea 平台</p>
      <button class="action-pill action-pill--primary" @click="$router.push('/settings/platforms')">
        <el-icon><Setting /></el-icon> 前往设置
      </button>
    </div>

    <div v-else class="content-area">
      <div class="provider-tabs">
        <button
          v-for="p in providers"
          :key="p.id"
          class="provider-tab"
          :class="{ active: activeProviderId === p.id }"
          @click="activeProviderId = p.id"
        >
          <span class="tab-dot" :style="{ background: platformMeta(p.platform).iconColor }"></span>
          {{ p.name }}
          <span class="tab-badge">{{ p.platform }}</span>
        </button>
      </div>

      <div v-if="loading" class="loading-card">
        <div class="loading-spinner"></div>
        <span>加载中...</span>
      </div>

      <div v-else-if="filteredRepos.length === 0" class="empty-card small">
        <div class="empty-text">当前平台下暂无远端仓库</div>
      </div>

      <div v-else class="repo-grid">
        <div v-for="repo in filteredRepos" :key="repo.id" class="repo-card">
          <div class="repo-card-header">
            <div class="repo-icon" :style="{ background: platformMeta(activeProvider?.platform || '').iconBg }">
              <el-icon :size="18" :style="{ color: platformMeta(activeProvider?.platform || '').iconColor }">
                <FolderOpened />
              </el-icon>
            </div>
            <div class="repo-card-info">
              <h3>{{ repo.name }}</h3>
              <span class="repo-path">{{ repo.full_name }}</span>
            </div>
            <span v-if="repo.linked" class="linked-badge">
              <span class="dot-green"></span> 已关联
            </span>
          </div>
          <div class="repo-card-meta">
            <div class="meta-item">
              <span class="meta-label">URL</span>
              <span class="meta-value mono">{{ repo.ssh_url || repo.http_url || '-' }}</span>
            </div>
            <div class="meta-item" v-if="repo.default_branch">
              <span class="meta-label">默认分支</span>
              <span class="meta-value">{{ repo.default_branch }}</span>
            </div>
            <div class="meta-item" v-if="repo.updated_at">
              <span class="meta-label">更新时间</span>
              <span class="meta-value">{{ formatDate(repo.updated_at) }}</span>
            </div>
          </div>
          <div class="repo-card-actions">
            <button class="act-btn act-btn--primary" @click="handleGoDetail(repo)">
              <el-icon><View /></el-icon> 查看详情
            </button>
            <button class="act-btn act-btn--outline" @click="handleGoCR(repo)">
              CR/MR
            </button>
            <button v-if="!repo.linked" class="act-btn act-btn--outline" @click="handleClone(repo)">
              <el-icon><Download /></el-icon> 克隆
            </button>
            <button v-if="!repo.linked" class="act-btn act-btn--outline" @click="handleLinkExisting(repo)">
              <el-icon><Link /></el-icon> 关联已有仓库
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Connection, Setting, FolderOpened, Download, View, Link } from '@element-plus/icons-vue'
import { listProviderRepos } from '@/api/modules/provider'
import { getRepoList } from '@/api/modules/repo'
import type { RepoDTO } from '@/types/repo'
import { createBinding } from '@/api/modules/binding'
import { useProviderStore } from '@/stores/useProviderStore'

const providerStore = useProviderStore()

interface RemoteRepo {
  id: string
  name: string
  full_name: string
  ssh_url: string
  http_url: string
  default_branch: string
  updated_at: string
  linked: boolean
  linked_repo_key?: string
}

const router = useRouter()
const loading = ref(false)
const providers = computed(() => providerStore.providers)
const activeProviderId = ref<number | null>(null)
const remoteRepos = ref<RemoteRepo[]>([])
const localRepos = ref<RepoDTO[]>([])

const PLATFORM_META: Record<string, { label: string; iconBg: string; iconColor: string }> = {
  gitlab: { label: 'GitLab', iconBg: '#FFF4E6', iconColor: '#FC6D26' },
  github: { label: 'GitHub', iconBg: '#F3F4F6', iconColor: '#24292F' },
  gitea: { label: 'Gitea', iconBg: '#ECFDF5', iconColor: '#609926' },
}

function platformMeta(p: string) {
  return PLATFORM_META[p] || { label: p, iconBg: '#F3F4F6', iconColor: '#6B7280' }
}

function formatDate(t: string) {
  if (!t) return '-'
  const d = new Date(t)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const activeProvider = computed(() => providers.value.find(p => p.id === activeProviderId.value))
const filteredRepos = computed(() => remoteRepos.value)

async function loadData() {
  loading.value = true
  try {
    await providerStore.fetchProviders()
    const repoList = await getRepoList().catch(() => [] as RepoDTO[])
    localRepos.value = repoList || []

    if (providers.value.length > 0 && !activeProviderId.value) {
      activeProviderId.value = providers.value[0]?.id ?? null
    }
    if (activeProviderId.value) {
      await loadRemoteRepos()
    }
  } finally {
    loading.value = false
  }
}

async function loadRemoteRepos() {
  if (!activeProvider.value) return
  loading.value = true
  remoteRepos.value = []
  try {
    const res = await listProviderRepos(activeProvider.value.id, { page: 1, per_page: 100 })
    const items = res || []
    remoteRepos.value = items.map(r => ({
      id: String(r.id),
      name: r.name,
      full_name: r.full_name,
      ssh_url: r.ssh_url,
      http_url: r.clone_url,
      default_branch: r.default_branch,
      updated_at: '',
      linked: localRepos.value.some(lr =>
        lr.provider_config_id === activeProvider.value!.id &&
        lr.platform_repo === r.name
      ),
      linked_repo_key: undefined,
    }))
    const matched = localRepos.value.filter(lr =>
      lr.provider_config_id === activeProvider.value!.id
    )
    for (const m of matched) {
      const idx = remoteRepos.value.findIndex(r => r.name === m.platform_repo)
      if (idx >= 0) {
        const repo = remoteRepos.value[idx]
        if (repo) {
          repo.linked = true
          repo.linked_repo_key = m.key
        }
      }
    }
  } catch {
    remoteRepos.value = []
  } finally {
    loading.value = false
  }
}

function handleClone(repo: RemoteRepo) {
  const url = repo.ssh_url || repo.http_url
  if (url) {
    const query: Record<string, string> = { url }
    const parts = repo.full_name.split('/')
    const owner = parts.slice(0, -1).join('/')
    const name = parts[parts.length - 1] || ''
    if (activeProviderId.value) query.provider_config_id = String(activeProviderId.value)
    if (owner) query.platform_owner = owner
    if (name) query.platform_repo = name
    router.push({ path: '/local-repos/clone', query })
  } else {
    ElMessage.warning('未找到仓库 URL')
  }
}

async function handleLinkExisting(remoteRepo: RemoteRepo) {
  const parts = remoteRepo.full_name.split('/')
  const owner = parts.slice(0, -1).join('/')
  const repoName = parts[parts.length - 1] || ''
  if (!activeProviderId.value || !owner || !repoName) {
    ElMessage.warning('缺少关联所需的信息')
    return
  }
  const repos = localRepos.value
  if (repos.length === 0) {
    ElMessage.warning('没有已注册的本地仓库可关联')
    return
  }
  const choices = repos.map((r: RepoDTO) => ({ label: `${r.name} (${r.path})`, value: r.key }))
  const { ElMessageBox } = await import('element-plus')
  try {
    await ElMessageBox({
      title: '选择本地仓库',
      message: `<p style="margin-bottom:10px">将 <strong>${remoteRepo.full_name}</strong> 关联到：</p>
        <select id="link-repo-select" style="width:100%;padding:8px;border:1px solid #dcdfe6;border-radius:4px;font-size:14px">
          ${choices.map((c: { label: string; value: string }) => `<option value="${c.value}">${c.label}</option>`).join('')}
        </select>`,
      dangerouslyUseHTMLString: true,
      confirmButtonText: '关联',
      cancelButtonText: '取消',
      showCancelButton: true,
    })
    const select = document.getElementById('link-repo-select') as HTMLSelectElement
    const selectedKey = select?.value
    if (!selectedKey) { ElMessage.warning('请选择一个仓库'); return }
    await createBinding({
      repo_key: selectedKey,
      provider_config_id: activeProviderId.value,
      platform_owner: owner,
      platform_repo: repoName,
      is_primary: true,
    })
    ElMessage.success('关联成功')
    loadRemoteRepos()
  } catch {}
}

function handleGoDetail(repo: RemoteRepo) {
  router.push({ name: 'RemoteRepoDetail', params: { providerId: String(activeProviderId.value), repoOwner: repo.full_name.split('/').slice(0, -1).join('/') || '-', repoName: repo.full_name.split('/').pop() || '' } })
}

function handleGoCR(repo: RemoteRepo) {
  router.push({ name: 'RemoteRepoDetail', params: { providerId: String(activeProviderId.value), repoOwner: repo.full_name.split('/').slice(0, -1).join('/') || '-', repoName: repo.full_name.split('/').pop() || '' }, query: { tab: 'cr' } })
}

import { watch } from 'vue'
watch(activeProviderId, () => {
  if (activeProviderId.value) loadRemoteRepos()
})

onMounted(loadData)
</script>

<style scoped>
.remote-repos-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-left h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.page-subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-color-secondary);
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

.action-pill--outline {
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  color: var(--text-color-secondary);
}

.action-pill--outline:hover:not(:disabled) {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.action-pill:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.provider-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.provider-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.provider-tab.active {
  background: var(--accent-primary, #6366F1);
  border-color: var(--accent-primary, #6366F1);
  color: #fff;
}

.provider-tab.active .tab-dot {
  background: #fff !important;
}

.provider-tab.active .tab-badge {
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.tab-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.tab-badge {
  padding: 1px 6px;
  border-radius: 4px;
  background: #F3F4F6;
  font-size: 10px;
  color: var(--text-color-secondary);
}

.repo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 16px;
}

.repo-card {
  border-radius: 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.repo-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.repo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  flex-shrink: 0;
}

.repo-card-info {
  flex: 1;
  min-width: 0;
}

.repo-card-info h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.repo-path {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.linked-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border-radius: 4px;
  background: #ECFDF5;
  color: #059669;
  font-size: 11px;
  font-weight: 500;
  flex-shrink: 0;
}

.dot-green {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #10B981;
  display: inline-block;
}

.repo-card-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.meta-item {
  display: flex;
  gap: 8px;
  align-items: baseline;
}

.meta-label {
  font-size: 11px;
  color: var(--text-color-placeholder);
  flex-shrink: 0;
  width: 60px;
}

.meta-value {
  font-size: 13px;
  color: var(--text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meta-value.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}

.repo-card-actions {
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

.act-btn--primary {
  border: 1px solid var(--accent-primary, #6366F1);
  background: var(--accent-primary, #6366F1);
  color: #fff;
}

.act-btn--primary:hover {
  background: #4F46E5;
}

.act-btn--outline {
  border: 1px solid var(--border-color, #e5e7eb);
  background: transparent;
  color: var(--text-color-secondary);
}

.act-btn--outline:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
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

.empty-card.small {
  padding: 40px 0;
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

.loading-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
  gap: 12px;
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
