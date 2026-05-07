<template>
  <div class="slim-section">
    <div class="slim-header">
      <SectionTitle title="仓库瘦身" />
      <div class="slim-actions">
        <div class="threshold-control">
          <span class="threshold-label">阈值</span>
          <el-input-number v-model="thresholdKB" :min="1" :max="1048576" :step="100" size="small" style="width: 110px" />
          <span class="threshold-unit">KB</span>
        </div>
        <el-button type="primary" :icon="Refresh" :loading="healthLoading" @click="loadHealth">开始体检</el-button>
        <el-button :icon="Delete" @click="handleGC" :loading="gcLoading" :disabled="!healthReport">垃圾回收</el-button>
        <el-button type="warning" @click="handleForcePush" :disabled="!!taskId && taskStatus === 'running'">强制推送远端</el-button>
      </div>
    </div>

    <div class="prefix-slim-section">
      <SectionTitle title="前缀定向瘦身" />
      <div class="prefix-input-row">
        <div class="prefix-tags-row">
          <el-tag v-for="(tag, idx) in prefixTags" :key="idx" closable size="small" class="prefix-tag" @close="removePrefix(idx)">{{ tag }}</el-tag>
          <el-input v-model="prefixInput" size="small" placeholder="输入前缀如 vendor/ node_modules/" @keyup.enter="addPrefix" style="width: 220px" />
          <el-button size="small" type="primary" :icon="Plus" @click="addPrefix">添加</el-button>
        </div>
        <div class="prefix-presets-row">
          <span class="prefix-preset-label">常用:</span>
          <el-tag v-for="p in prefixPresets" :key="p" size="small" class="preset-tag" effect="plain" @click="addPrefixPreset(p)">{{ p }}</el-tag>
        </div>
        <div class="prefix-action-row">
          <el-button type="info" :loading="prefixPreviewLoading" @click="previewPrefix" :disabled="prefixTags.length === 0">预览匹配文件</el-button>
          <el-checkbox v-model="prefixSlimForcePush" label="瘦身后强制推送远端" size="small" />
        </div>
      </div>

      <div v-if="prefixPreviewLoading" class="slim-loading">
        <el-icon class="is-loading" :size="24"><Refresh /></el-icon>
        <span>正在扫描匹配文件...</span>
      </div>

      <template v-if="prefixPreview && !prefixPreviewLoading">
        <div class="prefix-summary">
          <span>匹配 <strong>{{ prefixPreview.totalCount }}</strong> 个文件，总计 <strong>{{ prefixPreview.totalSize }}</strong></span>
          <el-button type="danger" @click="handlePrefixSlimConfirm" :disabled="prefixPreview.totalCount === 0">执行前缀瘦身</el-button>
        </div>
        <el-table :data="prefixPreview.files.slice(0, 100)" style="width: 100%" size="small" max-height="300" empty-text="未找到匹配文件">
          <el-table-column prop="path" label="文件路径" min-width="260">
            <template #default="{ row }"><span class="mono">{{ row.path }}</span></template>
          </el-table-column>
          <el-table-column prop="size" label="大小" width="100" />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.exists ? 'success' : 'info'" size="small">{{ row.exists ? '存在' : '已删除' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="commitCount" label="涉及提交" width="80" />
        </el-table>
        <div v-if="prefixPreview.totalCount > 100" class="prefix-more-hint">仅显示前 100 个文件（共 {{ prefixPreview.totalCount }} 个）</div>
      </template>
    </div>

    <div v-if="!healthReport && !healthLoading" class="slim-empty">
      <el-empty description="点击「开始体检」扫描仓库健康状态和历史大文件" />
    </div>

    <div v-if="healthLoading" class="slim-loading">
      <el-icon class="is-loading" :size="24"><Refresh /></el-icon>
      <span>正在扫描仓库...</span>
    </div>

    <template v-if="healthReport && !healthLoading">
      <div class="slim-config-bar">
        <div class="exclude-section">
          <span class="exclude-label">排除规则:</span>
          <div class="exclude-tags">
            <el-tag v-for="(pat, idx) in excludePatterns" :key="idx" closable size="small" class="exclude-tag" @close="excludePatterns.splice(idx, 1)">{{ pat }}</el-tag>
            <el-popover placement="bottom" :width="320" trigger="click">
              <template #reference>
                <el-button size="small" :icon="Plus">添加</el-button>
              </template>
              <div class="exclude-add">
                <el-input v-model="newExclude" size="small" placeholder="如 docs/  .png  dist/" @keyup.enter="addExclude" style="width: 200px" />
                <el-button size="small" type="primary" @click="addExclude">添加</el-button>
              </div>
              <div class="exclude-presets">
                <span class="preset-label">快捷:</span>
                <el-tag v-for="p in presetPatterns" :key="p" size="small" class="preset-tag" effect="plain" @click="addPreset(p)">{{ p }}</el-tag>
              </div>
            </el-popover>
          </div>
        </div>
      </div>

      <div class="slim-stats">
        <div class="stat-card"><span class="stat-value">{{ healthReport.gitDirSize }}</span><span class="stat-label">.git 目录</span></div>
        <div class="stat-card"><span class="stat-value">{{ healthReport.commitCount }}</span><span class="stat-label">提交数</span></div>
        <div class="stat-card"><span class="stat-value">{{ healthReport.looseObjects }}</span><span class="stat-label">松散对象</span></div>
        <div class="stat-card"><span class="stat-value">{{ healthReport.packFiles }}</span><span class="stat-label">Pack 文件</span></div>
        <div class="stat-card"><span class="stat-value">{{ healthReport.branchCount }}</span><span class="stat-label">分支</span></div>
        <div class="stat-card"><span class="stat-value">{{ healthReport.tagCount }}</span><span class="stat-label">标签</span></div>
      </div>

      <div v-if="healthReport.gitDirBreakdown" class="slim-breakdown">
        <SectionTitle title=".git 空间明细" />
        <div class="breakdown-grid">
          <div class="breakdown-item">
            <div class="breakdown-bar" :style="{ width: barWidth(healthReport.gitDirBreakdown.packDirSizeBytes) }">
              <span class="breakdown-inner-text" v-if="healthReport.gitDirBreakdown.packDirSizeBytes > 0">{{ healthReport.gitDirBreakdown.packDirSize }}</span>
            </div>
            <div class="breakdown-info">
              <span class="breakdown-label">Pack 文件</span>
              <span class="breakdown-size">{{ healthReport.gitDirBreakdown.packDirSize }}</span>
            </div>
          </div>
          <div class="breakdown-item">
            <div class="breakdown-bar loose" :style="{ width: barWidth(healthReport.gitDirBreakdown.looseObjSizeBytes) }">
              <span class="breakdown-inner-text" v-if="healthReport.gitDirBreakdown.looseObjSizeBytes > 0">{{ healthReport.gitDirBreakdown.looseObjSize }}</span>
            </div>
            <div class="breakdown-info">
              <span class="breakdown-label">松散对象</span>
              <span class="breakdown-size">{{ healthReport.gitDirBreakdown.looseObjSize }}</span>
            </div>
          </div>
          <div class="breakdown-item">
            <div class="breakdown-bar reflog" :style="{ width: barWidth(healthReport.gitDirBreakdown.reflogSizeBytes) }">
              <span class="breakdown-inner-text" v-if="healthReport.gitDirBreakdown.reflogSizeBytes > 0">{{ healthReport.gitDirBreakdown.reflogSize }}</span>
            </div>
            <div class="breakdown-info">
              <span class="breakdown-label">Reflog 日志</span>
              <span class="breakdown-size">{{ healthReport.gitDirBreakdown.reflogSize }}</span>
            </div>
          </div>
          <div class="breakdown-item">
            <div class="breakdown-bar other" :style="{ width: barWidth(healthReport.gitDirBreakdown.otherSizeBytes) }">
              <span class="breakdown-inner-text" v-if="healthReport.gitDirBreakdown.otherSizeBytes > 0">{{ healthReport.gitDirBreakdown.otherSize }}</span>
            </div>
            <div class="breakdown-info">
              <span class="breakdown-label">其他 (hooks/info/etc)</span>
              <span class="breakdown-size">{{ healthReport.gitDirBreakdown.otherSize }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="healthReport.stashEntries && healthReport.stashEntries.length > 0" class="slim-stash">
        <SectionTitle title="Stash 条目" />
        <el-table :data="healthReport.stashEntries" style="width: 100%" size="small">
          <el-table-column prop="index" label="#" width="50" />
          <el-table-column prop="message" label="消息" min-width="300" />
          <el-table-column prop="size" label="大小" width="120" />
        </el-table>
      </div>

      <div class="slim-files-header">
        <SectionTitle title="历史大文件" />
        <div class="slim-files-actions">
          <span class="slim-files-hint">阈值 &gt; {{ healthReport.thresholdHuman }}，来源：历史 / Stash / Reflog</span>
          <el-button type="warning" :loading="aiLoading" @click="analyzeWithAI" :disabled="healthReport.largeFiles.length === 0">AI 分析</el-button>
        </div>
      </div>

      <div v-if="aiResult" class="ai-result-bar">
        <div class="ai-summary">
          <el-icon :size="16" style="margin-right:4px"><svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg"><path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z" fill="currentColor"/><path d="M464 336a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm72 112h-48c-4.4 0-8 3.6-8 8v272c0 4.4 3.6 8 8 8h48c4.4 0 8-3.6 8-8V456c0-4.4-3.6-8-8-8z" fill="currentColor"/></svg></el-icon>
          <span>{{ aiResult.summary }}</span>
          <el-tag v-if="aiResult.totalSaveBytes > 0" type="success" size="small" class="ai-savings">预计可释放 {{ aiResult.totalSavings }}</el-tag>
        </div>
        <el-button size="small" type="success" @click="acceptAIRecommendations">采纳推荐</el-button>
      </div>

      <el-table :data="healthReport.largeFiles" style="width: 100%" @selection-change="handleSelection" empty-text="未发现大于阈值的大文件">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="path" label="文件路径" min-width="260">
          <template #default="{ row }"><span class="mono">{{ row.path }}</span></template>
        </el-table-column>
        <el-table-column prop="size" label="大小" width="100" />
        <el-table-column label="来源" width="80">
          <template #default="{ row }">
            <el-tag :type="sourceTagType(row.source)" size="small">{{ sourceLabel(row.source) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.exists ? 'success' : 'info'" size="small">{{ row.exists ? '存在' : '已删除' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="aiRecommendationMap.size > 0" label="AI 推荐" width="120">
          <template #default="{ row }">
            <template v-if="aiRecommendationMap.get(row.path)">
              <el-tag :type="recTagType(aiRecommendationMap.get(row.path)!.recommendation)" size="small">{{ recLabel(aiRecommendationMap.get(row.path)!.recommendation) }}</el-tag>
            </template>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column v-if="aiRecommendationMap.size > 0" label="AI 理由" min-width="200">
          <template #default="{ row }">
            <template v-if="aiRecommendationMap.get(row.path)">
              <span class="ai-reason">{{ aiRecommendationMap.get(row.path)!.reason }}</span>
              <el-tag size="small" effect="plain" class="ai-conf">{{ aiRecommendationMap.get(row.path)!.confidence }}</el-tag>
            </template>
          </template>
        </el-table-column>
        <el-table-column v-if="aiRecommendationMap.size === 0" prop="commitCount" label="涉及提交" width="80" />
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
              <span class="arrow">&rarr;</span>
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
import { onMounted, ref } from 'vue'
import { Refresh, Delete, Plus } from '@element-plus/icons-vue'
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
  aiLoading,
  aiResult,
  aiRecommendationMap,
  thresholdKB,
  excludePatterns,
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
  analyzeWithAI,
  acceptAIRecommendations,
  prefixInput,
  prefixTags,
  prefixPreview,
  prefixPreviewLoading,
  prefixSlimForcePush,
  addPrefix,
  removePrefix,
  previewPrefix,
  handlePrefixSlimConfirm,
  handleForcePush,
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

function sourceTagType(source: string) {
  switch (source) {
    case 'stash': return 'warning'
    case 'reflog': return 'danger'
    default: return ''
  }
}

function sourceLabel(source: string) {
  switch (source) {
    case 'stash': return 'Stash'
    case 'reflog': return 'Reflog'
    default: return '历史'
  }
}

function recTagType(rec: string) {
  switch (rec) {
    case 'safe_to_delete': return 'success'
    case 'caution': return 'warning'
    case 'keep': return 'danger'
    default: return 'info'
  }
}

function recLabel(rec: string) {
  switch (rec) {
    case 'safe_to_delete': return '可删除'
    case 'caution': return '需注意'
    case 'keep': return '保留'
    default: return rec
  }
}

const newExclude = ref('')
const presetPatterns = ['.png', '.jpg', '.jpeg', '.gif', '.svg', '.ico', '.woff', '.woff2', '.ttf', '.eot', '.pdf', '.zip', '.tar.gz', 'docs/', 'dist/', 'node_modules/', 'vendor/', 'build/', 'public/']

function addExclude() {
  const v = newExclude.value.trim()
  if (v && !excludePatterns.value.includes(v)) {
    excludePatterns.value.push(v)
  }
  newExclude.value = ''
}

function addPreset(p: string) {
  if (!excludePatterns.value.includes(p)) {
    excludePatterns.value.push(p)
  }
}

function barWidth(bytes: number) {
  if (!healthReport.value || healthReport.value.gitDirSizeBytes <= 0) return '0%'
  const pct = (bytes / healthReport.value.gitDirSizeBytes) * 100
  return Math.max(pct, 0).toFixed(1) + '%'
}

onMounted(() => {
  loadRecords()
})
</script>

<style scoped>
.slim-section { display: flex; flex-direction: column; gap: 20px; }
.slim-header { display: flex; justify-content: space-between; align-items: center; }
.slim-actions { display: flex; gap: 8px; align-items: center; }
.slim-config-bar { padding: 12px 16px; background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-md); }
.exclude-section { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.exclude-label { font-size: 13px; color: var(--text-color-secondary); white-space: nowrap; }
.exclude-tags { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; }
.exclude-tag { cursor: default; }
.exclude-add { display: flex; gap: 8px; align-items: center; }
.exclude-presets { margin-top: 8px; display: flex; flex-wrap: wrap; gap: 4px; align-items: center; }
.preset-label { font-size: 12px; color: var(--text-color-secondary); margin-right: 4px; }
.preset-tag { cursor: pointer; }
.threshold-control { display: flex; align-items: center; gap: 6px; }
.threshold-label { font-size: 13px; color: var(--text-color-secondary); }
.threshold-unit { font-size: 13px; color: var(--text-color-secondary); }
.slim-loading { display: flex; align-items: center; gap: 8px; padding: 40px 0; justify-content: center; color: var(--text-color-secondary); }
.slim-stats { display: grid; grid-template-columns: repeat(6, 1fr); gap: 12px; }
.stat-card { display: flex; flex-direction: column; align-items: center; padding: 16px 8px; background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-md); }
.stat-value { font-size: 18px; font-weight: 600; color: var(--text-color-primary); }
.stat-label { font-size: 12px; color: var(--text-color-secondary); margin-top: 4px; }
.slim-breakdown { margin-top: 4px; }
.breakdown-grid { display: flex; flex-direction: column; gap: 8px; }
.breakdown-item { display: flex; flex-direction: column; gap: 4px; }
.breakdown-bar { height: 24px; background: #409EFF; border-radius: 4px; min-width: 2px; transition: width 0.3s ease; display: flex; align-items: center; padding-left: 8px; }
.breakdown-bar.loose { background: #67C23A; }
.breakdown-bar.reflog { background: #E6A23C; }
.breakdown-bar.other { background: #909399; }
.breakdown-inner-text { color: #fff; font-size: 12px; font-weight: 500; white-space: nowrap; }
.breakdown-info { display: flex; justify-content: space-between; font-size: 13px; }
.breakdown-label { color: var(--text-color-regular); }
.breakdown-size { color: var(--text-color-secondary); font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; }
.slim-stash { margin-top: 4px; }
.slim-files-header { display: flex; justify-content: space-between; align-items: center; }
.slim-files-actions { display: flex; align-items: center; gap: 12px; }
.slim-files-hint { font-size: 12px; color: var(--text-color-secondary); }
.ai-result-bar { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; background: var(--bg-color-page); border: 1px solid var(--border-color); border-radius: var(--border-radius-md); margin-bottom: 8px; }
.ai-summary { display: flex; align-items: center; gap: 4px; font-size: 13px; color: var(--text-color-regular); flex: 1; }
.ai-savings { margin-left: 8px; }
.ai-reason { font-size: 12px; color: var(--text-color-secondary); }
.ai-conf { margin-left: 6px; font-size: 11px; }
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
