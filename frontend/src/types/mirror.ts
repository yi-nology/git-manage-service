export interface MirrorDTO {
  id: number
  repoId: number
  repoName: string
  repoKey: string
  mirrorType: 'pull' | 'push'
  remoteUrl: string
  remoteName: string
  credentialId: number | null
  credentialName?: string
  branchFilter: string
  syncInterval: number
  cronExpr: string
  syncOnPush: boolean
  gitForce: boolean
  gitPrune: boolean
  gitTags: boolean
  enabled: boolean
  status: 'active' | 'syncing' | 'failed' | 'paused'
  lastSyncAt: string | null
  lastError: string
  retryCount: number
  nextSyncAt: string | null
  webhookToken: string
  createdAt: string
  updatedAt: string
}

export interface CreateMirrorReq {
  repoId: number
  mirrorType: 'pull' | 'push'
  remoteUrl: string
  remoteName?: string
  credentialId?: number | null
  branchFilter?: string
  syncInterval?: number
  cronExpr?: string
  syncOnPush?: boolean
  gitForce?: boolean
  gitPrune?: boolean
  gitTags?: boolean
  enabled?: boolean
}

export interface UpdateMirrorReq {
  remoteUrl?: string
  remoteName?: string
  credentialId?: number | null
  branchFilter?: string
  syncInterval?: number
  cronExpr?: string
  syncOnPush?: boolean
  gitForce?: boolean
  gitPrune?: boolean
  gitTags?: boolean
  enabled?: boolean
}

export interface MirrorSyncLogDTO {
  id: number
  mirrorId: number
  triggerType: 'manual' | 'cron' | 'webhook' | 'push_event'
  status: 'pending' | 'running' | 'success' | 'failed'
  startedAt: string | null
  finishedAt: string | null
  durationMs: number
  branchesSynced: number
  commitsPushed: number
  errorMessage: string
  detailLog?: string
  createdAt: string
}

export interface AnalyzeRemoteResponse {
  reachable: boolean
  branches: string[]
  defaultBranch: string
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
