<template>
  <div class="slim-section">
    <div class="slim-header">
      <SectionTitle title="仓库瘦身" />
      <div class="slim-actions">
        <el-button type="primary" :icon="Refresh" :loading="healthLoading" @click="loadHealth">开始体检</el-button>
        <el-button :icon="Delete" @click="handleGC" :loading="gcLoading" :disabled="!healthReport">垃圾回收</el-button>
      </div>
    </div>

    <div v-if="!healthReport && !healthLoading" class="slim-empty">
      <el-empty description="点击「开始体检」扫描仓库健康状态和历史大文件" />
    </div>

    <div v-if="healthLoading" class="slim-loading">
      <el-icon class="is-loading" :size="24"><Refresh /></el-icon>
      <span>正在扫描仓库...</span>
    </div>

    <template v-if="healthReport && !healthLoading">
      <div class="slim-stats">
        <div class="stat-card"><span class="stat-value">{{ healthReport.gitDirSize }}</span><span class="stat-label">.git 目录</span></div>
        <div class="stat-card"><span class="stat-value">{{ healthReport.commitCount }}</span><span class="stat-label">提交数</span></div>
        <div class="stat-card"><span class="stat-value">{{ healthReport.looseObjects }}</span><span class="stat-label">松散对象</span></div>
        <div class="stat-card"><span class="stat-value">{{ healthReport.packFiles }}</span><span class="stat-label">Pack 文件</span></div>
        <div class="stat-card"><span class="stat-value">{{ healthReport.branchCount }}</span><span class="stat-label">分支</span></div>
        <div class="stat-card"><span class="stat-value">{{ healthReport.tagCount }}</span><span class="stat-label">标签</span></div>
      </div>

      <div class="slim-files-header">
        <SectionTitle title="历史大文件" />
        <span class="slim-files-hint">仅显示 &gt; 1MB 的文件</span>
      </div>

      <el-table :data="healthReport.largeFiles" style="width: 100%" @selection-change="handleSelection" empty-text="未发现大于 1MB 的文件">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="path" label="文件路径" min-width="300">
          <template #default="{ row }"><span class="mono">{{ row.path }}</span></template>
        </el-table-column>
        <el-table-column prop="size" label="大小" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.exists ? 'success' : 'info'" size="small">{{ row.exists ? '存在' : '已删除' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="commitCount" label="涉及提交" width="100" />
      </el-table>

      <div v-if="selectedFiles.length > 0" class="slim-bottom-bar">
        <span>已选 {{ selectedFiles.length }} 个文件，预计可清理 {{ formatFileSize(selectedFiles.reduce((s, f) => s + f.sizeBytes, 0)) }}</span>
        <div class="slim-bottom-actions">
          <el-button @click="handleAddGitignore" :loading="gitignoreLoading">加入 .gitignore</el-button>
          <el-button type="danger" @click="handleSlimConfirm">从历史中删除</el-button>
        </div>
      </div>

      <div v-if="taskId" class="slim-task-progress">
        <SectionTitle title="任务进度" />
        <el-alert v-if="taskStatus === 'failed'" :title="taskError" type="error" show-icon :closable="false" />
        <el-alert v-else-if="taskStatus === 'success'" title="操作完成" type="success" show-icon :closable="false" />
        <template v-else>
          <el-progress :percentage="100" :indeterminate="true" />
          <div class="slim-logs">
            <div v-for="(log, i) in taskLogs" :key="i" class="slim-log-line">{{ log }}</div>
          </div>
        </template>
      </div>
    </template>

    <div class="slim-history">
      <div class="slim-history-header">
        <SectionTitle title="操作历史" />
        <el-button size="small" :icon="Refresh" @click="loadRecords()" :loading="recordsLoading">刷新</el-button>
      </div>
      <el-table :data="records" v-loading="recordsLoading" border size="small" empty-text="暂无维护记录">
        <el-table-column prop="type" label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.type === 'slim' ? 'warning' : 'info'" size="small">{{ row.type === 'slim' ? '瘦身' : 'GC' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="清理效果" min-width="160">
          <template #default="{ row }">
            <template v-if="row.snapshotBefore && row.snapshotAfter">
              <span class="mono">{{ row.snapshotBefore.gitDirSize }}</span>
              <span class="arrow">→</span>
              <span class="mono" :class="{ 'text-success': row.savedBytes > 0 }">{{ row.snapshotAfter.gitDirSize }}</span>
              <el-tag v-if="row.savedBytes > 0" type="success" size="small" class="ml-4">-{{ formatFileSize(row.savedBytes) }} ({{ row.savedPercent.toFixed(1) }}%)</el-tag>
            </template>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="耗时" width="80">
          <template #default="{ row }">{{ row.duration || '-' }}</template>
        </el-table-column>
        <el-table-column prop="createdAt" label="时间" width="160">
          <template #default="{ row }">{{ row.createdAt }}</template>
        </el-table-column>
        <el-table-column prop="errorMessage" label="错误" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.errorMessage" class="text-danger">{{ row.errorMessage }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="recordsTotal > recordsPageSize" class="pagination-wrap">
        <el-pagination
          small
          layout="prev, pager, next"
          :total="recordsTotal"
          :page-size="recordsPageSize"
          v-model:current-page="recordsPage"
          @current-change="(p: number) => loadRecords(p)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { Refresh, Delete } from '@element-plus/icons-vue'
