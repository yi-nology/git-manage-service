<template>
  <div class="repo-detail-page" v-loading="loading">
    <div class="page-header-wrap">
      <PageHeader :title="repo?.name || '仓库详情'" show-back back-route="/local-repos">
        <template #title-suffix>
          <StatusBadge v-if="currentVersion" variant="success" :text="currentVersion" :show-dot="false" />
        </template>
         <template #actions>
           <ActionPill variant="ai" :icon="MagicStick" @click="showAIPanel = !showAIPanel">AI 助手</ActionPill>
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

        <div v-if="loadedTabs.branches" v-show="activeTab === 'branches'">
          <BranchOverviewPanel :repo-key="repoKey" />
        </div>

        <div v-if="loadedTabs.spec" v-show="activeTab === 'spec'" class="spec-full-area">
          <SpecEditor ref="specEditorRef" :repo-key="repoKey" />
        </div>

        <div v-if="loadedTabs.stats" v-show="activeTab === 'stats'">
          <RepoStatsTab
            :repo-key="repoKey"
            :stats-branches="statsBranches"
            :stats-authors="statsAuthors"
            :repo-name="repo?.name || ''"
          />
        </div>

        <div v-if="loadedTabs.lines" v-show="activeTab === 'lines'">
          <RepoLineStatsTab
            :repo-key="repoKey"
            :stats-branches="statsBranches"
            :stats-authors="statsAuthors"
            :repo-name="repo?.name || ''"
          />
        </div>

        <div v-if="loadedTabs.versions" v-show="activeTab === 'versions'">
          <RepoVersionsTab
            :repo-key="repoKey"
            :remote-names="remoteNames"
            :version-list="versionList"
            :versions-loading="versionsLoading"
            @reload="loadVersions"
            @version-changed="handleVersionChanged"
          />
        </div>

        <div v-if="loadedTabs.files" v-show="activeTab === 'files'" style="height: 100%; min-height: 600px;">
          <FileExplorer :repo-key="repoKey" />
        </div>

        <div v-if="loadedTabs.commits" v-show="activeTab === 'commits'">
          <CommitSearch :repo-key="repoKey" :branches="allRefs" :authors="statsAuthors" />
        </div>

        <div v-if="loadedTabs.stash" v-show="activeTab === 'stash'">
          <StashManager :repo-key="repoKey" />
        </div>

        <div v-if="loadedTabs.submodules" v-show="activeTab === 'submodules'">
          <SubmoduleManager :repo-key="repoKey" />
        </div>

        <div v-if="loadedTabs.patches" v-show="activeTab === 'patches'">
          <PatchManager :repo-key="repoKey" />
        </div>

        <div v-if="loadedTabs.sync" v-show="activeTab === 'sync'">
          <SyncConfigPanel :repo-key="repoKey" />
        </div>

        <div v-if="loadedTabs.slim" v-show="activeTab === 'slim'">
          <SlimManager :repo-key="repoKey" />
        </div>

        <div v-if="loadedTabs.author" v-show="activeTab === 'author'">
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

     <AIPanel
       ref="aiPanelRef"
       v-if="showAIPanel"
       title="AI 仓库助手"
        v-model:visible="showAIPanel"
       empty-hint="输入问题，AI 将分析仓库状态并给出建议"
       :ai-loading="aiLoading"
       @send="handleAISummary"
       @close="showAIPanel = false"
     />
   </div>
 </template>

<script setup lang="ts">
import { ref, computed, reactive, defineAsyncComponent, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Edit, Search, InfoFilled, Document, DataAnalysis, Files, Timer, Folder, Box, Link, Share, Operation, User, DocumentCopy, MagicStick } from '@element-plus/icons-vue'
import { getRepoDetail, scanRepo } from '@/api/modules/repo'
 import { getStatsBranches, getStatsAuthors, getStatsCommits } from '@/api/modules/stats'
