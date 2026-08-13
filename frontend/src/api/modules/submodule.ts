import request from '../request'

// Submodule管理相关API

export interface SubmoduleInfo {
  name: string
  path: string
  url: string
  branch: string
  commit: string
  status: 'initialized' | 'uninitialized' | 'modified' | 'unknown'
}

export interface SubmoduleStatusItem {
  path: string
  commit: string
  status: string  // +, -, U, 空
  description: string
}

// 列出submodule
export function listSubmodules(repo_key: string) {
  return request.get<unknown, { submodules: SubmoduleInfo[] }>('/submodule/list', {
    params: { repo_key: repo_key }
  })
}

// 获取submodule状态
export function getSubmoduleStatus(repo_key: string, recursive?: boolean) {
  return request.get<unknown, { items: SubmoduleStatusItem[] }>('/submodule/status', {
    params: { repo_key: repo_key, recursive }
  })
}

// 添加submodule
export function addSubmodule(repo_key: string, url: string, path: string, branch?: string) {
  return request.post('/submodule/add', {
    repo_key: repo_key,
    url,
    path,
    branch
  })
}

// 初始化submodule
export function initSubmodule(repo_key: string, path?: string) {
  return request.post('/submodule/init', {
    repo_key: repo_key,
    path
  })
}

// 更新submodule
export function updateSubmodule(repo_key: string, params: { path?: string; init?: boolean; recursive?: boolean; remote?: boolean }) {
  return request.post('/submodule/update', {
    repo_key: repo_key,
    ...params
  })
}

// 同步submodule URL
export function syncSubmodule(repo_key: string, path?: string, recursive?: boolean) {
  return request.post('/submodule/sync', {
    repo_key: repo_key,
    path,
    recursive
  })
}

// 移除submodule
export function removeSubmodule(repo_key: string, path: string, force?: boolean) {
  return request.post('/submodule/remove', {
    repo_key: repo_key,
    path,
    force
  })
}
