package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestProviderConfigDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewProviderConfigDAO()

	cfg := &po.ProviderConfig{
		Name:     "test-gitlab",
		Platform: "gitlab",
		BaseURL:  "https://gitlab.com",
	}
	if err := dao.Create(cfg); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindByID(cfg.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "test-gitlab" {
		t.Errorf("name mismatch: got %s", found.Name)
	}

	found.BaseURL = "https://gitlab.local"
	dao.Save(found)

	if err := dao.Delete(cfg.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestProviderConfigDAO_FindByPlatform(t *testing.T) {
	SetupTestDB(t)
	dao := NewProviderConfigDAO()
	dao.Create(&po.ProviderConfig{Name: "gl1", Platform: "gitlab", BaseURL: "https://gitlab.com"})
	dao.Create(&po.ProviderConfig{Name: "gh1", Platform: "github", BaseURL: "https://github.com"})
	dao.Create(&po.ProviderConfig{Name: "gl2", Platform: "gitlab", BaseURL: "https://gitlab.local"})

	gitlabs, err := dao.FindByPlatform("gitlab")
	if err != nil {
		t.Fatal(err)
	}
	if len(gitlabs) != 2 {
		t.Errorf("expected 2 gitlab configs, got %d", len(gitlabs))
	}
}

func TestProviderConfigDAO_ExistsByName(t *testing.T) {
	SetupTestDB(t)
	dao := NewProviderConfigDAO()
	dao.Create(&po.ProviderConfig{Name: "unique-config", Platform: "gitea"})

	exists, _ := dao.ExistsByName("unique-config")
	if !exists {
		t.Error("expected name to exist")
	}
}

func TestProviderConfigDAO_FindAll(t *testing.T) {
	SetupTestDB(t)
	dao := NewProviderConfigDAO()
	dao.Create(&po.ProviderConfig{Name: "c1", Platform: "gitlab"})
	dao.Create(&po.ProviderConfig{Name: "c2", Platform: "github"})

	all, err := dao.FindAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2, got %d", len(all))
	}
}
