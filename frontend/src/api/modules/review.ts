import request from '../request'

export interface ReviewTaskDTO {
  id: number
  repo_id: number
  repo_key: string
  repo_name: string
  provider_config_id: number
  platform: string
  event_type: string
  mr_iid: string
  source_branch: string
  target_branch: string
  commit_sha: string
  trigger_type: string
  trigger_user: string
  status: string
  risk_level: string
  summary: string
  error_message: string
  started_at: string | null
  finished_at: string | null
  created_at: string
  updated_at: string
  findings_count: number
}

export interface ReviewFindingDTO {
  id: number
  task_id: number
  source: string
  rule_id: string
  severity: string
  file_path: string
  old_line: number
  new_line: number
  title: string
  message: string
  suggestion: string
  fingerprint: string
}

export interface MergeCheckDTO {
  mergeable: boolean
  checks: MergeCheckItemDTO[]
}

export interface MergeCheckItemDTO {
  check_type: string
  status: string
  risk_level: string
  message: string
}

export function createReviewTask(data: {
  repo_key: string
  provider_config_id?: number
  mr_iid: string
  commit_sha?: string
  trigger_type?: string
}) {
  return request.post<unknown, ReviewTaskDTO>('/reviews/tasks', data)
}

export function getReviewTask(id: number) {
  return request.get<unknown, ReviewTaskDTO>(`/reviews/tasks/${id}`)
}

export function listReviewTasks(params: {
  repo_key: string
  status?: string
  page?: number
  page_size?: number
}) {
  return request.get<unknown, {
    tasks: ReviewTaskDTO[]
    pagination: { total: number; page: number; page_size: number }
  }>('/reviews/tasks', { params })
}

export function listReviewFindings(taskId: number, params?: { severity?: string; source?: string }) {
  return request.get<unknown, ReviewFindingDTO[]>(`/reviews/tasks/${taskId}/findings`, { params })
}

export function retryReviewTask(id: number) {
  return request.post<unknown, ReviewTaskDTO>(`/reviews/tasks/${id}/retry`)
}

export function checkMerge(params: { repo_key: string; mr_iid: string; commit_sha?: string }) {
  return request.get<unknown, MergeCheckDTO>('/merge-checks', { params })
}

export function getReviewConfig(repoKey: string) {
  return request.get<unknown, { config_yaml: string }>(`/reviews/config/${repoKey}`)
}

export function updateReviewConfig(repoKey: string, configYaml: string) {
  return request.put<unknown, { config_yaml: string }>(`/reviews/config/${repoKey}`, { config_yaml: configYaml })
}

export interface ReviewRepoConfigDTO {
  id: number
  provider_config_id: number
  platform_owner: string
  platform_repo: string
  enabled: boolean
  block_on_high: boolean
  auto_review_on_mr: boolean
  llm_provider: string
  max_files: number
  max_diff_lines: number
  rule_overrides_json: string
  scope_note: string
  linked_repos: LinkedRepoDTO[]
}

export interface LinkedRepoDTO {
  id: number
  key: string
  name: string
}

export function getRemoteRepoConfig(providerId: number, owner: string, repo: string) {
  return request.get<unknown, ReviewRepoConfigDTO>(`/review/remote-config/${providerId}/${owner}/${repo}`)
}

export function updateRemoteRepoConfig(providerId: number, owner: string, repo: string, data: Partial<ReviewRepoConfigDTO>) {
  return request.put<unknown, ReviewRepoConfigDTO>(`/review/remote-config/${providerId}/${owner}/${repo}`, data)
}

export function createReviewTaskByProvider(data: {
  provider_config_id: number
  owner: string
  repo: string
  mr_iid: string
  commit_sha?: string
  trigger_type?: string
}) {
  return request.post<unknown, ReviewTaskDTO>('/reviews/tasks/provider', data)
}

export function listReviewTasksByProvider(params: {
  provider_id: number
  mr_iid: string
  page?: number
  page_size?: number
}) {
  return request.get<unknown, {
    tasks: ReviewTaskDTO[]
    pagination: { total: number; page: number; page_size: number }
  }>('/reviews/tasks/provider', { params })
}
