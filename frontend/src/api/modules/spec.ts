import request from '../request'
import type {
  SpecFileNode,
  LintResult,
  LintRule,
  LintMode,
  SaveRequest,
  CommitRequest,
  CommitResponse,
  FormatOptions,
  FormatResult,
} from '@/types/spec'

export function getSpecTree(repo_key?: string) {
  const key = repo_key || 'test-repo' // 临时使用默认仓库进行测试
  return request.get<unknown, SpecFileNode[]>('/spec/tree', {
    params: { repo_key: key },
  })
}

export function getSpecContent(path: string, repo_key?: string) {
  const key = repo_key || 'test-repo'
  return request.get<unknown, { content: string }>('/spec/content', {
    params: { path, repo_key: key },
  })
}

export function saveSpecContent(path: string, data: SaveRequest, repo_key?: string) {
  const key = repo_key || 'test-repo'
  return request.put<unknown, { message: string }>(
    `/spec/content/${encodeURIComponent(path)}`,
    { ...data, repo_key: key }
  )
}

export function lintSpec(content: string, rules?: string[], mode?: LintMode) {
  return request.post<unknown, LintResult>('/spec/lint', { content, rules, mode })
}

export function aiFixSpec(content: string, issue: string, line?: number, severity?: string) {
  return request.post<unknown, { content: string }>('/spec/ai-fix', {
    content,
    issue,
    line,
    severity,
  })
}

export function getLintRules() {
  return request.get<unknown, LintRule[]>('/spec/rules')
}

export function updateLintRule(id: string, data: Partial<LintRule>) {
  return request.put<unknown, { message: string }>(`/spec/rules/${id}`, data)
}

export function createLintRule(data: LintRule) {
  return request.post<unknown, LintRule>('/spec/rules', data)
}

export function commitSpec(path: string, data: CommitRequest, repo_key?: string) {
  const key = repo_key || 'test-repo'
  return request.post<unknown, CommitResponse>(
    `/spec/commit/${encodeURIComponent(path)}`,
    { ...data, repo_key: key }
  )
}

export interface CreateSpecFileRequest {
  repo_key: string
  path: string
  name: string
  content?: string // 可选，如果提供则使用此内容
}

export function createSpecFile(data: CreateSpecFileRequest) {
  return request.post<unknown, { message: string; path: string }>('/spec/create', data)
}

// 删除spec文件
export function deleteSpecFile(repo_key: string, path: string, commitMessage?: string) {
  return request.post<unknown, { message: string }>('/spec/delete', {
    repo_key: repo_key,
    path,
    commit_message: commitMessage
  })
}

// 验证spec文件
export function validateSpec(content: string) {
  return request.post<unknown, { valid: boolean; issues: any[]; warnings: any[] }>('/spec/validate', {
    content
  })
}

// 格式化 spec 文件
export function formatSpec(content: string, options?: FormatOptions) {
  return request.post<unknown, FormatResult>('/spec/format', {
    content,
    ...options,
  })
}

export function getSpecConfig() {
  return request.get<unknown, any>('/spec/config')
}

export function saveSpecConfig(data: any) {
  return request.put<unknown, any>('/spec/config', data)
}


