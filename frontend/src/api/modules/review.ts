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
  raw_diff: string
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

export function listReviewFindings(task_id: number, params?: { severity?: string; source?: string }) {
  return request.get<unknown, ReviewFindingDTO[]>(`/reviews/tasks/${task_id}/findings`, { params })
}

export function retryReviewTask(id: number, data?: { owner?: string; repo?: string }) {
  return request.post<unknown, ReviewTaskDTO>(`/reviews/tasks/${id}/retry`, data || {})
}

export function checkMerge(params: { repo_key: string; mr_iid: string; commit_sha?: string }) {
  return request.get<unknown, MergeCheckDTO>('/merge-checks', { params })
}

export function getReviewConfig(repo_key: string) {
  return request.get<unknown, { config_yaml: string }>(`/reviews/config/${repo_key}`)
}

export function updateReviewConfig(repo_key: string, configYaml: string) {
  return request.put<unknown, { config_yaml: string }>(`/reviews/config/${repo_key}`, { config_yaml: configYaml })
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
  prompt_prefix_override: string
  prompt_intent_override: string
}

export interface LinkedRepoDTO {
  id: number
  key: string
  name: string
}

export function getRemoteRepoConfig(provider_id: number, owner: string, repo: string) {
  return request.get<unknown, ReviewRepoConfigDTO>(`/review/remote-config/${provider_id}/${owner}/${repo}`)
}

export function updateRemoteRepoConfig(provider_id: number, owner: string, repo: string, data: Partial<ReviewRepoConfigDTO>) {
  return request.put<unknown, ReviewRepoConfigDTO>(`/review/remote-config/${provider_id}/${owner}/${repo}`, data)
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

export interface ReviewStatsDTO {
  total_reviews: number
  total_findings: number
  pass_count: number
  blocked_count: number
  failed_count: number
  pass_rate: number
  by_risk_level: Record<string, number>
  by_severity: Record<string, number>
  by_source: Record<string, number>
  by_rule: Array<{ name: string; count: number }>
  top_files: Array<{ name: string; count: number }>
  daily_trend: Array<{
    date: string
    total: number
    passed: number
    rate: number
  }>
}

export function getReviewStats(params?: { repo_key?: string; period?: string }) {
  return request.get<unknown, ReviewStatsDTO>('/reviews/stats', { params })
}

export function submitFindingFeedback(findingId: number, feedback: 'useful' | 'false_positive') {
  return request.post<unknown, { status: string }>(`/reviews/findings/${findingId}/feedback`, { feedback })
}

export interface RAGIndexResult {
  repo_key: string
  chunk_count: number
  file_count: number
  duration: number
  error: string
}

export interface RAGStatsResult {
  stats: Record<string, number>
  available: boolean
}

export function indexRepoRAG(repo_key: string) {
  return request.post<unknown, RAGIndexResult>(`/reviews/rag/index/${repo_key}`)
}

export function getRAGStats() {
  return request.get<unknown, RAGStatsResult>('/reviews/rag/stats')
}
