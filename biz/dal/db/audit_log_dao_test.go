package db

import (
	"testing"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestAuditLogDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuditLogDAO()

	log := &po.AuditLog{
		Action:    "repo.create",
		Target:    "test-repo",
		Operator:  "admin",
		IPAddress: "127.0.0.1",
	}
	if err := dao.Create(log); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindByID(log.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Action != "repo.create" {
		t.Errorf("action mismatch: got %s", found.Action)
	}
}

func TestAuditLogDAO_FindLatest(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuditLogDAO()
	dao.Create(&po.AuditLog{Action: "a1", Operator: "u1"})
	dao.Create(&po.AuditLog{Action: "a2", Operator: "u2"})
	dao.Create(&po.AuditLog{Action: "a3", Operator: "u3"})

	logs, err := dao.FindLatest(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}
}

func TestAuditLogDAO_Count(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuditLogDAO()
	dao.Create(&po.AuditLog{Action: "a1"})
	dao.Create(&po.AuditLog{Action: "a2"})

	count, err := dao.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestAuditLogDAO_FindPage(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuditLogDAO()
	for i := 0; i < 10; i++ {
		dao.Create(&po.AuditLog{Action: "action"})
	}

	logs, err := dao.FindPage(1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 5 {
		t.Errorf("expected 5 logs, got %d", len(logs))
	}
}

func TestAuditLogDAO_FindPageWithFilters(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuditLogDAO()
	dao.Create(&po.AuditLog{Action: "repo.create", Target: "repo1"})
	dao.Create(&po.AuditLog{Action: "repo.delete", Target: "repo2"})
	dao.Create(&po.AuditLog{Action: "repo.create", Target: "repo3"})

	logs, err := dao.FindPageWithFilters(1, 10, "repo.create", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 filtered logs, got %d", len(logs))
	}
}

func TestAuditLogDAO_CountWithFilters(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuditLogDAO()
	dao.Create(&po.AuditLog{Action: "repo.create"})
	dao.Create(&po.AuditLog{Action: "repo.delete"})

	count, err := dao.CountWithFilters("repo.create", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}

func TestAuditLogDAO_FindByDateRange(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuditLogDAO()
	dao.Create(&po.AuditLog{Action: "a1"})

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)
	logs, err := dao.FindByDateRange(start, end)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log in range, got %d", len(logs))
	}
}

func TestAuditLogDAO_DeleteByDateRange(t *testing.T) {
	SetupTestDB(t)
	dao := NewAuditLogDAO()
	dao.Create(&po.AuditLog{Action: "a1"})

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)
	err := dao.DeleteByDateRange(start, end)
	if err != nil {
		t.Fatal(err)
	}
	count, _ := dao.Count()
	if count != 0 {
		t.Errorf("expected 0 after delete, got %d", count)
	}
}
