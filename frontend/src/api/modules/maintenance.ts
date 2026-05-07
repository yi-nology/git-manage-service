import request from '../request'

export interface LargeFileEntry {
  path: string
  size: string
  sizeBytes: number
  exists: boolean
  commitCount: number
  source: string
}

export interface GitDirBreakdown {
  packDirSize: string
  packDirSizeBytes: number
  looseObjSize: string
  looseObjSizeBytes: number
  reflogSize: string
  reflogSizeBytes: number
  stashCount: number
  otherSize: string
  otherSizeBytes: number
}

export interface StashEntry {
  index: number
  message: string
  size: string
  sizeBytes: number
}

export interface RepoHealthReport {
  gitDirSize: string
  gitDirSizeBytes: number
  looseObjects: number
  packFiles: number
  inPackObjects: number
  commitCount: number
  branchCount: number
  tagCount: number
  gitDirBreakdown: GitDirBreakdown | null
  stashEntries: StashEntry[]
  largeFiles: LargeFileEntry[]
  threshold: number
  thresholdHuman: string
  excludes: string[]
}

export interface MaintenanceTaskResponse {
  taskId: string
}

export interface TaskProgress {
  id: string
  status: 'queued' | 'running' | 'success' | 'failed'
  progress: string[]
  error: string
  startTime: string
  endTime: string
}

export interface MaintenanceSnapshot {
  gitDirSize: string
  gitDirSizeBytes: number
  looseObjects: number
  packFiles: number
  inPackObjects: number
  commitCount: number
  branchCount: number
  tagCount: number
}

export interface MaintenanceRecordDTO {
  id: number
  type: string
  status: string
  triggerBy: string
  paramsJson: string
  snapshotBefore: MaintenanceSnapshot | null
  snapshotAfter: MaintenanceSnapshot | null
  errorMessage: string
  startedAt: string | null
  finishedAt: string | null
  createdAt: string
  duration: string
  savedBytes: number
  savedPercent: number
}

export interface MaintenanceRecordListResponse {
  records: MaintenanceRecordDTO[]
  total: number
  page: number
  pageSize: number
}

export interface FileAIRecommendation {
  path: string
  size: string
  sizeBytes: number
  recommendation: 'safe_to_delete' | 'caution' | 'keep'
  category: string
  reason: string
  confidence: 'high' | 'medium' | 'low'
}

export interface MaintenanceAIAnalysisResponse {
  summary: string
  totalSavings: string
  totalSaveBytes: number
  recommendations: FileAIRecommendation[]
}

export function getRepoHealth(repoKey: string, threshold?: number, excludes?: string[]) {
  const params: Record<string, string> = {}
  if (threshold && threshold > 0) {
    params.threshold = String(threshold)
  }
  if (excludes && excludes.length > 0) {
    params.exclude = excludes.join(',')
  }
  return request.get<MaintenanceTaskResponse & RepoHealthReport>(`/repo/${repoKey}/maintenance/health`, { params })
}

export function slimRepo(repoKey: string, paths: string[], addGitignore = true) {
  return request.post<MaintenanceTaskResponse>(`/repo/${repoKey}/maintenance/slim`, { paths, addGitignore })
}

export function gcRepo(repoKey: string) {
  return request.post<MaintenanceTaskResponse>(`/repo/${repoKey}/maintenance/gc`)
}

export function addGitignore(repoKey: string, paths: string[]) {
  return request.post<MaintenanceTaskResponse>(`/repo/${repoKey}/maintenance/gitignore`, { paths })
}

export function getTaskStatus(taskId: string) {
  return request.get<TaskProgress>('/repo/task', { params: { task_id: taskId } })
}

export function getMaintenanceRecords(repoKey: string, page = 1, pageSize = 10) {
  return request.get<MaintenanceRecordListResponse>(`/repo/${repoKey}/maintenance/records`, {
    params: { page, page_size: pageSize }
  })
}

export function getMaintenanceRecord(repoKey: string, id: number) {
  return request.get<MaintenanceRecordDTO>(`/repo/${repoKey}/maintenance/records/${id}`)
}

export interface PrefixFileEntry {
  path: string
  size: string
  sizeBytes: number
  exists: boolean
  commitCount: number
}

export interface PrefixSlimPreview {
  files: PrefixFileEntry[]
  totalCount: number
  totalSize: string
  totalBytes: number
}

export interface ForcePushResult {
  remoteName: string
  platform: string
  branches: number
  success: boolean
  error?: string
}

export interface ForcePushResponse {
  results: ForcePushResult[]
  taskId: string
}

export function analyzeMaintenanceAI(repoKey: string, filePaths: string[], threshold?: number) {
  return request.post<MaintenanceAIAnalysisResponse>(`/repo/${repoKey}/maintenance/ai-analyze`, {
    file_paths: filePaths,
    threshold: threshold || undefined
  }, { timeout: 180000 })
}

export function previewPrefixSlim(repoKey: string, prefixes: string[]) {
  return request.post<PrefixSlimPreview>(`/repo/${repoKey}/maintenance/slim-prefix/preview`, { prefixes })
}

export function slimByPrefix(repoKey: string, prefixes: string[], addGitignore = true, forcePush = false) {
  return request.post<MaintenanceTaskResponse>(`/repo/${repoKey}/maintenance/slim-prefix`, {
    prefixes,
    addGitignore,
    forcePush
  })
}

export function forcePushRemotes(repoKey: string) {
  return request.post<ForcePushResponse>(`/repo/${repoKey}/maintenance/force-push`)
}
