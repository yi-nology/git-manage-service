package llm

import (
	"fmt"
	"log"
	"sync"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
)

var (
	providers   map[string]Provider
	providersMu sync.RWMutex
)

func init() {
	providers = make(map[string]Provider)
}

func InitProviders() {
	providersMu.Lock()
	defer providersMu.Unlock()

	cfg := configs.GlobalConfig
	for _, pCfg := range cfg.CodeReview.LLMProviders {
		if _, exists := providers[pCfg.Name]; exists {
			continue
		}
		var p Provider
		switch pCfg.Type {
		case "openai_compatible":
			if pCfg.BaseURL == "" || pCfg.APIKey == "" || pCfg.Model == "" {
				log.Printf("[LLM] Skipping provider %q: missing base_url/api_key/model", pCfg.Name)
				continue
			}
			maxTokens := pCfg.MaxTokens
			if maxTokens == 0 {
				maxTokens = 4096
			}
			p = NewOpenAICompatible(pCfg.BaseURL, pCfg.APIKey, pCfg.Model, maxTokens, pCfg.Name)
		case "ollama":
			if pCfg.BaseURL == "" || pCfg.Model == "" {
				log.Printf("[LLM] Skipping provider %q: missing base_url/model", pCfg.Name)
				continue
			}
			maxTokens := pCfg.MaxTokens
			if maxTokens == 0 {
				maxTokens = 4096
			}
			p = NewOllama(pCfg.BaseURL, pCfg.Model, maxTokens)
		case "anthropic":
			if pCfg.BaseURL == "" || pCfg.APIKey == "" || pCfg.Model == "" {
				log.Printf("[LLM] Skipping provider %q: missing base_url/api_key/model", pCfg.Name)
				continue
			}
			maxTokens := pCfg.MaxTokens
			if maxTokens == 0 {
				maxTokens = 4096
			}
			p = NewAnthropic(pCfg.BaseURL, pCfg.APIKey, pCfg.Model, maxTokens)
		case "gemini":
			if pCfg.BaseURL == "" || pCfg.APIKey == "" || pCfg.Model == "" {
				log.Printf("[LLM] Skipping provider %q: missing base_url/api_key/model", pCfg.Name)
				continue
			}
			maxTokens := pCfg.MaxTokens
			if maxTokens == 0 {
				maxTokens = 4096
			}
			p = NewGemini(pCfg.BaseURL, pCfg.APIKey, pCfg.Model, maxTokens)
		default:
			log.Printf("[LLM] Unknown provider type %q for %q", pCfg.Type, pCfg.Name)
			continue
		}
		providers[pCfg.Name] = p
		log.Printf("[LLM] Registered provider: %s (%s)", pCfg.Name, pCfg.Type)
	}
}

func GetProvider(name string) (Provider, error) {
	providersMu.RLock()
	defer providersMu.RUnlock()
	p, ok := providers[name]
	if !ok {
		available := make([]string, 0, len(providers))
		for k := range providers {
			available = append(available, k)
		}
		return nil, fmt.Errorf("LLM provider %q not found (available: %v)", name, available)
	}
	return p, nil
}

func GetDefaultProvider() (Provider, error) {
	cfg := configs.GlobalConfig
	name := cfg.CodeReview.DefaultLLM
	if name != "" {
		if p, err := GetProvider(name); err == nil {
			return p, nil
		}
	}
	return getDefaultDBProvider()
}

func getDefaultDBProvider() (Provider, error) {
	dao := db.NewLLMProviderDAO()
	p, err := dao.FindDefault()
	if err != nil {
		all, err2 := dao.FindAll()
		if err2 != nil || len(all) == 0 {
			return nil, fmt.Errorf("no LLM provider configured")
		}
		p = &all[0]
	}
	return buildProviderFromDB(p)
}

func buildProviderFromDB(p *po.LLMProvider) (Provider, error) {
	switch p.Type {
	case "openai_compatible":
		return NewOpenAICompatible(p.BaseURL, p.APIKey, p.AIModel, p.MaxTokens, p.Name), nil
	case "ollama":
		return NewOllama(p.BaseURL, p.AIModel, p.MaxTokens), nil
	case "anthropic":
		return NewAnthropic(p.BaseURL, p.APIKey, p.AIModel, p.MaxTokens), nil
	case "gemini":
		return NewGemini(p.BaseURL, p.APIKey, p.AIModel, p.MaxTokens), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", p.Type)
	}
}

func HasDefaultProvider() bool {
	cfg := configs.GlobalConfig
	name := cfg.CodeReview.DefaultLLM
	if name == "" {
		return false
	}
	providersMu.RLock()
	_, ok := providers[name]
	providersMu.RUnlock()
	if ok {
		return true
	}
	return hasDBProvider()
}

func hasDBProvider() bool {
	dao := db.NewLLMProviderDAO()
	def, err := dao.FindDefault()
	if err != nil {
		all, err2 := dao.FindAll()
		return err2 == nil && len(all) > 0
	}
	return def != nil
}
