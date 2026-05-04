import request from '../request'

export interface LLMProviderDTO {
  id: number
  name: string
  type: string
  base_url: string
  api_key?: string
  model: string
  max_tokens: number
  is_default: boolean
  preset_id?: string
  protocol?: string
  created_at: string
  updated_at: string
}

export interface LLMPresetModel {
  id: string
  display_name: string
}

export interface LLMPresetDTO {
  id: string
  display_name: string
  category: 'coding_plan' | 'direct' | 'local'
  type: string
  base_url: string
  anthropic_url?: string
  supports_anthropic: boolean
  models: LLMPresetModel[]
  default_model: string
  max_tokens: number
  icon: string
  region: 'cn' | 'global' | 'local'
  requires_key: boolean
  is_coding_plan: boolean
  coding_plan_price?: string
  warning?: string
  subscribe_url?: string
}

export interface CodeReviewGlobalSettingsDTO {
  enabled: boolean
  auto_review_on_mr: boolean
  block_on_high: boolean
  max_files: number
  max_diff_lines: number
}

export function listLLMProviders() {
  return request.get<unknown, LLMProviderDTO[]>('/settings/llm-providers')
}

export function fetchLLMPresets(category?: string) {
  const params = category ? { category } : {}
  return request.get<unknown, LLMPresetDTO[]>('/settings/llm-providers/presets', { params })
}

export function getLLMProvider(id: number) {
  return request.get<unknown, LLMProviderDTO>(`/settings/llm-providers/${id}`)
}

export function createLLMProvider(data: Partial<LLMProviderDTO>) {
  return request.post<unknown, LLMProviderDTO>('/settings/llm-providers', data)
}

export function updateLLMProvider(id: number, data: Partial<LLMProviderDTO>) {
  return request.put<unknown, LLMProviderDTO>(`/settings/llm-providers/${id}`, data)
}

export function deleteLLMProvider(id: number) {
  return request.delete<unknown, any>(`/settings/llm-providers/${id}`)
}

export function setDefaultLLMProvider(id: number) {
  return request.put<unknown, any>(`/settings/llm-providers/${id}/default`)
}

export function testLLMProvider(id: number) {
  return request.post<unknown, { status: string; message: string }>(`/settings/llm-providers/${id}/test`)
}

export function fetchOllamaModels(baseUrl?: string) {
  const params = baseUrl ? { base_url: baseUrl } : {}
  return request.get<unknown, string[]>('/settings/ollama-models', { params })
}

export function getCodeReviewSettings() {
  return request.get<unknown, CodeReviewGlobalSettingsDTO>('/settings/code-review')
}

export function updateCodeReviewSettings(data: CodeReviewGlobalSettingsDTO) {
  return request.put<unknown, CodeReviewGlobalSettingsDTO>('/settings/code-review', data)
}
