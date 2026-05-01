package llm

import (
	"context"
	"fmt"
	"log"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func ListProviders() ([]api.LLMProviderDTO, error) {
	providers, err := db.NewLLMProviderDAO().FindAll()
	if err != nil {
		return nil, err
	}
	dtos := make([]api.LLMProviderDTO, 0, len(providers))
	for _, p := range providers {
		dtos = append(dtos, api.NewLLMProviderDTO(p))
	}
	return dtos, nil
}

func GetProviderByID(id uint) (*api.LLMProviderDTO, error) {
	p, err := db.NewLLMProviderDAO().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}
	dto := api.NewLLMProviderDTO(*p)
	return &dto, nil
}

func CreateProvider(req api.LLMProviderDTO) (*api.LLMProviderDTO, error) {
	dao := db.NewLLMProviderDAO()
	exists, _ := dao.ExistsByName(req.Name)
	if exists {
		return nil, fmt.Errorf("provider name %q already exists", req.Name)
	}

	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}

	p := &po.LLMProvider{
		Name:      req.Name,
		Type:      req.Type,
		BaseURL:   req.BaseURL,
		APIKey:    req.APIKey,
		AIModel:   req.Model,
		MaxTokens: req.MaxTokens,
		IsDefault: req.IsDefault,
		PresetID:  req.PresetID,
		Protocol:  req.Protocol,
	}

	if p.IsDefault {
		dao.ClearAllDefault()
	}

	if err := dao.Create(p); err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	if p.IsDefault || isFirstProvider(p.ID) {
		dao.SetDefault(p.ID)
		p.IsDefault = true
	}

	RegisterProvider(p)
	dto := api.NewLLMProviderDTO(*p)
	return &dto, nil
}

func UpdateProvider(id uint, req api.LLMProviderDTO) (*api.LLMProviderDTO, error) {
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
	p.BaseURL = req.BaseURL
	p.AIModel = req.Model
	p.MaxTokens = req.MaxTokens
	p.IsDefault = req.IsDefault
	p.PresetID = req.PresetID
	p.Protocol = req.Protocol
	if req.APIKey != "" {
		p.APIKey = req.APIKey
	}

	if p.IsDefault {
		dao.ClearAllDefault()
	}

	if err := dao.Save(p); err != nil {
		return nil, fmt.Errorf("failed to save provider: %w", err)
	}

	RegisterProvider(p)
	dto := api.NewLLMProviderDTO(*p)
	return &dto, nil
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
