export interface RepoProviderBindingDTO {
  id: number
  repo_id: number
  repo_key: string
  repo_name: string
  provider_config_id: number
  provider_name: string
  platform: string
  platform_owner: string
  platform_repo: string
  platform_repo_id: string
  remote_name: string
  is_primary: boolean
  webhook_id: string
  webhook_url: string
  has_webhook: boolean
  status: string
  created_at: string
  updated_at: string
}

export interface CreateBindingReq {
  repo_key: string
  provider_config_id: number
  platform_owner: string
  platform_repo: string
  remote_name?: string
  is_primary?: boolean
  register_webhook?: boolean
}

export interface UpdateBindingReq {
  remote_name?: string
  is_primary?: boolean
  platform_repo_id?: string
}

export interface BindingSuggestion {
  provider_config_id: number
  platform: string
  platform_owner: string
  platform_repo: string
  remote_name: string
  remote_url: string
  confidence: 'high' | 'medium' | 'low'
  match_source: string
}

export interface AutoDetectResp {
  suggestions: BindingSuggestion[]
}
