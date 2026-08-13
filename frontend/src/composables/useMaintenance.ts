import { ref, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getRepoHealth,
  slimRepo,
  gcRepo,
  addGitignore as addGitignoreApi,
  getTaskStatus,
  getMaintenanceRecords,
  analyzeMaintenanceAI,
  previewPrefixSlim,
  slimByPrefix,
  forcePushRemotes,
} from '@/api/modules/maintenance'
import type { RepoHealthReport, LargeFileEntry, MaintenanceRecordDTO, MaintenanceRecordListResponse, MaintenanceAIAnalysisResponse, FileAIRecommendation, PrefixSlimPreview } from '@/api/modules/maintenance'

export function useMaintenance(repo_key: string) {
  const healthLoading = ref(false)
  const healthReport = ref<RepoHealthReport | null>(null)
  const selectedFiles = ref<LargeFileEntry[]>([])
  const gcLoading = ref(false)
  const gitignoreLoading = ref(false)
  const task_id = ref('')
  const task_status = ref('')
  const taskError = ref('')
  const taskLogs = ref<string[]>([])
  const aiLoading = ref(false)
  const aiResult = ref<MaintenanceAIAnalysisResponse | null>(null)
  const aiRecommendationMap = ref<Map<string, FileAIRecommendation>>(new Map())
  const thresholdKB = ref(100)
  const excludePatterns = ref<string[]>(['.png', '.jpg', '.jpeg', '.gif', '.svg', '.ico', '.woff', '.woff2', '.ttf', '.eot', 'docs/', 'dist/', 'node_modules/'])

  const recordsLoading = ref(false)
  const records = ref<MaintenanceRecordDTO[]>([])
  const recordsTotal = ref(0)
  const recordsPage = ref(1)
  const recordsPageSize = ref(10)

  let pollTimer: ReturnType<typeof setInterval> | null = null

  onUnmounted(() => {
    if (pollTimer) clearInterval(pollTimer)
  })

  async function analyzeWithAI() {
    if (!healthReport.value || healthReport.value.large_files.length === 0) {
      ElMessage.warning('请先体检并确保有大文件')
      return
    }
    aiLoading.value = true
    aiResult.value = null
    aiRecommendationMap.value = new Map()
    try {
      const paths = healthReport.value.large_files.map(f => f.path)
      const thresholdBytes = thresholdKB.value * 1024
      const res = await analyzeMaintenanceAI(repo_key, paths, thresholdBytes) as any as MaintenanceAIAnalysisResponse
      aiResult.value = res
      const map = new Map<string, FileAIRecommendation>()
      for (const r of res.recommendations || []) {
        map.set(r.path, r)
      }
      aiRecommendationMap.value = map
    } catch (e: any) {
      ElMessage.error('AI 分析失败: ' + (e.message || '未知错误'))
    } finally {
      aiLoading.value = false
    }
  }

  function acceptAIRecommendations() {
    if (!aiResult.value) return
    const safePaths = new Set(
      aiResult.value.recommendations
        .filter(r => r.recommendation === 'safe_to_delete')
        .map(r => r.path)
    )
    const matched = healthReport.value?.large_files.filter(f => safePaths.has(f.path)) || []
    selectedFiles.value = matched
    ElMessage.success(`已采纳 ${matched.length} 个安全删除建议`)
  }

  function formatFileSize(bytes: number) {
    if (bytes >= 1024 * 1024 * 1024) return (bytes / 1024 / 1024 / 1024).toFixed(1) + ' GB'
    if (bytes >= 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
    if (bytes >= 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return bytes + ' B'
  }

  async function loadHealth() {
    healthLoading.value = true
    task_id.value = ''
    task_status.value = ''
    try {
      const thresholdBytes = thresholdKB.value * 1024
      healthReport.value = await getRepoHealth(repo_key, thresholdBytes, excludePatterns.value) as any
    } catch (e: any) {
      ElMessage.error('体检失败: ' + (e.message || '未知错误'))
    } finally {
      healthLoading.value = false
    }
  }

  function handleSelection(rows: LargeFileEntry[]) {
    selectedFiles.value = rows
  }

  async function handleAddGitignore() {
    const paths = selectedFiles.value.map(f => f.path)
    gitignoreLoading.value = true
    try {
      await addGitignoreApi(repo_key, paths)
      ElMessage.success(`已将 ${paths.length} 个文件添加到 .gitignore`)
    } catch (e: any) {
      ElMessage.error('添加失败: ' + (e.message || '未知错误'))
    } finally {
      gitignoreLoading.value = false
    }
  }

  async function handleSlimConfirm() {
    const paths = selectedFiles.value.map(f => f.path)
    const total_size = selectedFiles.value.reduce((sum, f) => sum + f.size_bytes, 0)

    try {
      await ElMessageBox.confirm(
        `即将从历史中删除 ${paths.length} 个文件 (${formatFileSize(total_size)})，此操作会重写 git 历史，不可恢复！`,
        '确认瘦身',
        { confirmButtonText: '确认删除', cancelButtonText: '取消', type: 'warning' }
      )
    } catch { return }

    try {
      const res = await slimRepo(repo_key, paths, true) as any
      task_id.value = res.task_id
      task_status.value = 'running'
      taskLogs.value = []
      taskError.value = ''
      startPolling()
    } catch (e: any) {
      ElMessage.error('瘦身失败: ' + (e.message || '未知错误'))
    }
  }

  async function handleGC() {
    gcLoading.value = true
    try {
      await ElMessageBox.confirm('执行 git gc --aggressive --prune=now？', '垃圾回收', { type: 'warning' })
    } catch { gcLoading.value = false; return }

    try {
      const res = await gcRepo(repo_key) as any
      task_id.value = res.task_id
      task_status.value = 'running'
      taskLogs.value = []
      taskError.value = ''
      startPolling()
    } catch (e: any) {
      ElMessage.error('GC 失败: ' + (e.message || '未知错误'))
    } finally {
      gcLoading.value = false
    }
  }

  function startPolling() {
    if (pollTimer) clearInterval(pollTimer)
    pollTimer = setInterval(async () => {
      try {
        const task = await getTaskStatus(task_id.value) as any
        task_status.value = task.status
        taskLogs.value = task.progress || []
        taskError.value = task.error || ''
        if (task.status === 'success' || task.status === 'failed') {
          if (pollTimer) clearInterval(pollTimer)
          pollTimer = null
          if (task.status === 'success') {
            ElMessage.success('操作完成')
            loadHealth()
            loadRecords()
          }
        }
      } catch {
        if (pollTimer) clearInterval(pollTimer)
        pollTimer = null
      }
    }, 2000)
  }

  async function loadRecords(page?: number) {
    if (page !== undefined) recordsPage.value = page
    recordsLoading.value = true
    try {
      const res = await getMaintenanceRecords(repo_key, recordsPage.value, recordsPageSize.value) as any as MaintenanceRecordListResponse
      records.value = res.records || []
      recordsTotal.value = res.total || 0
    } catch {
      records.value = []
      recordsTotal.value = 0
    } finally {
      recordsLoading.value = false
    }
  }

  const prefixInput = ref('')
  const prefixTags = ref<string[]>([])
  const prefixPreview = ref<PrefixSlimPreview | null>(null)
  const prefixPreviewLoading = ref(false)
  const prefixSlimForcePush = ref(true)

  async function addPrefix() {
    const v = prefixInput.value.trim()
    if (v && !prefixTags.value.includes(v)) {
      prefixTags.value.push(v)
    }
    prefixInput.value = ''
  }

  function removePrefix(idx: number) {
    prefixTags.value.splice(idx, 1)
  }

  async function previewPrefix() {
    if (prefixTags.value.length === 0) {
      ElMessage.warning('请先添加前缀')
      return
    }
    prefixPreviewLoading.value = true
    prefixPreview.value = null
    try {
      prefixPreview.value = await previewPrefixSlim(repo_key, prefixTags.value) as any as PrefixSlimPreview
    } catch (e: any) {
      ElMessage.error('预览失败: ' + (e.message || '未知错误'))
    } finally {
      prefixPreviewLoading.value = false
    }
  }

  async function handlePrefixSlimConfirm() {
    if (prefixTags.value.length === 0) {
      ElMessage.warning('请先添加前缀')
      return
    }

    const total_size = prefixPreview.value ? formatFileSize(prefixPreview.value.total_bytes) : '未知'
    const count = prefixPreview.value ? prefixPreview.value.total_count : 0

    try {
      await ElMessageBox.confirm(
        `即将删除所有匹配前缀 [${prefixTags.value.join(', ')}] 的文件（共 ${count} 个，${total_size}），此操作会重写 git 历史，不可恢复！${prefixSlimForcePush.value ? '\n\n瘦身完成后将强制推送到所有绑定的远端。' : ''}`,
        '确认前缀瘦身',
        { confirmButtonText: '确认删除', cancelButtonText: '取消', type: 'warning' }
      )
    } catch { return }

    try {
      const res = await slimByPrefix(repo_key, prefixTags.value, true, prefixSlimForcePush.value) as any
      task_id.value = res.task_id
      task_status.value = 'running'
      taskLogs.value = []
      taskError.value = ''
      startPolling()
    } catch (e: any) {
      ElMessage.error('瘦身失败: ' + (e.message || '未知错误'))
    }
  }

  async function handleForcePush() {
    try {
      await ElMessageBox.confirm(
        '将强制推送所有本地分支到所有绑定的远端仓库，此操作会覆盖远端历史！',
        '确认强制推送',
        { confirmButtonText: '确认推送', cancelButtonText: '取消', type: 'warning' }
      )
    } catch { return }

    try {
      const res = await forcePushRemotes(repo_key) as any
      task_id.value = res.task_id
      task_status.value = 'running'
      taskLogs.value = []
      taskError.value = ''
      startPolling()
    } catch (e: any) {
      ElMessage.error('推送失败: ' + (e.message || '未知错误'))
    }
  }

  return {
    healthLoading,
    healthReport,
    selectedFiles,
    gcLoading,
    gitignoreLoading,
    task_id,
    task_status,
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
  }
}
