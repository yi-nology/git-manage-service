export interface AIRef {
  type: string
  id: string
  label: string
  filePath?: string
  startLine?: number
  endLine?: number
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
 	riskLevel?: string
 	references?: AIRef[]
 	suggestions?: string[]
 	actions?: AIAction[]
 	raw?: string
 	invocationId?: number
 }

 export interface AIDraftResponse {
 	summary: string
 	changeType?: string
 	riskLevel?: string
 	references?: AIRef[]
 	applyContent?: string
 	patch?: string
 	actions?: AIAction[]
 	raw?: string
 	invocationId?: number
 }

 export interface AIDiagnosisResponse {
 	rootCause: string
 	evidence?: string[]
 	recommendedActions?: string[]
 	canAutoFix?: boolean
 	riskLevel?: string
 	references?: AIRef[]
 	fixDraft?: string
 	raw?: string
 	invocationId?: number
 }

 export interface AIReviewFinding {
 	severity: string
 	category: string
 	message: string
 	filePath?: string
 	startLine?: number
 	endLine?: number
 	suggestion?: string
 	confidence?: string
 }

 export interface AIReviewResponse {
  	summary: string
  	blocking?: AIReviewFinding[]
  	highRisk?: AIReviewFinding[]
  	optional?: AIReviewFinding[]
  	riskLevel?: string
  	shouldMerge?: boolean
  	mergeNotes?: string
  	raw?: string
  	invocationId?: number
  }

 export interface AIFeedbackRequest {
 	invocationId: number
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
