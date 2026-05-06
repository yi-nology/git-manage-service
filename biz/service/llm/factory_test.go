package llm

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
)

func TestHasDefaultProviderFallsBackToDB(t *testing.T) {
	db.SetupTestDB(t)
	oldConfig := configs.GlobalConfig
	t.Cleanup(func() { configs.GlobalConfig = oldConfig })
	configs.GlobalConfig = configs.Config{}

	provider := &po.LLMProvider{
		Name:      "db-default",
		Type:      "ollama",
		BaseURL:   "http://localhost:11434",
		AIModel:   "llama3",
		MaxTokens: 1024,
		IsDefault: true,
	}
	if err := db.NewLLMProviderDAO().Create(provider); err != nil {
		t.Fatalf("Create provider: %v", err)
	}

	if !HasDefaultProvider() {
		t.Fatal("HasDefaultProvider() = false, want true")
	}
}

func TestGetDefaultProviderUsesDBNameFallback(t *testing.T) {
	db.SetupTestDB(t)
	oldConfig := configs.GlobalConfig
	t.Cleanup(func() { configs.GlobalConfig = oldConfig })
	configs.GlobalConfig = configs.Config{}
	configs.GlobalConfig.CodeReview.DefaultLLM = "db-named"

	provider := &po.LLMProvider{
		Name:      "db-named",
		Type:      "ollama",
		BaseURL:   "http://localhost:11434",
		AIModel:   "llama3",
		MaxTokens: 1024,
	}
	if err := db.NewLLMProviderDAO().Create(provider); err != nil {
		t.Fatalf("Create provider: %v", err)
	}

	got, err := GetDefaultProvider()
	if err != nil {
		t.Fatalf("GetDefaultProvider(): %v", err)
	}
	if got.Name() != "ollama:llama3" {
		t.Fatalf("provider name = %q, want ollama:llama3", got.Name())
	}
}

func TestGetDefaultProviderBuildsFromConfigWithoutRegistry(t *testing.T) {
	db.SetupTestDB(t)
	oldConfig := configs.GlobalConfig
	oldProviders := providers
	t.Cleanup(func() {
		configs.GlobalConfig = oldConfig
		providers = oldProviders
	})
	providers = map[string]Provider{}
	configs.GlobalConfig = configs.Config{}
	configs.GlobalConfig.CodeReview.DefaultLLM = "config-provider"
	configs.GlobalConfig.CodeReview.LLMProviders = []configs.LLMProviderConfig{{
		Name:      "config-provider",
		Type:      "ollama",
		BaseURL:   "http://localhost:11434",
		Model:     "llama3",
		MaxTokens: 1024,
	}}

	got, err := GetDefaultProvider()
	if err != nil {
		t.Fatalf("GetDefaultProvider(): %v", err)
	}
	if got.Name() != "ollama:llama3" {
		t.Fatalf("provider name = %q, want ollama:llama3", got.Name())
	}
}
