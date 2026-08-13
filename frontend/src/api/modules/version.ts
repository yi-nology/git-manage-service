import request from '../request'

export interface VersionTag {
  name: string
  hash: string
  date: string
  message: string
  tagger: string
}

export interface NextVersionInfo {
  current: string
  next_major: string
  next_minor: string
  next_patch: string
}

export function getVersionList(repo_key: string) {
  return request.get<unknown, VersionTag[]>('/version/list', { params: { repo_key: repo_key } })
}

export function getCurrentVersion(repo_key: string) {
  return request.get<unknown, string>('/version/current', { params: { repo_key: repo_key } })
}

export function getNextVersion(repo_key: string) {
  return request.get<unknown, NextVersionInfo>('/version/next', { params: { repo_key: repo_key } })
}
