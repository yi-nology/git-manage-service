<template>
  <div class="branch-detail-page" v-loading="loading">
    <PageHeader :title="branchName" show-back :back-route="`/local-repos/${repoKey}/branches`">
      <template #title-suffix>
        <StatusBadge v-if="isCurrent" variant="success" text="当前分支" :show-dot="false" />
      </template>
      <template #actions>
        <ActionPill variant="primary" :icon="Switch" @click="$router.push(`/local-repos/${repoKey}/compare`)">对比/合并</ActionPill>
        <ActionPill variant="green" :icon="Top" @click="handlePush">推送远端</ActionPill>
        <ActionPill v-if="hasUncommitted" variant="amber" :icon="Upload" @click="router.push({ name: 'RepoDetail', params: { repoKey }, query: { tab: 'workspace' } })">前往工作区</ActionPill>
        <ActionPill variant="outline" :icon="Refresh" @click="loadData">刷新</ActionPill>
        <ActionPill v-if="!isCurrent" variant="danger" :icon="Delete" @click="handleDelete">删除分支</ActionPill>
      </template>
    </PageHeader>

    <div v-if="hasUncommitted" class="uncommitted-alert">
      <span class="alert-title">检测到未提交的变更</span>
      <pre class="status-text">{{ repoStatus }}</pre>
    </div>

    <StatsRow :stats="branchStats" />

    <SectionTitle title="最近提交" />

    <DataTable :columns="commitColumns" :data="commits" row-key="hash">
      <template #cell-hash="{ row }">
        <span class="hash-text">{{ row.hash?.substring(0, 8) }}</span>
      </template>
      <template #cell-message="{ row }">
        <span class="td-message">{{ row.message }}</span>
      </template>
      <template #cell-author="{ row }">
        <span class="author-name">{{ row.author }}</span>
      </template>
      <template #cell-date="{ row }">{{ formatRelativeTime(row.date) }}</template>
      <template #empty>
        <EmptyState title="暂无提交记录" />
      </template>
    </DataTable>

    <el-dialog v-model="showPushDialog" :title="`推送分支: ${branchName}`" width="480px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="目标远端">
          <el-checkbox-group v-model="pushRemotes">
            <el-checkbox v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPushDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitPush">确认推送</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showSubmitDialog" title="提交变更" width="550px" destroy-on-close>
      <el-form :model="submitForm" label-width="110px">
        <el-form-item label="Author Name">
          <el-input v-model="submitForm.author_name" />
        </el-form-item>
        <el-form-item label="Author Email">
          <el-input v-model="submitForm.author_email" />
        </el-form-item>
        <el-form-item label="Commit 信息" required>
          <el-input v-model="submitForm.message" type="textarea" :rows="3" placeholder="请输入提交信息" />
        </el-form-item>
        <el-form-item label="提交后推送">
          <el-switch v-model="submitForm.push" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSubmitDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitChanges" :loading="submitting">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Switch, Top, Upload, Refresh, Delete, DataLine, Tickets, User, Files } from '@element-plus/icons-vue'
import { pushBranch, deleteBranch, getBranchList } from '@/api/modules/branch'
import { getRepoDetail, scanRepo } from '@/api/modules/repo'
import { getStatsAnalyze, getStatsCommits } from '@/api/modules/stats'
import { getRepoStatus, getRepoGitConfig, submitChanges } from '@/api/modules/system'
import type { StatsResponse } from '@/types/stats'
import { formatRelativeTime } from '@/utils/format'
import PageHeader from '@/components/common/PageHeader.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import StatsRow from '@/components/common/StatsRow.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'

interface CommitInfo {
  hash: string
  message: string
  author: string
  date: string
}

const route = useRoute()
const router = useRouter()
const repoKey = route.params.repoKey as string
const branchName = route.params.branchName as string

const loading = ref(false)
const isCurrent = ref(false)
const statsData = ref<StatsResponse | null>(null)
const commits = ref<CommitInfo[]>([])
const remoteNames = ref<string[]>([])
const repoStatus = ref('')
const hasUncommitted = ref(false)

const showPushDialog = ref(false)
const pushRemotes = ref<string[]>([])

const showSubmitDialog = ref(false)
const submitting = ref(false)
const submitForm = ref({
  author_name: '',
  author_email: '',
  message: '',
  push: false,
})

