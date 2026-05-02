package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestLLMProviderDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewLLMProviderDAO()

	p := &po.LLMProvider{
		Name:      "test-provider",
		Type:      "openai",
		BaseURL:   "https://api.openai.com",
		APIKey:    "sk-test-key",
		AIModel:   "gpt-4",
		MaxTokens: 4096,
	}
	if err := dao.Create(p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindByID(p.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "test-provider" {
		t.Errorf("name mismatch: got %s", found.Name)
	}

	found.AIModel = "gpt-4o"
	dao.Save(found)
	updated, _ := dao.FindByID(p.ID)
	if updated.AIModel != "gpt-4o" {
		t.Errorf("model mismatch: got %s", updated.AIModel)
	}

	if err := dao.Delete(p.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestLLMProviderDAO_SetDefault(t *testing.T) {
	SetupTestDB(t)
	dao := NewLLMProviderDAO()

	p1 := &po.LLMProvider{Name: "p1", Type: "openai", IsDefault: false}
	p2 := &po.LLMProvider{Name: "p2", Type: "anthropic", IsDefault: false}
	dao.Create(p1)
	dao.Create(p2)

	if err := dao.SetDefault(p1.ID); err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	def, err := dao.FindDefault()
	if err != nil {
		t.Fatalf("FindDefault failed: %v", err)
	}
	if def.ID != p1.ID {
		t.Errorf("expected p1 as default, got ID %d", def.ID)
	}

	dao.SetDefault(p2.ID)
	def, _ = dao.FindDefault()
	if def.ID != p2.ID {
		t.Errorf("expected p2 as default, got ID %d", def.ID)
	}
}

func TestLLMProviderDAO_ClearAllDefault(t *testing.T) {
	SetupTestDB(t)
	dao := NewLLMProviderDAO()
	p := &po.LLMProvider{Name: "p1", Type: "openai", IsDefault: true}
	dao.Create(p)

	dao.ClearAllDefault()
	_, err := dao.FindDefault()
	if err == nil {
		t.Error("expected error when no default")
	}
}

func TestLLMProviderDAO_ExistsByName(t *testing.T) {
	SetupTestDB(t)
	dao := NewLLMProviderDAO()
	dao.Create(&po.LLMProvider{Name: "unique-name", Type: "openai"})

	exists, _ := dao.ExistsByName("unique-name")
	if !exists {
		t.Error("expected name to exist")
	}
	exists, _ = dao.ExistsByName("no-name")
	if exists {
		t.Error("expected name not to exist")
	}
}

func TestLLMProviderDAO_FindByName(t *testing.T) {
	SetupTestDB(t)
	dao := NewLLMProviderDAO()
	dao.Create(&po.LLMProvider{Name: "my-provider", Type: "ollama"})

	found, err := dao.FindByName("my-provider")
	if err != nil {
		t.Fatalf("FindByName failed: %v", err)
	}
	if found.Type != "ollama" {
		t.Errorf("type mismatch: got %s", found.Type)
	}
}

func TestLLMProviderDAO_FindAll(t *testing.T) {
	SetupTestDB(t)
	dao := NewLLMProviderDAO()
	dao.Create(&po.LLMProvider{Name: "p1", Type: "openai"})
	dao.Create(&po.LLMProvider{Name: "p2", Type: "anthropic"})

	all, err := dao.FindAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 providers, got %d", len(all))
	}
}
