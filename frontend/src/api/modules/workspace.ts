import request from '../request'

export interface WorkspaceFileStatus {
  path: string
  status: string
  oldPath?: string
  additions: number
  deletions: number
  isBinary: boolean
}

export interface WorkspaceStatus {
  branch: string
  staged: WorkspaceFileStatus[]
  unstaged: WorkspaceFileStatus[]
  untracked: WorkspaceFileStatus[]
  conflicted: WorkspaceFileStatus[]
  isClean: boolean
  isMerging: boolean
  isRebasing: boolean
  ahead: number
  behind: number
}

export interface WorkspaceDiffFile {
  file: string
  diff: string
  additions: number
  deletions: number
  isBinary: boolean
}

export interface WorkspaceDiff {
  files: WorkspaceDiffFile[]
  totalAdditions: number
  totalDeletions: number
}

export interface CommitResult {
  commitHash: string
  pushed: boolean
}

export interface PullResult {
  status: string
  conflicts: string[]
  fetchLog: string
  behindPulled: boolean
}

export interface ConflictDetail {
  path: string
  oursContent: string
  theirsContent: string
  baseContent: string
  conflictMarker: string
}

export interface AIResolvedFile {
  filePath: string
  resolvedContent: string
  explanation: string
  confidence: number
}

export function getWorkspaceStatus(repoKey: string) {
  return request.get<unknown, WorkspaceStatus>('/workspace/status', {
    params: { repo_key: repoKey }
  })
}

export function getWorkspaceDiff(repoKey: string, params: { file?: string; stagedOnly?: boolean } = {}) {
  return request.get<unknown, WorkspaceDiff>('/workspace/diff', {
    params: { repo_key: repoKey, ...params }
  })
}

export function stageFiles(repoKey: string, files: string[], stageAll = false) {
  return request.post('/workspace/stage', {
    repo_key: repoKey,
    files,
    stage_all: stageAll
  })
}

export function unstageFiles(repoKey: string, files: string[], unstageAll = false) {
  return request.post('/workspace/unstage', {
    repo_key: repoKey,
    files,
    unstage_all: unstageAll
  })
}

export function commitChanges(data: {
  repo_key: string
  files?: string[]
  stage_all?: boolean
  message: string
  author_name?: string
  author_email?: string
  push?: boolean
  push_remote?: string
}) {
  return request.post<unknown, CommitResult>('/workspace/commit', data)
}

export function pullWithResolve(repoKey: string, remote = '', branch = '', fetchOnly = false) {
  return request.post<unknown, PullResult>('/workspace/pull', {
    repo_key: repoKey,
    remote,
    branch,
    fetch_only: fetchOnly
  })
}

export function getConflictDetail(repoKey: string, file: string) {
  return request.get<unknown, ConflictDetail>('/workspace/conflict-detail', {
    params: { repo_key: repoKey, file }
  })
}

export function markConflictResolved(repoKey: string, file: string, resolvedContent: string, stage = true) {
  return request.post('/workspace/resolve', {
    repo_key: repoKey,
    file,
    resolved_content: resolvedContent,
    stage
  })
}

export function aiResolveConflict(repoKey: string, file: string, oursContent: string, theirsContent: string, baseContent: string, hint = '') {
  return request.post<unknown, AIResolvedFile>('/workspace/ai-resolve', {
    repo_key: repoKey,
    file,
    ours_content: oursContent,
    theirs_content: theirsContent,
    base_content: baseContent,
    hint
  })
}

export function pushCurrent(repoKey: string, remote = '') {
  return request.post('/workspace/push', {
    repo_key: repoKey,
    remote,
  })
}

export function removeTracking(repoKey: string, files: string[]) {
  return request.post('/workspace/untrack', {
    repo_key: repoKey,
    files,
  })
}

export function addToGitignore(repoKey: string, patterns: string[]) {
  return request.post('/workspace/gitignore', {
    repo_key: repoKey,
    patterns,
  })
}

export function generateCommitMessage(repoKey: string) {
  return request.post<unknown, { message: string }>('/workspace/generate-commit-msg', {
    repo_key: repoKey,
  })
}
