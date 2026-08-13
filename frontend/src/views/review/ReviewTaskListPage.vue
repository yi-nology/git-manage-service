<template>
  <div class="review-page-wrapper">
    <PageHeader
      :title="`${repo_name} — 审查任务`"
      :show-back="true"
      :back-route="`/local-repos/${repo_key}/review`"
    >
      <template #actions>
        <ActionPill variant="primary" :icon="Plus" @click="showTriggerDialog = true">
          触发审查
        </ActionPill>
      </template>
    </PageHeader>

    <div class="review-layout">
      <RepoSidebar :repo-key="repo_key" active-key="review" />
      <div class="review-content">
        <div class="breadcrumb">
          <router-link :to="`/local-repos/${repo_key}/review`">总览</router-link>
          <span class="sep">/</span>
          <span class="current">审查任务</span>
        </div>

        <div class="filter-bar">
          <el-select v-model="filters.status" placeholder="全部状态" clearable size="default" @change="loadData">
            <el-option label="等待中" value="pending" />
            <el-option label="运行中" value="running" />
            <el-option label="通过" value="success" />
            <el-option label="失败" value="failed" />
            <el-option label="已阻止" value="blocked" />
          </el-select>
          <div class="filter-spacer"></div>
          <PaginationBar
            :total="total"
            :current-page="page"
            :page-size="page_size"
            @update:currentPage="handlePageChange"
          />
        </div>

        <DataTable
          v-if="tasks.length > 0 || loading"
          :columns="columns"
          :data="tasks"
          row-key="id"
          :loading="loading"
          class="table-card"
        >
          <template #cell-mr_iid="{ row }">
            <span class="td-mono">#{{ row.mr_iid }}</span>
          </template>
          <template #cell-commit_sha="{ row }">
            <span class="td-mono td-link" @click.stop="goDetail(row.id)">{{ shortSha(row.commit_sha) }}</span>
          </template>
          <template #cell-status="{ row }">
            <StatusBadge :variant="statusVariant(row.status)" :text="statusLabel(row.status)" :show-dot="true" />
          </template>
          <template #cell-risk_level="{ row }">
            <StatusBadge v-if="row.risk_level" :variant="riskVariant(row.risk_level)" :text="riskLabel(row.risk_level)" :show-dot="false" />
            <span v-else class="risk-none">—</span>
          </template>
          <template #cell-findings_count="{ row }">
            <span class="td-mono">{{ row.findings_count || 0 }}</span>
          </template>
          <template #cell-trigger_type="{ row }">
            {{ triggerLabel(row.trigger_type) }}
          </template>
          <template #cell-created_at="{ row }">
            {{ timeAgo(row.created_at) }}
          </template>
        </DataTable>

        <EmptyState
          v-else
          title="暂无审查记录"
          description="点击「触发审查」开始对 MR 进行代码审查"
        />
      </div>
    </div>

    <el-dialog v-model="showTriggerDialog" title="触发代码审查" width="480px">
      <el-form label-width="80px">
        <el-form-item label="MR 编号">
          <el-input v-model="triggerForm.mr_iid" placeholder="例如: 142" />
        </el-form-item>
        <el-form-item label="提交 SHA">
          <el-input v-model="triggerForm.commit_sha" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTriggerDialog = false">取消</el-button>
        <el-button type="primary" @click="handleTrigger" :loading="triggering">触发审查</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import RepoSidebar from '@/components/repo/RepoSidebar.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { listReviewTasks, createReviewTask, type ReviewTaskDTO } from '@/api/modules/review'

const route = useRoute()
const router = useRouter()
const repo_key = route.params.repo_key as string
const repo_name = ref(repo_key)

const loading = ref(false)
const tasks = ref<ReviewTaskDTO[]>([])
const page = ref(1)
const page_size = ref(20)
const total = ref(0)
const filters = ref({ status: '' })
const showTriggerDialog = ref(false)
const triggering = ref(false)
const triggerForm = ref({ mr_iid: '', commit_sha: '' })

