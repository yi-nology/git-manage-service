<template>
  <div class="remote-repos-page">
    <PageHeader title="远端仓库管理" subtitle="浏览和管理所有已关联平台的远端仓库">
      <template #actions>
        <ActionPill variant="outline" :icon="Refresh" @click="loadData" :disabled="loading">刷新</ActionPill>
      </template>
    </PageHeader>

    <EmptyState v-if="providers.length === 0 && !loading" title="暂无平台配置" description="请先在「设置 → 平台配置」中添加 GitLab/GitHub/Gitea 平台">
      <template #action>
        <ActionPill variant="primary" :icon="Setting" @click="$router.push('/settings/platforms')">前往设置</ActionPill>
      </template>
    </EmptyState>

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

      <LoadingState v-if="loading" />

      <EmptyState v-else-if="filteredRepos.length === 0" title="当前平台下暂无远端仓库" />

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
            <StatusBadge v-if="repo.linked" variant="success" text="已关联" />
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
            <ActionPill variant="primary" :icon="View" small @click="handleGoDetail(repo)">查看详情</ActionPill>
            <ActionPill variant="outline" small @click="handleGoCR(repo)">CR/MR</ActionPill>
            <ActionPill v-if="!repo.linked" variant="outline" :icon="Download" small @click="handleClone(repo)">克隆</ActionPill>
            <ActionPill v-if="!repo.linked" variant="outline" :icon="Link" small @click="handleLinkExisting(repo)">关联已有仓库</ActionPill>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Setting, FolderOpened, Download, View, Link } from '@element-plus/icons-vue'
import { listProviderRepos } from '@/api/modules/provider'
import { getRepoList } from '@/api/modules/repo'
import type { RepoDTO } from '@/types/repo'
import { createBinding } from '@/api/modules/binding'
import { useProviderStore } from '@/stores/useProviderStore'
import PageHeader from '@/components/common/PageHeader.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'

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
  gitee: { label: 'Gitee', iconBg: '#FEF2F2', iconColor: '#C71D23' },
  tencent_code: { label: '腾讯工蜂', iconBg: '#E8F5E9', iconColor: '#1B5E20' },
  forgejo: { label: 'Forgejo', iconBg: '#FFF7ED', iconColor: '#F97316' },
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

.content-area {
  display: flex;
  flex-direction: column;
  gap: 20px;
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
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.provider-tab.active {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
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
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
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

.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}

.repo-card-actions {
  display: flex;
  gap: 8px;
}
</style>