import { getVersionList, getCurrentVersion } from '@/api/modules/version'
import type { VersionTag } from '@/api/modules/version'
 import type { RepoDTO, ScanResult } from '@/types/repo'
 import { formatDate } from '@/utils/format'
 import { getProvider } from '@/api/modules/provider'
 import type { ProviderConfigDTO } from '@/api/modules/provider'
 import { listBindings, deleteBinding, setPrimaryBinding, registerBindingWebhook, deleteBindingWebhook } from '@/api/modules/binding'
 import type { RepoProviderBindingDTO } from '@/types/binding'
 import { useProviderStore } from '@/stores/useProviderStore'
 import { aiApi } from '@/api/modules/ai'
import { getWorkspaceStatus } from '@/api/modules/workspace'
import { getSyncHistory } from '@/api/modules/sync'

import PageHeader from '@/components/common/PageHeader.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'
import AIPanel from '@/components/ai/AIPanel.vue'

const FileExplorer = defineAsyncComponent(() => import('@/components/repo/FileExplorer.vue'))
const BranchOverviewPanel = defineAsyncComponent(() => import('@/components/branch/BranchOverviewPanel.vue'))
const CommitSearch = defineAsyncComponent(() => import('@/components/repo/CommitSearch.vue'))
const StashManager = defineAsyncComponent(() => import('@/components/repo/StashManager.vue'))
const SubmoduleManager = defineAsyncComponent(() => import('@/components/repo/SubmoduleManager.vue'))
const PatchManager = defineAsyncComponent(() => import('@/components/patch/PatchManager.vue'))
const SyncConfigPanel = defineAsyncComponent(() => import('@/components/sync/SyncConfigPanel.vue'))
const SpecEditor = defineAsyncComponent(() => import('@/components/spec/SpecEditor.vue'))
const SlimManager = defineAsyncComponent(() => import('@/components/repo/SlimManager.vue'))
const AuthorFix = defineAsyncComponent(() => import('@/components/repo/AuthorFix.vue'))
const RepoStatsTab = defineAsyncComponent(() => import('@/components/repo/RepoStatsTab.vue'))
const RepoLineStatsTab = defineAsyncComponent(() => import('@/components/repo/RepoLineStatsTab.vue'))
const RepoVersionsTab = defineAsyncComponent(() => import('@/components/repo/RepoVersionsTab.vue'))
const RepoEditDialog = defineAsyncComponent(() => import('@/components/repo/RepoEditDialog.vue'))
const BindingPanel = defineAsyncComponent(() => import('@/components/binding/BindingPanel.vue'))
const BindingDialog = defineAsyncComponent(() => import('@/components/binding/BindingDialog.vue'))

const providerStore = useProviderStore()

const route = useRoute()
const router = useRouter()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const repo = ref<RepoDTO | null>(null)
const scanData = ref<ScanResult | null>(null)
const activeTab = ref('info')
const loadedTabs = reactive<Record<string, boolean>>({
  info: true,
})
const currentVersion = ref('')
const providerInfo = ref<ProviderConfigDTO | null>(null)
const showBindingDialog = ref(false)
const availableProviders = ref<ProviderConfigDTO[]>([])
 const bindings = ref<RepoProviderBindingDTO[]>([])
 const specEditorRef = ref<{ refresh: () => void; clearEditor: () => void } | null>(null)
 const showAIPanel = ref(false)
 const aiLoading = ref(false)
const aiPanelRef = ref<{
  addResponse: (content: string) => void
  setDraft: (content: string, summary?: string, stats?: { added?: number; removed?: number }) => void
} | null>(null)

