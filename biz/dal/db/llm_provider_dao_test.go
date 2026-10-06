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

func TestLLMProviderDAO_UpsertWithDefault(t *testing.T) {
	SetupTestDB(t)
	dao := NewLLMProviderDAO()

	// 首个 provider 未声明默认：零默认时自动提升，保证不变量。
	p1 := &po.LLMProvider{Name: "up1", Type: "openai", IsDefault: false}
	if err := dao.UpsertWithDefault(p1); err != nil {
		t.Fatalf("UpsertWithDefault failed: %v", err)
	}
	if !p1.IsDefault {
		t.Error("sole provider should be promoted to default")
	}

	// 已有默认时保存非默认 provider：默认保持不变。
	p2 := &po.LLMProvider{Name: "up2", Type: "anthropic", IsDefault: false}
	if err := dao.UpsertWithDefault(p2); err != nil {
		t.Fatalf("UpsertWithDefault failed: %v", err)
	}
	if p2.IsDefault {
		t.Error("should not steal default when one exists")
	}
	def, _ := dao.FindDefault()
	if def.ID != p1.ID {
		t.Errorf("default changed unexpectedly: got ID %d", def.ID)
	}

	// 设 p2 为默认：p1 的默认被清除，全库仍有且仅有一个默认。
	p2.IsDefault = true
	if err := dao.UpsertWithDefault(p2); err != nil {
		t.Fatalf("UpsertWithDefault failed: %v", err)
	}
	def, _ = dao.FindDefault()
	if def.ID != p2.ID {
		t.Errorf("expected p2 as default, got ID %d", def.ID)
	}
	p1b, _ := dao.FindByID(p1.ID)
	if p1b.IsDefault {
		t.Error("old default should be cleared")
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