const fileTypeCount = computed(() => {
  if (!statsData.value?.authors) return 0
  const types = new Set<string>()
  for (const a of statsData.value.authors) {
    if (a.file_types) {
      Object.keys(a.file_types).forEach((t) => types.add(t))
    }
  }
  return types.size
})

const branchStats = computed(() => [
  { label: '总代码行数', value: statsData.value?.total_lines || 0, icon: DataLine, color: '#6366F1' },
  { label: '提交总数', value: commits.value.length, icon: Tickets, color: '#10B981' },
  { label: '贡献者数', value: statsData.value?.authors?.length || 0, icon: User, color: '#F59E0B' },
  { label: '文件类型', value: fileTypeCount.value, icon: Files, color: '#3B82F6' },
])

const commitColumns = [
  { key: 'hash', label: 'Hash', width: '100px' },
  { key: 'message', label: '信息', flex: 1 },
  { key: 'author', label: '作者', width: '120px' },
  { key: 'date', label: '时间', width: '140px' },
]

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    try {
      const res = await getBranchList(repoKey, { type: 'local', page_size: 500 })
      const branch = (res.list || []).find((b) => b.name === branchName)
      isCurrent.value = branch?.is_current || false
    } catch { /* ignore */ }

    try {
      statsData.value = await getStatsAnalyze(repoKey, { branch: branchName })
    } catch { /* ignore */ }

    try {
      const res = await getStatsCommits(repoKey, { branch: branchName })
      commits.value = (Array.isArray(res) ? res : []).slice(0, 20)
    } catch { /* ignore */ }

    try {
      const repo = await getRepoDetail(repoKey)
      if (repo?.path) {
        const scan = await scanRepo(repo.path)
        remoteNames.value = (scan.remotes || []).map((r: { name: string }) => r.name)
      }
    } catch { /* ignore */ }

    try {
      const status = await getRepoStatus(repoKey) as unknown as { status: string }
      repoStatus.value = status?.status || ''
      hasUncommitted.value = !!repoStatus.value && repoStatus.value.trim() !== ''
    } catch { /* ignore */ }

    try {
      const config = await getRepoGitConfig(repoKey) as unknown as { name: string; email: string }
      submitForm.value.author_name = config?.name || ''
      submitForm.value.author_email = config?.email || ''
    } catch { /* ignore */ }
  } finally {
    loading.value = false
  }
}

function handlePush() {
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
    await pushBranch(repoKey, branchName, pushRemotes.value)
    ElMessage.success('推送成功')
    showPushDialog.value = false
  } catch { /* handled */ }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`确定要删除分支 "${branchName}" 吗？`, '确认删除', { type: 'warning' })
    await deleteBranch(repoKey, branchName)
    ElMessage.success('分支已删除')
    router.push(`/local-repos/${repoKey}/branches`)
  } catch { /* cancelled */ }
}

async function handleSubmitChanges() {
  if (!submitForm.value.message) {
    ElMessage.warning('请输入提交信息')
    return
  }
  submitting.value = true
  try {
    await submitChanges({
      repo_key: repoKey,
      message: submitForm.value.message,
      push: submitForm.value.push,
      author_name: submitForm.value.author_name || undefined,
      author_email: submitForm.value.author_email || undefined,
    })
    ElMessage.success('提交成功')
    showSubmitDialog.value = false
    await loadData()
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.branch-detail-page {
  padding: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-height: 100vh;
  background: var(--bg-color);
}

.uncommitted-alert {
  background: #FFFBEB;
  border: 1px solid var(--warning-color);
  border-radius: var(--border-radius-lg);
  padding: 12px 16px;
}

.alert-title {
  font-weight: 600;
  font-size: 13px;
  color: var(--warning-color);
  display: block;
  margin-bottom: 8px;
}

.status-text {
  margin: 0;
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 12px;
  max-height: 200px;
  overflow-y: auto;
  color: var(--text-color-secondary);
}

.hash-text {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  color: var(--primary-color);
}

.td-message {
  color: var(--text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.author-name {
  color: var(--text-color-primary);
  font-size: 13px;
}

@media (max-width: 768px) {
  .branch-detail-page {
    padding: var(--spacing-md);
  }
}
</style>
