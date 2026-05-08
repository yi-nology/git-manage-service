import request from '../request'
import type {
  MirrorDTO,
  CreateMirrorReq,
  UpdateMirrorReq,
  MirrorSyncLogDTO,
  AnalyzeRemoteResponse,
} from '@/types/mirror'

export function getMirrors(repoId?: number, mirrorType?: string) {
  return request.get<unknown, MirrorDTO[]>('/mirrors', {
    params: { ...(repoId ? { repo_id: repoId } : {}), ...(mirrorType ? { mirror_type: mirrorType } : {}) },
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

export function triggerMirrorSync(id: number, triggerType?: string) {
  return request.post(`/mirror/${id}/sync`, { trigger_type: triggerType || 'manual' })
}

export function batchTriggerMirrorSync(mirrorIds: number[], triggerType?: string) {
  return request.post('/mirrors/sync', { mirror_ids: mirrorIds, trigger_type: triggerType || 'manual' })
}

export function previewMirrorSync(id: number) {
  return request.post<unknown, { preview: string }>(`/mirror/${id}/preview`)
}

export function getMirrorSyncLogs(mirrorId: number, limit?: number) {
  return request.get<unknown, MirrorSyncLogDTO[]>(`/mirror/${mirrorId}/logs`, {
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

export function analyzeRemote(remoteUrl: string, credentialId?: number) {
  return request.post<unknown, AnalyzeRemoteResponse>('/mirror/analyze', {
    remote_url: remoteUrl,
    credential_id: credentialId,
  })
}

export function validateCredential(credentialId: number, remoteUrl: string) {
  return request.post<unknown, { valid: boolean; message: string }>('/mirror/validate-credential', {
    credential_id: credentialId,
    remote_url: remoteUrl,
  })
}
