package api

type BranchRuleDTO struct {
	ID                uint     `json:"id"`
	Prefix            string   `json:"prefix"`
	DisplayName       string   `json:"display_name"`
	SourceBranches    []string `json:"source_branches"`
	TargetBranches    []string `json:"target_branches"`
	RequireTaskID     bool     `json:"require_task_id"`
	TaskIDPattern     string   `json:"task_id_pattern"`
	AutoDeleteOnMerge bool     `json:"auto_delete_on_merge"`
	AllowDirectPush   bool     `json:"allow_direct_push"`
	RequireCodeReview bool     `json:"require_code_review"`
	SortOrder         int      `json:"sort_order"`
}

type BranchRuleSetDTO struct {
	Enabled        bool           `json:"enabled"`
	Rules          []BranchRuleDTO `json:"rules"`
	Protected      []string       `json:"protected_branches"`
	UseCustomRules bool           `json:"use_custom_rules,omitempty"`
	ScopeType      string         `json:"scope_type,omitempty"`
	LinkedRepos    []LinkedRepoDTO `json:"linked_repos,omitempty"`
}

type BranchValidationResult struct {
	Valid    bool                    `json:"valid"`
	Errors   []BranchValidationError `json:"errors,omitempty"`
	RuleName string                  `json:"rule_name,omitempty"`
}

type BranchValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type RemoteRepoBranchRulesDTO struct {
	ProviderConfigID uint            `json:"provider_config_id"`
	PlatformOwner    string          `json:"platform_owner"`
	PlatformRepo     string          `json:"platform_repo"`
	UseCustomRules   bool            `json:"use_custom_rules"`
	Rules            []BranchRuleDTO `json:"rules"`
	Protected        []string        `json:"protected_branches"`
	LinkedRepos      []LinkedRepoDTO `json:"linked_repos"`
}
