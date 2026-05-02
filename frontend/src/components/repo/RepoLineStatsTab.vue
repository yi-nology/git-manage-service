<template>
  <el-card>
    <el-form inline class="filter-form">
      <el-form-item label="分支">
        <el-select v-model="lineStatsFilter.branch" placeholder="当前工作区" clearable @change="loadLineStats" style="width: 220px">
          <el-option v-for="b in statsBranches" :key="b" :label="b" :value="b" />
        </el-select>
      </el-form-item>
      <el-form-item label="提交人">
        <el-select v-model="lineStatsFilter.author" placeholder="全部" clearable filterable style="width: 220px">
          <el-option v-for="a in statsAuthors" :key="a.email" :label="`${a.name}(${a.email})`" :value="a.name" />
        </el-select>
      </el-form-item>
      <el-form-item label="开始日期">
        <el-date-picker v-model="lineStatsFilter.since" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
      </el-form-item>
      <el-form-item label="结束日期">
        <el-date-picker v-model="lineStatsFilter.until" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadLineStats">
          <el-icon><Search /></el-icon> 查询
        </el-button>
        <el-button @click="openExcludeConfig">
          <el-icon><Setting /></el-icon> 排除配置
        </el-button>
        <el-button @click="handleExportCsv">
          <el-icon><Download /></el-icon> 导出 CSV
        </el-button>
      </el-form-item>
    </el-form>

    <el-alert type="info" :closable="false" show-icon class="mb-4">
      选择分支/提交人/时间范围后将使用 git blame 分析代码归属，统计速度会较慢
    </el-alert>

    <div v-loading="lineStatsLoading">
      <div v-if="lineStatsData">
        <el-row :gutter="16" class="mb-4">
          <el-col :span="6"><el-statistic title="代码行数" :value="lineStatsData.code_lines" /></el-col>
          <el-col :span="6"><el-statistic title="注释行数" :value="lineStatsData.comment_lines" /></el-col>
          <el-col :span="6"><el-statistic title="空白行数" :value="lineStatsData.blank_lines" /></el-col>
          <el-col :span="6"><el-statistic title="文件总数" :value="lineStatsData.total_files" /></el-col>
        </el-row>

        <el-alert v-if="lineStatsData.status === 'processing'" title="正在统计中..." type="info" :closable="false" show-icon>
          {{ lineStatsData.progress }}
        </el-alert>

        <LineStatsCharts :line-stats-data="lineStatsData" />
      </div>
      <el-empty v-else description="点击查询按钮加载数据" />
    </div>
  </el-card>

  <el-dialog v-model="showExcludeDialog" title="排除配置" width="550px" destroy-on-close>
    <el-form label-width="100px">
      <el-form-item label="排除目录">
        <el-input v-model="excludeDirsText" type="textarea" :rows="4" placeholder="每行一个目录路径" />
      </el-form-item>
      <el-form-item label="排除规则">
        <el-input v-model="excludePatternsText" type="textarea" :rows="4" placeholder="每行一个 glob 规则" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="showExcludeDialog = false">取消</el-button>
      <el-button type="primary" @click="handleSaveExclude">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Download, Setting } from '@element-plus/icons-vue'
import { getLineStats, getLineStatsConfig, saveLineStatsConfig, exportStatsCsv } from '@/api/modules/stats'
import type { LineStatsResponse } from '@/types/stats'
import LineStatsCharts from '@/components/stats/LineStatsCharts.vue'

const props = defineProps<{
  repoKey: string
  statsBranches: string[]
  statsAuthors: { name: string; email: string }[]
  repoName: string
}>()

const lineStatsFilter = ref({ branch: '', author: '', since: '', until: '' })
const lineStatsData = ref<LineStatsResponse | null>(null)
const lineStatsLoading = ref(false)
const showExcludeDialog = ref(false)
const excludeDirsText = ref('')
const excludePatternsText = ref('')

async function loadLineStats() {
  try {
    lineStatsLoading.value = true
    const result = await getLineStats(props.repoKey, {
      branch: lineStatsFilter.value.branch || undefined,
      author: lineStatsFilter.value.author || undefined,
      since: lineStatsFilter.value.since || undefined,
      until: lineStatsFilter.value.until || undefined,
    })
    lineStatsData.value = result
    if (result && result.status === 'processing') {
      lineStatsLoading.value = false
      pollLineStats()
    } else {
      lineStatsLoading.value = false
    }
  } catch {
    lineStatsLoading.value = false
  }
}

function pollLineStats() {
  setTimeout(async () => {
    try {
      const result = await getLineStats(props.repoKey, {
        branch: lineStatsFilter.value.branch || undefined,
        author: lineStatsFilter.value.author || undefined,
        since: lineStatsFilter.value.since || undefined,
        until: lineStatsFilter.value.until || undefined,
      })
      lineStatsData.value = result
      if (result && result.status === 'processing') {
        pollLineStats()
      }
    } catch { /* ignore */ }
  }, 2000)
}

async function handleExportCsv() {
  try {
    const params: Record<string, string> = { type: 'lines' }
    if (lineStatsFilter.value.branch) params.branch = lineStatsFilter.value.branch
    const response = (await exportStatsCsv(props.repoKey, params)) as unknown as Blob
    const url = window.URL.createObjectURL(response)
    const a = document.createElement('a')
    a.href = url
    a.download = `${props.repoName || props.repoKey}-lines.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  } catch {
    ElMessage.error('导出失败')
  }
}

async function openExcludeConfig() {
  try {
    const config = await getLineStatsConfig(props.repoKey)
    excludeDirsText.value = (config.exclude_dirs || []).join('\n')
    excludePatternsText.value = (config.exclude_patterns || []).join('\n')
  } catch { /* ignore */ }
  showExcludeDialog.value = true
}

async function handleSaveExclude() {
  try {
    await saveLineStatsConfig(props.repoKey, {
      exclude_dirs: excludeDirsText.value.split('\n').map(s => s.trim()).filter(Boolean),
      exclude_patterns: excludePatternsText.value.split('\n').map(s => s.trim()).filter(Boolean),
    })
    ElMessage.success('排除配置已保存')
    showExcludeDialog.value = false
  } catch { /* handled */ }
}
</script>

<style scoped>
.filter-form {
  margin-bottom: var(--spacing-md);
}
.mb-4 {
  margin-bottom: var(--spacing-md);
}
</style>
