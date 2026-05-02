package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestAuthorIdentityDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuthorIdentityDAO()

	identity := &po.AuthorIdentity{
		CanonicalName:  "John Doe",
		CanonicalEmail: "john@example.com",
	}
	identity.SetAliases([]string{"jdoe", "john.doe"})
	if err := dao.Create(identity); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindByID(identity.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.CanonicalName != "John Doe" {
		t.Errorf("name mismatch: got %s", found.CanonicalName)
	}
	aliases := found.GetAliases()
	if len(aliases) != 2 || aliases[0] != "jdoe" {
		t.Errorf("aliases mismatch: %v", aliases)
	}

	found.CanonicalName = "Jane Doe"
	dao.Update(found)

	if err := dao.Delete(identity.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestAuthorIdentityDAO_SetDefault(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuthorIdentityDAO()
	i1 := &po.AuthorIdentity{CanonicalName: "A", CanonicalEmail: "a@x.com", IsDefault: false}
	i2 := &po.AuthorIdentity{CanonicalName: "B", CanonicalEmail: "b@x.com", IsDefault: false}
	dao.Create(i1)
	dao.Create(i2)

	if err := dao.SetDefault(i1.ID); err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	def, err := dao.GetDefault()
	if err != nil {
		t.Fatalf("GetDefault failed: %v", err)
	}
	if def.ID != i1.ID {
		t.Errorf("expected i1 as default, got ID %d", def.ID)
	}

	dao.SetDefault(i2.ID)
	def, _ = dao.GetDefault()
	if def.ID != i2.ID {
		t.Errorf("expected i2 as default after switch, got ID %d", def.ID)
	}
}

func TestAuthorIdentityDAO_ClearAllDefaults(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuthorIdentityDAO()
	dao.Create(&po.AuthorIdentity{CanonicalName: "A", CanonicalEmail: "a@x.com", IsDefault: true})

	dao.ClearAllDefaults()
	_, err := dao.GetDefault()
	if err == nil {
		t.Error("expected error when no default set")
	}
}

func TestAuthorIdentityDAO_ListAll(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuthorIdentityDAO()
	dao.Create(&po.AuthorIdentity{CanonicalName: "A", CanonicalEmail: "a@x.com", IsDefault: true})
	dao.Create(&po.AuthorIdentity{CanonicalName: "B", CanonicalEmail: "b@x.com", IsDefault: false})

	all, err := dao.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 identities, got %d", len(all))
	}
	if !all[0].IsDefault {
		t.Error("expected default identity first")
	}
}
