<template>
  <div class="remote-repo-detail-page">
    <div class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="$router.push('/remote-repos')">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <div class="repo-icon" :style="{ background: platformMeta(providerPlatform).iconBg }">
          <el-icon :size="18" :style="{ color: platformMeta(providerPlatform).iconColor }"><FolderOpened /></el-icon>
        </div>
        <div class="repo-title-info">
          <h2>{{ repoFullName }}</h2>
          <span class="platform-badge" :style="{ background: platformMeta(providerPlatform).iconBg, color: platformMeta(providerPlatform).iconColor }">{{ platformMeta(providerPlatform).label }}</span>
          <span v-if="linkedRepoKey" class="linked-badge"><span class="dot-green"></span> 已关联本地</span>
        </div>
      </div>
      <div class="header-actions">
        <button v-if="!linkedRepoKey" class="action-pill action-pill--primary" @click="handleClone">
          <el-icon><Download /></el-icon> 克隆到本地
        </button>
        <button v-else class="action-pill action-pill--outline" @click="$router.push(`/local-repos/${linkedRepoKey}`)">
          <el-icon><FolderOpened /></el-icon> 查看本地仓库
        </button>
      </div>
    </div>

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
      <button class="action-pill action-pill--primary" style="margin-top:16px" @click="handleClone">
        <el-icon><Download /></el-icon> 克隆到本地
      </button>
    </div>

    <div class="tab-bar">
      <button class="tab-btn" :class="{ active: activeTab === 'cr' }" @click="activeTab = 'cr'">CR / MR</button>
      <button class="tab-btn" :class="{ active: activeTab === 'webhooks' }" @click="activeTab = 'webhooks'">Webhook 事件</button>
      <button class="tab-btn" :class="{ active: activeTab === 'branches' }" @click="activeTab = 'branches'">远程分支</button>
    </div>

      <div v-show="activeTab === 'cr'" class="tab-content">
        <div class="content-header">
          <h3>Change Requests</h3>
          <div class="content-actions">
            <button class="action-pill action-pill--primary" @click="showCreateCRDialog = true">
              <el-icon><Plus /></el-icon> 创建 CR
            </button>
            <button class="action-pill action-pill--outline" @click="handleSyncCRs" :disabled="crSyncing">
              <el-icon><Refresh /></el-icon> {{ crSyncing ? '刷新中...' : '刷新' }}
            </button>
          </div>
        </div>

        <div v-if="crLoading" class="loading-state"><div class="spinner"></div> 加载中...</div>

        <div v-else-if="crs.length === 0" class="empty-state">
          <el-icon :size="24" style="color:var(--text-color-placeholder)"><Share /></el-icon>
          <span>暂无 CR</span>
        </div>

        <div v-else class="table-card">
          <div class="table-header">
            <span class="th" style="width:60px">#</span>
            <span class="th" style="flex:1">标题</span>
            <span class="th" style="width:100px">源分支</span>
            <span class="th" style="width:100px">目标分支</span>
            <span class="th" style="width:80px">状态</span>
            <span class="th" style="width:80px">操作</span>
          </div>
          <div v-for="cr in crs" :key="cr.id" class="table-row">
            <span class="td mono" style="width:60px">{{ cr.cr_number }}</span>
            <span class="td" style="flex:1">{{ cr.title }}</span>
            <span class="td mono" style="width:100px">{{ cr.source_branch }}</span>
            <span class="td mono" style="width:100px">{{ cr.target_branch }}</span>
            <span class="td" style="width:80px">
              <span class="status-pill" :class="'status-' + cr.state">{{ crStateLabel(cr.state) }}</span>
            </span>
            <span class="td" style="width:80px">
              <button v-if="cr.state === 'opened'" class="act-btn act-btn--primary" @click="handleMergeCR(cr)">合并</button>
              <button v-if="cr.state === 'opened'" class="act-btn act-btn--danger" @click="handleCloseCR(cr)">关闭</button>
            </span>
          </div>
        </div>
      </div>

      <div v-show="activeTab === 'webhooks'" class="tab-content">
        <div class="content-header">
          <h3>Webhook 事件</h3>
          <button class="action-pill action-pill--outline" @click="loadWebhookEvents">
            <el-icon><Refresh /></el-icon> 刷新
          </button>
        </div>

        <div v-if="whLoading" class="loading-state"><div class="spinner"></div> 加载中...</div>

        <div v-else-if="webhookEvents.length === 0" class="empty-state">
          <el-icon :size="24" style="color:var(--text-color-placeholder)"><Link /></el-icon>
          <span>暂无 Webhook 事件</span>
        </div>

        <div v-else class="table-card">
          <div class="table-header">
            <span class="th" style="width:120px">事件类型</span>
            <span class="th" style="width:80px">来源</span>
            <span class="th" style="flex:1">事件 ID</span>
            <span class="th" style="width:80px">状态</span>
            <span class="th" style="width:140px">时间</span>
            <span class="th" style="width:60px">操作</span>
          </div>
          <div v-for="ev in webhookEvents" :key="ev.id" class="table-row">
            <span class="td" style="width:120px"><el-icon :size="14" style="color:#6366F1"><Share /></el-icon> {{ ev.event_type }}</span>
            <span class="td" style="width:80px">{{ ev.source }}</span>
            <span class="td mono" style="flex:1">{{ ev.event_id?.substring(0, 16) }}...</span>
            <span class="td" style="width:80px">
              <span class="status-pill" :class="'status-' + ev.status">{{ whStatusLabel(ev.status) }}</span>
            </span>
            <span class="td" style="width:140px">{{ formatTime(ev.created_at) }}</span>
            <span class="td" style="width:60px">
              <button v-if="ev.status === 'failed'" class="act-btn act-btn--warn" @click="handleRetryEvent(ev)">重试</button>
            </span>
          </div>
        </div>
      </div>

      <div v-show="activeTab === 'branches'" class="tab-content">
        <div class="content-header">
          <h3>远程分支</h3>
          <div class="content-actions">
            <button class="action-pill action-pill--primary" @click="showCreateBranchDialog = true">
              <el-icon><Plus /></el-icon> 创建分支
            </button>
            <button class="action-pill action-pill--outline" @click="loadRemoteBranches">
              <el-icon><Refresh /></el-icon> 刷新
            </button>
          </div>
        </div>

        <div v-if="rbLoading" class="loading-state"><div class="spinner"></div> 加载中...</div>

        <div v-else-if="remoteBranches.length === 0" class="empty-state">
          <span>暂无远程分支数据</span>
        </div>

        <div v-else class="table-card">
          <div class="table-header">
            <span class="th" style="flex:1">分支名</span>
            <span class="th" style="width:160px">操作</span>
          </div>
          <div v-for="rb in remoteBranches" :key="rb.name" class="table-row">
            <span class="td mono" style="flex:1">
              <el-icon :size="14" style="color:#10B981"><Share /></el-icon> {{ rb.name }}
            </span>
            <span class="td" style="width:160px">
              <button class="act-btn act-btn--primary" @click="handleCheckoutRemote(rb.name)">检出本地</button>
              <button class="act-btn act-btn--danger" @click="handleDeleteRemoteBranch(rb.name)">删除</button>
            </span>
          </div>
        </div>
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
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, FolderOpened, Download, Link, Refresh, Share, Plus } from '@element-plus/icons-vue'
import { listProviderRepos, listRemoteBranches, createRemoteBranch, deleteRemoteBranch } from '@/api/modules/provider'
import { listRemoteCRs, createRemoteCR, mergeRemoteCR, closeRemoteCR } from '@/api/modules/cr'
import type { CRDTO } from '@/api/modules/cr'
import { listWebhookEvents, retryWebhookEvent } from '@/api/modules/webhook-event'
import type { WebhookEventDTO } from '@/api/modules/webhook-event'
import { getRepoList } from '@/api/modules/repo'
import { createBranch } from '@/api/modules/branch'
import { useProviderStore } from '@/stores/useProviderStore'

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

