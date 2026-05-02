<template>
  <div class="branch-list-page">
    <PageHeader title="分支管理" show-back :back-route="`/local-repos/${repoKey}`">
      <template #actions>
        <ActionPill variant="green" :icon="SetUp" @click="$router.push(`/local-repos/${repoKey}/branch-actions`)">分支操作</ActionPill>
        <ActionPill variant="green" :icon="Switch" @click="$router.push(`/local-repos/${repoKey}/compare`)">分支对比 & 合并</ActionPill>
        <ActionPill variant="outline" :icon="Download" :disabled="fetchLoading" @click="handleFetchAll">刷新远端 (Fetch)</ActionPill>
        <ActionPill variant="primary" :icon="Plus" @click="showCreateDialog = true">新建分支</ActionPill>
      </template>
    </PageHeader>

    <div class="tab-bar">
      <div
        class="tab-item"
        :class="{ active: activeTab === 'local' }"
        @click="handleTabChange('local')"
      >本地分支</div>
      <div
        v-for="remoteName in remoteNames"
        :key="remoteName"
        class="tab-item"
        :class="{ active: activeTab === `remote-${remoteName}` }"
        @click="handleTabChange(`remote-${remoteName}`)"
      >远程分支 - {{ remoteName }}</div>
    </div>

    <form @submit.prevent="loadBranches">
      <SearchBar v-model="searchQuery" placeholder="搜索分支名称..." />
    </form>

    <DataTable :columns="tableColumns" :data="branches" row-key="name" :loading="loading">
      <template #cell-name="{ row }">
        <template v-if="activeTab === 'local'">
          <span class="branch-name-cell" :class="{ 'branch-current': row.is_current }">
            <el-icon v-if="row.is_current" class="current-icon"><CircleCheck /></el-icon>
            {{ row.name }}
          </span>
        </template>
        <template v-else>
          <span class="branch-name-cell">{{ row.name.replace(`${activeTab.replace('remote-', '')}/`, '') }}</span>
        </template>
      </template>
      <template #cell-hash="{ row }">
        <span class="hash-text">{{ row.hash ? row.hash.substring(0, 8) : '-' }}</span>
        <span v-if="row.message" class="commit-msg">{{ row.message }}</span>
      </template>
      <template #cell-author="{ row }">
        <span class="author-name">{{ row.author }}</span>
        <span v-if="row.author_email" class="author-email">{{ row.author_email }}</span>
      </template>
      <template #cell-date="{ row }">{{ formatRelativeTime(row.date) }}</template>
      <template #cell-upstream="{ row }">
        <StatusBadge v-if="row.upstream" variant="info" :text="row.upstream" :show-dot="false" />
        <span v-else class="text-muted">无上游</span>
      </template>
      <template #cell-status="{ row }">
        <template v-if="row.upstream">
          <StatusBadge v-if="row.ahead > 0" variant="success" :text="`${row.ahead}↑`" :show-dot="false" />
          <StatusBadge v-if="row.behind > 0" variant="warning" :text="`${row.behind}↓`" :show-dot="false" />
          <StatusBadge v-if="row.ahead === 0 && row.behind === 0" variant="success" text="已同步" :show-dot="false" />
        </template>
        <span v-else class="text-muted">无上游</span>
      </template>
      <template #cell-localTracking="{ row }">
        <StatusBadge v-if="getLocalBranch(row.name)" variant="success" :text="getLocalBranch(row.name)!" :show-dot="false" />
        <span v-else class="text-muted">无关联</span>
      </template>
      <template #row-actions="{ row }">
        <template v-if="activeTab === 'local'">
          <el-dropdown @command="(cmd: string) => handleBranchCommand(cmd, row)">
            <ActionPill variant="primary" small>操作</ActionPill>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-if="!row.is_current" command="checkout">
                  <el-icon><Select /></el-icon> 切换
                </el-dropdown-item>
                <el-dropdown-item command="push">
                  <el-icon><Top /></el-icon> 推送
                </el-dropdown-item>
                <el-dropdown-item command="pull">
                  <el-icon><Bottom /></el-icon> 拉取
                </el-dropdown-item>
                <el-dropdown-item command="tag">
                  <el-icon><PriceTag /></el-icon> 打标签
                </el-dropdown-item>
                <el-dropdown-item command="detail">
                  <el-icon><View /></el-icon> 详情
                </el-dropdown-item>
                <el-dropdown-item command="rename">
                  <el-icon><Edit /></el-icon> 重命名
                </el-dropdown-item>
                <el-dropdown-item v-if="!row.is_current" command="delete" divided>
                  <el-text type="danger"><el-icon><Delete /></el-icon> 删除</el-text>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <template v-else>
          <el-dropdown @command="(cmd: string) => handleRemoteBranchCommand(cmd, row)">
            <ActionPill variant="primary" small>操作</ActionPill>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-if="!getLocalBranch(row.name)" command="checkout">
                  <el-icon><Download /></el-icon> 检出为本地
                </el-dropdown-item>
                <el-dropdown-item v-if="getLocalBranch(row.name)" command="update">
                  <el-icon><Bottom /></el-icon> 更新本地
                </el-dropdown-item>
                <el-dropdown-item v-if="getLocalBranch(row.name)" command="sync">
                  <el-icon><RefreshRight /></el-icon> 同步本地
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </template>
      <template #empty>
        <EmptyState :title="activeTab === 'local' ? '暂无本地分支' : '暂无远端分支'" />
      </template>
    </DataTable>

    <div class="table-footer">
      <span class="pag-info">
        {{ activeTab === 'local' ? `共 ${total} 个本地分支` : `共 ${total} 个远端分支 (${activeTab.replace('remote-', '')})` }}
      </span>
    </div>

    <el-dialog v-model="showCreateDialog" title="新建分支" width="480px" destroy-on-close>
      <el-form :model="createForm" label-width="110px">
        <el-form-item label="新分支名称" required>
          <el-input v-model="createForm.name" @input="handleBranchNameInput" />
          <div v-if="validationLoading" class="form-tip" style="color:var(--text-color-secondary)">校验中...</div>
          <div v-else-if="validationErrors.length > 0" class="validation-errors">
            <div v-for="(err, idx) in validationErrors" :key="idx" class="validation-error-item">
              <el-icon :size="12" style="color:#EF4444"><Close /></el-icon>
              <span>{{ err.message }}</span>
            </div>
          </div>
          <div v-else-if="createForm.name && validationValid" class="form-tip" style="color:#10B981">
            <el-icon :size="12"><CircleCheck /></el-icon> 分支名符合规则
          </div>
        </el-form-item>
        <el-form-item label="基于 (Base Ref)">
          <el-input v-model="createForm.base_ref" placeholder="默认为当前 HEAD" />
          <div class="form-tip">可以是分支名、标签或 Commit Hash</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :disabled="validationErrors.length > 0">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRenameDialog" title="重命名分支" width="480px" destroy-on-close>
      <el-form :model="renameForm" label-width="90px">
        <el-form-item label="当前名称">
          <el-input :model-value="renameForm.old_name" disabled />
        </el-form-item>
        <el-form-item label="新名称" required>
          <el-input v-model="renameForm.new_name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRenameDialog = false">取消</el-button>
        <el-button type="primary" @click="handleRename">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showPushDialog" :title="`推送分支: ${pushBranchName}`" width="480px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="目标远端">
          <el-checkbox-group v-model="pushRemotes">
            <el-checkbox v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <el-alert type="info" :closable="false" show-icon>
        推送操作将把本地分支更新推送到选定的远端仓库。
      </el-alert>
      <template #footer>
        <el-button @click="showPushDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitPush">确认推送</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showTagDialog" title="打标签 (Tag)" width="550px" destroy-on-close>
      <el-form :model="tagForm" label-width="100px">
        <el-form-item label="目标引用">
          <el-input :model-value="tagForm.ref" disabled />
        </el-form-item>
        <el-form-item label="版本类型">
          <el-radio-group v-model="tagForm.versionType" @change="handleTagVersionTypeChange">
            <el-radio-button value="patch">Patch</el-radio-button>
            <el-radio-button value="minor">Minor</el-radio-button>
            <el-radio-button value="major">Major</el-radio-button>
            <el-radio-button value="custom">自定义</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="当前版本" v-if="tagNextVersion">
          <el-tag type="info" size="small">{{ tagNextVersion.current || '无' }}</el-tag>
        </el-form-item>
        <el-form-item label="标签名" required>
          <el-input v-model="tagForm.name" :disabled="tagForm.versionType !== 'custom'" placeholder="v1.0.0" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="tagForm.message" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="推送到远端">
          <el-select v-model="tagForm.push_remote" placeholder="不推送" clearable>
            <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTagDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateTag">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit, Select, Top, Bottom, Switch, Download, CircleCheck, PriceTag, View, RefreshRight, Close, SetUp } from '@element-plus/icons-vue'
