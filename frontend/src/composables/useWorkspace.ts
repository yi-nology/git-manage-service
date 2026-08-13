import { ref } from 'vue'
import {
  getWorkspaceStatus,
  getWorkspaceDiff,
  stageFiles,
  unstageFiles,
  commitChanges,
  pullWithResolve,
  getConflictDetail,
  markConflictResolved,
  aiResolveConflict,
  type WorkspaceStatus,
  type WorkspaceDiff,
} from '@/api/modules/workspace'
import { ElMessage } from 'element-plus'

export function useWorkspace(repo_key: string) {
  const loading = ref(false)
  const status = ref<WorkspaceStatus | null>(null)
  const diff = ref<WorkspaceDiff | null>(null)
  const selectedFile = ref('')
  const pulling = ref(false)
  const committing = ref(false)

  async function loadStatus() {
    loading.value = true
    try {
      status.value = await getWorkspaceStatus(repo_key)
    } catch {
      status.value = null
    } finally {
      loading.value = false
    }
  }

  async function loadDiff(file?: string, stagedOnly = false) {
    try {
      diff.value = await getWorkspaceDiff(repo_key, { file, stagedOnly })
    } catch {
      diff.value = null
    }
  }

  async function handleStageAll() {
    try {
      await stageFiles(repo_key, [], true)
      await loadStatus()
      if (selectedFile.value) await loadDiff(selectedFile.value)
      ElMessage.success('全部文件已暂存')
    } catch (e: any) {
      ElMessage.error(e?.message || '暂存失败')
    }
  }

  async function handleStageFile(file: string) {
    try {
      await stageFiles(repo_key, [file])
      await loadStatus()
      if (selectedFile.value) await loadDiff(selectedFile.value)
    } catch (e: any) {
      ElMessage.error(e?.message || '暂存失败')
    }
  }

  async function handleUnstageFile(file: string) {
    try {
      await unstageFiles(repo_key, [file])
      await loadStatus()
      if (selectedFile.value) await loadDiff(selectedFile.value)
    } catch (e: any) {
      ElMessage.error(e?.message || '取消暂存失败')
    }
  }

  async function handleUnstageAll() {
    try {
      await unstageFiles(repo_key, [], true)
      await loadStatus()
      if (selectedFile.value) await loadDiff(selectedFile.value)
      ElMessage.success('已取消全部暂存')
    } catch (e: any) {
      ElMessage.error(e?.message || '取消暂存失败')
    }
  }

  async function handleCommit(message: string, author_name: string, author_email: string, push: boolean, pushRemote: string) {
    if (!message) {
      ElMessage.warning('请输入提交信息')
      return null
    }
    committing.value = true
    try {
      const result = await commitChanges({
        repo_key: repo_key,
        stage_all: true,
        message,
        author_name: author_name || undefined,
        author_email: author_email || undefined,
        push,
        push_remote: pushRemote || undefined,
      })
      ElMessage.success(push ? '提交并推送成功' : '提交成功')
      await loadStatus()
      diff.value = null
      selectedFile.value = ''
      return result
    } catch (e: any) {
      ElMessage.error(e?.message || '提交失败')
      return null
    } finally {
      committing.value = false
    }
  }

  async function handlePull(remote = '', branch = '') {
    pulling.value = true
    try {
      const result = await pullWithResolve(repo_key, remote, branch)
      if (result.status === 'conflicts') {
        ElMessage.warning(`拉取完成，检测到 ${result.conflicts.length} 个冲突`)
      } else if (result.status === 'success') {
        ElMessage.success('拉取成功')
      } else if (result.status === 'fetched') {
        ElMessage.success('已获取远程更新')
      }
      await loadStatus()
      return result
    } catch (e: any) {
      ElMessage.error(e?.message || '拉取失败')
      return null
    } finally {
      pulling.value = false
    }
  }

  async function handleGetConflictDetail(file: string) {
    try {
      return await getConflictDetail(repo_key, file)
    } catch (e: any) {
      ElMessage.error(e?.message || '获取冲突详情失败')
      return null
    }
  }

  async function handleResolveConflict(file: string, resolved_content: string, stage = true) {
    try {
      await markConflictResolved(repo_key, file, resolved_content, stage)
      ElMessage.success('冲突已解决')
      await loadStatus()
    } catch (e: any) {
      ElMessage.error(e?.message || '解决冲突失败')
    }
  }

  async function handleAIResolve(file: string, ours_content: string, theirs_content: string, base_content: string, hint = '') {
    try {
      return await aiResolveConflict(repo_key, file, ours_content, theirs_content, base_content, hint)
    } catch (e: any) {
      ElMessage.error(e?.message || 'AI 解决冲突失败')
      return null
    }
  }

  return {
    loading,
    status,
    diff,
    selectedFile,
    pulling,
    committing,
    loadStatus,
    loadDiff,
    handleStageAll,
    handleStageFile,
    handleUnstageFile,
    handleUnstageAll,
    handleCommit,
    handlePull,
    handleGetConflictDetail,
    handleResolveConflict,
    handleAIResolve,
  }
}
