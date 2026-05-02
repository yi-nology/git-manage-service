<template>
  <div class="repo-detail-page" v-loading="loading">
    <div class="page-header-wrap">
      <PageHeader :title="repo?.name || '仓库详情'" show-back back-route="/local-repos">
        <template #title-suffix>
          <StatusBadge v-if="currentVersion" variant="success" :text="currentVersion" :show-dot="false" />
        </template>
        <template #actions>
          <ActionPill variant="green" :icon="Share" @click="$router.push(`/local-repos/${repoKey}/branches`)">
            分支管理
          </ActionPill>
          <ActionPill variant="amber" :icon="Refresh" @click="$router.push(`/local-repos/${repoKey}/sync`)">
            同步任务
          </ActionPill>
        </template>
      </PageHeader>
    </div>

    <div class="layout-container">
      <div class="left-nav">
        <div class="sidebar-card">
          <div
            v-for="item in sidebarItems"
            :key="item.key"
            class="sidebar-item"
            :class="{ active: activeTab === item.key && !(item as any).route }"
            @click="handleNavSelect(item.key)"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </div>
        </div>
      </div>

      <div class="content-area">
        <div v-show="activeTab === 'info'">
          <div v-if="repo" class="info-card">
            <div class="info-top-row">
              <div class="info-left-col">
                <div class="info-section-header">
                  <SectionTitle title="基本信息" />
                  <ActionPill variant="outline" :icon="Edit" @click="showEditDialog = true">
                    编辑仓库
                  </ActionPill>
                </div>
                <div class="info-row">
                  <div class="info-field"><span class="info-label">名称</span><span class="info-value info-value--bold">{{ repo.name }}</span></div>
                  <div class="info-field"><span class="info-label">当前版本</span><StatusBadge v-if="currentVersion" variant="success" :text="currentVersion" :show-dot="false" /><span v-else class="info-value">-</span></div>
                </div>
                <div class="info-field"><span class="info-label">本地路径</span><span class="info-value mono">{{ repo.path }}</span></div>
                <div class="info-row">
                  <div class="info-field"><span class="info-label">Repo Key</span><span class="info-value info-value--accent">{{ repo.key }}<button class="copy-btn-sm" @click="copyKey">复制</button></span></div>
                  <div class="info-field"><span class="info-label">远程 URL</span><span class="info-value">{{ repo.remote_url || '-' }}</span></div>
                </div>
                <div class="info-row">
                  <div class="info-field"><span class="info-label">创建时间</span><span class="info-value">{{ formatDate(repo.created_at) }}</span></div>
                  <div class="info-field"><span class="info-label">更新时间</span><span class="info-value">{{ formatDate(repo.updated_at) }}</span></div>
                </div>
              </div>

              <div class="info-v-divider"></div>

              <div class="info-right-col">
                <BindingPanel
                  :bindings="bindings"
                  @add="openBindingDialog"
                  @delete="handleDeleteBinding"
                  @set-primary="handleSetPrimaryBinding"
                  @register-webhook="handleRegisterWebhook"
                  @delete-webhook="handleDeleteWebhook"
                />
              </div>
            </div>

            <template v-if="scanData">
              <div class="info-divider"></div>
              <div class="scan-section">
                <div class="info-section-header" style="margin-bottom:12px">
                  <SectionTitle title="远程配置" />
                  <span class="info-subtitle">来自 .git/config</span>
                </div>
                <div class="scan-remote-list">
                  <div v-for="r in scanData.remotes" :key="r.name" class="scan-remote-row">
                    <span class="remote-name">{{ r.name }}</span>
                    <span class="remote-url">{{ r.fetch_url }}</span>
                    <StatusBadge v-if="r.is_mirror" variant="warning" text="Mirror" :show-dot="false" />
                  </div>
                </div>
                <div v-if="scanData.branches?.length" class="tracking-tags">
                  <StatusBadge v-for="b in scanData.branches" :key="b.name" variant="info" :text="`${b.name} -> ${b.upstream_ref}`" :show-dot="false" />
                </div>
              </div>
            </template>
          </div>
        </div>

        <BindingDialog
          v-model:visible="showBindingDialog"
          :repo-key="repoKey"
          :providers="availableProviders"
          @created="loadBindings"
        />

        <div v-show="activeTab === 'spec'" class="spec-full-area">
          <SpecEditor ref="specEditorRef" :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'stats'">
          <RepoStatsTab
            :repo-key="repoKey"
            :stats-branches="statsBranches"
            :stats-authors="statsAuthors"
            :repo-name="repo?.name || ''"
          />
        </div>

        <div v-show="activeTab === 'lines'">
          <RepoLineStatsTab
            :repo-key="repoKey"
            :stats-branches="statsBranches"
            :stats-authors="statsAuthors"
            :repo-name="repo?.name || ''"
          />
        </div>

        <div v-show="activeTab === 'versions'">
          <RepoVersionsTab
            :repo-key="repoKey"
            :remote-names="remoteNames"
            :version-list="versionList"
            :versions-loading="versionsLoading"
            @reload="loadVersions"
            @version-changed="handleVersionChanged"
          />
        </div>

        <div v-show="activeTab === 'files'" style="height: 100%; min-height: 600px;">
          <FileExplorer :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'commits'">
          <CommitSearch :repo-key="repoKey" :branches="allRefs" :authors="statsAuthors" />
        </div>

        <div v-show="activeTab === 'stash'">
          <StashManager :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'submodules'">
          <SubmoduleManager :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'patches'">
          <PatchManager :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'slim'">
          <SlimManager :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'author'">
          <AuthorFix :repo-key="repoKey" :remotes="remoteNames" />
        </div>
      </div>
    </div>

    <RepoEditDialog
      v-model:visible="showEditDialog"
      :repo="repo"
      :repo-key="repoKey"
      @saved="handleEditSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Edit, Search, InfoFilled, Document, DataAnalysis, Files, Timer, Folder, Box, Link, Share, Operation, User, DocumentCopy } from '@element-plus/icons-vue'