import { getBranchList, createBranch, deleteBranch, updateBranch, checkoutBranch, pushBranch, pullBranch, createTag } from '@/api/modules/branch'
import { showGitError } from '@/utils/git'
import { fetchRepo, scanRepo } from '@/api/modules/repo'
import { validateBranchName } from '@/api/modules/branch-rule'
import { getRepoDetail } from '@/api/modules/repo'
import { getNextVersion } from '@/api/modules/version'
import type { NextVersionInfo } from '@/api/modules/version'
import type { BranchInfo } from '@/types/branch'
import { formatRelativeTime } from '@/utils/format'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchBar from '@/components/common/SearchBar.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'

const route = useRoute()
const router = useRouter()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const fetchLoading = ref(false)
const branches = ref<BranchInfo[]>([])
const localBranches = ref<BranchInfo[]>([])
const total = ref(0)
const activeTab = ref('local')
const searchQuery = ref('')
const remoteNames = ref<string[]>([])

const showCreateDialog = ref(false)
const createForm = ref({ name: '', base_ref: '' })

watch(showCreateDialog, (v) => {
  if (v) {
    validationErrors.value = []
    validationValid.value = false
  }
})

watch(searchQuery, (newVal, oldVal) => {
  if (oldVal && !newVal) loadBranches()
})

