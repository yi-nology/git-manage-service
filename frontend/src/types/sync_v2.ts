// ==================== Task Types ====================

export interface SyncTask {
  id: number
  key: string
  name: string
  source_repo_key: string
  source_branch: string
  target_repo_key: string
  target_branch: string
  sync_mode: 'single' | 'bidirectional'
  cron: string
  enabled: boolean
  git_tags: boolean
  git_force: boolean
  git_prune: boolean
  git_no_verify: boolean
  push_options: string
  last_run_at?: string
  last_status?: 'success' | 'failed' | 'running'
  created_at: string
  updated_at: string
}

export interface CreateTaskRequest {
  name: string
  source_repo_key: string
  source_branch: string
  target_repo_key: string
  target_branch: string
  sync_mode: 'single' | 'bidirectional'
  cron?: string
  enabled?: boolean
  git_tags?: boolean
  git_force?: boolean
  git_prune?: boolean
  git_no_verify?: boolean
  push_options?: string
}

export interface UpdateTaskRequest extends Partial<CreateTaskRequest> {
  key: string
}

// ==================== Run / History Types ====================

export interface SyncRun {
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

export interface PreviewSyncRequest {
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

export interface CreateRuleRequest {
  name: string
  repo_key: string
  event_type: string
  branch_pattern: string
  action: 'sync' | 'notify'
  sync_task_keys?: string
  min_interval?: number
  enabled?: boolean
  description?: string
}

export interface UpdateRuleRequest extends Partial<CreateRuleRequest> {
  id: number
}

// ==================== Webhook Event Types ====================

export interface WebhookEvent {
  id: number
  rule_id?: number
  repo_key: string
  event_type: string
  payload?: string
  status: 'pending' | 'processing' | 'success' | 'failed'
  triggered_at: string
  executed_at?: string
  error_message?: string
}

// ==================== Config Types ====================

export interface SyncConfigItem {
  key: string
  value: string
  scope: string
  description?: string
}