const sidebarItems = [
  { key: 'info', label: '基本信息', icon: InfoFilled },
  { key: 'branches', label: '分支管理', icon: Share },
  { key: 'spec', label: 'Spec 编辑器', icon: Document },
  { key: 'stats', label: 'Git 有效提交度量', icon: DataAnalysis },
  { key: 'lines', label: '真实工程代码度量', icon: Files },
  { key: 'versions', label: '版本历史', icon: Timer },
  { key: 'files', label: '文件', icon: Folder },
  { key: 'commits', label: 'Commit 搜索', icon: Search },
  { key: 'stash', label: 'Stash 管理', icon: Box },
  { key: 'submodules', label: 'Submodule', icon: Link },
  { key: 'patches', label: 'Patch 管理', icon: DocumentCopy },
  { key: 'sync', label: '同步任务', icon: Refresh },
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
  loadedTabs[activeTab.value] = true
  loading.value = true
  try {
    repo.value = await getRepoDetail(repoKey)
    if (repo.value?.path) {
      try {
        scanData.value = await scanRepo(repo.value.path)
        remoteNames.value = (scanData.value?.remotes || []).map((r: { name: string }) => r.name)
      } catch (_e) { /* ignore */ }
    }
    try {
      statsBranches.value = (await getStatsBranches(repoKey)) || []
    } catch (_e) { statsBranches.value = [] }
    try { statsAuthors.value = (await getStatsAuthors(repoKey)) || [] } catch (_e) { statsAuthors.value = [] }
    try { currentVersion.value = (await getCurrentVersion(repoKey)) || '' } catch (_e) { /* ignore */ }
    try { versionList.value = (await getVersionList(repoKey)) || [] } catch (_e) { versionList.value = [] }
    loadProviderInfo()
    loadBindings()
  } finally {
    loading.value = false
  }
})

watch(activeTab, (val, oldVal) => {
  const wasLoaded = !!loadedTabs[val]
  loadedTabs[val] = true

  if (val === 'versions' && (versionList.value || []).length === 0) {
    loadVersions()
  }
  if (val === 'spec' && oldVal && oldVal !== 'spec' && wasLoaded && specEditorRef.value) {
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
  } catch (_e) { /* ignore */ }
  finally {
    versionsLoading.value = false
  }
}

async function handleVersionChanged() {
  try { currentVersion.value = await getCurrentVersion(repoKey) || '' } catch (_e) { /* ignore */ }
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
  } catch (_e) { providerInfo.value = null }
}

async function loadBindings() {
  try {
    const result = await listBindings({ repo_key: repoKey })
    bindings.value = result || []
  } catch (_e) { bindings.value = [] }
}

async function openBindingDialog() {
  try {
    await providerStore.fetchProviders()
    availableProviders.value = providerStore.providers
  } catch (_e) { availableProviders.value = [] }
  showBindingDialog.value = true
}

async function handleDeleteBinding(id: number) {
  try {
    await ElMessageBox.confirm('确认取消此关联？', '取消关联', { type: 'warning' })
    await deleteBinding(id, true)
    ElMessage.success('关联已取消')
    loadBindings()
  } catch (_e) {}
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
    } catch (_e) { /* ignore */ }
  }
}

function handleNavSelect(key: string) {
  const item = sidebarItems.find(i => i.key === key)
  if (item && (item as any).route) {
    router.push(`/local-repos/${repoKey}/${key}`)
  } else {
    activeTab.value = key
    router.replace({ query: { ...route.query, tab: key } })
   }
 }

