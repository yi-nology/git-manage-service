package api

type SyncFailureRequest struct {
	RepoKey         string   `json:"repoKey"`
	Logs            string   `json:"logs"`
	Stderr          string   `json:"stderr"`
	CurrentBranch   string   `json:"currentBranch"`
	TrackingBranch  string   `json:"trackingBranch"`
	RecentActions   []string `json:"recentActions"`
	UserInstruction string   `json:"userInstruction"`
}

type RepoSummaryRequest struct {
	RepoKey         string                 `json:"repoKey"`
	Status          map[string]interface{} `json:"status"`
	Issues          []string               `json:"issues"`
	PendingChanges  int                    `json:"pendingChanges"`
	UserInstruction string                 `json:"userInstruction"`
}

type CommitMessageRequest struct {
	RepoKey         string `json:"repoKey"`
	Diff            string `json:"diff"`
	Style           string `json:"style"` // simple, conventional, detailed
	UserInstruction string `json:"userInstruction"`
}

type CodeReviewRequest struct {
	RepoKey          string   `json:"repoKey"`
	Diff             string   `json:"diff"`
	ChangedFiles     []string `json:"changedFiles"`
	ExistingFindings []string `json:"existingFindings"`
	Language         string   `json:"language"`
	UserInstruction  string   `json:"userInstruction"`
}

type ReviewReplyRequest struct {
	RepoKey          string   `json:"repoKey"`
	ReviewSummary    string   `json:"reviewSummary"`
	ReviewerComments []string `json:"reviewerComments"`
	Tone             string   `json:"tone"` // professional, friendly, concise
}

type ReviewSummaryRequest struct {
	RepoKey         string               `json:"repoKey"`
	TaskID          string               `json:"taskId"`
	TaskStatus      string               `json:"taskStatus"`
	Findings        []ReviewFindingInput `json:"findings"`
	ChangedFiles    []string             `json:"changedFiles"`
	RiskLevel       string               `json:"riskLevel,omitempty"`
	UserInstruction string               `json:"userInstruction,omitempty"`
}

type ReviewFindingInput struct {
	Severity string `json:"severity"`
	FilePath string `json:"filePath"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	RuleID   string `json:"ruleId,omitempty"`
}

type ConflictResolveRequest struct {
	RepoKey      string `json:"repoKey"`
	ConflictDiff string `json:"conflictDiff"`
	OursBranch   string `json:"oursBranch"`
	TheirsBranch string `json:"theirsBranch"`
}

type BranchRuleRequest struct {
	RepoKey          string   `json:"repoKey"`
	ExistingBranches []string `json:"existingBranches"`
	RepoType         string   `json:"repoType"`
}

type SpecTemplateRequest struct {
	RepoKey             string `json:"repoKey"`
	PackageName         string `json:"packageName"`
	SpecType            string `json:"specType"`
	ExistingSpecContent string `json:"existingSpecContent"`
}

type SpecRewriteRequest struct {
	RepoKey     string `json:"repoKey"`
	SpecContent string `json:"specContent"`
	SectionName string `json:"sectionName"`
	Instruction string `json:"instruction"`
}

type ProviderBindingRequest struct {
	RemoteRepos      []string          `json:"remoteRepos"`
	LocalRepos       []string          `json:"localRepos"`
	ExistingBindings map[string]string `json:"existingBindings"`
}

type PatchAnalysisRequest struct {
	PatchContent string   `json:"patchContent"`
	TargetBranch string   `json:"targetBranch"`
	FileList     []string `json:"fileList"`
}

type AuditSummaryRequest struct {
	Events    []string       `json:"events"`
	Stats     map[string]int `json:"stats"`
	Anomalies []string       `json:"anomalies"`
}

type StatsInsightRequest struct {
	Stats          map[string]interface{} `json:"stats"`
	Trends         map[string][]int       `json:"trends"`
	AuthorActivity map[string]int         `json:"authorActivity"`
}

type WebhookFailureRequest struct {
	Payload    string `json:"payload"`
	Response   string `json:"response"`
	StatusCode int    `json:"statusCode"`
	EventType  string `json:"eventType"`
}

type AIRef struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Label     string `json:"label"`
	FilePath  string `json:"filePath,omitempty"`
	StartLine int    `json:"startLine,omitempty"`
	EndLine   int    `json:"endLine,omitempty"`
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
	RiskLevel    string     `json:"riskLevel,omitempty"`
	References   []AIRef    `json:"references,omitempty"`
	Suggestions  []string   `json:"suggestions,omitempty"`
	Actions      []AIAction `json:"actions,omitempty"`
	Raw          string     `json:"raw,omitempty"`
	InvocationID uint       `json:"invocationId,omitempty"`
}

type AIDraftResponse struct {
	Summary      string     `json:"summary"`
	ChangeType   string     `json:"changeType,omitempty"`
	RiskLevel    string     `json:"riskLevel,omitempty"`
	References   []AIRef    `json:"references,omitempty"`
	ApplyContent string     `json:"applyContent,omitempty"`
	Patch        string     `json:"patch,omitempty"`
	Actions      []AIAction `json:"actions,omitempty"`
	Raw          string     `json:"raw,omitempty"`
	InvocationID uint       `json:"invocationId,omitempty"`
}

type AIDiagnosisResponse struct {
	RootCause          string   `json:"rootCause"`
	Evidence           []string `json:"evidence,omitempty"`
	RecommendedActions []string `json:"recommendedActions,omitempty"`
	CanAutoFix         bool     `json:"canAutoFix,omitempty"`
	RiskLevel          string   `json:"riskLevel,omitempty"`
	References         []AIRef  `json:"references,omitempty"`
	FixDraft           string   `json:"fixDraft,omitempty"`
	Raw                string   `json:"raw,omitempty"`
	InvocationID       uint     `json:"invocationId,omitempty"`
}

type AIReviewFinding struct {
	Severity   string `json:"severity"`
	Category   string `json:"category"`
	Message    string `json:"message"`
	FilePath   string `json:"filePath,omitempty"`
	StartLine  int    `json:"startLine,omitempty"`
	EndLine    int    `json:"endLine,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Confidence string `json:"confidence,omitempty"`
}

type AIReviewResponse struct {
	Summary      string            `json:"summary"`
	Blocking     []AIReviewFinding `json:"blocking,omitempty"`
	HighRisk     []AIReviewFinding `json:"highRisk,omitempty"`
	Optional     []AIReviewFinding `json:"optional,omitempty"`
	RiskLevel    string            `json:"riskLevel,omitempty"`
	ShouldMerge  bool              `json:"shouldMerge,omitempty"`
	MergeNotes   string            `json:"mergeNotes,omitempty"`
	Raw          string            `json:"raw,omitempty"`
	InvocationID uint              `json:"invocationId,omitempty"`
}

type AIChatRequest struct {
	TaskType string            `json:"taskType"`
	RepoKey  string            `json:"repoKey,omitempty"`
	Prompt   string            `json:"prompt"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type AIFeedbackRequest struct {
	InvocationID uint   `json:"invocationId"`
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
