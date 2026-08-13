import request from '../request'

export interface AliasEntry {
  name: string
  email: string
}

export interface AuthorIdentityDTO {
  id: number
  canonical_name: string
  canonical_email: string
  aliases: AliasEntry[]
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface RepoAuthorConfigDTO {
  repo_key: string
  identity_id: number | null
  identity: AuthorIdentityDTO | null
  source: string
}

export interface MismatchedCommit {
  hash: string
  short_hash: string
  message: string
  author_name: string
  author_email: string
  committer_name: string
  committer_email: string
  date: string
  match_type: 'exact' | 'email_only'
  target_name: string
  target_email: string
}

export interface AuthorScanResult {
  commits: MismatchedCommit[]
  total_commits: number
  match_count: number
}

export function listIdentities() {
  return request.get<unknown, AuthorIdentityDTO[]>('/author/identities')
}

export function createIdentity(data: { canonical_name: string; canonical_email: string; aliases: AliasEntry[] }) {
  return request.post<AuthorIdentityDTO>('/author/identities', {
    canonical_name: data.canonical_name,
    canonical_email: data.canonical_email,
    aliases_json: JSON.stringify(data.aliases),
  })
}

export function updateIdentity(id: number, data: { canonical_name: string; canonical_email: string; aliases: AliasEntry[] }) {
  return request.put<AuthorIdentityDTO>(`/author/identities/${id}`, {
    canonical_name: data.canonical_name,
    canonical_email: data.canonical_email,
    aliases_json: JSON.stringify(data.aliases),
  })
}

export function deleteIdentity(id: number) {
  return request.delete(`/author/identities/${id}`)
}

export function activateIdentity(id: number) {
  return request.post(`/author/identities/${id}/activate`)
}

export function getRepoAuthorConfig(repo_key: string) {
  return request.get<unknown, RepoAuthorConfigDTO>(`/repo/${repo_key}/author/config`)
}

export function setRepoAuthorConfig(repo_key: string, identity_id: number | null, clear = false) {
  return request.put(`/repo/${repo_key}/author/config`, {
    identity_id: identity_id,
    clear,
  })
}

export function scanAuthor(repo_key: string) {
  return request.get<AuthorScanResult>(`/repo/${repo_key}/author/scan`)
}

export function fixAuthorAll(repo_key: string, pushRemote = '') {
  return request.post<{ task_id: string }>(`/repo/${repo_key}/author/fix-all`, {
    push_remote: pushRemote,
  })
}

export function fixAuthor(repo_key: string, commitHashes: string[], pushRemote = '') {
  return request.post<{ task_id: string }>(`/repo/${repo_key}/author/fix`, {
    commit_hashes: commitHashes,
    push_remote: pushRemote,
  })
}

// --- AI ---

export interface AliasSuggestion {
  identity_id: number
  identity_name: string
  alias_name: string
  alias_email: string
  confidence: 'high' | 'medium' | 'low'
  reason: string
}

export interface AliasSuggestionResult {
  suggestions: AliasSuggestion[]
  summary: string
}

export interface MergeCandidate {
  keep_id: number
  keep_name: string
  merge_ids: number[]
  merge_names: string
  reason: string
}

export interface MergeSuggestionResult {
  merges: MergeCandidate[]
  summary: string
}

export interface RiskFactor {
  level: string
  description: string
  recommendation: string
}

export interface RiskAssessmentResult {
  risk_level: 'low' | 'medium' | 'high'
  summary: string
  factors: RiskFactor[]
  recommendations: string[]
}

export interface AuthorAIResponse {
  action: string
  result?: string
  suggest?: AliasSuggestionResult
  merge?: MergeSuggestionResult
  risk?: RiskAssessmentResult
}

export function authorAI(repo_key: string, action: string, data?: Record<string, unknown>) {
  return request.post<AuthorAIResponse>(`/repo/${repo_key}/author/ai`, { action, repo_key, ...data })
}

export interface ChatMessageDTO {
  role: string
  content: string
}

export function authorChat(repo_key: string, prompt: string, history: ChatMessageDTO[] = []) {
  return request.post<{ result: string }>(`/repo/${repo_key}/author/chat`, { repo_key, prompt, history })
}
