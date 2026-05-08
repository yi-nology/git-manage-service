package api

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewTaskDTO struct {
	ID               uint       `json:"id"`
	RepoID           uint       `json:"repo_id"`
	RepoKey          string     `json:"repo_key"`
	RepoName         string     `json:"repo_name"`
	ProviderConfigID uint       `json:"provider_config_id"`
	Platform         string     `json:"platform"`
	EventType        string     `json:"event_type"`
	MRIID            string     `json:"mr_iid"`
	SourceBranch     string     `json:"source_branch"`
	TargetBranch     string     `json:"target_branch"`
	CommitSHA        string     `json:"commit_sha"`
	TriggerType      string     `json:"trigger_type"`
	TriggerUser      string     `json:"trigger_user"`
	Status           string     `json:"status"`
	RiskLevel        string     `json:"risk_level"`
	Summary          string     `json:"summary"`
	ErrorMessage     string     `json:"error_message"`
	StartedAt        *time.Time `json:"started_at"`
	FinishedAt       *time.Time `json:"finished_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	FindingsCount    int        `json:"findings_count"`
	RawDiff          string     `json:"raw_diff"`
}

func NewReviewTaskDTO(t po.ReviewTask) ReviewTaskDTO {
	return ReviewTaskDTO{
		ID:               t.ID,
		RepoID:           t.RepoID,
		ProviderConfigID: t.ProviderConfigID,
		Platform:         t.Platform,
		EventType:        t.EventType,
		MRIID:            t.MRIID,
		SourceBranch:     t.SourceBranch,
		TargetBranch:     t.TargetBranch,
		CommitSHA:        t.CommitSHA,
		TriggerType:      t.TriggerType,
		TriggerUser:      t.TriggerUser,
		Status:           t.Status,
		RiskLevel:        t.RiskLevel,
		Summary:          t.Summary,
		ErrorMessage:     t.ErrorMessage,
		RawDiff:          t.RawDiff,
		StartedAt:        t.StartedAt,
		FinishedAt:       t.FinishedAt,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}

type ReviewFindingDTO struct {
	ID          uint   `json:"id"`
	TaskID      uint   `json:"task_id"`
	Source      string `json:"source"`
	RuleID      string `json:"rule_id"`
	Severity    string `json:"severity"`
	FilePath    string `json:"file_path"`
	OldLine     int    `json:"old_line"`
	NewLine     int    `json:"new_line"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion"`
	Fingerprint string `json:"fingerprint"`
}

func NewReviewFindingDTO(f po.ReviewFinding) ReviewFindingDTO {
	return ReviewFindingDTO{
		ID:          f.ID,
		TaskID:      f.TaskID,
		Source:      f.Source,
		RuleID:      f.RuleID,
		Severity:    f.Severity,
		FilePath:    f.FilePath,
		OldLine:     f.OldLine,
		NewLine:     f.NewLine,
		Title:       f.Title,
		Message:     f.Message,
		Suggestion:  f.Suggestion,
		Fingerprint: f.Fingerprint,
	}
}

type ReviewRepoConfigDTO struct {
	ID                   uint            `json:"id"`
	ProviderConfigID     uint            `json:"provider_config_id"`
	PlatformOwner        string          `json:"platform_owner"`
	PlatformRepo         string          `json:"platform_repo"`
	Enabled              bool            `json:"enabled"`
	BlockOnHigh          bool            `json:"block_on_high"`
	AutoReviewOnMR       bool            `json:"auto_review_on_mr"`
	LLMProvider          string          `json:"llm_provider"`
	MaxFiles             int             `json:"max_files"`
	MaxDiffLines         int             `json:"max_diff_lines"`
	RuleOverrides        string          `json:"rule_overrides_json"`
	ScopeNote            string          `json:"scope_note"`
	LinkedRepos          []LinkedRepoDTO `json:"linked_repos"`
	PromptPrefixOverride string          `json:"prompt_prefix_override"`
	PromptIntentOverride string          `json:"prompt_intent_override"`
}

type LinkedRepoDTO struct {
	ID   uint   `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

func NewReviewRepoConfigDTO(c po.ReviewRepoConfig, repos []LinkedRepoDTO) ReviewRepoConfigDTO {
	return ReviewRepoConfigDTO{
		ID:                   c.ID,
		ProviderConfigID:     c.ProviderConfigID,
		PlatformOwner:        c.PlatformOwner,
		PlatformRepo:         c.PlatformRepo,
		Enabled:              c.Enabled,
		BlockOnHigh:          c.BlockOnHigh,
		AutoReviewOnMR:       c.AutoReviewOnMR,
		LLMProvider:          c.LLMProvider,
		MaxFiles:             c.MaxFiles,
		MaxDiffLines:         c.MaxDiffLines,
		RuleOverrides:        c.RuleOverridesJSON,
		ScopeNote:            c.ScopeNote,
		LinkedRepos:          repos,
		PromptPrefixOverride: c.PromptPrefixOverride,
		PromptIntentOverride: c.PromptIntentOverride,
	}
}

type MergeCheckDTO struct {
	Mergeable bool                `json:"mergeable"`
	Checks    []MergeCheckItemDTO `json:"checks"`
}

type MergeCheckItemDTO struct {
	CheckType string `json:"check_type"`
	Status    string `json:"status"`
	RiskLevel string `json:"risk_level"`
	Message   string `json:"message"`
}
