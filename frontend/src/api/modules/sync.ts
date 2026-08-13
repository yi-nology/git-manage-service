import request from '../request'
import type { SyncTaskDTO, CreateSyncTaskReq, UpdateSyncTaskReq, ExecuteSyncReq, SyncRunDTO, PreviewSyncReq, PreviewSyncResponse } from '@/types/sync'

export function getSyncTasks(repo_key?: string) {
  return request.get<unknown, SyncTaskDTO[]>('/sync/tasks', {
    params: repo_key ? { repo_key: repo_key } : {},
  })
}

export function getSyncTask(key: string) {
  return request.get<unknown, SyncTaskDTO>('/sync/task', { params: { key } })
}

export function createSyncTask(data: CreateSyncTaskReq) {
  return request.post('/sync/task/create', data)
}

export function updateSyncTask(data: UpdateSyncTaskReq) {
  return request.post('/sync/task/update', data)
}

export function deleteSyncTask(key: string) {
  return request.post('/sync/task/delete', { key })
}

export function runSyncTask(task_key: string) {
  return request.post('/sync/run', { task_key: task_key })
}

export function executeSyncOnce(data: ExecuteSyncReq) {
  return request.post<unknown, { task_key: string }>('/sync/execute', data)
}

export function getSyncHistory(repo_key?: string) {
  return request.get<unknown, SyncRunDTO[]>('/sync/history', {
    params: repo_key ? { repo_key: repo_key } : {},
  })
}

export function deleteSyncHistory(id: number) {
  return request.post('/sync/history/delete', { id })
}

export function previewSync(data: PreviewSyncReq) {
  return request.post<unknown, PreviewSyncResponse>('/sync/preview', data)
}

export function batchSync(taskKeys: string[]) {
  return request.post('/sync/batch', { task_keys: taskKeys })
}

// 分析仓库以获取同步建议
export function analyzeRepoForSync(repoPath: string, repo_key: string) {
  return request.post<unknown, { message: string }>('/sync/analyze-repo', {
    repoPath,
    repo_key
  })
}
