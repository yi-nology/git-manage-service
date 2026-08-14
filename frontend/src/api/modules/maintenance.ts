import request from '../request'

export interface LargeFileEntry {
  path: string
  size: string
  size_bytes: number
  exists: boolean
  commit_count: number
  source: string
}

export interface GitDirBreakdown {
  pack_dir_size: string
  pack_dir_size_bytes: number
  loose_obj_size: string
  loose_obj_size_bytes: number
  reflog_size: string
  reflog_size_bytes: number
  stash_count: number
  other_size: string
  other_size_bytes: number
}

export interface StashEntry {
  index: number
  message: string
  size: string
  size_bytes: number
}

export interface RepoHealthReport {
  git_dir_size: string
  git_dir_size_bytes: number
  loose_objects: number
  pack_files: number
  in_pack_objects: number
  commit_count: number
  branch_count: number
  tag_count: number
  git_dir_breakdown: GitDirBreakdown | null
  stash_entries: StashEntry[]
  large_files: LargeFileEntry[]
  threshold: number
  threshold_human: string
  excludes: string[]
}

export interface MaintenanceTaskResponse {
  task_id: string
}

export interface TaskProgress {
  id: string
  status: 'queued' | 'running' | 'success' | 'failed'
  progress: string[]
  error: string
  start_time: string
  end_time: string
}

export interface MaintenanceSnapshot {
  git_dir_size: string
  git_dir_size_bytes: number
  loose_objects: number
  pack_files: number
  in_pack_objects: number
  commit_count: number
  branch_count: number
  tag_count: number
}

export interface MaintenanceRecordDTO {
  id: number
  type: string
  status: string
  trigger_by: string
  params_json: string
  snapshot_before: MaintenanceSnapshot | null
  snapshot_after: MaintenanceSnapshot | null
  error_message: string
  started_at: string | null
  finished_at: string | null
  created_at: string
  duration: string
  saved_bytes: number
  saved_percent: number
}

export interface MaintenanceRecordListResponse {
  records: MaintenanceRecordDTO[]
  total: number
  page: number
  page_size: number
}

export interface FileAIRecommendation {
  path: string
  size: string
  size_bytes: number
  recommendation: 'safe_to_delete' | 'caution' | 'keep'
  category: string
  reason: string
  confidence: 'high' | 'medium' | 'low'
}

export interface MaintenanceAIAnalysisResponse {
  summary: string
  total_savings: string
  total_save_bytes: number
  recommendations: FileAIRecommendation[]
}

export function getRepoHealth(repo_key: string, threshold?: number, excludes?: string[]) {
  const params: Record<string, string> = {}
  if (threshold && threshold > 0) {
    params.threshold = String(threshold)
  }
  if (excludes && excludes.length > 0) {
    params.exclude = excludes.join(',')
  }
  return request.get<MaintenanceTaskResponse & RepoHealthReport>(`/repo/${repo_key}/maintenance/health`, { params })
}

export function slimRepo(repo_key: string, paths: string[], addGitignore = true) {
  // SlimRequest proto binds json:"add_gitignore" — the key must be snake_case
  // (slimByPrefix below uses a raw-BindJSON handler that expects camelCase).
  return request.post<MaintenanceTaskResponse>(`/repo/${repo_key}/maintenance/slim`, { paths, add_gitignore: addGitignore })
}

export function gcRepo(repo_key: string) {
  return request.post<MaintenanceTaskResponse>(`/repo/${repo_key}/maintenance/gc`)
}

export function addGitignore(repo_key: string, paths: string[]) {
  return request.post<MaintenanceTaskResponse>(`/repo/${repo_key}/maintenance/gitignore`, { paths })
}

export function getTaskStatus(task_id: string) {
  return request.get<TaskProgress>('/repo/task', { params: { task_id: task_id } })
}

export function getMaintenanceRecords(repo_key: string, page = 1, page_size = 10) {
  return request.get<MaintenanceRecordListResponse>(`/repo/${repo_key}/maintenance/records`, {
    params: { page, page_size: page_size }
  })
}

export function getMaintenanceRecord(repo_key: string, id: number) {
  return request.get<MaintenanceRecordDTO>(`/repo/${repo_key}/maintenance/records/${id}`)
}

export interface PrefixFileEntry {
  path: string
  size: string
  size_bytes: number
  exists: boolean
  commit_count: number
}

export interface PrefixSlimPreview {
  files: PrefixFileEntry[]
  total_count: number
  total_size: string
  total_bytes: number
}

export interface ForcePushResult {
  remote_name: string
  platform: string
  branches: number
  success: boolean
  error?: string
}

export interface ForcePushResponse {
  results: ForcePushResult[]
  task_id: string
}

export function analyzeMaintenanceAI(repo_key: string, filePaths: string[], threshold?: number) {
  return request.post<MaintenanceAIAnalysisResponse>(`/repo/${repo_key}/maintenance/ai-analyze`, {
    file_paths: filePaths,
    threshold: threshold || undefined
  }, { timeout: 180000 })
}

export function previewPrefixSlim(repo_key: string, prefixes: string[]) {
  return request.post<PrefixSlimPreview>(`/repo/${repo_key}/maintenance/slim-prefix/preview`, { prefixes })
}

export function slimByPrefix(repo_key: string, prefixes: string[], addGitignore = true, forcePush = false) {
  return request.post<MaintenanceTaskResponse>(`/repo/${repo_key}/maintenance/slim-prefix`, {
    prefixes,
    addGitignore,
    forcePush
  })
}

export function forcePushRemotes(repo_key: string) {
  return request.post<ForcePushResponse>(`/repo/${repo_key}/maintenance/force-push`)
}
