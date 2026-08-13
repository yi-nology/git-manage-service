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
      <CRReviewDetail v-if="reviewTask" :task="reviewTask" :repo-owner="repoOwner" :repo-name="repo_name" @close="reviewTask = null" @retried="reviewTask = $event" />
      <CRTab v-else :active="activeTab === 'cr'" :provider-id="provider_id" :repo-owner="repoOwner" :repo-name="repo_name" @show-review="reviewTask = $event" />
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
          <button class="cfg-nav-btn" :class="{ active: cfgPanel === 'basic' }" @click="cfgPanel = 'basic'">基本设置</button>
          <button class="cfg-nav-btn" :class="{ active: cfgPanel === 'prompt' }" @click="switchToPrompt">提示词设计</button>
        </div>

        <div v-if="cfgPanel === 'basic'" class="config-form-area">
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

        <div v-if="cfgPanel === 'prompt'" class="config-form-area">
          <div class="form-section">
            <div class="prompt-section-header">
              <div class="form-label">
                <span>系统提示词 (System Prompt)</span>
                <span class="form-desc">自定义发送给 LLM 的系统提示词。留空则使用全局默认值。</span>
              </div>
            </div>

            <div class="prompt-block">
              <div class="prompt-block-header">
                <span class="prompt-block-title">角色与输出格式</span>
                <span class="prompt-block-badge">可编辑</span>
              </div>
              <textarea
                v-model="promptCfg.prefix"
                class="prompt-textarea"
                rows="10"
                placeholder="留空使用全局默认"
              />
              <div class="prompt-block-actions">
                <button class="prompt-reset-btn" @click="promptCfg.prefix = ''">恢复默认</button>
              </div>
            </div>

            <div class="prompt-block">
              <div class="prompt-block-header">
                <span class="prompt-block-title">变更意图分析</span>
                <span class="prompt-block-badge">可编辑</span>
              </div>
              <textarea
                v-model="promptCfg.intent"
                class="prompt-textarea"
                rows="8"
                placeholder="留空使用全局默认"
              />
              <div class="prompt-block-actions">
                <button class="prompt-reset-btn" @click="promptCfg.intent = ''">恢复默认</button>
              </div>
            </div>

            <div class="prompt-block prompt-block-readonly">
              <div class="prompt-block-header">
                <span class="prompt-block-title">约束规则</span>
                <span class="prompt-block-badge prompt-block-badge-locked">系统锁定</span>
              </div>
              <textarea
                :value="promptCfg.suffix"
                class="prompt-textarea prompt-textarea-readonly"
                rows="4"
                readonly
              />
            </div>
          </div>

          <div class="form-actions">
            <ActionPill variant="outline" @click="resetPromptCfg">重置为默认</ActionPill>
            <ActionPill variant="primary" @click="saveReviewConfig" :disabled="crCfgSaving">{{ crCfgSaving ? '保存中...' : '保存' }}</ActionPill>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'branchrules'" class="tab-content">
      <BranchRulesTab :active="activeTab === 'branchrules'" :provider-id="provider_id" :repo-owner="repoOwner" :repo-name="repo_name" />
    </div>

    <div v-show="activeTab === 'webhooks'" class="tab-content">
      <WebhookEventsTab :active="activeTab === 'webhooks'" />
    </div>

    <div v-show="activeTab === 'branches'" class="tab-content">
      <RemoteBranchesTab
        :active="activeTab === 'branches'"
        :provider-id="provider_id"
        :repo-owner="repoOwner"
        :repo-name="repo_name"
        :linked-repo-key="linkedRepoKey"
        :default-branch="repoData?.default_branch || ''"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { FolderOpened, Download, Link, Refresh } from '@element-plus/icons-vue'
import { listProviderRepos } from '@/api/modules/provider'
import { getRepoList } from '@/api/modules/repo'
import { listBindings } from '@/api/modules/binding'
import { getRemoteRepoConfig, updateRemoteRepoConfig } from '@/api/modules/review'
import type { ReviewRepoConfigDTO } from '@/api/modules/review'
import type { ReviewTaskDTO } from '@/api/modules/review'
import { getPromptStructure } from '@/api/modules/review-rules'
import type { PromptStructureDTO } from '@/api/modules/review-rules'
import { listLLMProviders } from '@/api/modules/llm-settings'
import type { LLMProviderDTO } from '@/api/modules/llm-settings'
import { useProviderStore } from '@/stores/useProviderStore'
import PageHeader from '@/components/common/PageHeader.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'
import CRTab from '@/components/remote/CRTab.vue'
import CRReviewDetail from '@/components/remote/CRReviewDetail.vue'
import BranchRulesTab from '@/components/remote/BranchRulesTab.vue'
import WebhookEventsTab from '@/components/remote/WebhookEventsTab.vue'
import RemoteBranchesTab from '@/components/remote/RemoteBranchesTab.vue'

const providerStore = useProviderStore()

const route = useRoute()
const router = useRouter()
const provider_id = Number(route.params.provider_id)
const repoOwner = route.params.repoOwner as string
const repo_name = route.params.repo_name as string
const repoFullName = computed(() => `${repoOwner}/${repo_name}`)

