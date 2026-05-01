import request from '../request'

export interface BranchRuleDTO {
  id: number
  prefix: string
  display_name: string
  source_branches: string[]
  target_branches: string[]
  require_task_id: boolean
  task_id_pattern: string
  auto_delete_on_merge: boolean
  allow_direct_push: boolean
  require_code_review: boolean
  sort_order: number
}

export interface BranchRuleSetDTO {
  enabled: boolean
  rules: BranchRuleDTO[]
  protected_branches: string[]
}

export interface RemoteRepoBranchRulesDTO {
  provider_config_id: number
  platform_owner: string
  platform_repo: string
  use_custom_rules: boolean
  rules: BranchRuleDTO[]
  protected_branches: string[]
  linked_repos: { id: number; key: string; name: string }[]
}

export interface BranchValidationResult {
  valid: boolean
  errors?: { field: string; message: string }[]
  rule_name?: string
}

export function getBranchRules() {
  return request.get<unknown, BranchRuleSetDTO>('/settings/branch-rules')
}

export function updateBranchRules(data: BranchRuleSetDTO) {
  return request.put<unknown, BranchRuleSetDTO>('/settings/branch-rules', data)
}

export function getRemoteRepoBranchRules(providerId: number, owner: string, repo: string) {
  return request.get<unknown, RemoteRepoBranchRulesDTO>(`/branch-rules/remote-config/${providerId}/${owner}/${repo}`)
}

export function updateRemoteRepoBranchRules(providerId: number, owner: string, repo: string, data: Partial<RemoteRepoBranchRulesDTO>) {
  return request.put<unknown, RemoteRepoBranchRulesDTO>(`/branch-rules/remote-config/${providerId}/${owner}/${repo}`, data)
}

export function validateBranchName(params: { repo_key: string; branch_name: string; base_ref?: string; skip_rules?: boolean }) {
  return request.get<unknown, BranchValidationResult>('/settings/branch-rules/validate', { params })
}
