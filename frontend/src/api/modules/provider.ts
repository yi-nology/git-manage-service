import request from '../request'

export interface ProviderConfigDTO {
  id: number
  name: string
  platform: string
  base_url: string
  credential_id: number
  credential_name?: string
  webhook_endpoint?: string
  has_webhook_secret: boolean
  skip_tls: boolean
  created_at: string
  updated_at: string
}

export interface CreateProviderConfigReq {
  name: string
  platform: string
  base_url: string
  credential_id: number
  webhook_secret?: string
  skip_tls?: boolean
}

export interface UpdateProviderConfigReq {
  name?: string
  base_url?: string
  credential_id?: number
  webhook_secret?: string
  skip_tls?: boolean
}

export interface TestProviderConfigResp {
  connected: boolean
  platform: string
  user_name: string
  message?: string
}

export function listProviders() {
  return request.get<unknown, ProviderConfigDTO[]>('/providers')
}

export function getProvider(id: number) {
  return request.get<unknown, ProviderConfigDTO>(`/providers/${id}`)
}

export function createProvider(data: CreateProviderConfigReq) {
  return request.post<unknown, ProviderConfigDTO>('/providers', data)
}

export function updateProvider(id: number, data: UpdateProviderConfigReq) {
  return request.put<unknown, ProviderConfigDTO>(`/providers/${id}`, data)
}

export function deleteProvider(id: number) {
  return request.delete<unknown, { message: string }>(`/providers/${id}`)
}

export function testProvider(id: number) {
  return request.post<unknown, TestProviderConfigResp>(`/providers/${id}/test`)
}

export interface RemoteRepoDTO {
  id: number
  name: string
  full_name: string
  owner: string
  description: string
  clone_url: string
  ssh_url: string
  default_branch: string
  private: boolean
  platform: string
}

export function listProviderRepos(id: number, params?: { page?: number; per_page?: number; owner?: string }) {
  return request.get<unknown, RemoteRepoDTO[]>(`/providers/${id}/repos`, { params })
}

export function listRemoteBranches(params: { provider_id: number; owner: string; repo: string }) {
  return request.get<unknown, { name: string }[]>('/providers/branches', { params })
}

export function createRemoteBranch(data: { provider_id: number; owner: string; repo: string; branch: string; ref: string }) {
  return request.post<unknown, { name: string }>('/providers/branches/create', data)
}

export function deleteRemoteBranch(data: { provider_id: number; owner: string; repo: string; branch: string }) {
  return request.post<unknown, { message: string }>('/providers/branches/delete', data)
}
