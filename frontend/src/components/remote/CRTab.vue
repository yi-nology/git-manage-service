<template>
  <div>
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
      <template #cell-review_status="{ row }">
        <template v-if="getReviewStatus(row.cr_number)">
          <StatusBadge
            :variant="reviewStatusVariant(getReviewStatus(row.cr_number)!.status)"
            :text="reviewStatusLabel(getReviewStatus(row.cr_number)!.status)"
          />
          <span v-if="getReviewStatus(row.cr_number)!.review_mode && getReviewStatus(row.cr_number)!.review_mode !== 'llm'" class="review-mode-tag">{{ reviewModeShort(getReviewStatus(row.cr_number)!.review_mode) }}</span>
        </template>
        <span v-else class="review-none">未审查</span>
      </template>
      <template #empty>
        <EmptyState title="暂无 CR" />
      </template>
      <template #row-actions="{ row }">
        <ActionPill v-if="row.state === 'opened' && !getReviewStatus(row.cr_number)" variant="primary" small @click="handleTriggerReview(row)" :disabled="reviewTriggering">审查</ActionPill>
        <ActionPill v-if="getReviewStatus(row.cr_number)" variant="outline" small @click="showReviewDetail(row)">详情</ActionPill>
        <ActionPill v-if="row.state === 'opened'" variant="primary" small @click="handleMergeCR(row)">合并</ActionPill>
        <ActionPill v-if="row.state === 'opened'" variant="danger" small @click="handleCloseCR(row)">关闭</ActionPill>
      </template>
    </DataTable>

    <el-dialog v-model="showCreateCRDialog" title="创建 CR / MR" width="520px" destroy-on-close @open="loadBranches">
      <el-form label-width="80px">
        <el-form-item label="标题"><el-input v-model="createCRForm.title" placeholder="CR 标题" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="createCRForm.description" type="textarea" :rows="3" placeholder="可选描述" /></el-form-item>
        <el-form-item label="源分支">
          <el-select v-model="createCRForm.source_branch" filterable placeholder="选择源分支" style="width: 100%" :loading="branchesLoading">
            <el-option v-for="b in branches" :key="b.name" :label="b.name" :value="b.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标分支">
          <el-select v-model="createCRForm.target_branch" filterable placeholder="选择目标分支" style="width: 100%" :loading="branchesLoading">
            <el-option v-for="b in branches" :key="b.name" :label="b.name" :value="b.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签"><el-input v-model="createCRForm.labels" placeholder="逗号分隔，如: bugfix,feature" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateCRDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateCR" :loading="createCRLoading">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { listRemoteCRs, createRemoteCR, mergeRemoteCR, closeRemoteCR } from '@/api/modules/cr'
import { listRemoteBranches } from '@/api/modules/provider'
import { listReviewTasksByProvider, createReviewTaskByProvider } from '@/api/modules/review'
import type { CRDTO } from '@/api/modules/cr'
import type { ReviewTaskDTO } from '@/api/modules/review'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const props = defineProps<{
  active: boolean
  providerId: number
  repoOwner: string
  repoName: string
}>()

const emit = defineEmits<{
  'show-review': [task: ReviewTaskDTO]
}>()

const crLoading = ref(false)
const crSyncing = ref(false)
const crs = ref<CRDTO[]>([])
const showCreateCRDialog = ref(false)
const createCRForm = ref({ title: '', description: '', source_branch: '', target_branch: '', labels: '' })
const createCRLoading = ref(false)
const loaded = ref(false)
const branches = ref<{ name: string }[]>([])
const branchesLoading = ref(false)
const reviewTasks = ref<ReviewTaskDTO[]>([])
const reviewTriggering = ref(false)

const crColumns: TableColumn[] = [
  { key: 'cr_number', label: '#', width: '60px' },
  { key: 'title', label: '标题' },
  { key: 'source_branch', label: '源分支', width: '100px' },
  { key: 'target_branch', label: '目标分支', width: '100px' },
  { key: 'state', label: '状态', width: '80px' },
  { key: 'review_status', label: 'Review', width: '100px' },
]

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

function getReviewStatus(crNumber: number): ReviewTaskDTO | null {
  const tasks = reviewTasks.value.filter(t => String(t.mr_iid) === String(crNumber))
  if (tasks.length === 0) return null
  return tasks.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())[0] ?? null
}

