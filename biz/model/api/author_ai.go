package api

type AuthorAIRequest struct {
	Action  string             `json:"action"`
	RepoKey string             `json:"repo_key,omitempty"`
	Scan    *AuthorScanResult  `json:"scan,omitempty"`
	Commits []MismatchedCommit `json:"commits,omitempty"`
}

type AuthorAIResponse struct {
	Action  string                 `json:"action"`
	Result  string                 `json:"result"`
	Suggest *AliasSuggestionResult `json:"suggest,omitempty"`
	Merge   *MergeSuggestionResult `json:"merge,omitempty"`
	Risk    *RiskAssessmentResult  `json:"risk,omitempty"`
}

type AliasSuggestion struct {
	IdentityID   uint   `json:"identity_id"`
	IdentityName string `json:"identity_name"`
	AliasName    string `json:"alias_name"`
	AliasEmail   string `json:"alias_email"`
	Confidence   string `json:"confidence"`
	Reason       string `json:"reason"`
}

type AliasSuggestionResult struct {
	Suggestions []AliasSuggestion `json:"suggestions"`
	Summary     string            `json:"summary"`
}

type MergeCandidate struct {
	KeepID     uint   `json:"keep_id"`
	KeepName   string `json:"keep_name"`
	MergeIDs   []uint `json:"merge_ids"`
	MergeNames string `json:"merge_names"`
	Reason     string `json:"reason"`
}

type MergeSuggestionResult struct {
	Merges  []MergeCandidate `json:"merges"`
	Summary string           `json:"summary"`
}

type RiskFactor struct {
	Level          string `json:"level"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

type RiskAssessmentResult struct {
	RiskLevel       string       `json:"risk_level"`
	Summary         string       `json:"summary"`
	Factors         []RiskFactor `json:"factors"`
	Recommendations []string     `json:"recommendations"`
}

type AuthorChatRequest struct {
	RepoKey string           `json:"repo_key"`
	Prompt  string           `json:"prompt"`
	History []ChatMessageDTO `json:"history"`
}

type ChatMessageDTO struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AuthorChatResponse struct {
	Result string `json:"result"`
}
