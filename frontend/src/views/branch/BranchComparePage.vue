<template>
  <div class="compare-page">
    <PageHeader title="分支对比 & 合并" show-back :back-route="branchBackRoute" />

    <div class="control-panel">
      <div class="branch-select-group">
        <div class="branch-box">
          <div class="branch-label">源分支 (Source/Feature)</div>
          <el-select v-model="sourceBranch" placeholder="选择源分支" filterable>
            <el-option-group label="本地分支">
              <el-option v-for="b in localBranches" :key="b" :label="b" :value="b" />
            </el-option-group>
            <el-option-group label="远程分支">
              <el-option v-for="b in remoteBranches" :key="b" :label="b" :value="b" />
            </el-option-group>
          </el-select>
        </div>
        <div class="arrow-box">
          <el-icon :size="20" color="#64748B"><Right /></el-icon>
        </div>
        <div class="branch-box">
          <div class="branch-label">目标分支 (Target/Base)</div>
          <el-select v-model="targetBranch" placeholder="选择目标分支" filterable>
            <el-option v-for="b in localBranches" :key="b" :label="b" :value="b" />
          </el-select>
        </div>
      </div>
      <div class="control-actions">
        <ActionPill variant="primary" :icon="Switch" :disabled="comparing" @click="handleCompare">对比</ActionPill>
        <ActionPill variant="green" :icon="Connection" :disabled="!compareResult || !canMerge" @click="showMergeDialog = true">合并</ActionPill>
      </div>
    </div>

    <el-alert
      v-if="targetBranch && isRemoteBranch(targetBranch)"
      title="目标分支不能是远程分支"
      type="warning"
      :closable="false"
      show-icon
      description="Git 合并只能在本地分支上执行，请选择本地分支作为目标分支。"
    />

    <div v-if="compareResult" class="compare-stats">
      <StatsRow :stats="compareStats" />
      <div class="stat-action-card">
        <ActionPill variant="outline" :icon="Download" @click="handleDownloadPatch">导出 Patch</ActionPill>
      </div>
    </div>

    <template v-if="compareResult">
      <SectionTitle title="变更文件列表" />

      <div class="file-table-card">
        <div class="table-header">
          <span class="th" style="width:80px">状态</span>
          <span class="th" style="flex:1">文件路径</span>
          <span class="th" style="width:120px">变更</span>
        </div>
        <div
          v-for="f in compareResult.files"
          :key="f.path"
          class="table-row"
          :class="{ active: f.path === currentFile }"
          @click="selectFile(f.path)"
        >
          <span class="td" style="width:80px">
            <StatusBadge :variant="getFileStatusVariant(f.status)" :text="f.status" :show-dot="false" />
          </span>
          <span class="td td-path" style="flex:1">{{ f.path }}</span>
          <span class="td" style="width:120px">
            <StatusBadge :variant="getFileStatusVariant(f.status)" :text="getChangeLabel(f.status)" :show-dot="false" />
          </span>
        </div>
      </div>

      <div v-if="currentFile" class="diff-section">
        <div class="diff-header">
          <span class="diff-title">{{ currentFile }}</span>
          <el-radio-group v-model="diffViewMode" size="small">
            <el-radio-button value="line-by-line">Line</el-radio-button>
            <el-radio-button value="side-by-side">Side</el-radio-button>
          </el-radio-group>
        </div>
        <div id="diff-viewer" v-html="diffHtml" class="diff-content"></div>
      </div>
    </template>

    <EmptyState v-if="!compareResult && !comparing" title="请选择分支进行对比" />

    <MergeDialog
      v-model:visible="showMergeDialog"
      :repo-key="repoKey"
      :source-branch="sourceBranch"
      :target-branch="targetBranch"
      @merged="handleCompare"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Right, Switch, Connection, Download } from '@element-plus/icons-vue'
import { getBranchList, compareBranches, getBranchDiff, getBranchPatch } from '@/api/modules/branch'
import type { BranchInfo } from '@/types/branch'
import * as Diff2Html from 'diff2html'
import 'diff2html/bundles/css/diff2html.min.css'
import PageHeader from '@/components/common/PageHeader.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatsRow from '@/components/common/StatsRow.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import MergeDialog from '@/components/branch/MergeDialog.vue'

const route = useRoute()
const repoKey = route.params.repoKey as string

const branchBackRoute = computed(() => `/local-repos/${repoKey}/branches`)

const allBranches = ref<BranchInfo[]>([])
const sourceBranch = ref('')
const targetBranch = ref('')
const comparing = ref(false)
const compareResult = ref<{ stat: { FilesChanged: number; Insertions: number; Deletions: number }; files: { path: string; status: string }[] } | null>(null)

