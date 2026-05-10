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
  review_mode: string
  process_log: string
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

export function retryReviewTask(id: number, data?: { owner?: string; repo?: string }) {
  return request.post<unknown, ReviewTaskDTO>(`/reviews/tasks/${id}/retry`, data || {})
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
  review_mode: string
  cli_config_json: string
  custom_prompt: string
  use_custom_prompt: boolean
  exclude_file_types: string
  ignore_patterns: string
  linked_repos: LinkedRepoDTO[]
  prompt_prefix_override: string
  prompt_intent_override: string
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

export function indexRepoRAG(repoKey: string) {
  return request.post<unknown, RAGIndexResult>(`/reviews/rag/index/${repoKey}`)
}

export function getRAGStats() {
  return request.get<unknown, RAGStatsResult>('/reviews/rag/stats')
}

// CLI 配置管理
export interface ReviewCLIConfigDTO {
  id: number
  name: string
  cliType: string
  execPath: string
  configJson: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export type ReviewMode = 'llm' | 'claude_cli' | 'opencode_cli' | 'qoder_cli' | 'codex_cli' | 'hybrid'

export function listCLIConfigs() {
  return request.get<unknown, ReviewCLIConfigDTO[]>('/reviews/cli-configs')
}

export function scanCLIs() {
  return request.get<unknown, ScannedCLI[]>('/reviews/cli-configs/scan')
}

export interface ScannedCLI {
  cliType: string
  name: string
  execPath: string
  version: string
  isInstalled: boolean
}

export function getCLIConfig(id: number) {
  return request.get<unknown, { config: ReviewCLIConfigDTO }>(`/reviews/cli-configs/${id}`)
}

export function createCLIConfig(data: {
  name: string
  cli_type: string
  exec_path: string
  config_json?: string
  is_active?: boolean
}) {
  return request.post<unknown, { config: ReviewCLIConfigDTO }>('/reviews/cli-configs', data)
}

export function updateCLIConfig(id: number, data: {
  name?: string
  cli_type?: string
  exec_path?: string
  config_json?: string
  is_active?: boolean
}) {
  return request.put<unknown, { config: ReviewCLIConfigDTO }>(`/reviews/cli-configs/${id}`, data)
}

export function deleteCLIConfig(id: number) {
  return request.delete(`/reviews/cli-configs/${id}`)
}

export function testCLIConfig(id: number) {
  return request.post<unknown, { success: boolean; message: string; version?: string }>(`/reviews/cli-configs/${id}/test`)
}

// 审查审计日志
export interface ReviewAuditLogDTO {
  id: number
  task_id: number
  action: string
  status: string
  error_message: string
  duration: number
  metadata: string
  created_at: string
}

export function listReviewAuditLogs(params?: {
  task_id?: number
  action?: string
  status?: string
  start_time?: string
  end_time?: string
  page?: number
  page_size?: number
}) {
  return request.get<unknown, {
    logs: ReviewAuditLogDTO[]
    total: number
  }>('/reviews/audit-logs', { params })
}

// Webhook 事件规则
export interface WebhookEventRuleDTO {
  id: number
  name: string
  event_type: string
  description: string
  match_rules: string
  is_active: boolean
  priority: number
  created_at: string
  updated_at: string
}

export function listEventRules(params?: {
  event_type?: string
  is_active?: boolean
  page?: number
  page_size?: number
}) {
  return request.get<unknown, {
    rules: WebhookEventRuleDTO[]
    total: number
  }>('/webhook/event-rules', { params })
}

export function createEventRule(data: {
  name: string
  event_type: string
  description?: string
  match_rules: string
  is_active?: boolean
  priority?: number
}) {
  return request.post<unknown, { rule: WebhookEventRuleDTO }>('/webhook/event-rules', data)
}

export function updateEventRule(id: number, data: {
  name?: string
  event_type?: string
  description?: string
  match_rules?: string
  is_active?: boolean
  priority?: number
}) {
  return request.put<unknown, { rule: WebhookEventRuleDTO }>(`/webhook/event-rules/${id}`, data)
}

export function deleteEventRule(id: number) {
  return request.delete(`/webhook/event-rules/${id}`)
}
