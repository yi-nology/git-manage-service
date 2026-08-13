import request from '../request'
import type {
  AIAdviceResponse,
  AIDraftResponse,
  AIDiagnosisResponse,
  AIReviewResponse,
} from '../../types/ai'

 export interface SyncFailureRequest {
   repo_key: string
   logs: string
   stderr: string
   current_branch: string
   tracking_branch: string
   recent_actions: string[]
   user_instruction?: string
 }

 export interface RepoSummaryRequest {
   repo_key: string
   status: Record<string, unknown>
   issues: string[]
   pending_changes: number
   user_instruction?: string
 }

 export interface CommitMessageRequest {
   repo_key: string
   diff: string
   style?: 'simple' | 'conventional' | 'detailed'
   user_instruction?: string
 }

 export interface CodeReviewRequest {
   repo_key: string
   diff: string
   changed_files: string[]
   existing_findings: string[]
   language: string
   user_instruction?: string
 }

export interface ReviewReplyRequest {
  repo_key: string
  review_summary: string
  reviewer_comments: string[]
  tone?: 'professional' | 'friendly' | 'concise'
}

export interface ReviewFindingInput {
  severity: string
  file_path: string
  title: string
  message: string
  rule_id?: string
}

export interface ReviewSummaryRequest {
  repo_key: string
  task_id: string
  task_status: string
  findings: ReviewFindingInput[]
  changed_files: string[]
  risk_level?: string
  user_instruction?: string
}

export interface ConflictResolveRequest {
  repo_key: string
  conflict_diff: string
  ours_branch: string
  theirs_branch: string
}

export interface BranchRuleRequest {
  repo_key: string
  existing_branches: string[]
  repo_type?: string
}

export interface SpecTemplateRequest {
  repo_key: string
  package_name: string
  spec_type: string
  existing_spec_content?: string
}

export interface SpecRewriteRequest {
  repo_key: string
  spec_content: string
  section_name: string
  instruction: string
}

export interface ProviderBindingRequest {
  remote_repos: string[]
  local_repos: string[]
  existing_bindings: Record<string, string>
}

export interface PatchAnalysisRequest {
  patch_content: string
  target_branch: string
  file_list: string[]
}

export interface AuditSummaryRequest {
  events: string[]
  stats: Record<string, number>
  anomalies: string[]
}

export interface StatsInsightRequest {
  stats: Record<string, unknown>
  trends: Record<string, number[]>
  author_activity: Record<string, number>
}

export interface WebhookFailureRequest {
  payload: string
  response: string
  status_code: number
  event_type: string
}

 export const aiApi = {
   diagnose_sync_failure: async (data: SyncFailureRequest) =>
      await request.post<any, AIDiagnosisResponse>('/ai/sync/failure', data),

    generate_repo_summary: async (data: RepoSummaryRequest) =>
      await request.post<any, AIAdviceResponse>('/ai/repo/summary', data),

    generateCommitMessage: async (data: CommitMessageRequest) =>
      await request.post<any, AIDraftResponse>('/ai/commit/message', data),

    code_review: async (data: CodeReviewRequest) =>
      await request.post<any, AIReviewResponse>('/ai/review', data),

    review_reply_draft: async (data: ReviewReplyRequest) =>
       await request.post<any, AIDraftResponse>('/ai/review/reply', data),

     review_summary: async (data: ReviewSummaryRequest) =>
       await request.post<any, AIReviewResponse>('/ai/review/summary', data),

     resolve_conflict: async (data: ConflictResolveRequest) =>
      await request.post<any, AIDraftResponse>('/ai/conflict/resolve', data),

    explain_conflict: async (data: ConflictResolveRequest) =>
      await request.post<any, AIAdviceResponse>('/ai/conflict/explain', data),

    generate_branch_rule: async (data: BranchRuleRequest) =>
      await request.post<any, AIDraftResponse>('/ai/branch/rule', data),

    generate_spec_template: async (data: SpecTemplateRequest) =>
      await request.post<any, AIDraftResponse>('/ai/spec/template', data),

    rewrite_spec_section: async (data: SpecRewriteRequest) =>
      await request.post<any, AIDraftResponse>('/ai/spec/rewrite', data),

    recommend_provider_binding: async (data: ProviderBindingRequest) =>
      await request.post<any, AIAdviceResponse>('/ai/provider/binding', data),

    analyze_patch_risk: async (data: PatchAnalysisRequest) =>
      await request.post<any, AIDiagnosisResponse>('/ai/patch/analyze', data),

    summarize_audit_logs: async (data: AuditSummaryRequest) =>
      await request.post<any, AIAdviceResponse>('/ai/audit/summary', data),

    analyze_stats_insight: async (data: StatsInsightRequest) =>
      await request.post<any, AIAdviceResponse>('/ai/stats/insight', data),

     analyze_webhook_failure: async (data: WebhookFailureRequest) =>
       await request.post<any, AIDiagnosisResponse>('/ai/webhook/failure', data),

     submit_feedback: async (data: { invocation_id: number; feedback: string }) =>
       await request.post<any, { success: boolean }>('/ai/feedback', data),
  }
