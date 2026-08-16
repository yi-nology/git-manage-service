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
      <SidebarNav :items="sidebarItems" :active-tab="activeTab" @select="handleNavSelect" />

      <div class="content-area">
        <div v-show="activeTab === 'info'">
          <RepoInfoTab
            v-if="repo"
            :repo="repo"
            :current-version="currentVersion"
            :scan-data="scanData"
            :bindings="bindings"
            @open-edit="showEditDialog = true"
            @open-binding-dialog="openBindingDialog"
            @delete-binding="handleDeleteBinding"
            @set-primary-binding="handleSetPrimaryBinding"
            @register-webhook="handleRegisterWebhook"
            @delete-webhook="handleDeleteWebhook"
          />
        </div>

        <BindingDialog
          v-model:visible="showBindingDialog"
          :repo-key="repo_key"
          :providers="availableProviders"
          @created="loadBindings"
        />

        <div v-if="loadedTabs.branches" v-show="activeTab === 'branches'">
          <BranchOverviewPanel :repo-key="repo_key" />
        </div>

        <div v-if="loadedTabs.spec" v-show="activeTab === 'spec'" class="spec-full-area">
          <SpecEditor ref="specEditorRef" :repo-key="repo_key" />
        </div>

        <div v-if="loadedTabs.stats" v-show="activeTab === 'stats'">
          <RepoStatsTab
            :repo-key="repo_key"
            :stats-branches="statsBranches"
            :stats-authors="statsAuthors"
            :repo-name="repo?.name || ''"
          />
        </div>

        <div v-if="loadedTabs.lines" v-show="activeTab === 'lines'">
          <RepoLineStatsTab
            :repo-key="repo_key"
            :stats-branches="statsBranches"
            :stats-authors="statsAuthors"
            :repo-name="repo?.name || ''"
          />
        </div>

        <div v-if="loadedTabs.versions" v-show="activeTab === 'versions'">
          <RepoVersionsTab
            :repo-key="repo_key"
            :remote-names="remoteNames"
            :version-list="versionList"
            :versions-loading="versionsLoading"
            @reload="loadVersions"
            @version-changed="handleVersionChanged"
          />
        </div>

        <div v-if="loadedTabs.files" v-show="activeTab === 'files'" style="height: 100%; min-height: 600px;">
          <FileExplorer :repo-key="repo_key" />
        </div>

        <div v-if="loadedTabs.commits" v-show="activeTab === 'commits'">
          <CommitSearch :repo-key="repo_key" :branches="allRefs" :authors="statsAuthors" />
        </div>

        <div v-if="loadedTabs.stash" v-show="activeTab === 'stash'">
          <StashManager :repo-key="repo_key" />
        </div>

        <div v-if="loadedTabs.submodules" v-show="activeTab === 'submodules'">
          <SubmoduleManager :repo-key="repo_key" />
        </div>

        <div v-if="loadedTabs.patches" v-show="activeTab === 'patches'">
          <PatchManager :repo-key="repo_key" />
        </div>

        <div v-if="loadedTabs.sync" v-show="activeTab === 'sync'">
          <SyncConfigPanel :repo-key="repo_key" />
        </div>

        <div v-if="loadedTabs.slim" v-show="activeTab === 'slim'">
          <SlimManager :repo-key="repo_key" />
        </div>

        <div v-if="loadedTabs.author" v-show="activeTab === 'author'">
          <AuthorFix :repo-key="repo_key" :remotes="remoteNames" />
        </div>
      </div>
    </div>

     <RepoEditDialog
       v-model:visible="showEditDialog"
       :repo="repo"
       :repo-key="repo_key"
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
import AIPanel from '@/components/ai/AIPanel.vue'
import SidebarNav from './detail/SidebarNav.vue'
import RepoInfoTab from './detail/RepoInfoTab.vue'

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
const BindingDialog = defineAsyncComponent(() => import('@/components/binding/BindingDialog.vue'))

const providerStore = useProviderStore()

