package llm

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	settingsModel "github.com/yi-nology/git-manage-service/biz/model/settings"
	"gorm.io/gorm"
)

func ListProviders() ([]*settingsModel.LLMProviderInfo, error) {
	providers, err := db.NewLLMProviderDAO().FindAll()
	if err != nil {
		return nil, err
	}
	dtos := make([]*settingsModel.LLMProviderInfo, 0, len(providers))
	for _, p := range providers {
		dtos = append(dtos, convertToProtoLLMProvider(&p))
	}
	return dtos, nil
}

func GetProviderByID(id uint) (*settingsModel.LLMProviderInfo, error) {
	p, err := db.NewLLMProviderDAO().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}
	return convertToProtoLLMProvider(p), nil
}

func CreateProvider(req *settingsModel.LLMProviderInfo) (*settingsModel.LLMProviderInfo, error) {
	dao := db.NewLLMProviderDAO()

	maxTokens := int(req.MaxTokens)
	if maxTokens == 0 {
		maxTokens = 4096
	}

	p := &po.LLMProvider{
		Name:           req.Name,
		Type:           req.Type,
		BaseURL:        req.BaseUrl,
		APIKey:         req.ApiKey,
		AIModel:        req.Model,
		MaxTokens:      maxTokens,
		IsDefault:      req.IsDefault,
		IsEmbedding:    req.IsEmbedding,
		EmbeddingModel: req.EmbeddingModel,
		PresetID:       req.PresetId,
		Protocol:       req.Protocol,
	}

	if p.IsDefault {
		dao.ClearAllDefault()
	}

	existing, err := dao.FindByNameUnscoped(req.Name)
	if err == nil && existing != nil {
		existing.Type = p.Type
		existing.BaseURL = p.BaseURL
		existing.APIKey = p.APIKey
		existing.AIModel = p.AIModel
		existing.MaxTokens = p.MaxTokens
		existing.IsDefault = p.IsDefault
		existing.IsEmbedding = p.IsEmbedding
		existing.EmbeddingModel = p.EmbeddingModel
		existing.PresetID = p.PresetID
		existing.Protocol = p.Protocol
		existing.DeletedAt = gorm.DeletedAt{}
		if err := dao.Save(existing); err != nil {
			return nil, fmt.Errorf("failed to restore provider: %w", err)
		}
		return convertToProtoLLMProvider(existing), nil
	}

	if err := dao.Create(p); err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	if p.IsDefault || isFirstProvider(p.ID) {
		dao.SetDefault(p.ID)
		p.IsDefault = true
	}

	RegisterProvider(p)
	return convertToProtoLLMProvider(p), nil
}

func UpdateProvider(id uint, req *settingsModel.LLMProviderInfo) (*settingsModel.LLMProviderInfo, error) {
	dao := db.NewLLMProviderDAO()
	p, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	exists, _ := dao.ExistsByNameExcludeID(req.Name, id)
	if exists {
		return nil, fmt.Errorf("provider name %q already exists", req.Name)
	}

	p.Name = req.Name
	p.Type = req.Type
	p.BaseURL = req.BaseUrl
	p.AIModel = req.Model
	p.MaxTokens = int(req.MaxTokens)
	p.IsDefault = req.IsDefault
	p.IsEmbedding = req.IsEmbedding
	p.EmbeddingModel = req.EmbeddingModel
	p.PresetID = req.PresetId
	p.Protocol = req.Protocol
	if req.ApiKey != "" {
		p.APIKey = req.ApiKey
	}

	if p.IsDefault {
		dao.ClearAllDefault()
	}

	if err := dao.Save(p); err != nil {
		return nil, fmt.Errorf("failed to save provider: %w", err)
	}

	RegisterProvider(p)
	return convertToProtoLLMProvider(p), nil
}

