import request from '../request'

export interface WorkspaceFileStatus {
  path: string
  status: string
  old_path?: string
  additions: number
  deletions: number
  is_binary: boolean
}

export interface WorkspaceStatus {
  branch: string
  staged: WorkspaceFileStatus[]
  unstaged: WorkspaceFileStatus[]
  untracked: WorkspaceFileStatus[]
  conflicted: WorkspaceFileStatus[]
  is_clean: boolean
  is_merging: boolean
  is_rebasing: boolean
  ahead: number
  behind: number
}

export interface WorkspaceDiffFile {
  file: string
  diff: string
  additions: number
  deletions: number
  is_binary: boolean
}

export interface WorkspaceDiff {
  files: WorkspaceDiffFile[]
  total_additions: number
  total_deletions: number
}

export interface CommitResult {
  commit_hash: string
  pushed: boolean
}

export interface PullResult {
  status: string
  conflicts: string[]
  fetch_log: string
  behind_pulled: boolean
}

export interface ConflictDetail {
  path: string
  ours_content: string
  theirs_content: string
  base_content: string
  conflict_marker: string
}

export interface AIResolvedFile {
  file_path: string
  resolved_content: string
  explanation: string
  confidence: number
}

export function getWorkspaceStatus(repo_key: string) {
  return request.get<unknown, WorkspaceStatus>('/workspace/status', {
    params: { repo_key: repo_key }
  })
}

export function getWorkspaceDiff(repo_key: string, params: { file?: string; stagedOnly?: boolean } = {}) {
  return request.get<unknown, WorkspaceDiff>('/workspace/diff', {
    params: { repo_key: repo_key, ...params }
  })
}

export function stageFiles(repo_key: string, files: string[], stageAll = false) {
  return request.post('/workspace/stage', {
    repo_key: repo_key,
    files,
    stage_all: stageAll
  })
}

export function unstageFiles(repo_key: string, files: string[], unstageAll = false) {
  return request.post('/workspace/unstage', {
    repo_key: repo_key,
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

export function pullWithResolve(repo_key: string, remote = '', branch = '', fetchOnly = false) {
  return request.post<unknown, PullResult>('/workspace/pull', {
    repo_key: repo_key,
    remote,
    branch,
    fetch_only: fetchOnly
  })
}

export function getConflictDetail(repo_key: string, file: string) {
  return request.get<unknown, ConflictDetail>('/workspace/conflict-detail', {
    params: { repo_key: repo_key, file }
  })
}

export function markConflictResolved(repo_key: string, file: string, resolved_content: string, stage = true) {
  return request.post('/workspace/resolve', {
    repo_key: repo_key,
    file,
    resolved_content: resolved_content,
    stage
  })
}

export function aiResolveConflict(repo_key: string, file: string, ours_content: string, theirs_content: string, base_content: string, hint = '') {
  return request.post<unknown, AIResolvedFile>('/workspace/ai-resolve', {
    repo_key: repo_key,
    file,
    ours_content: ours_content,
    theirs_content: theirs_content,
    base_content: base_content,
    hint
  })
}

export function pushCurrent(repo_key: string, remote = '') {
  return request.post('/workspace/push', {
    repo_key: repo_key,
    remote,
  })
}

export function removeTracking(repo_key: string, files: string[]) {
  return request.post('/workspace/untrack', {
    repo_key: repo_key,
    files,
  })
}

export function addToGitignore(repo_key: string, patterns: string[]) {
  return request.post('/workspace/gitignore', {
    repo_key: repo_key,
    patterns,
  })
}

export function generateCommitMessage(repo_key: string) {
  return request.post<unknown, { message: string }>('/workspace/generate-commit-msg', {
    repo_key: repo_key,
  })
}
