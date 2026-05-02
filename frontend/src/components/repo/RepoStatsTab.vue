<template>
  <el-card>
    <el-form inline class="filter-form">
      <el-form-item label="分支">
        <el-select v-model="statsFilter.branch" placeholder="全部" clearable @change="loadStats" style="width: 220px">
          <el-option v-for="b in statsBranches" :key="b" :label="b" :value="b" />
        </el-select>
      </el-form-item>
      <el-form-item label="提交人">
        <el-select v-model="statsFilter.author" placeholder="全部" clearable filterable @change="loadStats" style="width: 220px">
          <el-option v-for="a in statsAuthors" :key="a.email" :label="`${a.name}(${a.email})`" :value="a.name" />
        </el-select>
      </el-form-item>
      <el-form-item label="开始日期">
        <el-date-picker v-model="statsFilter.since" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
      </el-form-item>
      <el-form-item label="结束日期">
        <el-date-picker v-model="statsFilter.until" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadStats">
          <el-icon><Search /></el-icon> 查询
        </el-button>
        <el-button @click="handleExportCsv">
          <el-icon><Download /></el-icon> 导出 CSV
        </el-button>
      </el-form-item>
    </el-form>

    <div v-if="statsData">
      <el-row :gutter="16" class="mb-4">
        <el-col :span="12">
          <el-statistic title="总有效行数" :value="statsData.total_lines" />
        </el-col>
        <el-col :span="12">
          <el-statistic title="活跃贡献者" :value="statsData.authors?.length || 0" />
        </el-col>
      </el-row>

      <GitStatsCharts :stats-data="statsData" />

      <el-card shadow="never" class="mt-4">
        <template #header><span style="font-weight:600;font-size:14px">提交历史（最近100条）</span></template>
        <el-table :data="commitHistory" border size="small" max-height="400">
          <el-table-column prop="hash" label="Hash" width="100">
            <template #default="{ row }">
              <el-text class="mono-text" size="small">{{ row.hash?.substring(0, 8) }}</el-text>
            </template>
          </el-table-column>
          <el-table-column prop="author" label="作者" width="120" />
          <el-table-column prop="date" label="时间" width="160">
            <template #default="{ row }">{{ formatRelativeTime(row.date) }}</template>
          </el-table-column>
          <el-table-column prop="message" label="信息" />
        </el-table>
      </el-card>
    </div>
    <el-empty v-else description="点击查询按钮加载数据" />
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Download } from '@element-plus/icons-vue'
import { getStatsAnalyze, getStatsCommits, exportStatsCsv } from '@/api/modules/stats'
import type { StatsResponse } from '@/types/stats'
import { formatRelativeTime } from '@/utils/format'
import GitStatsCharts from '@/components/stats/GitStatsCharts.vue'

const props = defineProps<{
  repoKey: string
  statsBranches: string[]
  statsAuthors: { name: string; email: string }[]
  repoName: string
}>()

const statsFilter = ref({ branch: '', author: '', since: '', until: '' })
const statsData = ref<StatsResponse | null>(null)
const commitHistory = ref<{ hash: string; author: string; date: string; message: string }[]>([])

async function loadStats() {
  try {
    statsData.value = await getStatsAnalyze(props.repoKey, {
      branch: statsFilter.value.branch || undefined,
      author: statsFilter.value.author || undefined,
      since: statsFilter.value.since || undefined,
      until: statsFilter.value.until || undefined,
    })
    const res = await getStatsCommits(props.repoKey, {
      branch: statsFilter.value.branch || undefined,
      author: statsFilter.value.author || undefined,
      since: statsFilter.value.since || undefined,
      until: statsFilter.value.until || undefined,
    })
    commitHistory.value = (Array.isArray(res) ? res : []).slice(0, 100)
  } catch { /* ignore */ }
}

async function handleExportCsv() {
  try {
    const params: Record<string, string> = { type: 'stats' }
    if (statsFilter.value.branch) params.branch = statsFilter.value.branch
    if (statsFilter.value.author) params.author = statsFilter.value.author
    if (statsFilter.value.since) params.since = statsFilter.value.since
    if (statsFilter.value.until) params.until = statsFilter.value.until
    const response = (await exportStatsCsv(props.repoKey, params)) as unknown as Blob
    const url = window.URL.createObjectURL(response)
    const a = document.createElement('a')
    a.href = url
    a.download = `${props.repoName || props.repoKey}-stats.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  } catch {
    ElMessage.error('导出失败')
  }
}
</script>

<style scoped>
.filter-form {
  margin-bottom: var(--spacing-md);
}
.mono-text {
  font-family: monospace;
}
.mt-4 {
  margin-top: var(--spacing-md);
}
.mb-4 {
  margin-bottom: var(--spacing-md);
}
</style>
