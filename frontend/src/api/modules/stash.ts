import request from '../request'

// Stash管理相关API

export interface StashEntry {
  index: number
  ref: string
  message: string
  branch: string
  date: string
}

// 列出stash
export function listStash(repo_key: string) {
  return request.get<unknown, { stashes: StashEntry[] }>('/stash/list', {
    params: { repo_key: repo_key }
  })
}

// 保存stash
export function saveStash(repo_key: string, message?: string, includeUntracked?: boolean) {
  return request.post('/stash/save', {
    repo_key: repo_key,
    message,
    include_untracked: includeUntracked
  })
}

// 应用stash
export function applyStash(repo_key: string, index: number) {
  return request.post('/stash/apply', {
    repo_key: repo_key,
    index
  })
}

// 弹出stash
export function popStash(repo_key: string, index: number) {
  return request.post('/stash/pop', {
    repo_key: repo_key,
    index
  })
}

// 删除stash
export function dropStash(repo_key: string, index: number) {
  return request.post('/stash/drop', {
    repo_key: repo_key,
    index
  })
}

// 清空stash
export function clearStash(repo_key: string) {
  return request.post('/stash/clear', {
    repo_key: repo_key
  })
}
