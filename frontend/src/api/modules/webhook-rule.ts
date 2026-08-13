import request from '../request'

// WebhookRule mirrors the backend webhook_rule management DTO (snake_case).
export interface WebhookRule {
  id: number
  name: string
  provider_config_id: number
  event_type_pattern: string
  repo_pattern: string
  action: string
  action_config?: Record<string, unknown>
  enabled: boolean
  created_at?: string
  updated_at?: string
}

export interface WebhookRulePayload {
  name: string
  provider_config_id: number
  event_type_pattern: string
  repo_pattern: string
  action: string
  action_config?: Record<string, unknown>
  enabled: boolean
}

// 列出所有 webhook 规则
export function listWebhookRules() {
  return request.get<unknown, { items: WebhookRule[] }>('/webhook/rules')
}

// 创建 webhook 规则
export function createWebhookRule(data: WebhookRulePayload) {
  return request.post<unknown, WebhookRule>('/webhook/rules', data)
}

// 更新 webhook 规则
export function updateWebhookRule(id: number, data: WebhookRulePayload) {
  return request.put<unknown, WebhookRule>(`/webhook/rules/${id}`, data)
}

// 删除 webhook 规则
export function deleteWebhookRule(id: number) {
  return request.delete<unknown, { message: string }>(`/webhook/rules/${id}`)
}
