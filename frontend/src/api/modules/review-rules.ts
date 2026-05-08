import request from '../request'

export interface ReviewRuleDTO {
  id: string
  name: string
  description: string
  severity: string
  category: string
  enabled: boolean
  sort_order: number
  rule_type: string
  prompt_text: string
}

export function listReviewRules() {
  return request.get<unknown, ReviewRuleDTO[]>('/settings/review-rules')
}

export function getReviewRule(ruleId: string) {
  return request.get<unknown, ReviewRuleDTO>(`/settings/review-rules/${ruleId}`)
}

export function createReviewRule(data: Partial<ReviewRuleDTO>) {
  return request.post<unknown, ReviewRuleDTO>('/settings/review-rules', data)
}

export function updateReviewRule(ruleId: string, data: Partial<ReviewRuleDTO>) {
  return request.put<unknown, ReviewRuleDTO>(`/settings/review-rules/${ruleId}`, data)
}

export function deleteReviewRule(ruleId: string) {
  return request.delete<unknown, any>(`/settings/review-rules/${ruleId}`)
}

export function batchUpdateReviewRules(rules: Partial<ReviewRuleDTO>[]) {
  return request.put<unknown, ReviewRuleDTO[]>('/settings/review-rules/batch', rules)
}

export interface PromptStructureDTO {
  prefix: string
  intent: string
  suffix: string
}

export function getPromptStructure() {
  return request.get<unknown, PromptStructureDTO>('/reviews/prompt-structure')
}