function reviewStatusVariant(s: string): 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default' {
  const m: Record<string, 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default'> = {
    pending: 'warning', running: 'running', success: 'success', failed: 'danger', blocked: 'danger'
  }
  return m[s] || 'default'
}

function reviewStatusLabel(s: string) {
  const m: Record<string, string> = { pending: '审查中', running: '审查中', success: '通过', failed: '失败', blocked: '阻塞' }
  return m[s] || s
}

function reviewModeShort(mode: string) {
  const m: Record<string, string> = { claude_cli: 'Claude', opencode_cli: 'OpenCode', qoder_cli: 'Qoder', codex_cli: 'Codex', hybrid: '混合' }
  return m[mode] || mode
}

async function loadBranches() {
  branchesLoading.value = true
  try {
    const res = await listRemoteBranches({ provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName })
    branches.value = res || []
  } catch { branches.value = [] }
  finally { branchesLoading.value = false }
}

async function loadCRs() {
  crLoading.value = true
  try {
    const res = await listRemoteCRs({ provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName, page: 1, per_page: 100 })
    crs.value = res?.items || []
  } catch { crs.value = [] }
  finally { crLoading.value = false }
}

async function loadReviewTasks() {
  try {
    const allTasks: ReviewTaskDTO[] = []
    for (const cr of crs.value) {
      const res = await listReviewTasksByProvider({ provider_id: props.providerId, mr_iid: String(cr.cr_number), page: 1, page_size: 100 })
      if (res?.tasks) allTasks.push(...res.tasks)
    }
    reviewTasks.value = allTasks
  } catch { reviewTasks.value = [] }
}

async function handleSyncCRs() {
  crSyncing.value = true
  try {
    const res = await listRemoteCRs({ provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName, page: 1, per_page: 100 })
    crs.value = res?.items || []
    await loadReviewTasks()
    ElMessage.success(`已刷新，共 ${crs.value.length} 个 CR`)
  } catch (e: any) { ElMessage.error('刷新失败: ' + (e?.message || '')) }
  finally { crSyncing.value = false }
}

async function handleMergeCR(cr: CRDTO) {
  try { await ElMessageBox.confirm(`确定合并 CR #${cr.cr_number}？`, '确认合并', { type: 'info' }) } catch { return }
  try {
    await mergeRemoteCR(props.providerId, props.repoOwner, props.repoName, cr.cr_number)
    ElMessage.success('合并成功')
    loadCRs()
  } catch (e: any) { ElMessage.error('合并失败: ' + (e?.message || '')) }
}

async function handleCloseCR(cr: CRDTO) {
  try { await ElMessageBox.confirm(`确定关闭 CR #${cr.cr_number}？`, '确认关闭', { type: 'warning' }) } catch { return }
  try {
    await closeRemoteCR(props.providerId, props.repoOwner, props.repoName, cr.cr_number)
    ElMessage.success('已关闭')
    loadCRs()
  } catch (e: any) { ElMessage.error('关闭失败: ' + (e?.message || '')) }
}

async function handleTriggerReview(cr: CRDTO) {
  reviewTriggering.value = true
  try {
    await createReviewTaskByProvider({
      provider_config_id: props.providerId,
      owner: props.repoOwner,
      repo: props.repoName,
      mr_iid: String(cr.cr_number),
      trigger_type: 'manual',
    })
    ElMessage.success(`已触发 MR #${cr.cr_number} 的代码审查`)
    setTimeout(loadReviewTasks, 2000)
  } catch (e: any) {
    ElMessage.error('触发审查失败: ' + (e?.message || ''))
  } finally {
    reviewTriggering.value = false
  }
}

function showReviewDetail(cr: CRDTO) {
  const task = getReviewStatus(cr.cr_number)
  if (task) {
    emit('show-review', task)
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
      provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName,
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

watch(() => props.active, async (val) => {
  if (val && !loaded.value) {
    await loadCRs()
    await loadReviewTasks()
    loaded.value = true
  }
}, { immediate: true })
</script>

<style scoped>
.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.content-actions {
  display: flex;
  gap: 8px;
}
.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}

.review-none {
  font-size: 12px;
  color: var(--text-color-placeholder);
}

.review-mode-tag {
  display: inline-block;
  margin-left: 4px;
  padding: 0 4px;
  font-size: 10px;
  line-height: 16px;
  border-radius: 3px;
  background: var(--el-fill-color-light);
  color: var(--text-color-secondary);
  vertical-align: middle;
}
</style>
