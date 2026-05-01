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
  return request.get<TaskProgress>('/repo/task', { task_id: taskId })
}
