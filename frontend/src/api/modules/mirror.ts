import request from '../request'
import type {
  MirrorDTO,
  CreateMirrorReq,
  UpdateMirrorReq,
  MirrorSyncLogDTO,
  AnalyzeRemoteResponse,
} from '@/types/mirror'

export function getMirrors(repo_id?: number, mirror_type?: string) {
  return request.get<unknown, MirrorDTO[]>('/mirrors', {
    params: { ...(repo_id ? { repo_id: repo_id } : {}), ...(mirror_type ? { mirror_type: mirror_type } : {}) },
  })
}

export function getMirror(id: number) {
  return request.get<unknown, MirrorDTO>(`/mirror/${id}`)
}

export function createMirror(data: CreateMirrorReq) {
  return request.post<unknown, MirrorDTO>('/mirror', data)
}

export function updateMirror(id: number, data: UpdateMirrorReq) {
  return request.post<unknown, MirrorDTO>(`/mirror/${id}/update`, data)
}

export function deleteMirror(id: number) {
  return request.post(`/mirror/${id}/delete`)
}

export function triggerMirrorSync(id: number, trigger_type?: string) {
  return request.post(`/mirror/${id}/sync`, { trigger_type: trigger_type || 'manual' })
}

export function batchTriggerMirrorSync(mirrorIds: number[], trigger_type?: string) {
  return request.post('/mirrors/sync', { mirror_ids: mirrorIds, trigger_type: trigger_type || 'manual' })
}

export function previewMirrorSync(id: number) {
  return request.post<unknown, { preview: string }>(`/mirror/${id}/preview`)
}

export function getMirrorSyncLogs(mirror_id: number, limit?: number) {
  return request.get<unknown, MirrorSyncLogDTO[]>(`/mirror/${mirror_id}/logs`, {
    params: limit ? { limit } : {},
  })
}

export function getSyncLogDetail(logId: number) {
  return request.get<unknown, MirrorSyncLogDTO>(`/mirror/log/${logId}`)
}

export function deleteSyncLog(logId: number) {
  return request.post(`/mirror/log/${logId}/delete`)
}

export function pauseMirror(id: number) {
  return request.post<unknown, MirrorDTO>(`/mirror/${id}/pause`)
}

export function resumeMirror(id: number) {
  return request.post<unknown, MirrorDTO>(`/mirror/${id}/resume`)
}

export function analyzeRemote(remote_url: string, credential_id?: number) {
  return request.post<unknown, AnalyzeRemoteResponse>('/mirror/analyze', {
    remote_url: remote_url,
    credential_id: credential_id,
  })
}

export function validateCredential(credential_id: number, remote_url: string) {
  return request.post<unknown, { valid: boolean; message: string }>('/mirror/validate-credential', {
    credential_id: credential_id,
    remote_url: remote_url,
  })
}