const providerPlatform = ref('')
const linkedRepoKey = ref<string | null>(null)
const activeTab = ref((route.query.tab as string) || 'cr')
const reviewTask = ref<ReviewTaskDTO | null>(null)
const repoData = ref<{ clone_url?: string; ssh_url?: string; default_branch?: string; private?: boolean } | null>(null)

const crCfgLoading = ref(false)
const crCfgSaving = ref(false)
const cfgPanel = ref<'basic' | 'prompt'>('basic')
const globalPrompt = ref<PromptStructureDTO>({ prefix: '', intent: '', suffix: '' })
const promptCfg = ref<{ prefix: string; intent: string; suffix: string }>({ prefix: '', intent: '', suffix: '' })
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
  prompt_prefix_override: '',
  prompt_intent_override: '',
})

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

async function loadInitial() {
  const [, , bindings] = await Promise.all([
    providerStore.fetchProviders(),
    getRepoList().catch(() => []),
    listBindings({ provider_config_id: provider_id }).catch(() => []),
  ])
  const prov = providerStore.getProviderById(provider_id)
  if (prov) providerPlatform.value = prov.platform

  const linked = (bindings || []).find((b: any) =>
    b.platform_owner === repoOwner &&
    b.platform_repo === repo_name &&
    b.status === 'active'
  )
  if (linked) linkedRepoKey.value = linked.repo_key

  const remote_repos = await listProviderRepos(provider_id, { page: 1, per_page: 100 }).catch(() => [])
  const found = (remote_repos || []).find((r: any) => r.full_name === repoFullName.value)
  if (found) repoData.value = found
}

function handleClone() {
  const url = repoData.value?.ssh_url || repoData.value?.clone_url
  if (url) {
    const query: Record<string, string> = { url }
    if (provider_id) query.provider_config_id = String(provider_id)
    if (repoOwner) query.platform_owner = repoOwner
    if (repo_name) query.platform_repo = repo_name
    router.push({ path: '/local-repos/clone', query })
  }
}

async function loadReviewConfig() {
  crCfgLoading.value = true
  try {
    const [res, provs, promptRes] = await Promise.all([
      getRemoteRepoConfig(provider_id, repoOwner, repo_name),
      listLLMProviders().catch(() => []),
      getPromptStructure().catch(() => null),
    ])
    if (res) reviewCfg.value = res
    globalProviders.value = provs as LLMProviderDTO[] || []
    if (promptRes) {
      globalPrompt.value = promptRes
      promptCfg.value.suffix = promptRes.suffix
    }
    promptCfg.value.prefix = reviewCfg.value.prompt_prefix_override || ''
    promptCfg.value.intent = reviewCfg.value.prompt_intent_override || ''
  } catch { /* use defaults */ }
  finally { crCfgLoading.value = false }
}

async function saveReviewConfig() {
  crCfgSaving.value = true
  try {
    const res = await updateRemoteRepoConfig(provider_id, repoOwner, repo_name, {
      enabled: reviewCfg.value.enabled,
      block_on_high: reviewCfg.value.block_on_high,
      auto_review_on_mr: reviewCfg.value.auto_review_on_mr,
      llm_provider: reviewCfg.value.llm_provider,
      max_files: reviewCfg.value.max_files,
      max_diff_lines: reviewCfg.value.max_diff_lines,
      scope_note: reviewCfg.value.scope_note,
      prompt_prefix_override: promptCfg.value.prefix,
      prompt_intent_override: promptCfg.value.intent,
    })
    if (res) reviewCfg.value = res
    ElMessage.success('配置已保存')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || ''))
  } finally {
    crCfgSaving.value = false
  }
}

function switchToPrompt() {
  cfgPanel.value = 'prompt'
}

function resetPromptCfg() {
  promptCfg.value.prefix = ''
  promptCfg.value.intent = ''
}

watch(activeTab, (tab) => {
  if (tab !== 'cr') reviewTask.value = null
  if (tab === 'codereview' && !crCfgLoading.value) loadReviewConfig()
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

.prompt-section-header { margin-bottom: 16px; }

.prompt-block {
  margin-bottom: 20px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.prompt-block-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background: var(--bg-color-tertiary);
  border-bottom: 1px solid var(--border-color);
}

.prompt-block-title { font-size: 13px; font-weight: 500; color: var(--text-color-primary); }

.prompt-block-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  background: #EEF2FF;
  color: #6366F1;
  font-weight: 500;
}

.prompt-block-badge-locked {
  background: #FEF2F2;
  color: #EF4444;
}

.prompt-block-readonly {
  opacity: 0.7;
}

.prompt-textarea {
  width: 100%;
  padding: 12px 14px;
  border: none;
  font-size: 13px;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  line-height: 1.6;
  color: var(--text-color-primary);
  background: transparent;
  resize: vertical;
  outline: none;
}

.prompt-textarea:focus {
  background: rgba(99, 102, 241, 0.03);
}

.prompt-textarea-readonly {
  color: var(--text-color-secondary);
  cursor: not-allowed;
  resize: none;
}

.prompt-block-actions {
  display: flex;
  justify-content: flex-end;
  padding: 6px 14px;
}

.prompt-reset-btn {
  border: none;
  background: transparent;
  color: var(--text-color-placeholder);
  font-size: 12px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
}

.prompt-reset-btn:hover {
  color: var(--accent-primary);
  background: var(--accent-bg);
}
</style>
