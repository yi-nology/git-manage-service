import request from '../request'

export interface CRDTO {
  id: number
  repo_id: number
  repo_name?: string
  provider_id: number
  platform?: string
  cr_number: number
  title: string
  description: string
  state: string
  source_branch: string
  target_branch: string
  author_name: string
  author_username: string
  web_url: string
  merge_status: string
  labels: string[]
  created_at: string
  updated_at: string
  merged_at?: string
}

export interface ListCRsReq {
  repo_key: string
  state?: string
  source_branch?: string
  target_branch?: string
  page?: number
  page_size?: number
}

export function listCRs(params: ListCRsReq) {
  return request.get<unknown, { items: CRDTO[]; total: number }>('/cr/list', { params })
}

export function getCR(repoKey: string, crNumber: number) {
  return request.get<unknown, CRDTO>('/cr/detail', { params: { repo_key: repoKey, cr_number: crNumber } })
}

export function syncCRs(repoKey: string, state?: string) {
  return request.post<unknown, { synced_count: number }>('/cr/sync', { repo_key: repoKey, state })
}

export function mergeCR(repoKey: string, crNumber: number, mergeCommitMessage?: string, squash?: boolean, removeSourceBranch?: boolean) {
  return request.post<unknown, CRDTO>('/cr/merge', { repo_key: repoKey, cr_number: crNumber, merge_commit_message: mergeCommitMessage, squash, remove_source_branch: removeSourceBranch })
}

export function closeCR(repoKey: string, crNumber: number) {
  return request.post<unknown, CRDTO>('/cr/close', { repo_key: repoKey, cr_number: crNumber })
}

export function detectCR(repoKey: string) {
  return request.get<unknown, { provider_config_id: number; platform_owner: string; platform_repo: string }>('/cr/detect', { params: { repo_key: repoKey } })
}

export function createCR(data: { repo_key: string; title: string; description?: string; source_branch: string; target_branch: string; labels?: string[]; remove_source_branch?: boolean }) {
  return request.post<unknown, CRDTO>('/cr/create', data)
}

export function listRemoteCRs(params: { provider_id: number; owner: string; repo: string; state?: string; page?: number; per_page?: number }) {
  return request.get<unknown, { items: CRDTO[]; total: number }>('/cr/remote/list', { params })
}

export function createRemoteCR(data: { provider_id: number; owner: string; repo: string; title: string; description?: string; source_branch: string; target_branch: string; labels?: string[]; remove_source_branch?: boolean }) {
  return request.post<unknown, CRDTO>('/cr/remote/create', data)
}

export function mergeRemoteCR(providerId: number, owner: string, repo: string, crNumber: number, mergeCommitMessage?: string, squash?: boolean, removeSourceBranch?: boolean) {
  return request.post<unknown, CRDTO>('/cr/remote/merge', { provider_id: providerId, owner, repo, cr_number: crNumber, merge_commit_message: mergeCommitMessage, squash, remove_source_branch: removeSourceBranch })
}

export function closeRemoteCR(providerId: number, owner: string, repo: string, crNumber: number) {
  return request.post<unknown, CRDTO>('/cr/remote/close', { provider_id: providerId, owner, repo, cr_number: crNumber })
}
