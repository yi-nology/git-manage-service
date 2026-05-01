<template>
  <div class="remote-repo-detail-page">
    <PageHeader showBack backRoute="/remote-repos">
      <template #title-suffix>
        <div class="repo-icon" :style="{ background: platformMeta(providerPlatform).iconBg }">
          <el-icon :size="18" :style="{ color: platformMeta(providerPlatform).iconColor }"><FolderOpened /></el-icon>
        </div>
        <div class="repo-title-info">
          <h2>{{ repoFullName }}</h2>
          <span class="platform-badge" :style="{ background: platformMeta(providerPlatform).iconBg, color: platformMeta(providerPlatform).iconColor }">{{ platformMeta(providerPlatform).label }}</span>
          <StatusBadge v-if="linkedRepoKey" variant="success" text="已关联本地" />
        </div>
      </template>
      <template #actions>
        <ActionPill v-if="!linkedRepoKey" variant="primary" :icon="Download" @click="handleClone">克隆到本地</ActionPill>
        <ActionPill v-else variant="outline" :icon="FolderOpened" @click="$router.push(`/local-repos/${linkedRepoKey}`)">查看本地仓库</ActionPill>
      </template>
    </PageHeader>

    <div v-if="!linkedRepoKey" class="unlinked-card">
      <div class="unlinked-icon"><el-icon :size="32"><Link /></el-icon></div>
      <h3>该仓库尚未关联本地</h3>
      <div class="unlinked-info">
        <div class="info-row">
          <span class="info-label">HTTP</span>
          <span class="info-value mono">{{ repoData?.clone_url || '-' }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">SSH</span>
          <span class="info-value mono">{{ repoData?.ssh_url || '-' }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">默认分支</span>
          <span class="info-value">{{ repoData?.default_branch || '-' }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">可见性</span>
          <span class="info-value">{{ repoData?.private ? '私有' : '公开' }}</span>
        </div>
      </div>
      <ActionPill variant="primary" :icon="Download" @click="handleClone" style="margin-top:16px">克隆到本地</ActionPill>
    </div>

    <div class="tab-bar">
      <button class="tab-btn" :class="{ active: activeTab === 'cr' }" @click="activeTab = 'cr'">CR / MR</button>
      <button class="tab-btn" :class="{ active: activeTab === 'codereview' }" @click="activeTab = 'codereview'">代码审查</button>
      <button class="tab-btn" :class="{ active: activeTab === 'branchrules' }" @click="activeTab = 'branchrules'">分支规则</button>
      <button class="tab-btn" :class="{ active: activeTab === 'webhooks' }" @click="activeTab = 'webhooks'">Webhook 事件</button>
      <button class="tab-btn" :class="{ active: activeTab === 'branches' }" @click="activeTab = 'branches'">远程分支</button>
    </div>

    <div v-show="activeTab === 'cr'" class="tab-content">
      <div class="content-header">
        <SectionTitle title="Change Requests" />
        <div class="content-actions">
          <ActionPill variant="primary" :icon="Plus" @click="showCreateCRDialog = true">创建 CR</ActionPill>
          <ActionPill variant="outline" :icon="Refresh" @click="handleSyncCRs" :disabled="crSyncing">{{ crSyncing ? '刷新中...' : '刷新' }}</ActionPill>
        </div>
      </div>

      <DataTable :columns="crColumns" :data="crs" :loading="crLoading" row-key="id">
        <template #cell-cr_number="{ row }">
          <span class="mono">{{ row.cr_number }}</span>
        </template>
        <template #cell-source_branch="{ row }">
          <span class="mono">{{ row.source_branch }}</span>
        </template>
        <template #cell-target_branch="{ row }">
          <span class="mono">{{ row.target_branch }}</span>
        </template>
        <template #cell-state="{ row }">
          <StatusBadge :variant="crStateVariant(row.state)" :text="crStateLabel(row.state)" :showDot="false" />
        </template>
        <template #empty>
          <EmptyState title="暂无 CR" />
        </template>
        <template #row-actions="{ row }">
          <ActionPill v-if="row.state === 'opened'" variant="primary" small @click="handleMergeCR(row)">合并</ActionPill>
          <ActionPill v-if="row.state === 'opened'" variant="danger" small @click="handleCloseCR(row)">关闭</ActionPill>
        </template>
      </DataTable>
    </div>

    <div v-show="activeTab === 'codereview'" class="tab-content">
      <div class="content-header">
        <SectionTitle title="代码审查配置" />
        <div class="content-actions">
          <ActionPill variant="outline" :icon="Refresh" @click="loadReviewConfig">刷新</ActionPill>
        </div>
      </div>

      <LoadingState v-if="crCfgLoading" />

      <div v-else class="config-panel">
        <div class="config-sidebar">
          <button class="cfg-nav-btn active">基本设置</button>
        </div>

        <div class="config-form-area">
          <div class="form-section">
            <div class="form-row">
              <div class="form-label">
                <span>启用代码审查</span>
                <span class="form-desc">开启后，MR/CR 创建时将自动进行代码审查</span>
              </div>
              <el-switch v-model="reviewCfg.enabled" />
            </div>
            <div class="form-row">
              <div class="form-label">
                <span>自动审查 MR</span>
                <span class="form-desc">MR 创建或更新时自动触发审查</span>
              </div>
              <el-switch v-model="reviewCfg.auto_review_on_mr" />
            </div>
            <div class="form-row">
              <div class="form-label">
                <span>高危阻止合并</span>
                <span class="form-desc">当审查发现高危问题时阻止合并</span>
              </div>
              <el-switch v-model="reviewCfg.block_on_high" />
            </div>
            <div class="form-row-inline">
              <div class="form-field">
                <label>LLM 提供商</label>
                <el-select v-model="reviewCfg.llm_provider" placeholder="留空使用全局默认" clearable style="width:100%">
                  <el-option v-for="p in globalProviders" :key="p.name" :label="p.name + (p.is_default ? '（默认）' : '')" :value="p.name" />
                  <template #empty><span style="padding:8px;font-size:12px;color:#999">请先在系统设置中配置 LLM 提供商</span></template>
                </el-select>
              </div>
            </div>
            <div class="form-row-inline">
              <div class="form-field">
                <label>最大文件数</label>
                <el-input-number v-model="reviewCfg.max_files" :min="1" :max="500" />
              </div>
              <div class="form-field">
                <label>最大差异行数</label>
                <el-input-number v-model="reviewCfg.max_diff_lines" :min="100" :max="50000" :step="500" />
              </div>
            </div>
          </div>

          <div v-if="reviewCfg.linked_repos && reviewCfg.linked_repos.length > 0" class="scope-card">
            <h4>生效范围</h4>
            <p class="scope-desc">以下本地仓库将使用此远端仓库的代码审查配置：</p>
            <div class="scope-repos">
              <div v-for="r in reviewCfg.linked_repos" :key="r.id" class="scope-repo-item">
                <el-icon :size="14" style="color:#6366F1"><FolderOpened /></el-icon>
                <span class="scope-repo-name">{{ r.name }}</span>
                <span class="scope-repo-key">{{ r.key }}</span>
              </div>
            </div>
          </div>
          <div v-else class="scope-card scope-card--empty">
            <p>暂无关联的本地仓库，此配置将在关联本地仓库后生效。</p>
          </div>

          <div class="form-actions">
            <ActionPill variant="outline" @click="loadReviewConfig">取消</ActionPill>
            <ActionPill variant="primary" @click="saveReviewConfig" :disabled="crCfgSaving">{{ crCfgSaving ? '保存中...' : '保存' }}</ActionPill>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'branchrules'" class="tab-content">
      <div class="content-header">
        <SectionTitle title="分支规则配置" />
        <div class="content-actions">
          <ActionPill variant="outline" :icon="Refresh" @click="loadBranchRules">刷新</ActionPill>
        </div>
      </div>

      <LoadingState v-if="brLoading" />

      <div v-else class="config-panel">
        <div class="config-sidebar">
          <button class="cfg-nav-btn active">分支规则</button>
        </div>

        <div class="config-form-area">
          <div class="form-section">
            <div class="form-row">
              <div class="form-label">
                <span>使用自定义规则</span>
                <span class="form-desc">开启后将覆盖全局分支规则，仅对此远端仓库生效</span>
              </div>
              <el-switch v-model="branchRuleCfg.use_custom_rules" />
            </div>

            <template v-if="branchRuleCfg.use_custom_rules">
              <div class="form-section" style="margin-top:12px">
                <h4 style="margin:0 0 12px;font-size:14px;color:var(--text-color-primary)">分支类型规则</h4>
                <div v-for="(rule, idx) in branchRuleCfg.rules" :key="idx" class="br-rule-card">
                  <div class="br-rule-header">
                    <span class="br-rule-prefix">{{ rule.prefix || '/' }}</span>
                    <input v-model="rule.display_name" class="br-rule-name-input" placeholder="显示名称" />
                    <ActionPill variant="danger" small @click="branchRuleCfg.rules.splice(idx, 1)">删除</ActionPill>
                  </div>
                  <div class="br-rule-grid">
                    <div class="form-field">
                      <label>前缀</label>
                      <input v-model="rule.prefix" class="field-input" placeholder="feature/" />
                    </div>
                    <div class="form-field">
                      <label>任务 ID</label>
                      <div class="switch-row">
                        <el-switch v-model="rule.require_task_id" />
                        <span class="toggle-label-sm">{{ rule.require_task_id ? '必需' : '可选' }}</span>
                      </div>
                    </div>
                    <div class="form-field">
                      <label>允许直接推送</label>
                      <div class="switch-row">
                        <el-switch v-model="rule.allow_direct_push" />
                        <span class="toggle-label-sm">{{ rule.allow_direct_push ? '允许' : '禁止' }}</span>
                      </div>
                    </div>
                    <div class="form-field">
                      <label>需要代码审查</label>
                      <div class="switch-row">
                        <el-switch v-model="rule.require_code_review" />
                        <span class="toggle-label-sm">{{ rule.require_code_review ? '必需' : '可选' }}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <ActionPill variant="outline" :icon="Plus" @click="branchRuleCfg.rules.push({ id:0, prefix:'', display_name:'', source_branches:[], target_branches:[], require_task_id:false, task_id_pattern:'', auto_delete_on_merge:false, allow_direct_push:true, require_code_review:false, sort_order:branchRuleCfg.rules.length })" style="margin-top:8px">添加规则</ActionPill>
              </div>

              <div class="form-section" style="margin-top:16px">
                <h4 style="margin:0 0 8px;font-size:14px;color:var(--text-color-primary)">保护分支</h4>
                <div class="protected-tags">
                  <div v-for="(name, idx) in branchRuleCfg.protected_branches" :key="idx" class="protected-tag-sm">
                    <span>{{ name }}</span>
                    <button class="tag-remove" @click="branchRuleCfg.protected_branches.splice(idx, 1)">&times;</button>
                  </div>
                  <span v-if="branchRuleCfg.protected_branches.length === 0" class="text-muted">暂无</span>
                </div>
              </div>
            </template>
          </div>

          <div v-if="branchRuleCfg.linked_repos && branchRuleCfg.linked_repos.length > 0" class="scope-card">
            <h4>生效范围</h4>
            <p class="scope-desc">以下本地仓库将使用此远端仓库的分支规则配置：</p>
            <div class="scope-repos">
              <div v-for="r in branchRuleCfg.linked_repos" :key="r.id" class="scope-repo-item">
                <el-icon :size="14" style="color:#6366F1"><FolderOpened /></el-icon>
                <span class="scope-repo-name">{{ r.name }}</span>
                <span class="scope-repo-key">{{ r.key }}</span>
              </div>
            </div>
          </div>

          <div class="form-actions">
            <ActionPill variant="outline" @click="loadBranchRules">取消</ActionPill>
            <ActionPill variant="primary" @click="saveBranchRules" :disabled="brSaving">{{ brSaving ? '保存中...' : '保存' }}</ActionPill>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'webhooks'" class="tab-content">
      <div class="content-header">
        <SectionTitle title="Webhook 事件" />
        <ActionPill variant="outline" :icon="Refresh" @click="loadWebhookEvents">刷新</ActionPill>
      </div>

      <DataTable :columns="whColumns" :data="webhookEvents" :loading="whLoading" row-key="id">
        <template #cell-event_type="{ row }">
          <span class="event-type-cell"><el-icon :size="14" style="color:#6366F1"><Share /></el-icon> {{ row.event_type }}</span>
        </template>
        <template #cell-event_id="{ row }">
          <span class="mono">{{ row.event_id?.substring(0, 16) }}...</span>
        </template>
        <template #cell-status="{ row }">
          <StatusBadge :variant="whStatusVariant(row.status)" :text="whStatusLabel(row.status)" :showDot="false" />
        </template>
        <template #cell-created_at="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
        <template #empty>
          <EmptyState title="暂无 Webhook 事件" />
        </template>
        <template #row-actions="{ row }">
          <ActionPill v-if="row.status === 'failed'" variant="amber" small @click="handleRetryEvent(row)">重试</ActionPill>
        </template>
      </DataTable>
    </div>

    <div v-show="activeTab === 'branches'" class="tab-content">
      <div class="content-header">
        <SectionTitle title="远程分支" />
        <div class="content-actions">
          <ActionPill variant="primary" :icon="Plus" @click="showCreateBranchDialog = true">创建分支</ActionPill>
          <ActionPill variant="outline" :icon="Refresh" @click="loadRemoteBranches">刷新</ActionPill>
        </div>
      </div>

      <DataTable :columns="branchColumns" :data="remoteBranches" :loading="rbLoading" row-key="name">
        <template #cell-name="{ row }">
          <span class="mono"><el-icon :size="14" style="color:#10B981"><Share /></el-icon> {{ row.name }}</span>
        </template>
        <template #empty>
          <EmptyState title="暂无远程分支数据" />
        </template>
        <template #row-actions="{ row }">
          <ActionPill variant="primary" small @click="handleCheckoutRemote(row.name)">检出本地</ActionPill>
          <ActionPill variant="danger" small @click="handleDeleteRemoteBranch(row.name)">删除</ActionPill>
        </template>
      </DataTable>
    </div>

    <el-dialog v-model="showCreateCRDialog" title="创建 CR / MR" width="520px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="标题"><el-input v-model="createCRForm.title" placeholder="CR 标题" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="createCRForm.description" type="textarea" :rows="3" placeholder="可选描述" /></el-form-item>
        <el-form-item label="源分支"><el-input v-model="createCRForm.source_branch" placeholder="feature-branch" /></el-form-item>
        <el-form-item label="目标分支"><el-input v-model="createCRForm.target_branch" placeholder="main" /></el-form-item>
        <el-form-item label="标签"><el-input v-model="createCRForm.labels" placeholder="逗号分隔，如: bugfix,feature" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateCRDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateCR" :loading="createCRLoading">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showCreateBranchDialog" title="创建远程分支" width="480px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="分支名"><el-input v-model="createBranchForm.branch" placeholder="如: feature/new-api" /></el-form-item>
        <el-form-item label="基于">
          <el-input v-model="createBranchForm.ref" :placeholder="repoData?.default_branch || 'main'" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateBranchDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateBranch" :loading="createBranchLoading">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { FolderOpened, Download, Link, Refresh, Share, Plus } from '@element-plus/icons-vue'
import { listProviderRepos, listRemoteBranches, createRemoteBranch, deleteRemoteBranch } from '@/api/modules/provider'
import { listRemoteCRs, createRemoteCR, mergeRemoteCR, closeRemoteCR } from '@/api/modules/cr'
import type { CRDTO } from '@/api/modules/cr'
import { listWebhookEvents, retryWebhookEvent } from '@/api/modules/webhook-event'
import type { WebhookEventDTO } from '@/api/modules/webhook-event'
import { getRepoList } from '@/api/modules/repo'
import { createBranch } from '@/api/modules/branch'
import { getRemoteRepoConfig, updateRemoteRepoConfig } from '@/api/modules/review'
import type { ReviewRepoConfigDTO } from '@/api/modules/review'
import { listLLMProviders } from '@/api/modules/llm-settings'
import type { LLMProviderDTO } from '@/api/modules/llm-settings'
import { getRemoteRepoBranchRules, updateRemoteRepoBranchRules } from '@/api/modules/branch-rule'
import type { BranchRuleDTO } from '@/api/modules/branch-rule'
import { useProviderStore } from '@/stores/useProviderStore'
import PageHeader from '@/components/common/PageHeader.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const providerStore = useProviderStore()

const route = useRoute()
const router = useRouter()
const providerId = Number(route.params.providerId)
const repoOwner = route.params.repoOwner as string
const repoName = route.params.repoName as string
const repoFullName = computed(() => `${repoOwner}/${repoName}`)

const providerPlatform = ref('')
const linkedRepoKey = ref<string | null>(null)
const activeTab = ref((route.query.tab as string) || 'cr')
const repoData = ref<{ clone_url?: string; ssh_url?: string; default_branch?: string; private?: boolean } | null>(null)

const crLoading = ref(false)
const crSyncing = ref(false)
const crs = ref<CRDTO[]>([])
const showCreateCRDialog = ref(false)
const createCRForm = ref({ title: '', description: '', source_branch: '', target_branch: '', labels: '' })
const createCRLoading = ref(false)

const whLoading = ref(false)
const webhookEvents = ref<WebhookEventDTO[]>([])

const rbLoading = ref(false)
const remoteBranches = ref<{ name: string }[]>([])
const showCreateBranchDialog = ref(false)
const createBranchForm = ref({ branch: '', ref: '' })
const createBranchLoading = ref(false)

const crCfgLoading = ref(false)
const crCfgSaving = ref(false)
const globalProviders = ref<LLMProviderDTO[]>([])
const reviewCfg = ref<ReviewRepoConfigDTO>({
  id: 0,
  provider_config_id: 0,
  platform_owner: '',
  platform_repo: '',
  enabled: true,
  block_on_high: true,
  auto_review_on_mr: true,
  llm_provider: '',
  max_files: 50,
  max_diff_lines: 3000,
  rule_overrides_json: '',
  scope_note: '',
  linked_repos: [],
})

const brLoading = ref(false)
const brSaving = ref(false)
const branchRuleCfg = ref<{
  use_custom_rules: boolean
  rules: BranchRuleDTO[]
  protected_branches: string[]
  linked_repos: { id: number; key: string; name: string }[]
}>({
  use_custom_rules: false,
  rules: [],
  protected_branches: [],
  linked_repos: [],
})

const PLATFORM_META: Record<string, { label: string; iconBg: string; iconColor: string }> = {
  gitlab: { label: 'GitLab', iconBg: '#FFF4E6', iconColor: '#FC6D26' },
  github: { label: 'GitHub', iconBg: '#F3F4F6', iconColor: '#24292F' },
  gitea: { label: 'Gitea', iconBg: '#ECFDF5', iconColor: '#609926' },
}

function platformMeta(p: string) {
  return PLATFORM_META[p] || { label: p, iconBg: '#F3F4F6', iconColor: '#6B7280' }
}

function crStateLabel(s: string) {
  if (s === 'opened') return '开启'
  if (s === 'merged') return '已合并'
  if (s === 'closed') return '已关闭'
  return s
}

function crStateVariant(s: string): 'success' | 'info' | 'danger' | 'default' {
  if (s === 'opened') return 'success'
  if (s === 'merged') return 'info'
  if (s === 'closed') return 'danger'
  return 'default'
}

function whStatusLabel(s: string) {
  if (s === 'processed') return '已处理'
  if (s === 'received') return '待处理'
  if (s === 'failed') return '失败'
  return s
}

function whStatusVariant(s: string): 'success' | 'warning' | 'danger' | 'default' {
  if (s === 'processed') return 'success'
  if (s === 'received') return 'warning'
  if (s === 'failed') return 'danger'
  return 'default'
}

function formatTime(t: string) {
  if (!t) return '-'
  const d = new Date(t)
  return `${d.getMonth() + 1}-${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const crColumns: TableColumn[] = [
  { key: 'cr_number', label: '#', width: '60px' },
  { key: 'title', label: '标题' },
  { key: 'source_branch', label: '源分支', width: '100px' },
  { key: 'target_branch', label: '目标分支', width: '100px' },
  { key: 'state', label: '状态', width: '80px' },
]

const whColumns: TableColumn[] = [
  { key: 'event_type', label: '事件类型', width: '120px' },
  { key: 'source', label: '来源', width: '80px' },
  { key: 'event_id', label: '事件 ID' },
  { key: 'status', label: '状态', width: '80px' },
  { key: 'created_at', label: '时间', width: '140px' },
]

const branchColumns: TableColumn[] = [
  { key: 'name', label: '分支名' },
]

async function loadInitial() {
  const [localRepos] = await Promise.all([
    providerStore.fetchProviders(),
    getRepoList().catch(() => []),
  ])
  const prov = providerStore.getProviderById(providerId)
  if (prov) providerPlatform.value = prov.platform

  const repos = (localRepos || []) as any[]
  const linked = repos.find((r: any) =>
    r.provider_config_id === providerId &&
    r.platform_owner === repoOwner &&
    r.platform_repo === repoName
  )
  if (linked) linkedRepoKey.value = linked.key

  const remoteRepos = await listProviderRepos(providerId, { page: 1, per_page: 100 }).catch(() => [])
  const found = (remoteRepos || []).find((r: any) => r.full_name === repoFullName.value)
  if (found) repoData.value = found

  if (linkedRepoKey.value) {
    loadCRs()
  } else {
    loadCRs()
  }
}

async function loadCRs() {
  crLoading.value = true
  try {
    const res = await listRemoteCRs({ provider_id: providerId, owner: repoOwner, repo: repoName, page: 1, per_page: 100 })
    crs.value = res?.items || []
  } catch { crs.value = [] }
  finally { crLoading.value = false }
}

async function handleSyncCRs() {
  crSyncing.value = true
  try {
    const res = await listRemoteCRs({ provider_id: providerId, owner: repoOwner, repo: repoName, page: 1, per_page: 100 })
    crs.value = res?.items || []
    ElMessage.success(`已刷新，共 ${crs.value.length} 个 CR`)
  } catch (e: any) { ElMessage.error('刷新失败: ' + (e?.message || '')) }
  finally { crSyncing.value = false }
}

async function handleMergeCR(cr: CRDTO) {
  try { await ElMessageBox.confirm(`确定合并 CR #${cr.cr_number}？`, '确认合并', { type: 'info' }) } catch { return }
  try {
    await mergeRemoteCR(providerId, repoOwner, repoName, cr.cr_number)
    ElMessage.success('合并成功')
    loadCRs()
  } catch (e: any) { ElMessage.error('合并失败: ' + (e?.message || '')) }
}

async function handleCloseCR(cr: CRDTO) {
  try { await ElMessageBox.confirm(`确定关闭 CR #${cr.cr_number}？`, '确认关闭', { type: 'warning' }) } catch { return }
  try {
    await closeRemoteCR(providerId, repoOwner, repoName, cr.cr_number)
    ElMessage.success('已关闭')
    loadCRs()
  } catch (e: any) { ElMessage.error('关闭失败: ' + (e?.message || '')) }
}

async function loadWebhookEvents() {
  whLoading.value = true
  try {
    const res = await listWebhookEvents({ page: 1, page_size: 50 })
    webhookEvents.value = res?.items || []
  } catch { webhookEvents.value = [] }
  finally { whLoading.value = false }
}

async function handleRetryEvent(ev: WebhookEventDTO) {
  try {
    await retryWebhookEvent(ev.id)
    ElMessage.success('已重试')
    loadWebhookEvents()
  } catch (e: any) { ElMessage.error('重试失败: ' + (e?.message || '')) }
}

async function loadRemoteBranches() {
  rbLoading.value = true
  remoteBranches.value = []
  try {
    const res = await listRemoteBranches({ provider_id: providerId, owner: repoOwner, repo: repoName })
    remoteBranches.value = (res || []) as any[]
  } catch { remoteBranches.value = [] }
  finally { rbLoading.value = false }
}

async function handleCheckoutRemote(branchName: string) {
  if (!linkedRepoKey.value) {
    ElMessage.warning('请先克隆到本地再检出分支')
    return
  }
  try {
    await createBranch({ repo_key: linkedRepoKey.value!, name: branchName, base_ref: `origin/${branchName}` })
    ElMessage.success(`已检出分支 ${branchName}`)
  } catch (e: any) { ElMessage.error('检出失败: ' + (e?.message || '')) }
}

async function handleDeleteRemoteBranch(branchName: string) {
  try { await ElMessageBox.confirm(`确定删除远程分支 ${branchName}？此操作不可恢复！`, '确认删除', { type: 'warning' }) } catch { return }
  try {
    await deleteRemoteBranch({ provider_id: providerId, owner: repoOwner, repo: repoName, branch: branchName })
    ElMessage.success(`已删除分支 ${branchName}`)
    loadRemoteBranches()
  } catch (e: any) { ElMessage.error('删除失败: ' + (e?.message || '')) }
}

async function handleCreateBranch() {
  const f = createBranchForm.value
  if (!f.branch) { ElMessage.warning('请输入分支名'); return }
  if (!f.ref) { ElMessage.warning('请输入基于的分支/标签'); return }
  createBranchLoading.value = true
  try {
    await createRemoteBranch({ provider_id: providerId, owner: repoOwner, repo: repoName, branch: f.branch, ref: f.ref })
    ElMessage.success(`分支 ${f.branch} 创建成功`)
    showCreateBranchDialog.value = false
    createBranchForm.value = { branch: '', ref: '' }
    loadRemoteBranches()
  } catch (e: any) { ElMessage.error('创建失败: ' + (e?.message || '')) }
  finally { createBranchLoading.value = false }
}

function handleClone() {
  const url = repoData.value?.ssh_url || repoData.value?.clone_url
  if (url) {
    const query: Record<string, string> = { url }
    if (providerId) query.provider_config_id = String(providerId)
    if (repoOwner) query.platform_owner = repoOwner
    if (repoName) query.platform_repo = repoName
    router.push({ path: '/local-repos/clone', query })
  }
}

async function handleCreateCR() {
  const f = createCRForm.value
  if (!f.title || !f.source_branch || !f.target_branch) {
    ElMessage.warning('请填写标题、源分支和目标分支')
    return
  }
  createCRLoading.value = true
  try {
    await createRemoteCR({
      provider_id: providerId, owner: repoOwner, repo: repoName,
      title: f.title, description: f.description,
      source_branch: f.source_branch, target_branch: f.target_branch,
      labels: f.labels ? f.labels.split(',').map(s => s.trim()).filter(Boolean) : undefined,
    })
    ElMessage.success('CR 创建成功')
    showCreateCRDialog.value = false
    createCRForm.value = { title: '', description: '', source_branch: '', target_branch: '', labels: '' }
    loadCRs()
  } catch (e: any) {
    ElMessage.error('创建失败: ' + (e?.message || ''))
  } finally {
    createCRLoading.value = false
  }
}

async function loadReviewConfig() {
  crCfgLoading.value = true
  try {
    const [res, provs] = await Promise.all([
      getRemoteRepoConfig(providerId, repoOwner, repoName),
      listLLMProviders().catch(() => []),
    ])
    if (res) reviewCfg.value = res
    globalProviders.value = provs as LLMProviderDTO[] || []
  } catch { /* use defaults */ }
  finally { crCfgLoading.value = false }
}

async function saveReviewConfig() {
  crCfgSaving.value = true
  try {
    const res = await updateRemoteRepoConfig(providerId, repoOwner, repoName, {
      enabled: reviewCfg.value.enabled,
      block_on_high: reviewCfg.value.block_on_high,
      auto_review_on_mr: reviewCfg.value.auto_review_on_mr,
      llm_provider: reviewCfg.value.llm_provider,
      max_files: reviewCfg.value.max_files,
      max_diff_lines: reviewCfg.value.max_diff_lines,
      scope_note: reviewCfg.value.scope_note,
    })
    if (res) reviewCfg.value = res
    ElMessage.success('配置已保存')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || ''))
  } finally {
    crCfgSaving.value = false
  }
}

async function loadBranchRules() {
  brLoading.value = true
  try {
    const res = await getRemoteRepoBranchRules(providerId, repoOwner, repoName)
    if (res) {
      branchRuleCfg.value = {
        use_custom_rules: res.use_custom_rules,
        rules: res.rules || [],
        protected_branches: res.protected_branches || [],
        linked_repos: res.linked_repos || [],
      }
    }
  } catch { /* use defaults */ }
  finally { brLoading.value = false }
}

async function saveBranchRules() {
  brSaving.value = true
  try {
    const res = await updateRemoteRepoBranchRules(providerId, repoOwner, repoName, {
      use_custom_rules: branchRuleCfg.value.use_custom_rules,
      rules: branchRuleCfg.value.rules,
      protected_branches: branchRuleCfg.value.protected_branches,
    })
    if (res) {
      branchRuleCfg.value = {
        use_custom_rules: res.use_custom_rules,
        rules: res.rules || [],
        protected_branches: res.protected_branches || [],
        linked_repos: res.linked_repos || [],
      }
    }
    ElMessage.success('分支规则已保存')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || ''))
  } finally {
    brSaving.value = false
  }
}