import { getRepoDetail, scanRepo } from '@/api/modules/repo'
import { getStatsBranches, getStatsAuthors } from '@/api/modules/stats'
import { getVersionList, getCurrentVersion } from '@/api/modules/version'
import type { VersionTag } from '@/api/modules/version'
import type { RepoDTO, ScanResult } from '@/types/repo'
import { formatDate } from '@/utils/format'
import { getProvider } from '@/api/modules/provider'
import type { ProviderConfigDTO } from '@/api/modules/provider'
import { listBindings, deleteBinding, setPrimaryBinding, registerBindingWebhook, deleteBindingWebhook } from '@/api/modules/binding'
import type { RepoProviderBindingDTO } from '@/types/binding'
import { useProviderStore } from '@/stores/useProviderStore'

import FileExplorer from '@/components/repo/FileExplorer.vue'
import CommitSearch from '@/components/repo/CommitSearch.vue'
import StashManager from '@/components/repo/StashManager.vue'
import SubmoduleManager from '@/components/repo/SubmoduleManager.vue'
import PatchManager from '@/components/patch/PatchManager.vue'
import SpecEditor from '@/components/spec/SpecEditor.vue'
import SlimManager from '@/components/repo/SlimManager.vue'
import AuthorFix from '@/components/repo/AuthorFix.vue'
import RepoStatsTab from '@/components/repo/RepoStatsTab.vue'
import RepoLineStatsTab from '@/components/repo/RepoLineStatsTab.vue'
import RepoVersionsTab from '@/components/repo/RepoVersionsTab.vue'
import RepoEditDialog from '@/components/repo/RepoEditDialog.vue'
import BindingPanel from '@/components/binding/BindingPanel.vue'
import BindingDialog from '@/components/binding/BindingDialog.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const providerStore = useProviderStore()

const route = useRoute()
const router = useRouter()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const repo = ref<RepoDTO | null>(null)
const scanData = ref<ScanResult | null>(null)
const activeTab = ref('info')
const currentVersion = ref('')
const providerInfo = ref<ProviderConfigDTO | null>(null)
const showBindingDialog = ref(false)
const availableProviders = ref<ProviderConfigDTO[]>([])
const bindings = ref<RepoProviderBindingDTO[]>([])
const specEditorRef = ref<{ refresh: () => void; clearEditor: () => void } | null>(null)

