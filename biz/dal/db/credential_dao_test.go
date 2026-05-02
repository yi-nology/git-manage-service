package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestCredentialDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewCredentialDAO()

	cred := &po.Credential{
		Name:        "test-cred",
		Type:        "ssh_key",
		Description: "Test credential",
	}
	if err := dao.Create(cred); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if cred.ID == 0 {
		t.Error("expected ID set after create")
	}

	found, err := dao.FindByID(cred.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "test-cred" {
		t.Errorf("name mismatch: got %s", found.Name)
	}

	found.Description = "Updated"
	dao.Save(found)
	updated, _ := dao.FindByID(cred.ID)
	if updated.Description != "Updated" {
		t.Errorf("description mismatch: got %s", updated.Description)
	}

	if err := dao.Delete(cred.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = dao.FindByID(cred.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestCredentialDAO_ExistsByName(t *testing.T) {
	SetupTestDB(t)
	dao := NewCredentialDAO()
	dao.Create(&po.Credential{Name: "my-cred", Type: "http_basic"})

	exists, _ := dao.ExistsByName("my-cred")
	if !exists {
		t.Error("expected name to exist")
	}
	exists, _ = dao.ExistsByName("nonexistent")
	if exists {
		t.Error("expected name not to exist")
	}
}

func TestCredentialDAO_ExistsByNameExcludeID(t *testing.T) {
	SetupTestDB(t)
	dao := NewCredentialDAO()
	c := &po.Credential{Name: "my-cred", Type: "http_basic"}
	dao.Create(c)

	exists, _ := dao.ExistsByNameExcludeID("my-cred", c.ID)
	if exists {
		t.Error("should not exist when excluding own ID")
	}
	exists, _ = dao.ExistsByNameExcludeID("my-cred", 9999)
	if !exists {
		t.Error("should exist when excluding different ID")
	}
}

func TestCredentialDAO_FindByType(t *testing.T) {
	SetupTestDB(t)
	dao := NewCredentialDAO()
	dao.Create(&po.Credential{Name: "c1", Type: "ssh_key"})
	dao.Create(&po.Credential{Name: "c2", Type: "http_basic"})
	dao.Create(&po.Credential{Name: "c3", Type: "ssh_key"})

	sshCreds, err := dao.FindByType("ssh_key")
	if err != nil {
		t.Fatal(err)
	}
	if len(sshCreds) != 2 {
		t.Errorf("expected 2 ssh_key, got %d", len(sshCreds))
	}
}

func TestCredentialDAO_FindAll(t *testing.T) {
	SetupTestDB(t)
	dao := NewCredentialDAO()
	dao.Create(&po.Credential{Name: "c1", Type: "ssh_key"})
	dao.Create(&po.Credential{Name: "c2", Type: "http_basic"})

	all, err := dao.FindAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2, got %d", len(all))
	}
}

func TestCredentialDAO_UpdateLastUsed(t *testing.T) {
	SetupTestDB(t)
	dao := NewCredentialDAO()
	c := &po.Credential{Name: "c1", Type: "ssh_key"}
	dao.Create(c)

	if err := dao.UpdateLastUsed(c.ID); err != nil {
		t.Fatalf("UpdateLastUsed failed: %v", err)
	}
	found, _ := dao.FindByID(c.ID)
	if found.LastUsedAt == nil {
		t.Error("expected LastUsedAt to be set")
	}
}