watch(activeTab, (tab) => {
  if (tab === 'cr' && crs.value.length === 0) loadCRs()
  if (tab === 'codereview' && !crCfgLoading.value) loadReviewConfig()
  if (tab === 'branchrules' && !brLoading.value) loadBranchRules()
  if (tab === 'webhooks' && webhookEvents.value.length === 0) loadWebhookEvents()
  if (tab === 'branches' && remoteBranches.value.length === 0) loadRemoteBranches()
})

onMounted(loadInitial)
</script>

<style scoped>
.remote-repo-detail-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.repo-title-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.repo-title-info h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.platform-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
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

.unlinked-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
}

.unlinked-icon { color: var(--text-color-placeholder); }
.unlinked-card h3 { margin: 0; font-size: 16px; color: var(--text-color-primary); }
.unlinked-card p { margin: 0; font-size: 13px; color: var(--text-color-secondary); }

.unlinked-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  max-width: 600px;
  margin-top: 16px;
}

.info-row {
  display: flex;
  gap: 12px;
  align-items: baseline;
}

.info-label {
  width: 70px;
  font-size: 12px;
  color: var(--text-color-placeholder);
  flex-shrink: 0;
}

.info-value {
  font-size: 13px;
  color: var(--text-color-primary);
}

.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}

.tab-bar {
  display: flex;
  gap: 4px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 0;
}