async function handleAISummary(message: string) {
   aiLoading.value = true
   try {
     let workspaceStatus: any = null
     try {
       workspaceStatus = await getWorkspaceStatus(repoKey)
     } catch (_e) { /* ignore */ }

     let failedRuns: any[] = []
     try {
       const syncHistory = await getSyncHistory(repoKey)
       failedRuns = (syncHistory || []).filter((r: any) => r.status === 'failed').slice(0, 5)
     } catch (_e) { /* ignore */ }

     const issues: string[] = []
     if (workspaceStatus?.branch) {
       issues.push(`当前分支: ${workspaceStatus.branch}`)
     }
     if (workspaceStatus && workspaceStatus.behind > 0) {
       issues.push(`当前分支落后远端 ${workspaceStatus.behind} 个提交`)
     }
     if (workspaceStatus && workspaceStatus.ahead > 0) {
       issues.push(`当前分支有 ${workspaceStatus.ahead} 个未推送提交`)
     }
     if (failedRuns.length > 0) {
       issues.push(`最近有 ${failedRuns.length} 次同步失败`)
     }

     const pendingChanges = (workspaceStatus?.staged?.length || 0) + (workspaceStatus?.unstaged?.length || 0) + (workspaceStatus?.untracked?.length || 0)

     let commitCount = 0
     try {
        const commitStats = await getStatsCommits(repoKey) as unknown as any[]
       commitCount = commitStats?.length || versionList.value?.length || 0
     } catch (_e) { /* ignore */ }

     const response = await aiApi.generateRepoSummary({
       repoKey,
       status: {
         name: repo.value?.name || '',
         currentBranch: workspaceStatus?.branch || '',
         defaultBranch: workspaceStatus?.branch || '',
         branchCount: statsBranches.value?.length || 0,
         tagCount: versionList.value?.length || 0,
         commitCount,
         stagedCount: workspaceStatus?.staged?.length || 0,
         unstagedCount: workspaceStatus?.unstaged?.length || 0,
         untrackedCount: workspaceStatus?.untracked?.length || 0,
         conflictedCount: workspaceStatus?.conflicted?.length || 0,
         ahead: workspaceStatus?.ahead || 0,
         behind: workspaceStatus?.behind || 0,
         isClean: workspaceStatus?.isClean ?? true,
         isMerging: workspaceStatus?.isMerging ?? false,
         isRebasing: workspaceStatus?.isRebasing ?? false,
         remoteCount: remoteNames.value.length,
         hasRecentSyncFailure: failedRuns.length > 0,
         recentFailureCount: failedRuns.length,
       },
       issues,
       pendingChanges,
       userInstruction: message,
     })

     aiPanelRef.value?.addResponse(
       `## 仓库健康分析\n\n${response.summary}\n\n**风险等级：** ${response.riskLevel || 'unknown'}\n\n**建议操作：**\n${(response.suggestions || []).map((s: string) => `- ${s}`).join('\n')}`
     )
   } catch (e) {
     aiPanelRef.value?.addResponse('AI 分析失败，请稍后重试。')
     ElMessage.error('AI 分析失败，请稍后重试')
   } finally {
     aiLoading.value = false
   }
 }
 </script>

<style scoped>
.repo-detail-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header-wrap {
  padding: 0 0 4px;
  border-bottom: 1px solid var(--border-color);
}

.info-card {
  border-radius: var(--border-radius-md);
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 0;
  box-shadow: var(--box-shadow-sm);
}

.info-top-row {
  display: flex;
  gap: 0;
}

.info-left-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding-right: 20px;
}

.info-v-divider {
  width: 1px;
  background: var(--border-color);
  align-self: stretch;
}

.info-right-col {
  width: 340px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-left: 20px;
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
  min-width: 0;
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
  min-width: 0;
  overflow-wrap: anywhere;
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
  border-radius: var(--border-radius-sm);
  background: var(--surface-card);
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
  gap: 16px;
  padding: 0;
  min-height: 0;
}

.left-nav {
  width: 200px;
  flex-shrink: 0;
}

.sidebar-card {
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: var(--surface-sidebar);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  padding: 8px;
  height: calc(100vh - 156px);
  overflow-y: auto;
  position: sticky;
  top: calc(var(--header-height) + 16px);
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 10px;
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  color: var(--text-color-primary);
  font-size: var(--font-size-sm);
  min-height: 34px;
  white-space: nowrap;
}

.sidebar-item:hover {
  background: var(--bg-color-page);
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
  min-width: 0;
  min-height: calc(100vh - 156px);
}

.spec-full-area {
  height: calc(100vh - 156px);
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
  }

  .left-nav {
    width: 100%;
  }

  .sidebar-card {
    height: auto;
    max-height: 300px;
    flex-direction: row;
    flex-wrap: wrap;
    position: static;
  }

  .sidebar-item {
    flex-shrink: 0;
  }

  .content-area {
    min-height: auto;
  }

  .info-top-row,
  .info-row {
    flex-direction: column;
  }

  .info-left-col,
  .info-right-col {
    width: auto;
    padding: 0;
  }

  .info-v-divider {
    width: auto;
    height: 1px;
    margin: 16px 0;
  }
}
</style>
