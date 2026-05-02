import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listIdentities,
  createIdentity,
  updateIdentity,
  deleteIdentity,
  activateIdentity,
  getRepoAuthorConfig,
  setRepoAuthorConfig,
  scanAuthor,
  fixAuthorAll,
  fixAuthor,
} from '@/api/modules/author'
import type { AuthorIdentityDTO, AliasEntry, RepoAuthorConfigDTO, MismatchedCommit } from '@/api/modules/author'
import { getTaskStatus } from '@/api/modules/maintenance'

export function useAuthorIdentity() {
  const identities = ref<AuthorIdentityDTO[]>([])
  const loading = ref(false)

  async function loadIdentities() {
    loading.value = true
    try {
      identities.value = (await listIdentities() as any) || []
    } catch { identities.value = [] }
    finally { loading.value = false }
  }

  async function handleCreate(data: { canonicalName: string; canonicalEmail: string; aliases: AliasEntry[] }) {
    try {
      await createIdentity(data) as any
      ElMessage.success('身份创建成功')
      await loadIdentities()
    } catch (e: any) {
      ElMessage.error('创建失败: ' + (e.message || '未知错误'))
    }
  }

  async function handleUpdate(id: number, data: { canonicalName: string; canonicalEmail: string; aliases: AliasEntry[] }) {
    try {
      await updateIdentity(id, data) as any
      ElMessage.success('身份更新成功')
      await loadIdentities()
    } catch (e: any) {
      ElMessage.error('更新失败: ' + (e.message || '未知错误'))
    }
  }

  async function handleDelete(id: number) {
    try {
      await ElMessageBox.confirm('确认删除此身份？关联的仓库将恢复使用全局默认。', '删除身份', { type: 'warning' })
    } catch { return }
    try {
      await deleteIdentity(id) as any
      ElMessage.success('删除成功')
      await loadIdentities()
    } catch (e: any) {
      ElMessage.error('删除失败: ' + (e.message || '未知错误'))
    }
  }

  async function handleActivate(id: number) {
    try {
      await activateIdentity(id) as any
      ElMessage.success('已激活并更新 ~/.gitconfig')
      await loadIdentities()
    } catch (e: any) {
      ElMessage.error('激活失败: ' + (e.message || '未知错误'))
    }
  }

  return {
    identities,
    loading,
    loadIdentities,
    handleCreate,
    handleUpdate,
    handleDelete,
    handleActivate,
  }
}

export function useAuthorFix(repoKey: string) {
  const repoConfig = ref<RepoAuthorConfigDTO | null>(null)
  const configLoading = ref(false)

  const scanResult = ref<MismatchedCommit[]>([])
  const scanLoading = ref(false)
  const totalCommits = ref(0)
  const selectedCommits = ref<MismatchedCommit[]>([])

  const taskId = ref('')
  const taskStatus = ref('')
  const taskLogs = ref<string[]>([])
  const taskError = ref('')

  let pollTimer: ReturnType<typeof setInterval> | null = null

  async function loadRepoConfig() {
    configLoading.value = true
    try {
      repoConfig.value = await getRepoAuthorConfig(repoKey) as any
    } catch { repoConfig.value = null }
    finally { configLoading.value = false }
  }

  async function setConfig(identityId: number | null) {
    try {
      if (identityId === null) {
        await setRepoAuthorConfig(repoKey, null, true) as any
      } else {
        await setRepoAuthorConfig(repoKey, identityId) as any
      }
      ElMessage.success('仓库作者身份已更新')
      await loadRepoConfig()
    } catch (e: any) {
      ElMessage.error('设置失败: ' + (e.message || '未知错误'))
    }
  }

  async function scan() {
    scanLoading.value = true
    try {
      const result = await scanAuthor(repoKey) as any
      scanResult.value = result.commits || []
      totalCommits.value = result.totalCommits || 0
    } catch (e: any) {
      ElMessage.error('扫描失败: ' + (e.message || '未知错误'))
    } finally {
      scanLoading.value = false
    }
  }

  async function fixAll(pushRemote = '') {
    try {
      await ElMessageBox.confirm(
        '即将重写所有匹配提交的作者信息。此操作不可恢复！' + (pushRemote ? '\n\n修复后将 force push 到 ' + pushRemote : ''),
        '确认一键修复',
        { confirmButtonText: '确认修复', cancelButtonText: '取消', type: 'warning' }
      )
    } catch { return }
    try {
      const res = await fixAuthorAll(repoKey, pushRemote) as any
      taskId.value = res.taskId
      taskStatus.value = 'running'
      taskLogs.value = []
      taskError.value = ''
      startPolling()
    } catch (e: any) {
      ElMessage.error('修复失败: ' + (e.message || '未知错误'))
    }
  }

  async function fixSelected(pushRemote = '') {
    if (selectedCommits.value.length === 0) {
      ElMessage.warning('请先选择要修复的提交')
      return
    }
    try {
      await ElMessageBox.confirm(
        `即将修复 ${selectedCommits.value.length} 个提交的作者信息。此操作不可恢复！` + (pushRemote ? '\n\n修复后将 force push 到 ' + pushRemote : ''),
        '确认修复选中',
        { confirmButtonText: '确认修复', cancelButtonText: '取消', type: 'warning' }
      )
    } catch { return }
    try {
      const hashes = selectedCommits.value.map(c => c.hash)
      const res = await fixAuthor(repoKey, hashes, pushRemote) as any
      taskId.value = res.taskId
      taskStatus.value = 'running'
      taskLogs.value = []
      taskError.value = ''
      startPolling()
    } catch (e: any) {
      ElMessage.error('修复失败: ' + (e.message || '未知错误'))
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
            ElMessage.success('作者修复完成')
            scan()
          }
        }
      } catch {
        if (pollTimer) clearInterval(pollTimer)
        pollTimer = null
      }
    }, 2000)
  }

  function handleSelection(rows: MismatchedCommit[]) {
    selectedCommits.value = rows
  }

  return {
    repoConfig,
    configLoading,
    scanResult,
    scanLoading,
    totalCommits,
    selectedCommits,
    taskId,
    taskStatus,
    taskLogs,
    taskError,
    loadRepoConfig,
    setConfig,
    scan,
    fixAll,
    fixSelected,
    handleSelection,
  }
}
