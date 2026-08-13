import request from '../request'
import type { StatsResponse, LineStatsResponse, LineStatsConfig } from '@/types/stats'

export function getStatsAnalyze(repo_key: string, params?: { branch?: string; author?: string; since?: string; until?: string }) {
  return request.get<unknown, StatsResponse>('/stats/analyze', {
    params: { repo_key: repo_key, ...params },
  })
}

export function getStatsAuthors(repo_key: string) {
  return request.get<unknown, { name: string; email: string }[]>('/stats/authors', { params: { repo_key: repo_key } })
}

export function getStatsBranches(repo_key: string) {
  return request.get<unknown, string[]>('/stats/branches', { params: { repo_key: repo_key } })
}

export function getStatsCommits(repo_key: string, params?: { branch?: string; author?: string; since?: string; until?: string }) {
  return request.get('/stats/commits', { params: { repo_key: repo_key, ...params } })
}

export function getLineStats(repo_key: string, params?: { branch?: string; author?: string; since?: string; until?: string }) {
  return request.get<unknown, LineStatsResponse>('/stats/lines', {
    params: { repo_key: repo_key, ...params },
  })
}

export function getLineStatsConfig(repo_key: string) {
  return request.get<unknown, LineStatsConfig>('/stats/lines/config', { params: { repo_key: repo_key } })
}

export function saveLineStatsConfig(repo_key: string, data: LineStatsConfig) {
  return request.post('/stats/lines/config', { repo_key: repo_key, ...data })
}

export function exportStatsCsv(repo_key: string, params?: Record<string, string>) {
  return request.get('/stats/export/csv', {
    params: { repo_key: repo_key, ...params },
    responseType: 'blob',
  })
}