const showRenameDialog = ref(false)
const renameForm = ref({ old_name: '', new_name: '' })

const showPushDialog = ref(false)
const pushBranchName = ref('')
const pushRemotes = ref<string[]>([])

const showTagDialog = ref(false)
const tagForm = ref({ ref: '', name: '', message: '', push_remote: '', versionType: 'patch' as 'patch' | 'minor' | 'major' | 'custom' })
const tagNextVersion = ref<NextVersionInfo | null>(null)

const validationErrors = ref<{ field: string; message: string }[]>([])
const validationValid = ref(false)
const validationLoading = ref(false)
let validationTimer: ReturnType<typeof setTimeout> | null = null

const tableColumns = computed<TableColumn[]>(() => {
  const cols: TableColumn[] = [
    { key: 'name', label: '分支名称', width: '180px' },
    { key: 'hash', label: '最新提交', width: '200px' },
    { key: 'author', label: '提交人', width: '140px' },
    { key: 'date', label: '提交时间', width: '140px' },
  ]
  if (activeTab.value === 'local') {
    cols.push(
      { key: 'upstream', label: '上游分支', width: '160px' },
      { key: 'status', label: '状态', width: '120px' },
    )
  } else {
    cols.push(
      { key: 'localTracking', label: '本地关联', width: '120px' },
    )
  }
  return cols
})

function handleBranchNameInput() {
  if (validationTimer) clearTimeout(validationTimer)
  validationErrors.value = []
  validationValid.value = false
  if (!createForm.value.name.trim()) return
  validationTimer = setTimeout(async () => {
    validationLoading.value = true
    try {
      const res = await validateBranchName({
        repo_key: repoKey,
        branch_name: createForm.value.name,
        base_ref: createForm.value.base_ref || undefined,
      })
      if (res.valid) {
        validationValid.value = true
        validationErrors.value = []
      } else {
        validationValid.value = false
        validationErrors.value = res.errors || []
      }
    } catch {
      validationValid.value = false
    } finally {
      validationLoading.value = false
    }
  }, 400)
}

onMounted(async () => {
  try {
    const repo = await getRepoDetail(repoKey)
    if (repo?.path) {
      const scan = await scanRepo(repo.path)
      remoteNames.value = (scan.remotes || []).map((r: { name: string }) => r.name)
    }
  } catch { /* ignore */ }

  try {
    const res = await getBranchList(repoKey, { type: 'local', page_size: 500 })
    localBranches.value = res.list || []
  } catch { /* ignore */ }

  await loadBranches()
})