func DeleteProvider(id uint) error {
	dao := db.NewLLMProviderDAO()
	p, err := dao.FindByID(id)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}
	if err := dao.Delete(id); err != nil {
		return err
	}
	UnregisterProvider(p.Name)
	if p.IsDefault {
		providers, _ := dao.FindAll()
		if len(providers) > 0 {
			dao.SetDefault(providers[0].ID)
		}
	}
	return nil
}

func SetDefaultProvider(id uint) error {
	dao := db.NewLLMProviderDAO()
	if _, err := dao.FindByID(id); err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}
	return dao.SetDefault(id)
}

func TestProvider(id uint) error {
	p, err := db.NewLLMProviderDAO().FindByID(id)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	var provider Provider
	switch p.Type {
	case "openai_compatible":
		provider = NewOpenAICompatible(p.BaseURL, p.APIKey, p.AIModel, p.MaxTokens, p.Name)
	case "ollama":
		provider = NewOllama(p.BaseURL, p.AIModel, p.MaxTokens)
	case "anthropic":
		provider = NewAnthropic(p.BaseURL, p.APIKey, p.AIModel, p.MaxTokens)
	case "gemini":
		provider = NewGemini(p.BaseURL, p.APIKey, p.AIModel, p.MaxTokens)
	default:
		return fmt.Errorf("unknown provider type: %s", p.Type)
	}

	_, err = provider.Chat(context.Background(), &ChatRequest{
		Messages:     []ChatMessage{{Role: "user", Content: "Hi, reply with OK."}},
		MaxTokens:    16,
		SystemPrompt: "Reply with exactly: OK",
	})
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	return nil
}

func isFirstProvider(id uint) bool {
	all, err := db.NewLLMProviderDAO().FindAll()
	if err != nil || len(all) == 0 {
		return false
	}
	return len(all) == 1 && all[0].ID == id
}

func convertToProtoLLMProvider(p *po.LLMProvider) *settingsModel.LLMProviderInfo {
	return &settingsModel.LLMProviderInfo{
		Id:             uint64(p.ID),
		Name:           p.Name,
		Type:           p.Type,
		BaseUrl:        p.BaseURL,
		ApiKey:         p.APIKey,
		Model:          p.AIModel,
		MaxTokens:      int32(p.MaxTokens),
		IsDefault:      p.IsDefault,
		IsEmbedding:    p.IsEmbedding,
		EmbeddingModel: p.EmbeddingModel,
		PresetId:       p.PresetID,
		Protocol:       p.Protocol,
		CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func RegisterProvider(p *po.LLMProvider) {
	providersMu.Lock()
	defer providersMu.Unlock()
	var prov Provider
	switch p.Type {
	case "openai_compatible":
		prov = NewOpenAICompatible(p.BaseURL, p.APIKey, p.AIModel, p.MaxTokens, p.Name)
	case "ollama":
		prov = NewOllama(p.BaseURL, p.AIModel, p.MaxTokens)
	case "anthropic":
		prov = NewAnthropic(p.BaseURL, p.APIKey, p.AIModel, p.MaxTokens)
	case "gemini":
		prov = NewGemini(p.BaseURL, p.APIKey, p.AIModel, p.MaxTokens)
	default:
		log.Printf("[LLM] Unknown provider type %q for %q", p.Type, p.Name)
		return
	}
	providers[p.Name] = prov
	log.Printf("[LLM] Registered provider: %s (%s)", p.Name, p.Type)
}

func UnregisterProvider(name string) {
	providersMu.Lock()
	defer providersMu.Unlock()
	delete(providers, name)
	log.Printf("[LLM] Unregistered provider: %s", name)
}

func InitProvidersFromDB() {
	providerList, err := db.NewLLMProviderDAO().FindAll()
	if err != nil {
		log.Printf("[LLM] Failed to load providers from DB: %v", err)
		return
	}
	for _, p := range providerList {
		RegisterProvider(&p)
	}
}

func GetProviderNames() []string {
	providersMu.RLock()
	defer providersMu.RUnlock()
	names := make([]string, 0, len(providers))
	for k := range providers {
		names = append(names, k)
	}
	return names
}

func init() {
	_ = strconv.Itoa
}