function whStatusLabel(s: string) {
  if (s === 'processed') return '已处理'
  if (s === 'received') return '待处理'
  if (s === 'failed') return '失败'
  return s
}

function formatTime(t: string) {
  if (!t) return '-'
  const d = new Date(t)
  return `${d.getMonth() + 1}-${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

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

import { watch } from 'vue'
watch(activeTab, (tab) => {
  if (tab === 'cr' && crs.value.length === 0) loadCRs()
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

.repo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  flex-shrink: 0;
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

.linked-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 4px;
  background: #ECFDF5;
  color: #059669;
  font-size: 11px;
  font-weight: 500;
}

.dot-green {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #10B981;
  display: inline-block;
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

.action-pill--primary:hover { background: #4F46E5; }

.action-pill--outline {
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  color: var(--text-color-secondary);
}

.action-pill--outline:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.unlinked-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
  border-radius: 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
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

.info-value.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}

.tab-bar {
  display: flex;
  gap: 4px;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
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
  color: var(--accent-primary, #6366F1);
  border-bottom-color: var(--accent-primary, #6366F1);
}

.tab-btn:hover {
  color: var(--accent-primary, #6366F1);
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

.content-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.content-actions {
  display: flex;
  gap: 8px;
}

.loading-state, .empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: var(--text-color-secondary);
  font-size: 13px;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color, #e5e7eb);
  border-top-color: var(--accent-primary, #6366F1);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.table-card {
  border-radius: 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  overflow: hidden;
}

.table-header {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  background: #EEF2FF;
}

.th { font-size: 12px; font-weight: 600; color: var(--text-color-secondary); }

.table-row {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
  transition: background 0.15s;
}

.table-row:last-child { border-bottom: none; }
.table-row:hover { background: #F8FAFC; }

.td { font-size: 13px; color: var(--text-color-secondary); display: inline-flex; align-items: center; gap: 4px; }
.td.mono { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; font-size: 12px; }

.status-pill {
  display: inline-block;
  padding: 3px 8px;
  border-radius: 9999px;
  font-size: 11px;
  font-weight: 500;
}

.status-opened { background: #ECFDF5; color: #059669; }
.status-merged { background: #EEF2FF; color: #6366F1; }
.status-closed { background: #FEF2F2; color: #DC2626; }
.status-processed { background: #ECFDF5; color: #059669; }
.status-received { background: #FFFBEB; color: #D97706; }
.status-failed { background: #FEF2F2; color: #DC2626; }

.act-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid var(--border-color, #e5e7eb);
  background: transparent;
  color: var(--text-color-secondary);
}

.act-btn--primary { border-color: #6366F1; color: #6366F1; }
.act-btn--primary:hover { background: #EEF2FF; }
.act-btn--danger { border-color: #EF4444; color: #EF4444; }
.act-btn--danger:hover { background: #FEF2F2; }
.act-btn--warn { border-color: #F59E0B; color: #F59E0B; }
.act-btn--warn:hover { background: #FFFBEB; }
</style>