const sidebarItems = [
  { key: 'info', label: '基本信息', icon: InfoFilled },
  { key: 'spec', label: 'Spec 编辑器', icon: Document },
  { key: 'stats', label: 'Git 有效提交度量', icon: DataAnalysis },
  { key: 'lines', label: '真实工程代码度量', icon: Files },
  { key: 'versions', label: '版本历史', icon: Timer },
  { key: 'files', label: '文件', icon: Folder },
  { key: 'commits', label: 'Commit 搜索', icon: Search },
  { key: 'stash', label: 'Stash 管理', icon: Box },
  { key: 'submodules', label: 'Submodule', icon: Link },
  { key: 'patches', label: 'Patch 管理', icon: DocumentCopy },
  { key: 'slim', label: '仓库瘦身', icon: Operation },
  { key: 'author', label: '作者修复', icon: User },
]

const statsBranches = ref<string[]>([])
const statsAuthors = ref<{ name: string; email: string }[]>([])

const versionList = ref<VersionTag[]>([])
const versionsLoading = ref(false)
const remoteNames = ref<string[]>([])

const allRefs = computed(() => {
  const tags = (versionList.value || []).map(v => v.name)
  return [...(statsBranches.value || []), ...tags]
})

const showEditDialog = ref(false)

onMounted(async () => {
  if (route.query.tab && typeof route.query.tab === 'string') {
    activeTab.value = route.query.tab === 'workspace' ? 'files' : route.query.tab
  }
  loading.value = true
  try {
    repo.value = await getRepoDetail(repoKey)
    if (repo.value?.path) {
      try {
        scanData.value = await scanRepo(repo.value.path)
        remoteNames.value = (scanData.value?.remotes || []).map((r: { name: string }) => r.name)
      } catch { /* ignore */ }
    }
    try {
      statsBranches.value = (await getStatsBranches(repoKey)) || []
    } catch { statsBranches.value = [] }
    try { statsAuthors.value = (await getStatsAuthors(repoKey)) || [] } catch { statsAuthors.value = [] }
    try { currentVersion.value = (await getCurrentVersion(repoKey)) || '' } catch { /* ignore */ }
    try { versionList.value = (await getVersionList(repoKey)) || [] } catch { versionList.value = [] }
    loadProviderInfo()
    loadBindings()
  } finally {
    loading.value = false
  }
})

watch(activeTab, (val) => {
  if (val === 'versions' && (versionList.value || []).length === 0) {
    loadVersions()
  }
  if (val === 'spec') {
    nextTick(() => {
      specEditorRef.value?.clearEditor()
      specEditorRef.value?.refresh()
    })
  }
})

async function loadVersions() {
  versionsLoading.value = true
  try {
    versionList.value = (await getVersionList(repoKey)) || []
  } catch { /* ignore */ }
  finally {
    versionsLoading.value = false
  }
}

async function handleVersionChanged() {
  try { currentVersion.value = await getCurrentVersion(repoKey) || '' } catch { /* ignore */ }
}

function copyKey() {
  if (repo.value?.key) {
    navigator.clipboard.writeText(repo.value.key)
    ElMessage.success('已复制 Repo Key')
  }
}

async function loadProviderInfo() {
  if (!repo.value?.provider_config_id) { providerInfo.value = null; return }
  const cached = providerStore.getProviderById(repo.value.provider_config_id)
  if (cached) { providerInfo.value = cached; return }
  try {
    providerInfo.value = await getProvider(repo.value.provider_config_id)
  } catch { providerInfo.value = null }
}

async function loadBindings() {
  try {
    const result = await listBindings({ repo_key: repoKey })
    bindings.value = result || []
  } catch { bindings.value = [] }
}

async function openBindingDialog() {
  try {
    await providerStore.fetchProviders()
    availableProviders.value = providerStore.providers
  } catch { availableProviders.value = [] }
  showBindingDialog.value = true
}

async function handleDeleteBinding(id: number) {
  try {
    await ElMessageBox.confirm('确认取消此关联？', '取消关联', { type: 'warning' })
    await deleteBinding(id, true)
    ElMessage.success('关联已取消')
    loadBindings()
  } catch {}
}

async function handleSetPrimaryBinding(id: number) {
  try {
    await setPrimaryBinding(id)
    ElMessage.success('已设为主关联')
    loadBindings()
  } catch (e: any) {
    ElMessage.error('操作失败: ' + (e?.message || ''))
  }
}

