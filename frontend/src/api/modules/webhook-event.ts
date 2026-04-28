import request from '../request'

export interface WebhookEventDTO {
  id: number
  event_id: string
  event_type: string
  source: string
  repo_id?: number
  cr_id?: number
  platform_cr_number?: number
  actor_name: string
  actor_username: string
  status: string
  processed_at?: string
  created_at: string
  error_message?: string
}

export interface ListWebhookEventsReq {
  event_type?: string
  source?: string
  status?: string
  page?: number
  page_size?: number
}

export function listWebhookEvents(params?: ListWebhookEventsReq) {
  return request.get<unknown, { items: WebhookEventDTO[]; total: number }>('/webhook/events', { params })
}

export function retryWebhookEvent(eventId: number) {
  return request.post<unknown, { message: string }>('/webhook/events/retry', { event_id: eventId })
}