const columns: TableColumn[] = [
  { key: 'mr_iid', label: 'MR', width: '60px' },
  { key: 'commit_sha', label: '提交', width: '100px' },
  { key: 'status', label: '状态', width: '100px' },
  { key: 'risk_level', label: '风险', width: '80px' },
  { key: 'findings_count', label: '问题', width: '70px' },
  { key: 'trigger_type', label: '触发', width: '80px' },
  { key: 'created_at', label: '时间', flex: 1 },
]

function statusLabel(s: string) { const m: Record<string, string> = { pending: '等待中', running: '运行中', success: '通过', failed: '失败', blocked: '已阻止' }; return m[s] || s }
function statusVariant(s: string): 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default' {
  const m: Record<string, 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default'> = { pending: 'warning', running: 'running', success: 'success', failed: 'danger', blocked: 'danger' }
  return m[s] || 'default'
}
function riskLabel(r: string) { const m: Record<string, string> = { critical: '严重', high: '高危', medium: '中等', low: '低危', info: '提示' }; return m[r] || r }
function riskVariant(r: string): 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default' {
  const m: Record<string, 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default'> = { critical: 'danger', high: 'danger', medium: 'warning', low: 'info', info: 'default' }
  return m[r] || 'default'
}
function triggerLabel(t: string) { const m: Record<string, string> = { manual: '手动', webhook: 'Webhook', api: 'API' }; return m[t] || t }
function shortSha(sha: string) { return sha ? sha.substring(0, 7) : '—' }
function timeAgo(d: string) {
  if (!d) return '—'
  const diff = Date.now() - new Date(d).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins} 分钟前`
  const h = Math.floor(mins / 60)
  if (h < 24) return `${h} 小时前`
  return `${Math.floor(h / 24)} 天前`
}

function handlePageChange(p: number) {
  page.value = p
  loadData()
}

function goDetail(task_id: number) {
  router.push(`/local-repos/${repo_key}/review/tasks/${task_id}`)
}

async function loadData() {
  loading.value = true
  try {
    const res = await listReviewTasks({
      repo_key: repo_key,
      status: filters.value.status || undefined,
      page: page.value,
      page_size: page_size.value,
    })
    tasks.value = res?.tasks || []
    total.value = res?.pagination?.total || 0
    if (tasks.value.length > 0 && tasks.value[0]?.repo_name) repo_name.value = tasks.value[0].repo_name
  } catch (e) { console.error(e) } finally { loading.value = false }
}

async function handleTrigger() {
  if (!triggerForm.value.mr_iid) { ElMessage.warning('请输入 MR 编号'); return }
  triggering.value = true
  try {
    await createReviewTask({ repo_key: repo_key, mr_iid: triggerForm.value.mr_iid, commit_sha: triggerForm.value.commit_sha || undefined, trigger_type: 'manual' })
    ElMessage.success('审查任务已创建')
    showTriggerDialog.value = false
    triggerForm.value = { mr_iid: '', commit_sha: '' }
    loadData()
  } catch (e) { console.error(e) } finally { triggering.value = false }
}

onMounted(loadData)
</script>

<style scoped>
.review-page-wrapper { min-height: 100%; }
.review-layout { display: flex; gap: 20px; padding: 20px 24px; }
.review-content { flex: 1; min-width: 0; }
.breadcrumb { font-size: var(--font-size-sm); margin-bottom: var(--spacing-md); display: flex; align-items: center; gap: var(--spacing-sm); color: var(--text-color-secondary); }
.breadcrumb a { color: var(--primary-color); text-decoration: none; }
.breadcrumb .sep { color: var(--text-color-placeholder); }
.breadcrumb .current { font-weight: 600; color: var(--text-color-primary); }
.filter-bar { display: flex; align-items: center; gap: 12px; margin-bottom: var(--spacing-md); }
.filter-spacer { flex: 1; }
.td-mono { font-family: 'IBM Plex Mono', monospace; }
.td-link { color: var(--primary-color); cursor: pointer; }
.risk-none { color: var(--text-color-placeholder); }
</style>
