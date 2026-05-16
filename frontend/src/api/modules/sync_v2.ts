import request from '../request'
import type {
  SyncTaskDTO,
  SyncRunDTO,
  SyncStats,
  WebhookRule,
  WebhookEvent,
  SyncConfigItem,
  CreateSyncTaskReq,
  UpdateSyncTaskReq,
  PreviewSyncReq,
  PreviewSyncResponse
} from '@/types/sync_v2'

export function getSyncStats() {
  return request.get<unknown, SyncStats>('/sync/stats')
}

export function getSyncTasks(repoKey?: string) {
  return request.get<unknown, SyncTaskDTO[]>('/sync/tasks', {
    params: repoKey ? { repo_key: repoKey } : {},
  })
}

export function getSyncTask(key: string) {
  return request.get<unknown, SyncTaskDTO>('/sync/task', { params: { key } })
}

export function createSyncTask(data: CreateSyncTaskReq) {
  return request.post('/sync/task', data)
}

export function updateSyncTask(data: UpdateSyncTaskReq) {
  return request.put('/sync/task', data)
}

export function deleteSyncTask(key: string) {
  return request.delete('/sync/task', { data: { key } })
}

export function runSyncTask(taskKey: string) {
  return request.post('/sync/task/run', { key: taskKey })
}

export function batchRunTasks(taskKeys: string[]) {
  return request.post('/sync/tasks/batch-run', { task_keys: taskKeys })
}

export function previewSync(data: PreviewSyncReq) {
  return request.post<unknown, PreviewSyncResponse>('/sync/preview', data)
}

export function getSyncHistory(taskKey?: string, limit?: number) {
  return request.get<unknown, SyncRunDTO[]>('/sync/history', {
    params: {
      ...(taskKey ? { task_key: taskKey } : {}),
      ...(limit ? { limit } : {}),
    },
  })
}

export function deleteSyncHistory(id: number) {
  return request.delete('/sync/history', { data: { id } })
}

export function getWebhookRules(repoKey?: string) {
  return request.get<unknown, WebhookRule[]>('/sync/webhook/rules', {
    params: repoKey ? { repo_key: repoKey } : {},
  })
}

export function getWebhookRule(id: number) {
  return request.get<unknown, WebhookRule>('/sync/webhook/rule', { params: { id } })
}

export function createWebhookRule(data: Partial<WebhookRule>) {
  return request.post('/sync/webhook/rule', data)
}

export function updateWebhookRule(data: Partial<WebhookRule>) {
  return request.put('/sync/webhook/rule', data)
}

export function deleteWebhookRule(id: number) {
  return request.delete('/sync/webhook/rule', { data: { id } })
}

export function getWebhookEvents(repoKey?: string, limit?: number) {
  return request.get<unknown, WebhookEvent[]>('/sync/webhook/events', {
    params: {
      ...(repoKey ? { repo_key: repoKey } : {}),
      ...(limit ? { limit } : {}),
    },
  })
}

export function retryWebhookEvent(id: number) {
  return request.post('/sync/webhook/event/retry', { id })
}

export function getSyncConfig() {
  return request.get<unknown, SyncConfigItem[]>('/sync/config')
}

export function updateSyncConfig(items: SyncConfigItem[]) {
  return request.put('/sync/config', items)
}
