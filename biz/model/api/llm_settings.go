package api

import "github.com/yi-nology/git-manage-service/biz/model/po"

type LLMProviderDTO struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	BaseURL        string `json:"base_url"`
	APIKey         string `json:"api_key,omitempty"`
	Model          string `json:"model"`
	MaxTokens      int    `json:"max_tokens"`
	IsDefault      bool   `json:"is_default"`
	IsEmbedding    bool   `json:"is_embedding"`
	EmbeddingModel string `json:"embedding_model"`
	PresetID       string `json:"preset_id,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func NewLLMProviderDTO(p po.LLMProvider) LLMProviderDTO {
	return LLMProviderDTO{
		ID:             p.ID,
		Name:           p.Name,
		Type:           p.Type,
		BaseURL:        p.BaseURL,
		APIKey:         p.APIKey,
		Model:          p.AIModel,
		MaxTokens:      p.MaxTokens,
		IsDefault:      p.IsDefault,
		IsEmbedding:    p.IsEmbedding,
		EmbeddingModel: p.EmbeddingModel,
		PresetID:       p.PresetID,
		Protocol:       p.Protocol,
		CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type CodeReviewGlobalSettingsDTO struct {
	Enabled        bool `json:"enabled"`
	AutoReviewOnMR bool `json:"auto_review_on_mr"`
	BlockOnHigh    bool `json:"block_on_high"`
	MaxFiles       int  `json:"max_files"`
	MaxDiffLines   int  `json:"max_diff_lines"`
	RAGEnabled     bool `json:"rag_enabled"`
}