async function handleRegisterWebhook(id: number) {
  try {
    await registerBindingWebhook(id)
    ElMessage.success('Webhook 已注册')
    loadBindings()
  } catch (e: any) {
    ElMessage.error('注册失败: ' + (e?.message || ''))
  }
}

async function handleDeleteWebhook(id: number) {
  try {
    await deleteBindingWebhook(id)
    ElMessage.success('Webhook 已删除')
    loadBindings()
  } catch (e: any) {
    ElMessage.error('删除失败: ' + (e?.message || ''))
  }
}

async function handleEditSaved() {
  repo.value = await getRepoDetail(repoKey)
  if (repo.value?.path) {
    try {
      scanData.value = await scanRepo(repo.value.path)
      remoteNames.value = (scanData.value?.remotes || []).map((r: { name: string }) => r.name)
    } catch { /* ignore */ }
  }
}

function handleNavSelect(key: string) {
  const item = sidebarItems.find(i => i.key === key)
  if (item && (item as any).route) {
    router.push(`/local-repos/${repoKey}/${key}`)
  } else {
    activeTab.value = key
  }
}
</script>

<style scoped>
.page-header-wrap {
  padding: 16px 32px;
  border-bottom: 1px solid var(--border-color);
}

.info-card {
  border-radius: 12px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.info-top-row {
  display: flex;
  gap: 0;
}

.info-left-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-right: 16px;
}

.info-v-divider {
  width: 1px;
  background: var(--border-color);
  align-self: stretch;
}

.info-right-col {
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-left: 16px;
}

.info-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-subtitle {
  font-size: 12px;
  color: var(--text-color-secondary, #94A3B8);
}

.info-row {
  display: flex;
  gap: 20px;
}

.info-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.info-label {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.info-value {
  font-size: 14px;
  color: var(--text-color-primary);
}
.info-value--bold { font-weight: 500; }
.info-value--accent { color: var(--accent-primary); font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; display: flex; align-items: center; gap: 8px; }
.info-value.mono { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; font-size: 13px; }

.copy-btn-sm {
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid var(--border-color);
  background: transparent;
  font-size: 11px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}
.copy-btn-sm:hover { border-color: var(--accent-primary); color: var(--accent-primary); }

.info-divider {
  height: 1px;
  background: var(--border-color);
  margin: 16px 0;
}

.scan-section { display: flex; flex-direction: column; }

.scan-remote-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.scan-remote-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 6px;
  background: #F8F9FC;
  font-size: 13px;
}

.remote-name {
  font-weight: 600;
  color: var(--accent-primary);
  min-width: 60px;
}

.remote-url {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  color: var(--text-color-secondary);
  flex: 1;
}

.tracking-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

.layout-container {
  display: flex;
  gap: 20px;
  padding: 20px;
}

.left-nav {
  width: 220px;
  flex-shrink: 0;
}

.sidebar-card {
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  padding: 8px;
  height: calc(100vh - 180px);
  overflow-y: auto;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: var(--border-radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  color: var(--text-color-primary);
  font-size: var(--font-size-sm);
}

.sidebar-item:hover {
  background: var(--border-color-extra-light);
}

.sidebar-item.active {
  background: var(--accent-bg);
  color: var(--accent-primary);
}

.sidebar-item.active .el-icon {
  color: var(--accent-primary);
}

.sidebar-item .el-icon {
  color: var(--text-color-secondary);
  font-size: 16px;
}

.content-area {
  flex: 1;
  min-height: calc(100vh - 180px);
}

.spec-full-area {
  height: calc(100vh - 180px);
}

.spec-full-area :deep(.spec-editor-container) {
  height: 100%;
}

@media (max-width: 1024px) {
  .left-nav {
    width: 200px;
  }
}

@media (max-width: 768px) {
  .layout-container {
    flex-direction: column;
    padding: var(--spacing-md);
  }

  .left-nav {
    width: 100%;
  }

  .sidebar-card {
    height: auto;
    max-height: 300px;
    flex-direction: row;
    flex-wrap: wrap;
  }

  .content-area {
    min-height: auto;
  }
}
</style>
