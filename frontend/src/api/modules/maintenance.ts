import request from '../request'

export interface LargeFileEntry {
  path: string
  size: string
  sizeBytes: number
  exists: boolean
  commitCount: number
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
  largeFiles: LargeFileEntry[]
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

export function getRepoHealth(repoKey: string) {
  return request.get<MaintenanceTaskResponse & RepoHealthReport>(`/repo/${repoKey}/maintenance/health`)
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
