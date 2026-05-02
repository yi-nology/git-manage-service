package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestSSHKeyDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewSSHKeyDAO()

	key := &po.SSHKey{
		Name:        "test-key",
		PrivateKey:  "private-key-content",
		PublicKey:   "ssh-rsa AAAA...",
		Description: "Test SSH key",
	}
	if err := dao.Create(key); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if key.ID == 0 {
		t.Error("expected ID set after create")
	}

	found, err := dao.FindByID(key.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "test-key" {
		t.Errorf("name mismatch: got %s", found.Name)
	}

	found.Description = "Updated"
	dao.Update(found)
	updated, _ := dao.FindByID(key.ID)
	if updated.Description != "Updated" {
		t.Errorf("description mismatch: got %s", updated.Description)
	}

	if err := dao.Delete(key.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = dao.FindByID(key.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestSSHKeyDAO_FindByName(t *testing.T) {
	SetupTestDB(t)
	dao := NewSSHKeyDAO()
	dao.Create(&po.SSHKey{Name: "my-key", PrivateKey: "pk"})

	found, err := dao.FindByName("my-key")
	if err != nil {
		t.Fatalf("FindByName failed: %v", err)
	}
	if found.Name != "my-key" {
		t.Errorf("name mismatch: got %s", found.Name)
	}
}

func TestSSHKeyDAO_ExistsByName(t *testing.T) {
	SetupTestDB(t)
	dao := NewSSHKeyDAO()
	dao.Create(&po.SSHKey{Name: "unique-key", PrivateKey: "pk"})

	exists, _ := dao.ExistsByName("unique-key")
	if !exists {
		t.Error("expected name to exist")
	}
	exists, _ = dao.ExistsByName("no-key")
	if exists {
		t.Error("expected name not to exist")
	}
}

func TestSSHKeyDAO_FindAll(t *testing.T) {
	SetupTestDB(t)
	dao := NewSSHKeyDAO()
	dao.Create(&po.SSHKey{Name: "k1", PrivateKey: "pk1"})
	dao.Create(&po.SSHKey{Name: "k2", PrivateKey: "pk2"})

	all, err := dao.FindAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 keys, got %d", len(all))
	}
}
