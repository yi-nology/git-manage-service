export interface AIRef {
  type: string
  id: string
  label: string
  file_path?: string
  start_line?: number
  end_line?: number
  url?: string
}

export interface AIAction {
  id: string
  label: string
  type: string
  description?: string
}

 export interface AIAdviceResponse {
 	summary: string
 	confidence?: string
 	risk_level?: string
 	references?: AIRef[]
 	suggestions?: string[]
 	actions?: AIAction[]
 	raw?: string
 	invocation_id?: number
 }

 export interface AIDraftResponse {
 	summary: string
 	change_type?: string
 	risk_level?: string
 	references?: AIRef[]
 	apply_content?: string
 	patch?: string
 	actions?: AIAction[]
 	raw?: string
 	invocation_id?: number
 }

 export interface AIDiagnosisResponse {
 	root_cause: string
 	evidence?: string[]
 	recommended_actions?: string[]
 	can_auto_fix?: boolean
 	risk_level?: string
 	references?: AIRef[]
 	fix_draft?: string
 	raw?: string
 	invocation_id?: number
 }

 export interface AIReviewFinding {
 	severity: string
 	category: string
 	message: string
 	file_path?: string
 	start_line?: number
 	end_line?: number
 	suggestion?: string
 	confidence?: string
 }

 export interface AIReviewResponse {
  	summary: string
  	blocking?: AIReviewFinding[]
  	high_risk?: AIReviewFinding[]
  	optional?: AIReviewFinding[]
  	risk_level?: string
  	should_merge?: boolean
  	merge_notes?: string
  	raw?: string
  	invocation_id?: number
  }

 export interface AIFeedbackRequest {
 	invocation_id: number
 	feedback: string
 }

 export interface AIChatMessage {
   role: 'user' | 'assistant' | 'system'
   content: string
 }

 export interface QuickAction {
   key: string
   label: string
   prompt: string
 }
