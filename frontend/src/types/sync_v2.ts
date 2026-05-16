// ==================== Task Types ====================

export interface SyncTask {
  id: number
  key: string
  name: string
  sourceRepoKey: string
  sourceBranch: string
  targetRepoKey: string
  targetBranch: string
  syncMode: 'single' | 'bidirectional'
  cron: string
  enabled: boolean
  gitTags: boolean
  gitForce: boolean
  gitPrune: boolean
  gitNoVerify: boolean
  pushOptions: string
  lastRunAt?: string
  lastStatus?: 'success' | 'failed' | 'running'
  createdAt: string
  updatedAt: string
}

export interface CreateTaskRequest {
  name: string
  sourceRepoKey: string
  sourceBranch: string
  targetRepoKey: string
  targetBranch: string
  syncMode: 'single' | 'bidirectional'
  cron?: string
  enabled?: boolean
  gitTags?: boolean
  gitForce?: boolean
  gitPrune?: boolean
  gitNoVerify?: boolean
  pushOptions?: string
}

export interface UpdateTaskRequest extends Partial<CreateTaskRequest> {
  key: string
}

// ==================== Run / History Types ====================

export interface SyncRun {
  id: number
  taskKey: string
  triggerSource: 'manual' | 'cron' | 'webhook'
  status: 'running' | 'success' | 'failed' | 'conflict'
  startTime: string
  endTime?: string
  details?: string
  errorMessage?: string
  createdAt: string
}

// ==================== Stats Types ====================

export interface SyncStats {
  totalTasks: number
  enabledTasks: number
  todayRuns: number
  failedRuns: number
  runningTasks: number
}

// ==================== Preview Types ====================

export interface PreviewSyncRequest {
  sourceRepoKey: string
  sourceBranch: string
  targetRepoKey: string
  targetBranch: string
  gitForce?: boolean
}

export interface PreviewSyncResponse {
  canSync: boolean
  sourceExists: boolean
  targetExists: boolean
  isFastForward: boolean
  commitsCount: number
  message: string
  changes?: string[]
}

// ==================== Webhook Rule Types ====================

export interface WebhookRule {
  id: number
  name: string
  repoKey: string
  eventType: string
  branchPattern: string
  action: 'sync' | 'notify'
  syncTaskKeys: string
  minInterval: number
  enabled: boolean
  description?: string
  createdAt: string
  updatedAt: string
}

export interface CreateRuleRequest {
  name: string
  repoKey: string
  eventType: string
  branchPattern: string
  action: 'sync' | 'notify'
  syncTaskKeys?: string
  minInterval?: number
  enabled?: boolean
  description?: string
}

export interface UpdateRuleRequest extends Partial<CreateRuleRequest> {
  id: number
}

// ==================== Webhook Event Types ====================

export interface WebhookEvent {
  id: number
  ruleId?: number
  repoKey: string
  eventType: string
  payload?: string
  status: 'pending' | 'processing' | 'success' | 'failed'
  triggeredAt: string
  executedAt?: string
  errorMessage?: string
}

// ==================== Config Types ====================

export interface SyncConfigItem {
  key: string
  value: string
  scope: string
  description?: string
}
