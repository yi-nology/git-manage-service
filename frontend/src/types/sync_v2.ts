// ==================== Task Types ====================

export interface SyncTaskDTO {
  id: number
  key: string
  name: string
  source_repo_key: string
  source_branch: string
  target_repo_key: string
  target_branch: string
  sync_mode: 'single' | 'all-branch'
  cron: string
  enabled: boolean
  git_tags: boolean
  git_force: boolean
  git_prune: boolean
  git_no_verify: boolean
  last_run_at?: string
  last_status?: 'success' | 'failed' | 'running'
  created_at: string
  updated_at: string
}

export interface CreateSyncTaskReq {
  name: string
  source_repo_key: string
  source_branch: string
  target_repo_key: string
  target_branch: string
  sync_mode: 'single' | 'all-branch'
  cron?: string
  enabled?: boolean
  git_tags?: boolean
  git_force?: boolean
  git_prune?: boolean
  git_no_verify?: boolean
}

export interface UpdateSyncTaskReq extends Partial<CreateSyncTaskReq> {
  key: string
}

// ==================== Run / History Types ====================

export interface SyncRunDTO {
  id: number
  task_key: string
  trigger_source: 'manual' | 'cron' | 'webhook'
  status: 'running' | 'success' | 'failed' | 'conflict'
  start_time: string
  end_time?: string
  details?: string
  error_message?: string
  created_at: string
}

// ==================== Stats Types ====================

export interface SyncStats {
  total_tasks: number
  enabled_tasks: number
  today_runs: number
  failed_runs: number
  running_tasks: number
}

// ==================== Preview Types ====================

export interface PreviewSyncReq {
  source_repo_key: string
  source_branch: string
  target_repo_key: string
  target_branch: string
  git_force?: boolean
}

export interface PreviewSyncResponse {
  can_sync: boolean
  source_exists: boolean
  target_exists: boolean
  is_fast_forward: boolean
  commits_count: number
  message: string
  changes?: string[]
}

// ==================== Webhook Rule Types ====================

export interface WebhookRule {
  id: number
  name: string
  repo_key: string
  event_type: string
  branch_pattern: string
  action: 'sync' | 'notify'
  sync_task_keys: string
  min_interval: number
  enabled: boolean
  description?: string
  created_at: string
  updated_at: string
}

// ==================== Webhook Event Types ====================

export interface WebhookEvent {
  id: number
  event_id: string
  repo_key: string
  event_type: string
  source: string
  actor_name: string
  branch: string
  commit_sha: string
  payload: Record<string, any>
  status: 'pending' | 'processing' | 'success' | 'failed'
  error_message?: string
  processed_at?: string
  created_at: string
}

// ==================== Config Types ====================

export interface SyncConfigItem {
  key: string
  value: string
  scope: 'global' | 'repo' | 'task'
  description?: string
}
