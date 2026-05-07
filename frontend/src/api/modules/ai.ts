import request from '../request'
import type {
  AIAdviceResponse,
  AIDraftResponse,
  AIDiagnosisResponse,
  AIReviewResponse,
} from '../../types/ai'

 export interface SyncFailureRequest {
   repoKey: string
   logs: string
   stderr: string
   currentBranch: string
   trackingBranch: string
   recentActions: string[]
   userInstruction?: string
 }

 export interface RepoSummaryRequest {
   repoKey: string
   status: Record<string, unknown>
   issues: string[]
   pendingChanges: number
   userInstruction?: string
 }

 export interface CommitMessageRequest {
   repoKey: string
   diff: string
   style?: 'simple' | 'conventional' | 'detailed'
   userInstruction?: string
 }

 export interface CodeReviewRequest {
   repoKey: string
   diff: string
   changedFiles: string[]
   existingFindings: string[]
   language: string
   userInstruction?: string
 }

export interface ReviewReplyRequest {
  repoKey: string
  reviewSummary: string
  reviewerComments: string[]
  tone?: 'professional' | 'friendly' | 'concise'
}

export interface ReviewFindingInput {
  severity: string
  filePath: string
  title: string
  message: string
  ruleId?: string
}

export interface ReviewSummaryRequest {
  repoKey: string
  taskId: string
  taskStatus: string
  findings: ReviewFindingInput[]
  changedFiles: string[]
  riskLevel?: string
  userInstruction?: string
}

export interface ConflictResolveRequest {
  repoKey: string
  conflictDiff: string
  oursBranch: string
  theirsBranch: string
}

export interface BranchRuleRequest {
  repoKey: string
  existingBranches: string[]
  repoType?: string
}

export interface SpecTemplateRequest {
  repoKey: string
  packageName: string
  specType: string
  existingSpecContent?: string
}

export interface SpecRewriteRequest {
  repoKey: string
  specContent: string
  sectionName: string
  instruction: string
}

export interface ProviderBindingRequest {
  remoteRepos: string[]
  localRepos: string[]
  existingBindings: Record<string, string>
}

export interface PatchAnalysisRequest {
  patchContent: string
  targetBranch: string
  fileList: string[]
}

export interface AuditSummaryRequest {
  events: string[]
  stats: Record<string, number>
  anomalies: string[]
}

export interface StatsInsightRequest {
  stats: Record<string, unknown>
  trends: Record<string, number[]>
  authorActivity: Record<string, number>
}

export interface WebhookFailureRequest {
  payload: string
  response: string
  statusCode: number
  eventType: string
}

 export const aiApi = {
   diagnoseSyncFailure: async (data: SyncFailureRequest) =>
      await request.post<any, AIDiagnosisResponse>('/ai/sync/failure', data),

    generateRepoSummary: async (data: RepoSummaryRequest) =>
      await request.post<any, AIAdviceResponse>('/ai/repo/summary', data),

    generateCommitMessage: async (data: CommitMessageRequest) =>
      await request.post<any, AIDraftResponse>('/ai/commit/message', data),

    codeReview: async (data: CodeReviewRequest) =>
      await request.post<any, AIReviewResponse>('/ai/review', data),

    reviewReplyDraft: async (data: ReviewReplyRequest) =>
       await request.post<any, AIDraftResponse>('/ai/review/reply', data),

     reviewSummary: async (data: ReviewSummaryRequest) =>
       await request.post<any, AIReviewResponse>('/ai/review/summary', data),

     resolveConflict: async (data: ConflictResolveRequest) =>
      await request.post<any, AIDraftResponse>('/ai/conflict/resolve', data),

    explainConflict: async (data: ConflictResolveRequest) =>
      await request.post<any, AIAdviceResponse>('/ai/conflict/explain', data),

    generateBranchRule: async (data: BranchRuleRequest) =>
      await request.post<any, AIDraftResponse>('/ai/branch/rule', data),

    generateSpecTemplate: async (data: SpecTemplateRequest) =>
      await request.post<any, AIDraftResponse>('/ai/spec/template', data),

    rewriteSpecSection: async (data: SpecRewriteRequest) =>
      await request.post<any, AIDraftResponse>('/ai/spec/rewrite', data),

    recommendProviderBinding: async (data: ProviderBindingRequest) =>
      await request.post<any, AIAdviceResponse>('/ai/provider/binding', data),

    analyzePatchRisk: async (data: PatchAnalysisRequest) =>
      await request.post<any, AIDiagnosisResponse>('/ai/patch/analyze', data),

    summarizeAuditLogs: async (data: AuditSummaryRequest) =>
      await request.post<any, AIAdviceResponse>('/ai/audit/summary', data),

    analyzeStatsInsight: async (data: StatsInsightRequest) =>
      await request.post<any, AIAdviceResponse>('/ai/stats/insight', data),

     analyzeWebhookFailure: async (data: WebhookFailureRequest) =>
       await request.post<any, AIDiagnosisResponse>('/ai/webhook/failure', data),

     submitFeedback: async (data: { invocationId: number; feedback: string }) =>
       await request.post<any, { success: boolean }>('/ai/feedback', data),
  }
