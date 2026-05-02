package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestRepoDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewRepoDAO()

	repo := &po.Repo{
		Key:  "test-repo-1",
		Name: "Test Repo",
		Path: "/tmp/test-repo",
	}
	if err := dao.Create(repo); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if repo.ID == 0 {
		t.Error("expected ID to be set after create")
	}

	found, err := dao.FindByKey("test-repo-1")
	if err != nil {
		t.Fatalf("FindByKey failed: %v", err)
	}
	if found.Name != "Test Repo" {
		t.Errorf("name mismatch: got %s", found.Name)
	}

	found.Name = "Updated Repo"
	if err := dao.Save(found); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	updated, _ := dao.FindByKey("test-repo-1")
	if updated.Name != "Updated Repo" {
		t.Errorf("expected Updated Repo, got %s", updated.Name)
	}

	all, err := dao.FindAll()
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 repo, got %d", len(all))
	}

	if err := dao.Delete(repo); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = dao.FindByKey("test-repo-1")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestRepoDAO_FindByPath(t *testing.T) {
	SetupTestDB(t)
	dao := NewRepoDAO()
	dao.Create(&po.Repo{Key: "r1", Name: "R1", Path: "/unique/path/repo"})

	found, err := dao.FindByPath("/unique/path/repo")
	if err != nil {
		t.Fatalf("FindByPath failed: %v", err)
	}
	if found.Key != "r1" {
		t.Errorf("key mismatch: got %s", found.Key)
	}
}

func TestRepoDAO_FindByID(t *testing.T) {
	SetupTestDB(t)
	dao := NewRepoDAO()
	r := &po.Repo{Key: "r2", Name: "R2", Path: "/path2"}
	dao.Create(r)

	found, err := dao.FindByID(r.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Key != "r2" {
		t.Errorf("key mismatch: got %s", found.Key)
	}
}

func TestRepoDAO_FindByKey_NotFound(t *testing.T) {
	SetupTestDB(t)
	dao := NewRepoDAO()
	_, err := dao.FindByKey("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}
