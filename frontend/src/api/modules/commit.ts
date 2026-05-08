import request from '../request'

// Commit搜索相关API

export interface CommitDetail {
  hash: string
  shortHash: string
  message: string
  authorName: string
  authorEmail: string
  authorDate: string
  committerName: string
  committerEmail: string
  committerDate: string
  parentHashes: string[]
  filesChanged: number
  additions: number
  deletions: number
}

export interface FileChange {
  path: string
  status: 'added' | 'modified' | 'deleted' | 'renamed'
  additions: number
  deletions: number
  old_path?: string
}

export interface SearchCommitsParams {
  ref?: string
  author?: string
  keyword?: string
  since?: string
  until?: string
  path?: string
  page?: number
  page_size?: number
}

// 搜索提交
export function searchCommits(repoKey: string, params: SearchCommitsParams) {
  return request.get<unknown, { commits: CommitDetail[]; total: number }>('/commit/search', {
    params: { repo_key: repoKey, ...params }
  })
}

// 获取提交详情
export function getCommitDetail(repoKey: string, hash: string) {
  return request.get<unknown, { commit: CommitDetail; files: FileChange[] }>('/commit/detail', {
    params: { repo_key: repoKey, hash }
  })
}

// 获取提交diff
export function getCommitDiff(repoKey: string, hash: string, file?: string) {
  return request.get<unknown, string>('/commit/diff', {
    params: { repo_key: repoKey, hash, file }
  })
}
