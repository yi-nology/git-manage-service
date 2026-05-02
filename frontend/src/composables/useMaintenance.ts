import { ref, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getRepoHealth,
  slimRepo,
  gcRepo,
  addGitignore as addGitignoreApi,
  getTaskStatus,
  getMaintenanceRecords,
} from '@/api/modules/maintenance'
import type { RepoHealthReport, LargeFileEntry, MaintenanceRecordDTO, MaintenanceRecordListResponse } from '@/api/modules/maintenance'

export function useMaintenance(repoKey: string) {
  const healthLoading = ref(false)
  const healthReport = ref<RepoHealthReport | null>(null)
  const selectedFiles = ref<LargeFileEntry[]>([])
  const gcLoading = ref(false)
  const gitignoreLoading = ref(false)
  const taskId = ref('')
  const taskStatus = ref('')
  const taskError = ref('')
  const taskLogs = ref<string[]>([])

  const recordsLoading = ref(false)
  const records = ref<MaintenanceRecordDTO[]>([])
  const recordsTotal = ref(0)
  const recordsPage = ref(1)
  const recordsPageSize = ref(10)

  let pollTimer: ReturnType<typeof setInterval> | null = null

  onUnmounted(() => {
    if (pollTimer) clearInterval(pollTimer)
  })

  function formatFileSize(bytes: number) {
    if (bytes >= 1024 * 1024 * 1024) return (bytes / 1024 / 1024 / 1024).toFixed(1) + ' GB'
    if (bytes >= 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
    if (bytes >= 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return bytes + ' B'
  }

  async function loadHealth() {
    healthLoading.value = true
    taskId.value = ''
    taskStatus.value = ''
    try {
      healthReport.value = await getRepoHealth(repoKey) as any
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
      await addGitignoreApi(repoKey, paths)
      ElMessage.success(`已将 ${paths.length} 个文件添加到 .gitignore`)
    } catch (e: any) {
      ElMessage.error('添加失败: ' + (e.message || '未知错误'))
    } finally {
      gitignoreLoading.value = false
    }
  }

  async function handleSlimConfirm() {
    const paths = selectedFiles.value.map(f => f.path)
    const totalSize = selectedFiles.value.reduce((sum, f) => sum + f.sizeBytes, 0)

    try {
      await ElMessageBox.confirm(
        `即将从历史中删除 ${paths.length} 个文件 (${formatFileSize(totalSize)})，此操作会重写 git 历史，不可恢复！`,
        '确认瘦身',
        { confirmButtonText: '确认删除', cancelButtonText: '取消', type: 'warning' }
      )
    } catch { return }

    try {
      const res = await slimRepo(repoKey, paths, true) as any
      taskId.value = res.taskId
      taskStatus.value = 'running'
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
      const res = await gcRepo(repoKey) as any
      taskId.value = res.taskId
      taskStatus.value = 'running'
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
        const task = await getTaskStatus(taskId.value) as any
        taskStatus.value = task.status
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
      const res = await getMaintenanceRecords(repoKey, recordsPage.value, recordsPageSize.value) as any as MaintenanceRecordListResponse
      records.value = res.records || []
      recordsTotal.value = res.total || 0
    } catch {
      records.value = []
      recordsTotal.value = 0
    } finally {
      recordsLoading.value = false
    }
  }

  return {
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
  }
}