const route = useRoute()
const router = useRouter()
const repo_key = route.params.repo_key as string

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
    repo.value = await getRepoDetail(repo_key)
    if (repo.value?.path) {
      try {
        scanData.value = await scanRepo(repo.value.path)
        remoteNames.value = (scanData.value?.remotes || []).map((r: { name: string }) => r.name)
      } catch (_e) { /* ignore */ }
    }
    try {
      statsBranches.value = (await getStatsBranches(repo_key)) || []
    } catch (_e) { statsBranches.value = [] }
    try { statsAuthors.value = (await getStatsAuthors(repo_key)) || [] } catch (_e) { statsAuthors.value = [] }
    try { currentVersion.value = (await getCurrentVersion(repo_key)) || '' } catch (_e) { /* ignore */ }
    try { versionList.value = (await getVersionList(repo_key)) || [] } catch (_e) { versionList.value = [] }
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
    versionList.value = (await getVersionList(repo_key)) || []
  } catch (_e) { /* ignore */ }
  finally {
    versionsLoading.value = false
  }
}

async function handleVersionChanged() {
  try { currentVersion.value = await getCurrentVersion(repo_key) || '' } catch (_e) { /* ignore */ }
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
    const result = await listBindings({ repo_key: repo_key })
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
  repo.value = await getRepoDetail(repo_key)
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
    router.push(`/local-repos/${repo_key}/${key}`)
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
       workspaceStatus = await getWorkspaceStatus(repo_key)
     } catch (_e) { /* ignore */ }

     let failed_runs: any[] = []
     try {
       const syncHistory = await getSyncHistory(repo_key)
       failed_runs = (syncHistory || []).filter((r: any) => r.status === 'failed').slice(0, 5)
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
     if (failed_runs.length > 0) {
       issues.push(`最近有 ${failed_runs.length} 次同步失败`)
     }

     const pending_changes = (workspaceStatus?.staged?.length || 0) + (workspaceStatus?.unstaged?.length || 0) + (workspaceStatus?.untracked?.length || 0)

     let commit_count = 0
     try {
        const commitStats = await getStatsCommits(repo_key) as unknown as any[]
       commit_count = commitStats?.length || versionList.value?.length || 0
     } catch (_e) { /* ignore */ }

     const response = await aiApi.generate_repo_summary({
       repo_key,
       status: {
         name: repo.value?.name || '',
         currentBranch: workspaceStatus?.branch || '',
         defaultBranch: workspaceStatus?.branch || '',
         branchCount: statsBranches.value?.length || 0,
         tagCount: versionList.value?.length || 0,
         commitCount: commit_count,
         stagedCount: workspaceStatus?.staged?.length || 0,
         unstagedCount: workspaceStatus?.unstaged?.length || 0,
         untrackedCount: workspaceStatus?.untracked?.length || 0,
         conflictedCount: workspaceStatus?.conflicted?.length || 0,
         ahead: workspaceStatus?.ahead || 0,
         behind: workspaceStatus?.behind || 0,
         isClean: workspaceStatus?.is_clean ?? true,
         isMerging: workspaceStatus?.is_merging ?? false,
         isRebasing: workspaceStatus?.is_rebasing ?? false,
         remoteCount: remoteNames.value.length,
         hasRecentSyncFailure: failed_runs.length > 0,
         recentFailureCount: failed_runs.length,
       },
       issues,
       pending_changes,
       user_instruction: message,
     })

     aiPanelRef.value?.addResponse(
       `## 仓库健康分析\n\n${response.summary}\n\n**风险等级：** ${response.risk_level || 'unknown'}\n\n**建议操作：**\n${(response.suggestions || []).map((s: string) => `- ${s}`).join('\n')}`
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

.layout-container {
  display: flex;
  gap: 16px;
  padding: 0;
  min-height: 0;
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

@media (max-width: 768px) {
  .layout-container {
    flex-direction: column;
  }

  .content-area {
    min-height: auto;
  }
}
</style>