import { useMaintenance } from '@/composables/useMaintenance'
import SectionTitle from '@/components/common/SectionTitle.vue'

const props = defineProps<{ repoKey: string }>()

const {
  healthLoading,
  healthReport,
  selectedFiles,
  gcLoading,
  gitignoreLoading,
  taskId,
  taskStatus,
  taskError,
  taskLogs,
  recordsLoading,
  records,
  recordsTotal,
  recordsPage,
  recordsPageSize,
  formatFileSize,
  loadHealth,
  handleSelection,
  handleAddGitignore,
  handleSlimConfirm,
  handleGC,
  loadRecords,
} = useMaintenance(props.repoKey)

function statusTagType(status: string) {
  switch (status) {
    case 'success': return 'success'
    case 'failed': return 'danger'
    case 'running': return 'warning'
    default: return 'info'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'success': return '完成'
    case 'failed': return '失败'
    case 'running': return '运行中'
    case 'pending': return '等待中'
    default: return status
  }
}

onMounted(() => {
  loadRecords()
})
</script>

<style scoped>
.slim-section { display: flex; flex-direction: column; gap: 20px; }
.slim-header { display: flex; justify-content: space-between; align-items: center; }
.slim-actions { display: flex; gap: 8px; }
.slim-loading { display: flex; align-items: center; gap: 8px; padding: 40px 0; justify-content: center; color: var(--text-color-secondary); }
.slim-stats { display: grid; grid-template-columns: repeat(6, 1fr); gap: 12px; }
.stat-card { display: flex; flex-direction: column; align-items: center; padding: 16px 8px; background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-md); }
.stat-value { font-size: 18px; font-weight: 600; color: var(--text-color-primary); }
.stat-label { font-size: 12px; color: var(--text-color-secondary); margin-top: 4px; }
.slim-files-header { display: flex; justify-content: space-between; align-items: center; }
.slim-files-hint { font-size: 12px; color: var(--text-color-secondary); }
.slim-bottom-bar { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; background: var(--bg-color-page); border: 1px solid var(--border-color); border-radius: var(--border-radius-md); font-size: 14px; color: var(--text-color-regular); }
.slim-bottom-actions { display: flex; align-items: center; gap: 12px; }
.slim-task-progress { margin-top: 8px; }
.slim-logs { max-height: 200px; overflow-y: auto; background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-md); padding: 12px; margin-top: 8px; font-family: monospace; font-size: 13px; }
.slim-log-line { padding: 2px 0; color: var(--text-color-regular); }
.slim-history { margin-top: 8px; }
.slim-history-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.mono { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; font-size: 13px; }
.arrow { margin: 0 6px; color: var(--text-color-secondary); }
.text-success { color: #10B981; }
.text-danger { color: var(--danger-color); }
.text-muted { color: var(--text-color-secondary); }
.ml-4 { margin-left: 8px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 12px; }
</style>
