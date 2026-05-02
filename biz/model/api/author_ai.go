package api

type AuthorAIRequest struct {
	Action  string        `json:"action"`
	RepoKey string        `json:"repoKey,omitempty"`
	Scan    *AuthorScanResult `json:"scan,omitempty"`
	Commits []MismatchedCommit `json:"commits,omitempty"`
}

type AuthorAIResponse struct {
	Action   string               `json:"action"`
	Result   string               `json:"result"`
	Suggest  *AliasSuggestionResult  `json:"suggest,omitempty"`
	Merge    *MergeSuggestionResult  `json:"merge,omitempty"`
	Risk     *RiskAssessmentResult   `json:"risk,omitempty"`
}

type AliasSuggestion struct {
	IdentityID  uint   `json:"identityId"`
	IdentityName string `json:"identityName"`
	AliasName   string `json:"aliasName"`
	AliasEmail  string `json:"aliasEmail"`
	Confidence  string `json:"confidence"`
	Reason      string `json:"reason"`
}

type AliasSuggestionResult struct {
	Suggestions []AliasSuggestion `json:"suggestions"`
	Summary     string            `json:"summary"`
}

type MergeCandidate struct {
	KeepID     uint   `json:"keepId"`
	KeepName   string `json:"keepName"`
	MergeIDs   []uint `json:"mergeIds"`
	MergeNames string `json:"mergeNames"`
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
	RiskLevel       string       `json:"riskLevel"`
	Summary         string       `json:"summary"`
	Factors         []RiskFactor `json:"factors"`
	Recommendations []string     `json:"recommendations"`
}

type AuthorChatRequest struct {
	RepoKey string            `json:"repoKey"`
	Prompt  string            `json:"prompt"`
	History []ChatMessageDTO  `json:"history"`
}

type ChatMessageDTO struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AuthorChatResponse struct {
	Result string `json:"result"`
}