function handleTabChange(tab: string) {
  activeTab.value = tab
  loadBranches()
}

async function loadBranches() {
  loading.value = true
  try {
    let branchType = 'local'
    let remoteName = ''

    if (activeTab.value.startsWith('remote-')) {
      branchType = 'remote'
      remoteName = activeTab.value.replace('remote-', '')
    }

    const res = await getBranchList(repoKey, {
      type: branchType,
      keyword: searchQuery.value || undefined,
      page_size: 500,
    })

    let filteredBranches = res.list || []
    if (remoteName) {
      filteredBranches = filteredBranches.filter(branch => branch.name.startsWith(`${remoteName}/`))
    }

    branches.value = filteredBranches
    total.value = filteredBranches.length
  } finally {
    loading.value = false
  }
}

function getLocalBranch(remoteName: string): string | null {
  const parts = remoteName.split('/')
  if (parts.length < 2) return null
  const localName = parts.slice(1).join('/')
  const found = localBranches.value.find(b => b.name === localName)
  return found ? found.name : null
}

function goDetail(branchName: string) {
  router.push(`/local-repos/${repoKey}/branches/${encodeURIComponent(branchName)}`)
}

async function handleFetchAll() {
  fetchLoading.value = true
  try {
    await fetchRepo(repoKey)
    ElMessage.success('远端数据已刷新')
    await loadBranches()
  } finally {
    fetchLoading.value = false
  }
}

async function handleCheckout(name: string) {
  try {
    await checkoutBranch(repoKey, name)
    ElMessage.success(`已切换到 ${name}`)
    await loadBranches()
  } catch { /* handled */ }
}

async function handleCheckoutRemote(remoteName: string) {
  const parts = remoteName.split('/')
  const localName = parts.length >= 2 ? parts.slice(1).join('/') : remoteName
  try {
    await createBranch({
      repo_key: repoKey,
      name: localName,
      base_ref: remoteName,
    })
    ElMessage.success(`已检出为本地分支 ${localName}`)
    const res = await getBranchList(repoKey, { type: 'local', page_size: 500 })
    localBranches.value = res.list || []
    await loadBranches()
  } catch { /* handled */ }
}

async function handleFfRemote(remoteName: string) {
  const localName = getLocalBranch(remoteName)
  if (!localName) return
  try {
    await pullBranch(repoKey, localName)
    ElMessage.success(`已更新本地分支 ${localName}`)
    await loadBranches()
  } catch { /* handled */ }
}

async function handlePullRemote(remoteName: string) {
  const localName = getLocalBranch(remoteName)
  if (!localName) return
  try {
    await pullBranch(repoKey, localName)
    ElMessage.success(`已同步本地分支 ${localName}`)
    await loadBranches()
  } catch (e) {
    showGitError(e, '同步分支')
  }
}

function handlePush(name: string) {
  pushBranchName.value = name
  const first = remoteNames.value[0]
  pushRemotes.value = first ? [first] : []
  showPushDialog.value = true
}

async function handleSubmitPush() {
  if (!pushRemotes.value.length) {
    ElMessage.warning('请选择目标远端')
    return
  }
  try {
    await pushBranch(repoKey, pushBranchName.value, pushRemotes.value)
    ElMessage.success('推送成功')
    showPushDialog.value = false
    await loadBranches()
  } catch (e) {
    showGitError(e, '推送分支')
  }
}

async function handlePull(name: string) {
  try {
    await pullBranch(repoKey, name)
    ElMessage.success('拉取成功')
    await loadBranches()
  } catch (e) {
    showGitError(e, '拉取分支')
  }
}

async function handleCreate() {
  if (!createForm.value.name) {
    ElMessage.warning('请输入分支名称')
    return
  }
  try {
    await createBranch({
      repo_key: repoKey,
      name: createForm.value.name,
      base_ref: createForm.value.base_ref || undefined,
    })
    ElMessage.success('分支创建成功')
    showCreateDialog.value = false
    createForm.value = { name: '', base_ref: '' }
    await loadBranches()
  } catch { /* handled */ }
}

function openRenameDialog(branch: BranchInfo) {
  renameForm.value = { old_name: branch.name, new_name: branch.name }
  showRenameDialog.value = true
}

