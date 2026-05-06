package po

import "gorm.io/gorm"

type AIInvocation struct {
	gorm.Model
	TaskType      string `gorm:"size:64;index" json:"taskType"`
	ProviderName  string `gorm:"size:128;index" json:"providerName"`
	PromptVersion string `gorm:"size:64;index" json:"promptVersion"`
	RepoKey       string `gorm:"size:256;index" json:"repoKey,omitempty"`
	OperatorID    uint   `gorm:"index" json:"operatorId,omitempty"`
	RelatedID     string `gorm:"size:128;index" json:"relatedId,omitempty"`
	InputHash     string `gorm:"size:64;index" json:"inputHash"`
	OutputHash    string `gorm:"size:64;index" json:"outputHash"`
	Status        string `gorm:"size:32;index" json:"status"`
	ErrorMessage  string `gorm:"type:text" json:"errorMessage"`
	LatencyMs     int64  `json:"latencyMs"`
	InputChars    int    `json:"inputChars"`
	OutputChars   int    `json:"outputChars"`
	UserFeedback  string `gorm:"size:32;index" json:"userFeedback,omitempty"`
	MetadataJSON  string `gorm:"type:text" json:"metadataJson"`
}

func (AIInvocation) TableName() string {
	return "ai_invocations"
}
