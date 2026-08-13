package api

type SyncFailureRequest struct {
	RepoKey         string   `json:"repo_key"`
	Logs            string   `json:"logs"`
	Stderr          string   `json:"stderr"`
	CurrentBranch   string   `json:"current_branch"`
	TrackingBranch  string   `json:"tracking_branch"`
	RecentActions   []string `json:"recent_actions"`
	UserInstruction string   `json:"user_instruction"`
}

type RepoSummaryRequest struct {
	RepoKey         string                 `json:"repo_key"`
	Status          map[string]interface{} `json:"status"`
	Issues          []string               `json:"issues"`
	PendingChanges  int                    `json:"pending_changes"`
	UserInstruction string                 `json:"user_instruction"`
}

type CommitMessageRequest struct {
	RepoKey         string `json:"repo_key"`
	Diff            string `json:"diff"`
	Style           string `json:"style"` // simple, conventional, detailed
	UserInstruction string `json:"user_instruction"`
}

type CodeReviewRequest struct {
	RepoKey          string   `json:"repo_key"`
	Diff             string   `json:"diff"`
	ChangedFiles     []string `json:"changed_files"`
	ExistingFindings []string `json:"existing_findings"`
	Language         string   `json:"language"`
	UserInstruction  string   `json:"user_instruction"`
}

type ReviewReplyRequest struct {
	RepoKey          string   `json:"repo_key"`
	ReviewSummary    string   `json:"review_summary"`
	ReviewerComments []string `json:"reviewer_comments"`
	Tone             string   `json:"tone"` // professional, friendly, concise
}

type ReviewSummaryRequest struct {
	RepoKey         string               `json:"repo_key"`
	TaskID          string               `json:"task_id"`
	TaskStatus      string               `json:"task_status"`
	Findings        []ReviewFindingInput `json:"findings"`
	ChangedFiles    []string             `json:"changed_files"`
	RiskLevel       string               `json:"risk_level,omitempty"`
	UserInstruction string               `json:"user_instruction,omitempty"`
}

type ReviewFindingInput struct {
	Severity string `json:"severity"`
	FilePath string `json:"file_path"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	RuleID   string `json:"rule_id,omitempty"`
}

type ConflictResolveRequest struct {
	RepoKey      string `json:"repo_key"`
	ConflictDiff string `json:"conflict_diff"`
	OursBranch   string `json:"ours_branch"`
	TheirsBranch string `json:"theirs_branch"`
}

type BranchRuleRequest struct {
	RepoKey          string   `json:"repo_key"`
	ExistingBranches []string `json:"existing_branches"`
	RepoType         string   `json:"repo_type"`
}

type SpecTemplateRequest struct {
	RepoKey             string `json:"repo_key"`
	PackageName         string `json:"package_name"`
	SpecType            string `json:"spec_type"`
	ExistingSpecContent string `json:"existing_spec_content"`
}

type SpecRewriteRequest struct {
	RepoKey     string `json:"repo_key"`
	SpecContent string `json:"spec_content"`
	SectionName string `json:"section_name"`
	Instruction string `json:"instruction"`
}

type ProviderBindingRequest struct {
	RemoteRepos      []string          `json:"remote_repos"`
	LocalRepos       []string          `json:"local_repos"`
	ExistingBindings map[string]string `json:"existing_bindings"`
}

type PatchAnalysisRequest struct {
	PatchContent string   `json:"patch_content"`
	TargetBranch string   `json:"target_branch"`
	FileList     []string `json:"file_list"`
}

type AuditSummaryRequest struct {
	Events    []string       `json:"events"`
	Stats     map[string]int `json:"stats"`
	Anomalies []string       `json:"anomalies"`
}

type StatsInsightRequest struct {
	Stats          map[string]interface{} `json:"stats"`
	Trends         map[string][]int       `json:"trends"`
	AuthorActivity map[string]int         `json:"author_activity"`
}

type WebhookFailureRequest struct {
	Payload    string `json:"payload"`
	Response   string `json:"response"`
	StatusCode int    `json:"status_code"`
	EventType  string `json:"event_type"`
}

type AIRef struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Label     string `json:"label"`
	FilePath  string `json:"file_path,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	URL       string `json:"url,omitempty"`
}

