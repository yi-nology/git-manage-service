import request from '../request'
import type { BranchInfo, CreateBranchReq, MergeReq, MergeCheckResult, CreateTagReq } from '@/types/branch'
import type { PaginationResponse } from '@/types/common'

export function getBranchList(repo_key: string, params?: { page?: number; page_size?: number; keyword?: string; type?: string }) {
  return request.get<unknown, PaginationResponse<BranchInfo>>('/branch/list', {
    params: { repo_key: repo_key, ...params },
  })
}

export function createBranch(data: CreateBranchReq) {
  return request.post('/branch/create', data)
}

export function deleteBranch(repo_key: string, name: string) {
  return request.post('/branch/delete', { repo_key: repo_key, name })
}

export function updateBranch(repo_key: string, name: string, newName: string, desc?: string) {
  return request.post('/branch/update', { repo_key: repo_key, name, new_name: newName, desc })
}

export function checkoutBranch(repo_key: string, name: string) {
  return request.post('/branch/checkout', { repo_key: repo_key, name })
}

export function pushBranch(repo_key: string, name: string, remotes: string[]) {
  return request.post('/branch/push', { repo_key: repo_key, name, remotes })
}

export function pullBranch(repo_key: string, name: string) {
  return request.post('/branch/pull', { repo_key: repo_key, name })
}

export function compareBranches(repo_key: string, base: string, target: string) {
  return request.get<unknown, { stat: { FilesChanged: number; Insertions: number; Deletions: number }; files: { path: string; status: string }[] }>('/branch/compare', {
    params: { repo_key: repo_key, base, target },
  })
}

export function getBranchDiff(repo_key: string, base: string, target: string, file?: string) {
  return request.get<unknown, { diff: string }>('/branch/diff', {
    params: { repo_key: repo_key, base, target, file },
  })
}

export function getBranchPatch(repo_key: string, base: string, target: string) {
  return request.get('/branch/patch', {
    params: { repo_key: repo_key, base, target },
    responseType: 'blob',
  })
}

export function checkMerge(repo_key: string, base: string, target: string) {
  return request.get<unknown, MergeCheckResult>('/branch/merge/check', {
    params: { repo_key: repo_key, base, target },
  })
}

export function mergeBranch(data: MergeReq) {
  return request.post('/branch/merge', data)
}

export function getTagList(repo_key: string) {
  return request.get<unknown, string[]>('/tag/list', { params: { repo_key: repo_key } })
}

export function createTag(data: CreateTagReq) {
  return request.post('/tag/create', data)
}

export function deleteTag(data: { repo_key: string; name: string; delete_remote?: boolean; remote_name?: string }) {
  return request.post('/tag/delete', data)
}

export function pushTag(data: { repo_key: string; tag_name: string; remote_name: string }) {
  return request.post('/tag/push', data)
}

// Cherry-pick提交
export function cherryPick(repo_key: string, commit_hash: string, noCommit?: boolean) {
  return request.post<unknown, { success: boolean; new_commit?: string; conflicts?: string[] }>('/branch/cherry-pick', {
    repo_key: repo_key,
    commit_hash: commit_hash,
    no_commit: noCommit
  })
}

// Rebase分支
export function rebaseBranch(repo_key: string, upstream: string, onto?: string, interactive?: boolean) {
  return request.post<unknown, { success: boolean; in_progress?: boolean; conflicts?: string[]; current_commit?: string }>('/branch/rebase', {
    repo_key: repo_key,
    upstream,
    onto,
    interactive
  })
}

// 中止Rebase
export function rebaseAbort(repo_key: string) {
  return request.post('/branch/rebase/abort', {
    repo_key: repo_key
  })
}

// 继续Rebase
export function rebaseContinue(repo_key: string) {
  return request.post<unknown, { success: boolean; in_progress?: boolean; conflicts?: string[]; current_commit?: string }>('/branch/rebase/continue', {
    repo_key: repo_key
  })
}