const currentFile = ref('')
const diffHtml = ref('')
const diffViewMode = ref<'line-by-line' | 'side-by-side'>('line-by-line')

const showMergeDialog = ref(false)

const localBranches = computed(() =>
  allBranches.value.filter(b => b.type === 'local').map(b => b.name)
)
const remoteBranches = computed(() =>
  allBranches.value.filter(b => b.type === 'remote').map(b => b.name)
)

function isRemoteBranch(name: string): boolean {
  return remoteBranches.value.includes(name)
}

const canMerge = computed(() => {
  return targetBranch.value && !isRemoteBranch(targetBranch.value)
})

const compareStats = computed(() => {
  if (!compareResult.value) return []
  return [
    { label: '变更文件', value: String(compareResult.value.stat.FilesChanged) },
    { label: '新增行数', value: `+${compareResult.value.stat.Insertions}` },
    { label: '删除行数', value: `-${compareResult.value.stat.Deletions}` },
  ]
})

onMounted(async () => {
  try {
    const res = await getBranchList(repoKey, { page_size: 1000 })
    allBranches.value = res.list || []
  } catch {}
})

watch(diffViewMode, () => {
  if (currentFile.value) selectFile(currentFile.value)
})

function getFileStatusVariant(status: string): 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default' {
  if (status === 'A') return 'success'
  if (status === 'D') return 'danger'
  if (status === 'M') return 'info'
  if (status === 'R') return 'warning'
  return 'default'
}

function getChangeLabel(status: string): string {
  if (status === 'A') return 'Added'
  if (status === 'D') return 'Deleted'
  if (status === 'M') return 'Modified'
  return status
}

async function handleCompare() {
  if (!sourceBranch.value || !targetBranch.value) {
    ElMessage.warning('请选择源分支和目标分支')
    return
  }
  comparing.value = true
  compareResult.value = null
  currentFile.value = ''
  diffHtml.value = ''
  try {
    compareResult.value = await compareBranches(repoKey, sourceBranch.value, targetBranch.value)
  } finally {
    comparing.value = false
  }
}

async function selectFile(path: string) {
  currentFile.value = path
  try {
    const res = await getBranchDiff(repoKey, sourceBranch.value, targetBranch.value, path)
    diffHtml.value = Diff2Html.html(res.diff || '', {
      drawFileList: false,
      matching: 'lines',
      outputFormat: diffViewMode.value,
    })
  } catch {
    diffHtml.value = '<p>加载差异失败</p>'
  }
}

async function handleDownloadPatch() {
  try {
    const response = await getBranchPatch(repoKey, sourceBranch.value, targetBranch.value)
    const blob = response.data instanceof Blob ? response.data : new Blob([response.data], { type: 'application/octet-stream' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${sourceBranch.value}-to-${targetBranch.value}.patch`
    a.click()
    window.URL.revokeObjectURL(url)
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('导出 Patch 失败: ' + (err.message || '未知错误'))
  }
}
</script>

<style scoped>
.compare-page {
  padding: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-height: 100vh;
  background: var(--bg-color);
}

.control-panel {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
}

.branch-select-group {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.branch-box {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.branch-label {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.branch-box :deep(.el-select) {
  width: 100%;
}

.arrow-box {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 8px;
}

.control-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.compare-stats {
  display: flex;
  gap: 16px;
}

.compare-stats :deep(.stats-row) {
  flex: 1;
}

.stat-action-card {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
}

.file-table-card {
  border-radius: var(--border-radius-lg);
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  overflow: hidden;
}

.table-header {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  background: var(--accent-bg);
}

.th {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color-secondary);
}

.table-row {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: var(--border-color-extra-light);
}

.table-row.active {
  background: var(--accent-bg);
}

.td {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.td-path {
  color: var(--text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.diff-section {
  border-radius: var(--border-radius-lg);
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  overflow: hidden;
}

.diff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color);
}

.diff-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.diff-content {
  overflow-x: auto;
  padding: 0;
}

@media (max-width: 768px) {
  .compare-page {
    padding: var(--spacing-md);
  }

  .control-panel {
    flex-direction: column;
    align-items: stretch;
  }

  .branch-select-group {
    flex-direction: column;
  }

  .arrow-box {
    transform: rotate(90deg);
  }

  .compare-stats {
    flex-wrap: wrap;
  }

  .compare-stats :deep(.stat-card) {
    min-width: calc(50% - 12px);
  }
}
</style>