type AIAction struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type AIAdviceResponse struct {
	Summary      string     `json:"summary"`
	Confidence   string     `json:"confidence,omitempty"`
	RiskLevel    string     `json:"risk_level,omitempty"`
	References   []AIRef    `json:"references,omitempty"`
	Suggestions  []string   `json:"suggestions,omitempty"`
	Actions      []AIAction `json:"actions,omitempty"`
	Raw          string     `json:"raw,omitempty"`
	InvocationID uint       `json:"invocation_id,omitempty"`
}

type AIDraftResponse struct {
	Summary      string     `json:"summary"`
	ChangeType   string     `json:"change_type,omitempty"`
	RiskLevel    string     `json:"risk_level,omitempty"`
	References   []AIRef    `json:"references,omitempty"`
	ApplyContent string     `json:"apply_content,omitempty"`
	Patch        string     `json:"patch,omitempty"`
	Actions      []AIAction `json:"actions,omitempty"`
	Raw          string     `json:"raw,omitempty"`
	InvocationID uint       `json:"invocation_id,omitempty"`
}

type AIDiagnosisResponse struct {
	RootCause          string   `json:"root_cause"`
	Evidence           []string `json:"evidence,omitempty"`
	RecommendedActions []string `json:"recommended_actions,omitempty"`
	CanAutoFix         bool     `json:"can_auto_fix,omitempty"`
	RiskLevel          string   `json:"risk_level,omitempty"`
	References         []AIRef  `json:"references,omitempty"`
	FixDraft           string   `json:"fix_draft,omitempty"`
	Raw                string   `json:"raw,omitempty"`
	InvocationID       uint     `json:"invocation_id,omitempty"`
}

type AIReviewFinding struct {
	Severity   string `json:"severity"`
	Category   string `json:"category"`
	Message    string `json:"message"`
	FilePath   string `json:"file_path,omitempty"`
	StartLine  int    `json:"start_line,omitempty"`
	EndLine    int    `json:"end_line,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Confidence string `json:"confidence,omitempty"`
}

type AIReviewResponse struct {
	Summary      string            `json:"summary"`
	Blocking     []AIReviewFinding `json:"blocking,omitempty"`
	HighRisk     []AIReviewFinding `json:"high_risk,omitempty"`
	Optional     []AIReviewFinding `json:"optional,omitempty"`
	RiskLevel    string            `json:"risk_level,omitempty"`
	ShouldMerge  bool              `json:"should_merge,omitempty"`
	MergeNotes   string            `json:"merge_notes,omitempty"`
	Raw          string            `json:"raw,omitempty"`
	InvocationID uint              `json:"invocation_id,omitempty"`
}

type AIChatRequest struct {
	TaskType string            `json:"task_type"`
	RepoKey  string            `json:"repo_key,omitempty"`
	Prompt   string            `json:"prompt"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type AIFeedbackRequest struct {
	InvocationID uint   `json:"invocation_id"`
	Feedback     string `json:"feedback"`
}

func (r *AIAdviceResponse) SetRaw(raw string)       { r.Raw = raw }
func (r *AIAdviceResponse) SetInvocationID(id uint) { r.InvocationID = id }

func (r *AIDraftResponse) SetRaw(raw string)       { r.Raw = raw }
func (r *AIDraftResponse) SetInvocationID(id uint) { r.InvocationID = id }

func (r *AIDiagnosisResponse) SetRaw(raw string)       { r.Raw = raw }
func (r *AIDiagnosisResponse) SetInvocationID(id uint) { r.InvocationID = id }

func (r *AIReviewResponse) SetRaw(raw string)       { r.Raw = raw }
func (r *AIReviewResponse) SetInvocationID(id uint) { r.InvocationID = id }
