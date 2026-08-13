export interface MirrorDTO {
  id: number
  repo_id: number
  repo_name: string
  repo_key: string
  mirror_type: 'pull' | 'push'
  remote_url: string
  remote_name: string
  credential_id: number | null
  credential_name?: string
  branch_filter: string
  sync_interval: number
  cron_expr: string
  sync_on_push: boolean
  git_force: boolean
  git_prune: boolean
  git_tags: boolean
  enabled: boolean
  status: 'active' | 'syncing' | 'failed' | 'paused'
  last_sync_at: string | null
  last_error: string
  retry_count: number
  next_sync_at: string | null
  webhook_token: string
  created_at: string
  updated_at: string
}

export interface CreateMirrorReq {
  repo_id: number
  mirror_type: 'pull' | 'push'
  remote_url: string
  remote_name?: string
  credential_id?: number | null
  branch_filter?: string
  sync_interval?: number
  cron_expr?: string
  sync_on_push?: boolean
  git_force?: boolean
  git_prune?: boolean
  git_tags?: boolean
  enabled?: boolean
}

export interface UpdateMirrorReq {
  remote_url?: string
  remote_name?: string
  credential_id?: number | null
  branch_filter?: string
  sync_interval?: number
  cron_expr?: string
  sync_on_push?: boolean
  git_force?: boolean
  git_prune?: boolean
  git_tags?: boolean
  enabled?: boolean
}

export interface MirrorSyncLogDTO {
  id: number
  mirror_id: number
  trigger_type: 'manual' | 'cron' | 'webhook' | 'push_event'
  status: 'pending' | 'running' | 'success' | 'failed'
  started_at: string | null
  finished_at: string | null
  duration_ms: number
  branches_synced: number
  commits_pushed: number
  error_message: string
  detail_log?: string
  created_at: string
}

export interface AnalyzeRemoteResponse {
  reachable: boolean
  branches: string[]
  default_branch: string
  protocol: string
}

export const MIRROR_STATUS_MAP: Record<string, { label: string; type: string }> = {
  active: { label: '活跃', type: 'success' },
  syncing: { label: '同步中', type: 'warning' },
  failed: { label: '失败', type: 'danger' },
  paused: { label: '已暂停', type: 'info' },
}

export const TRIGGER_TYPE_MAP: Record<string, string> = {
  manual: '手动',
  cron: '定时',
  webhook: 'Webhook',
  push_event: 'Push事件',
}
