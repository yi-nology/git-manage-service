package po

import (
	"database/sql"
	"encoding/json"

	"gorm.io/gorm"
)

type BranchRule struct {
	ID                uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Prefix            string `gorm:"size:64;not null" json:"prefix"`
	DisplayName       string `gorm:"size:128" json:"display_name"`
	SourceBranches    string `gorm:"size:500" json:"source_branches"`
	TargetBranches    string `gorm:"size:500" json:"target_branches"`
	RequireTaskID     bool   `gorm:"default:false" json:"require_task_id"`
	TaskIDPattern     string `gorm:"size:200" json:"task_id_pattern"`
	AutoDeleteOnMerge bool   `gorm:"default:true" json:"auto_delete_on_merge"`
	AllowDirectPush   bool   `gorm:"default:false" json:"allow_direct_push"`
	RequireCodeReview bool   `gorm:"default:false" json:"require_code_review"`
	SortOrder         int    `gorm:"default:0" json:"sort_order"`
}

func (BranchRule) TableName() string { return "branch_rules_meta" }

func (r *BranchRule) GetSourceBranches() []string {
	if r.SourceBranches == "" {
		return nil
	}
	var list []string
	json.Unmarshal([]byte(r.SourceBranches), &list)
	return list
}

func (r *BranchRule) SetSourceBranches(list []string) {
	b, _ := json.Marshal(list)
	r.SourceBranches = string(b)
}

func (r *BranchRule) GetTargetBranches() []string {
	if r.TargetBranches == "" {
		return nil
	}
	var list []string
	json.Unmarshal([]byte(r.TargetBranches), &list)
	return list
}

func (r *BranchRule) SetTargetBranches(list []string) {
	b, _ := json.Marshal(list)
	r.TargetBranches = string(b)
}

type BranchRuleSet struct {
	gorm.Model
	ScopeType        string `gorm:"size:32;not null;index"`
	ScopeID          string `gorm:"size:128;not null;index"`
	Enabled          bool   `gorm:"default:true" json:"enabled"`
	RulesJSON        string `gorm:"type:text" json:"rules_json"`
	ProtectedJSON    string `gorm:"type:text" json:"protected_json"`
	UseCustomRules   bool   `gorm:"default:false" json:"use_custom_rules"`
}

func (BranchRuleSet) TableName() string { return "branch_rule_sets" }

func (s *BranchRuleSet) GetRules() []BranchRule {
	if s.RulesJSON == "" {
		return nil
	}
	var rules []BranchRule
	json.Unmarshal([]byte(s.RulesJSON), &rules)
	return rules
}

func (s *BranchRuleSet) SetRules(rules []BranchRule) {
	b, _ := json.Marshal(rules)
	s.RulesJSON = string(b)
}

func (s *BranchRuleSet) GetProtected() []string {
	if s.ProtectedJSON == "" {
		return nil
	}
	var list []string
	json.Unmarshal([]byte(s.ProtectedJSON), &list)
	return list
}

func (s *BranchRuleSet) SetProtected(list []string) {
	b, _ := json.Marshal(list)
	s.ProtectedJSON = string(b)
}

type BranchRuleOverride struct {
	gorm.Model
	ProviderConfigID uint   `gorm:"index;not null"`
	PlatformOwner    string `gorm:"size:200;not null"`
	PlatformRepo     string `gorm:"size:200;not null"`
	UseCustomRules   bool   `gorm:"default:false"`
	RulesJSON        string `gorm:"type:text"`
	ProtectedJSON    string `gorm:"type:text"`
}

func (BranchRuleOverride) TableName() string { return "branch_rule_overrides" }

func (o *BranchRuleOverride) GetRules() []BranchRule {
	if o.RulesJSON == "" {
		return nil
	}
	var rules []BranchRule
	json.Unmarshal([]byte(o.RulesJSON), &rules)
	return rules
}

func (o *BranchRuleOverride) SetRules(rules []BranchRule) {
	b, _ := json.Marshal(rules)
	o.RulesJSON = string(b)
}

func (o *BranchRuleOverride) GetProtected() []string {
	if o.ProtectedJSON == "" {
		return nil
	}
	var list []string
	json.Unmarshal([]byte(o.ProtectedJSON), &list)
	return list
}

func (o *BranchRuleOverride) SetProtected(list []string) {
	b, _ := json.Marshal(list)
	o.ProtectedJSON = string(b)
}

func init() {
	_ = sql.NullInt32{}
}
