import request from '../request'
import type { RepoProviderBindingDTO, CreateBindingReq, UpdateBindingReq, AutoDetectResp } from '@/types/binding'

export function listBindings(params?: { repo_key?: string; provider_config_id?: number }) {
  return request.get<unknown, RepoProviderBindingDTO[]>('/bindings', { params })
}

export function getBinding(id: number) {
  return request.get<unknown, RepoProviderBindingDTO>(`/bindings/${id}`)
}

export function createBinding(data: CreateBindingReq) {
  return request.post<unknown, RepoProviderBindingDTO>('/bindings', data)
}

export function updateBinding(id: number, data: UpdateBindingReq) {
  return request.put<unknown, RepoProviderBindingDTO>(`/bindings/${id}`, data)
}

export function deleteBinding(id: number, cleanupWebhook = false) {
  return request.delete<unknown, { message: string }>(`/bindings/${id}`, {
    params: { cleanup_webhook: cleanupWebhook },
  })
}

export function setPrimaryBinding(id: number) {
  return request.post<unknown, RepoProviderBindingDTO>(`/bindings/${id}/set-primary`)
}

export function autoDetectBindings(repoKey: string) {
  return request.post<unknown, AutoDetectResp>('/bindings/auto-detect', { repo_key: repoKey })
}

export function registerBindingWebhook(id: number) {
  return request.post<unknown, RepoProviderBindingDTO>(`/bindings/${id}/webhook`)
}

export function deleteBindingWebhook(id: number) {
  return request.delete<unknown, RepoProviderBindingDTO>(`/bindings/${id}/webhook`)
}
