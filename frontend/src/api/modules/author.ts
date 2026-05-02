import request from '../request'

export interface AliasEntry {
  name: string
  email: string
}

export interface AuthorIdentityDTO {
  id: number
  canonicalName: string
  canonicalEmail: string
  aliases: AliasEntry[]
  isDefault: boolean
  createdAt: string
  updatedAt: string
}

export interface RepoAuthorConfigDTO {
  repoKey: string
  identityId: number | null
  identity: AuthorIdentityDTO | null
  source: string
}

export interface MismatchedCommit {
  hash: string
  shortHash: string
  message: string
  authorName: string
  authorEmail: string
  committerName: string
  committerEmail: string
  date: string
  matchType: 'exact' | 'email_only'
  targetName: string
  targetEmail: string
}

export interface AuthorScanResult {
  commits: MismatchedCommit[]
  totalCommits: number
  matchCount: number
}

export function listIdentities() {
  return request.get<AuthorIdentityDTO[]>('/author/identities')
}

export function createIdentity(data: { canonicalName: string; canonicalEmail: string; aliases: AliasEntry[] }) {
  return request.post<AuthorIdentityDTO>('/author/identities', {
    canonical_name: data.canonicalName,
    canonical_email: data.canonicalEmail,
    aliases_json: JSON.stringify(data.aliases),
  })
}

export function updateIdentity(id: number, data: { canonicalName: string; canonicalEmail: string; aliases: AliasEntry[] }) {
  return request.put<AuthorIdentityDTO>(`/author/identities/${id}`, {
    canonical_name: data.canonicalName,
    canonical_email: data.canonicalEmail,
    aliases_json: JSON.stringify(data.aliases),
  })
}

export function deleteIdentity(id: number) {
  return request.delete(`/author/identities/${id}`)
}

export function activateIdentity(id: number) {
  return request.post(`/author/identities/${id}/activate`)
}

export function getRepoAuthorConfig(repoKey: string) {
  return request.get<RepoAuthorConfigDTO>(`/repo/${repoKey}/author/config`)
}

export function setRepoAuthorConfig(repoKey: string, identityId: number | null, clear = false) {
  return request.put(`/repo/${repoKey}/author/config`, {
    identity_id: identityId,
    clear,
  })
}

export function scanAuthor(repoKey: string) {
  return request.get<AuthorScanResult>(`/repo/${repoKey}/author/scan`)
}

export function fixAuthorAll(repoKey: string, pushRemote = '') {
  return request.post<{ taskId: string }>(`/repo/${repoKey}/author/fix-all`, {
    push_remote: pushRemote,
  })
}

export function fixAuthor(repoKey: string, commitHashes: string[], pushRemote = '') {
  return request.post<{ taskId: string }>(`/repo/${repoKey}/author/fix`, {
    commit_hashes: commitHashes,
    push_remote: pushRemote,
  })
}

// --- AI ---

export interface AliasSuggestion {
  identityId: number
  identityName: string
  aliasName: string
  aliasEmail: string
  confidence: 'high' | 'medium' | 'low'
  reason: string
}

export interface AliasSuggestionResult {
  suggestions: AliasSuggestion[]
  summary: string
}

export interface MergeCandidate {
  keepId: number
  keepName: string
  mergeIds: number[]
  mergeNames: string
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
  riskLevel: 'low' | 'medium' | 'high'
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

export function authorAI(repoKey: string, action: string, data?: Record<string, unknown>) {
  return request.post<AuthorAIResponse>(`/repo/${repoKey}/author/ai`, { action, repoKey, ...data })
}

export interface ChatMessageDTO {
  role: string
  content: string
}

export function authorChat(repoKey: string, prompt: string, history: ChatMessageDTO[] = []) {
  return request.post<{ result: string }>(`/repo/${repoKey}/author/chat`, { repoKey, prompt, history })
}
