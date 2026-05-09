<template>
  <div class="branch-overview-panel">
    <div class="bo-toolbar">
      <ActionPill variant="outline" :icon="Refresh" @click="loadBranches" :disabled="loading">刷新</ActionPill>
      <ActionPill variant="outline" :icon="Download" @click="handleFetchAll" :disabled="fetchLoading">Fetch All</ActionPill>
      <ActionPill variant="primary" :icon="Plus" @click="showCreateDialog = true">新建分支</ActionPill>
      <ActionPill variant="green" :icon="Share" @click="$router.push(`/local-repos/${repoKey}/branches`)">
        完整管理
      </ActionPill>
    </div>

    <div v-loading="loading" class="bo-branch-list">
      <el-empty v-if="!loading && branches.length === 0" description="暂无分支" :image-size="48" />

      <div v-for="branch in branches" :key="branch.name" class="bo-branch-item" :class="{ current: branch.is_current }">
        <div class="bo-branch-main">
          <el-icon v-if="branch.is_current" class="bo-current-icon"><CircleCheck /></el-icon>
          <span class="bo-branch-name" :title="branch.name">{{ branch.name }}</span>
          <div class="bo-branch-status">
            <StatusBadge v-if="branch.upstream" variant="info" :text="branch.upstream" :show-dot="false" />
            <span v-if="branch.ahead > 0" class="bo-ahead">↑{{ branch.ahead }}</span>
            <span v-if="branch.behind > 0" class="bo-behind">↓{{ branch.behind }}</span>
            <StatusBadge v-if="branch.ahead === 0 && branch.behind === 0 && branch.upstream" variant="success" text="synced" :show-dot="false" />
          </div>
        </div>
        <div class="bo-branch-meta">
          <span class="bo-commit-msg" :title="branch.message">{{ branch.message || '-' }}</span>
          <span class="bo-commit-hash">{{ branch.hash ? branch.hash.substring(0, 7) : '' }}</span>
        </div>
        <div class="bo-branch-actions">
          <el-dropdown trigger="click">
            <el-button size="small" text>
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-if="!branch.is_current" @click="handleCheckout(branch.name)">
                  <el-icon><Select /></el-icon> 切换
                </el-dropdown-item>
                <el-dropdown-item @click="handlePush(branch.name)">
                  <el-icon><Top /></el-icon> 推送
                </el-dropdown-item>
                <el-dropdown-item v-if="branch.upstream && !branch.is_current" @click="handlePull(branch.name)">
                  <el-icon><Bottom /></el-icon> 拉取
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </div>

    <el-dialog v-model="showCreateDialog" title="新建分支" width="450px" destroy-on-close>
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="源分支">
          <el-select v-model="createForm.base_ref" style="width: 100%" placeholder="选择源分支">
            <el-option v-for="b in branches" :key="b.name" :label="b.name" :value="b.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="分支名称">
          <el-input v-model="createForm.name" placeholder="feature/new-branch" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :disabled="!createForm.name">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Download, Share, CircleCheck, MoreFilled, Select, Top, Bottom } from '@element-plus/icons-vue'
import { getBranchList, createBranch, checkoutBranch, pushBranch, pullBranch } from '@/api/modules/branch'
import { getRepoDetail, scanRepo, fetchRepo } from '@/api/modules/repo'
import type { BranchInfo } from '@/types/branch'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'

const props = defineProps<{
  repoKey: string
}>()

const loading = ref(false)
const fetchLoading = ref(false)
const branches = ref<BranchInfo[]>([])
const remoteNames = ref<string[]>([])
const showCreateDialog = ref(false)
const createForm = ref({ name: '', base_ref: 'main' })

onMounted(async () => {
  await loadBranches()
  try {
    const repo = await getRepoDetail(props.repoKey)
    if (repo?.path) {
      const scan = await scanRepo(repo.path)
      remoteNames.value = (scan.remotes || []).map((r: { name: string }) => r.name)
    }
  } catch { /* ignore */ }
})

async function loadBranches() {
  loading.value = true
  try {
    const result = await getBranchList(props.repoKey, { page_size: 500 })
    branches.value = result?.list || []
  } finally {
    loading.value = false
  }
}

async function handleFetchAll() {
  fetchLoading.value = true
  try {
    await fetchRepo(props.repoKey)
    ElMessage.success('Fetch 完成')
    await loadBranches()
  } catch {
    ElMessage.error('Fetch 失败')
  } finally {
    fetchLoading.value = false
  }
}

async function handleCreate() {
  if (!createForm.value.name) return
  try {
    await createBranch({
      repo_key: props.repoKey,
      name: createForm.value.name,
      base_ref: createForm.value.base_ref,
    })
    ElMessage.success('分支创建成功')
    showCreateDialog.value = false
    createForm.value = { name: '', base_ref: 'main' }
    await loadBranches()
  } catch {
    ElMessage.error('创建失败')
  }
}

async function handleCheckout(name: string) {
  try {
    await checkoutBranch(props.repoKey, name)
    ElMessage.success(`已切换到 ${name}`)
    await loadBranches()
  } catch {
    ElMessage.error('切换失败')
  }
}

async function handlePush(name: string) {
  try {
    await pushBranch(props.repoKey, name, remoteNames.value)
    ElMessage.success('推送成功')
  } catch {
    ElMessage.error('推送失败')
  }
}

async function handlePull(name: string) {
  try {
    await pullBranch(props.repoKey, name)
    ElMessage.success('拉取成功')
    await loadBranches()
  } catch {
    ElMessage.error('拉取失败')
  }
}
</script>

<style scoped>
.branch-overview-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.bo-toolbar {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.bo-branch-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.bo-branch-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 14px;
  border-radius: var(--border-radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  transition: all var(--transition-fast);
}

.bo-branch-item:hover {
  border-color: var(--primary-color);
  box-shadow: var(--box-shadow-sm);
}

.bo-branch-item.current {
  border-color: var(--primary-color);
  background: rgba(99, 102, 241, 0.04);
}

.bo-branch-main {
  display: flex;
  align-items: center;
  gap: 8px;
}

.bo-current-icon {
  color: var(--primary-color);
  font-size: 14px;
  flex-shrink: 0;
}

.bo-branch-name {
  font-weight: 500;
  font-size: var(--font-size-md);
  color: var(--text-color-primary);
  font-family: 'SF Mono', 'Consolas', monospace;
}

.bo-branch-status {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-left: auto;
  flex-shrink: 0;
}

.bo-ahead {
  font-size: var(--font-size-xs);
  color: var(--success-color);
  font-weight: 600;
}

.bo-behind {
  font-size: var(--font-size-xs);
  color: var(--warning-color);
  font-weight: 600;
}

.bo-branch-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-left: 22px;
}

.bo-commit-msg {
  flex: 1;
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bo-commit-hash {
  font-size: 11px;
  color: var(--text-color-placeholder);
  font-family: monospace;
  flex-shrink: 0;
}

.bo-branch-actions {
  position: absolute;
  right: 12px;
  top: 10px;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.bo-branch-item {
  position: relative;
}

.bo-branch-item:hover .bo-branch-actions {
  opacity: 1;
}
</style>
