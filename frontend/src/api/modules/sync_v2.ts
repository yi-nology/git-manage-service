import request from '../request'
import type {
  SyncTask,
  SyncRun,
  SyncStats,
  WebhookRule,
  WebhookEvent,
  SyncConfigItem,
  CreateTaskRequest,
  UpdateTaskRequest,
  PreviewSyncRequest,
  PreviewSyncResponse
} from '@/types/sync_v2'

const V2_PREFIX = '/v2'

export function getSyncStats() {
  return request.get<unknown, SyncStats>(`${V2_PREFIX}/sync/stats`)
}

export function getSyncTasks(repoKey?: string) {
  return request.get<unknown, SyncTask[]>(`${V2_PREFIX}/sync/tasks`, {
    params: repoKey ? { repo_key: repoKey } : {},
  })
}

export function getSyncTask(key: string) {
  return request.get<unknown, SyncTask>(`${V2_PREFIX}/sync/task`, { params: { key } })
}

export function createSyncTask(data: CreateTaskRequest) {
  return request.post(`${V2_PREFIX}/sync/task`, data)
}

export function updateSyncTask(data: UpdateTaskRequest) {
  return request.put(`${V2_PREFIX}/sync/task`, data)
}

export function deleteSyncTask(key: string) {
  return request.delete(`${V2_PREFIX}/sync/task`, { data: { key } })
}

export function runSyncTask(taskKey: string) {
  return request.post(`${V2_PREFIX}/sync/task/run`, { key: taskKey })
}

export function batchRunTasks(taskKeys: string[]) {
  return request.post(`${V2_PREFIX}/sync/tasks/batch-run`, { task_keys: taskKeys })
}

export function previewSync(data: PreviewSyncRequest) {
  return request.post<unknown, PreviewSyncResponse>(`${V2_PREFIX}/sync/preview`, data)
}

export function getSyncHistory(taskKey?: string, limit?: number) {
  return request.get<unknown, SyncRun[]>(`${V2_PREFIX}/sync/history`, {
    params: {
      ...(taskKey ? { task_key: taskKey } : {}),
      ...(limit ? { limit } : {}),
    },
  })
}

export function deleteSyncHistory(id: number) {
  return request.delete(`${V2_PREFIX}/sync/history`, { data: { id } })
}

export function getWebhookRules(repoKey?: string) {
  return request.get<unknown, WebhookRule[]>(`${V2_PREFIX}/sync/webhook/rules`, {
    params: repoKey ? { repo_key: repoKey } : {},
  })
}

export function getWebhookRule(id: number) {
  return request.get<unknown, WebhookRule>(`${V2_PREFIX}/sync/webhook/rule`, { params: { id } })
}

export function createWebhookRule(data: Partial<WebhookRule>) {
  return request.post(`${V2_PREFIX}/sync/webhook/rule`, data)
}

export function updateWebhookRule(data: Partial<WebhookRule>) {
  return request.put(`${V2_PREFIX}/sync/webhook/rule`, data)
}

export function deleteWebhookRule(id: number) {
  return request.delete(`${V2_PREFIX}/sync/webhook/rule`, { data: { id } })
}

export function getWebhookEvents(repoKey?: string, limit?: number) {
  return request.get<unknown, WebhookEvent[]>(`${V2_PREFIX}/sync/webhook/events`, {
    params: {
      ...(repoKey ? { repo_key: repoKey } : {}),
      ...(limit ? { limit } : {}),
    },
  })
}

export function retryWebhookEvent(id: number) {
  return request.post(`${V2_PREFIX}/sync/webhook/event/retry`, { id })
}

export function getSyncConfig() {
  return request.get<unknown, SyncConfigItem[]>(`${V2_PREFIX}/sync/config`)
}

export function updateSyncConfig(items: SyncConfigItem[]) {
  return request.put(`${V2_PREFIX}/sync/config`, items)
}

export const syncV2Api = {
  getStats: getSyncStats,
  listTasks: getSyncTasks,
  getTask: getSyncTask,
  createTask: createSyncTask,
  updateTask: updateSyncTask,
  deleteTask: deleteSyncTask,
  runTask: runSyncTask,
  batchRunTasks,
  previewSync,
  listHistory: getSyncHistory,
  deleteHistory: deleteSyncHistory,
  listRules: getWebhookRules,
  getRule: getWebhookRule,
  createRule: createWebhookRule,
  updateRule: updateWebhookRule,
  deleteRule: deleteWebhookRule,
  listEvents: getWebhookEvents,
  retryEvent: retryWebhookEvent,
  getConfig: getSyncConfig,
  updateConfig: updateSyncConfig,
}