.tab-btn {
  padding: 10px 20px;
  border: none;
  background: transparent;
  font-size: 14px;
  color: var(--text-color-secondary);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
  font-weight: 500;
}

.tab-btn.active {
  color: var(--accent-primary);
  border-bottom-color: var(--accent-primary);
}

.tab-btn:hover {
  color: var(--accent-primary);
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.content-actions {
  display: flex;
  gap: 8px;
}

.event-type-cell {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.config-panel {
  display: flex;
  gap: 20px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  overflow: hidden;
}

.config-sidebar {
  width: 180px;
  border-right: 1px solid var(--border-color);
  padding: 16px 0;
  flex-shrink: 0;
}

.cfg-nav-btn {
  display: block;
  width: 100%;
  padding: 10px 20px;
  border: none;
  background: transparent;
  text-align: left;
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  border-left: 3px solid transparent;
  transition: all 0.2s;
}

.cfg-nav-btn.active {
  color: var(--accent-primary);
  border-left-color: var(--accent-primary);
  background: var(--accent-bg);
  font-weight: 500;
}

.cfg-nav-btn:hover { color: var(--accent-primary); }

.config-form-area {
  flex: 1;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.form-section { display: flex; flex-direction: column; gap: 16px; }

.form-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}

.form-label { display: flex; flex-direction: column; gap: 2px; }
.form-label span:first-child { font-size: 14px; font-weight: 500; color: var(--text-color-primary); }
.form-desc { font-size: 12px; color: var(--text-color-placeholder); }

.form-row-inline { display: flex; gap: 16px; }
.form-field { display: flex; flex-direction: column; gap: 4px; flex: 1; }
.form-field label { font-size: 13px; font-weight: 500; color: var(--text-color-secondary); }

.scope-card {
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #EEF2FF;
  background: #F8F9FF;
}

.scope-card h4 { margin: 0 0 4px 0; font-size: 14px; color: var(--text-color-primary); }
.scope-desc { margin: 0 0 12px 0; font-size: 12px; color: var(--text-color-placeholder); }
.scope-card--empty { background: #F9FAFB; border-color: var(--border-color); }
.scope-card--empty p { margin: 0; font-size: 13px; color: var(--text-color-placeholder); }

.scope-repos { display: flex; flex-direction: column; gap: 8px; }

.scope-repo-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  background: #fff;
  border: 1px solid var(--border-color);
}

.scope-repo-name { font-size: 13px; font-weight: 500; color: var(--text-color-primary); }
.scope-repo-key { font-size: 11px; color: var(--text-color-placeholder); font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; }

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.br-rule-card {
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  margin-bottom: 8px;
}
.br-rule-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}
.br-rule-prefix {
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--accent-bg);
  color: #6366F1;
  font-size: 12px;
  font-weight: 600;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}
.br-rule-name-input {
  border: none;
  background: transparent;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
  outline: none;
  flex: 1;
}
.br-rule-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 16px;
}
.switch-row {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 32px;
}
.toggle-label-sm { font-size: 11px; color: var(--text-color-secondary); }
.field-input {
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 6px 8px;
  font-size: 13px;
  outline: none;
  background: var(--bg-color-page);
  color: var(--text-color-primary);
}
.field-input:focus {
  border-color: var(--accent-primary);
}
.protected-tags { display: flex; flex-wrap: wrap; gap: 6px; }
.protected-tag-sm {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 4px;
  background: #FEF2F2;
  border: 1px solid #FECACA;
  font-size: 12px;
  color: #DC2626;
  font-weight: 500;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}
.tag-remove {
  border: none;
  background: none;
  color: #DC2626;
  cursor: pointer;
  font-size: 14px;
  padding: 0;
  line-height: 1;
  opacity: 0.6;
}
.tag-remove:hover { opacity: 1; }
.text-muted { font-size: 12px; color: var(--text-color-placeholder); }
</style>