async function handleRename() {
  if (!renameForm.value.new_name) return
  try {
    await updateBranch(repoKey, renameForm.value.old_name, renameForm.value.new_name)
    ElMessage.success('重命名成功')
    showRenameDialog.value = false
    await loadBranches()
  } catch { /* handled */ }
}

async function handleDeleteBranch(name: string) {
  try {
    await ElMessageBox.confirm(`确定要删除分支 "${name}" 吗？`, '确认删除', { type: 'warning' })
    await deleteBranch(repoKey, name)
    ElMessage.success('分支已删除')
    await loadBranches()
  } catch { /* cancelled or handled */ }
}

function openTagDialog(branchName: string) {
  tagForm.value = { ref: branchName, name: '', message: '', push_remote: '', versionType: 'patch' }
  tagNextVersion.value = null
  showTagDialog.value = true
  getNextVersion(repoKey).then(info => {
    tagNextVersion.value = info
    handleTagVersionTypeChange('patch')
  }).catch(() => { /* ignore */ })
}

function handleTagVersionTypeChange(type: string | number | boolean) {
  if (!tagNextVersion.value) return
  switch (type) {
    case 'patch':
      tagForm.value.name = tagNextVersion.value.next_patch
      break
    case 'minor':
      tagForm.value.name = tagNextVersion.value.next_minor
      break
    case 'major':
      tagForm.value.name = tagNextVersion.value.next_major
      break
    case 'custom':
      tagForm.value.name = ''
      break
  }
}

async function handleCreateTag() {
  if (!tagForm.value.name) {
    ElMessage.warning('请输入标签名')
    return
  }
  try {
    await createTag({
      repo_key: repoKey,
      name: tagForm.value.name,
      ref: tagForm.value.ref,
      message: tagForm.value.message || undefined,
      push_remote: tagForm.value.push_remote || undefined,
    })
    ElMessage.success('标签创建成功')
    showTagDialog.value = false
  } catch { /* handled */ }
}

function handleBranchCommand(command: string, row: BranchInfo) {
  switch (command) {
    case 'checkout':
      handleCheckout(row.name)
      break
    case 'push':
      handlePush(row.name)
      break
    case 'pull':
      handlePull(row.name)
      break
    case 'tag':
      openTagDialog(row.name)
      break
    case 'detail':
      goDetail(row.name)
      break
    case 'rename':
      openRenameDialog(row)
      break
    case 'delete':
      handleDeleteBranch(row.name)
      break
  }
}

function handleRemoteBranchCommand(command: string, row: BranchInfo) {
  switch (command) {
    case 'checkout':
      handleCheckoutRemote(row.name)
      break
    case 'update':
      handleFfRemote(row.name)
      break
    case 'sync':
      handlePullRemote(row.name)
      break
  }
}
</script>

<style scoped>
.branch-list-page {
  padding: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-height: 100vh;
  background: var(--bg-color);
}

.tab-bar {
  display: flex;
  border-bottom: 1px solid var(--border-color);
}

.tab-item {
  padding: 10px 20px;
  font-size: 14px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
  border-bottom: 2px solid transparent;
}

.tab-item:hover {
  color: var(--primary-color);
}

.tab-item.active {
  color: var(--primary-color);
  font-weight: 500;
  border-bottom-color: var(--primary-color);
}

.branch-name-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.branch-current {
  color: var(--success-color);
}

.current-icon {
  color: var(--success-color);
}

.hash-text {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  color: var(--primary-color);
  font-size: 12px;
}

.commit-msg {
  margin-left: 6px;
  color: var(--text-color-secondary);
  font-size: 12px;
}

.author-name {
  display: block;
  color: var(--text-color-primary);
  font-size: 13px;
}

.author-email {
  display: block;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.text-muted {
  font-size: 12px;
  color: var(--text-color-placeholder);
}

.table-footer {
  padding: 8px 0;
  text-align: left;
}

.pag-info {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.form-tip {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  margin-top: var(--spacing-xs);
}

.validation-errors {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 6px;
}

.validation-error-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #EF4444;
}

@media (max-width: 768px) {
  .branch-list-page {
    padding: var(--spacing-md);
  }
}
</style>
