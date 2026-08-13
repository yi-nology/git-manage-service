<template>
  <div class="review-page-wrapper">
    <PageHeader
      :title="`代码审查 · ${repo_name}`"
      :back-route="`/local-repos/${repo_key}`"
      :show-back="true"
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
        <StatsRow :stats="statsData" />

        <div class="tab-bar">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="tab-btn"
            :class="{ active: activeTab === tab.key }"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>

        <DataTable
          :columns="columns"
          :data="displayedTasks"
          row-key="id"
          :loading="loading"
        >
          <template #cell-mr="{ row }">
            <span class="cell-link" @click="goDetail(row.id)">#{{ row.mr_iid }}</span>
          </template>
          <template #cell-commit_sha="{ row }">
            <span class="cell-link cell-mono" @click="goDetail(row.id)">{{ shortSha(row.commit_sha) }}</span>
          </template>
          <template #cell-status="{ row }">
            <StatusBadge :variant="statusVariant(row.status)" :text="statusLabel(row.status)" />
          </template>
          <template #cell-risk_level="{ row }">
            <StatusBadge v-if="row.risk_level" :variant="riskVariant(row.risk_level)" :text="riskLabel(row.risk_level)" :show-dot="false" />
            <span v-else>—</span>
          </template>
          <template #cell-findings_count="{ row }">
            {{ row.findings_count || 0 }}
          </template>
          <template #cell-trigger_type="{ row }">
            {{ triggerLabel(row.trigger_type) }}
          </template>
          <template #cell-created_at="{ row }">
            {{ timeAgo(row.created_at) }}
          </template>
          <template #empty>
            <EmptyState
              title="暂无审查记录"
              description="点击「触发审查」开始对 MR 进行代码审查"
            />
          </template>
        </DataTable>
      </div>
    </div>

    <el-dialog v-model="showTriggerDialog" title="触发代码审查" width="480px">
      <el-form label-width="80px">
        <el-form-item label="MR 编号">
          <el-input v-model="triggerForm.mr_iid" placeholder="例如: 142" />
        </el-form-item>
        <el-form-item label="提交 SHA">
          <el-input v-model="triggerForm.commit_sha" placeholder="可选，留空自动获取" />
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
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import RepoSidebar from '@/components/repo/RepoSidebar.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatsRow from '@/components/common/StatsRow.vue'
import DataTable from '@/components/common/DataTable.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import { listReviewTasks, createReviewTask, type ReviewTaskDTO } from '@/api/modules/review'

const route = useRoute()
const router = useRouter()
const repo_key = route.params.repo_key as string
const repo_name = ref(repo_key)

const loading = ref(false)
const tasks = ref<ReviewTaskDTO[]>([])
const showTriggerDialog = ref(false)
const triggering = ref(false)
const triggerForm = ref({ mr_iid: '', commit_sha: '' })
const activeTab = ref('recent')

const tabs = [
  { key: 'recent', label: '最近任务' },
  { key: 'all', label: '全部任务' },
]

const displayedTasks = computed(() => {
  if (activeTab.value === 'recent') {
    return tasks.value.slice(0, 10)
  }
  return tasks.value
})

const columns: TableColumn[] = [
  { key: 'mr', label: 'MR', width: '60px' },
  { key: 'commit_sha', label: '提交', width: '100px' },
  { key: 'status', label: '状态', width: '100px' },
  { key: 'risk_level', label: '风险', width: '80px' },
  { key: 'findings_count', label: '问题数', width: '70px' },
  { key: 'trigger_type', label: '触发', width: '80px' },
  { key: 'created_at', label: '时间', flex: 1 },
]

const stats = computed(() => {
  const total = tasks.value.length
  const critical = tasks.value.filter(t => t.risk_level === 'critical' || t.risk_level === 'high').length
  const blocked = tasks.value.filter(t => t.status === 'blocked').length
  const passed = tasks.value.filter(t => t.status === 'success').length
  return { total, critical, blocked, passed }
})

const passRate = computed(() => {
  if (stats.value.total === 0) return 0
  return Math.round((stats.value.passed / stats.value.total) * 100)
})

const statsData = computed(() => [
  { label: '审查总数', value: stats.value.total },
  { label: '严重问题', value: stats.value.critical, color: 'var(--danger-color)' },
  { label: '合并被阻止', value: stats.value.blocked, color: '#EA580C' },
  { label: `通过审查 (${passRate.value}%)`, value: stats.value.passed, color: 'var(--success-color)' },
])

function statusVariant(s: string): 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default' {
  const m: Record<string, 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default'> = { pending: 'warning', running: 'running', success: 'success', failed: 'danger', blocked: 'danger' }
  return m[s] || 'default'
}

function statusLabel(s: string) {
  const m: Record<string, string> = { pending: '等待中', running: '运行中', success: '通过', failed: '失败', blocked: '已阻止' }
  return m[s] || s
}

function riskVariant(r: string): 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default' {
  const m: Record<string, 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default'> = { critical: 'danger', high: 'danger', medium: 'warning', low: 'info', info: 'default' }
  return m[r] || 'default'
}

function riskLabel(r: string) {
  const m: Record<string, string> = { critical: '严重', high: '高危', medium: '中等', low: '低危', info: '提示' }
  return m[r] || r
}

function triggerLabel(t: string) {
  const m: Record<string, string> = { manual: '手动', webhook: 'Webhook', api: 'API' }
  return m[t] || t
}

function shortSha(sha: string) {
  if (!sha) return '—'
  return sha.substring(0, 7)
}

function timeAgo(dateStr: string) {
  if (!dateStr) return '—'
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins} 分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  return `${days} 天前`
}

function goDetail(task_id: number) {
  router.push(`/local-repos/${repo_key}/review/tasks/${task_id}`)
}

async function loadData() {
  loading.value = true
  try {
    const res = await listReviewTasks({ repo_key: repo_key, page: 1, page_size: 10 })
    tasks.value = res?.tasks || []
    if (tasks.value.length > 0 && tasks.value[0]?.repo_name) {
      repo_name.value = tasks.value[0].repo_name
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function handleTrigger() {
  if (!triggerForm.value.mr_iid) {
    ElMessage.warning('请输入 MR 编号')
    return
  }
  triggering.value = true
  try {
    await createReviewTask({
      repo_key: repo_key,
      mr_iid: triggerForm.value.mr_iid,
      commit_sha: triggerForm.value.commit_sha || undefined,
      trigger_type: 'manual',
    })
    ElMessage.success('审查任务已创建')
    showTriggerDialog.value = false
    triggerForm.value = { mr_iid: '', commit_sha: '' }
    loadData()
  } catch (e) {
    console.error(e)
  } finally {
    triggering.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.review-page-wrapper { min-height: 100%; }
.review-layout { display: flex; gap: 20px; padding: 20px 24px; }
.review-content { flex: 1; min-width: 0; }
.review-content > .stats-row { margin-bottom: var(--spacing-lg); }
.tab-bar {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
  margin-bottom: 16px;
}
.tab-btn {
  padding: 10px 20px;
  border: none;
  border-bottom: 2px solid transparent;
  background: transparent;
  font-size: 14px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}
.tab-btn.active {
  color: var(--accent-primary);
  border-bottom-color: var(--accent-primary);
  font-weight: 500;
}
.tab-btn:hover {
  color: var(--accent-primary);
}
.cell-link { color: var(--primary-color); cursor: pointer; }
.cell-mono { font-family: 'IBM Plex Mono', monospace; }
</style>
