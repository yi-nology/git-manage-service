package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestRepoProviderBindingDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	repoDAO := NewRepoDAO()
	repoDAO.Create(&po.Repo{Key: "bind-repo", Name: "Bind Repo", Path: "/tmp/bind"})
	providerDAO := NewProviderConfigDAO()
	providerDAO.Create(&po.ProviderConfig{Name: "bind-gl", Platform: "gitlab", BaseURL: "https://gitlab.com"})

	repo, _ := repoDAO.FindByKey("bind-repo")
	providers, _ := providerDAO.FindAll()
	provider := providers[0]

	dao := NewRepoProviderBindingDAO()

	binding := &po.RepoProviderBinding{
		RepoID:           repo.ID,
		ProviderConfigID: provider.ID,
		PlatformOwner:    "myorg",
		PlatformRepo:     "myrepo",
		RemoteName:       "origin",
		IsPrimary:        true,
	}
	if err := dao.Create(binding); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindByID(binding.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.PlatformOwner != "myorg" {
		t.Errorf("owner mismatch: got %s", found.PlatformOwner)
	}

	byRepo, err := dao.FindByRepoID(repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(byRepo) != 1 {
		t.Errorf("expected 1 binding, got %d", len(byRepo))
	}

	primary, err := dao.FindPrimaryByRepoID(repo.ID)
	if err != nil {
		t.Fatalf("FindPrimaryByRepoID failed: %v", err)
	}
	if primary.PlatformRepo != "myrepo" {
		t.Errorf("expected myrepo, got %s", primary.PlatformRepo)
	}

	exists, _ := dao.ExistsByRepoAndProvider(repo.ID, provider.ID)
	if !exists {
		t.Error("expected binding to exist")
	}

	existsPlatform, _ := dao.ExistsByPlatformRepo(provider.ID, "myorg", "myrepo")
	if !existsPlatform {
		t.Error("expected platform repo to exist")
	}
}

func TestRepoProviderBindingDAO_SoftDelete(t *testing.T) {
	SetupTestDB(t)
	repoDAO := NewRepoDAO()
	repoDAO.Create(&po.Repo{Key: "soft-repo", Name: "Soft Repo", Path: "/tmp/soft"})
	repo, _ := repoDAO.FindByKey("soft-repo")
	providerDAO := NewProviderConfigDAO()
	providerDAO.Create(&po.ProviderConfig{Name: "p1", Platform: "gitlab"})
	provider, _ := providerDAO.FindAll()

	dao := NewRepoProviderBindingDAO()
	dao.Create(&po.RepoProviderBinding{
		RepoID:           repo.ID,
		ProviderConfigID: provider[0].ID,
		PlatformOwner:    "org",
		PlatformRepo:     "repo",
		IsPrimary:        true,
	})

	bindings, _ := dao.FindByRepoID(repo.ID)
	if err := dao.SoftDelete(bindings[0].ID); err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}

	active, _ := dao.FindByRepoID(repo.ID)
	if len(active) != 0 {
		t.Errorf("expected 0 active after soft delete, got %d", len(active))
	}
}

func TestRepoProviderBindingDAO_ClearPrimaryByRepoID(t *testing.T) {
	SetupTestDB(t)
	repoDAO := NewRepoDAO()
	repoDAO.Create(&po.Repo{Key: "cp-repo", Name: "CP Repo", Path: "/tmp/cp"})
	repo, _ := repoDAO.FindByKey("cp-repo")
	providerDAO := NewProviderConfigDAO()
	providerDAO.Create(&po.ProviderConfig{Name: "p1", Platform: "gitlab"})
	provider, _ := providerDAO.FindAll()

	dao := NewRepoProviderBindingDAO()
	dao.Create(&po.RepoProviderBinding{
		RepoID:           repo.ID,
		ProviderConfigID: provider[0].ID,
		PlatformOwner:    "org",
		PlatformRepo:     "repo",
		IsPrimary:        true,
	})

	dao.ClearPrimaryByRepoID(repo.ID)
	bindings, _ := dao.FindByRepoIDWithProvider(repo.ID)
	for _, b := range bindings {
		if b.IsPrimary {
			t.Error("expected IsPrimary=false after clear")
		}
	}
}
