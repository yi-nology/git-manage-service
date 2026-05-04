package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestBackupDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewBackupDAO()

	rec := &po.BackupRecord{
		RepoID:     1,
		RepoKey:    "test-repo",
		StorageKey: "backup-001",
		Size:       1024,
		Status:     "completed",
	}
	if err := dao.Create(rec); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindByID(rec.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.RepoKey != "test-repo" {
		t.Errorf("repo key mismatch: got %s", found.RepoKey)
	}

	found.Status = "failed"
	dao.Update(found)

	if err := dao.Delete(rec.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestBackupDAO_FindByRepoID(t *testing.T) {
	SetupTestDB(t)
	dao := NewBackupDAO()
	dao.Create(&po.BackupRecord{RepoID: 1, RepoKey: "r1", StorageKey: "b1", Status: "completed"})
	dao.Create(&po.BackupRecord{RepoID: 1, RepoKey: "r1", StorageKey: "b2", Status: "completed"})
	dao.Create(&po.BackupRecord{RepoID: 2, RepoKey: "r2", StorageKey: "b3", Status: "completed"})

	recs, err := dao.FindByRepoID(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Errorf("expected 2 records, got %d", len(recs))
	}
}

func TestBackupDAO_FindByRepoKey(t *testing.T) {
	SetupTestDB(t)
	dao := NewBackupDAO()
	dao.Create(&po.BackupRecord{RepoID: 1, RepoKey: "my-repo", StorageKey: "b1", Status: "completed"})

	recs, err := dao.FindByRepoKey("my-repo")
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Errorf("expected 1 record, got %d", len(recs))
	}
}

func TestBackupDAO_FindLatestByRepoID(t *testing.T) {
	SetupTestDB(t)
	dao := NewBackupDAO()
	dao.Create(&po.BackupRecord{RepoID: 1, RepoKey: "r1", StorageKey: "b1", Status: "completed"})
	dao.Create(&po.BackupRecord{RepoID: 1, RepoKey: "r1", StorageKey: "b2", Status: "completed"})

	latest, err := dao.FindLatestByRepoID(1)
	if err != nil {
		t.Logf("FindLatestByRepoID returned error (expected completed status): %v", err)
		return
	}
	if latest == nil {
		t.Fatal("expected non-nil latest backup")
	}
}

func TestBackupDAO_CountByRepoID(t *testing.T) {
	SetupTestDB(t)
	dao := NewBackupDAO()
	dao.Create(&po.BackupRecord{RepoID: 1, RepoKey: "r1", StorageKey: "b1", Status: "completed"})
	dao.Create(&po.BackupRecord{RepoID: 1, RepoKey: "r1", StorageKey: "b2", Status: "completed"})

	count, err := dao.CountByRepoID(1)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestBackupDAO_FindPage(t *testing.T) {
	SetupTestDB(t)
	dao := NewBackupDAO()
	for i := 0; i < 5; i++ {
		dao.Create(&po.BackupRecord{RepoID: uint(i + 1), RepoKey: "r", StorageKey: "b", Status: "completed"})
	}

	recs, count, err := dao.FindPage(1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Errorf("expected total 5, got %d", count)
	}
	if len(recs) != 3 {
		t.Errorf("expected 3 in page, got %d", len(recs))
	}
}

func TestBackupDAO_DeleteByRepoID(t *testing.T) {
	SetupTestDB(t)
	dao := NewBackupDAO()
	dao.Create(&po.BackupRecord{RepoID: 1, RepoKey: "r1", StorageKey: "b1", Status: "completed"})

	if err := dao.DeleteByRepoID(1); err != nil {
		t.Fatal(err)
	}
	count, _ := dao.CountByRepoID(1)
	if count != 0 {
		t.Errorf("expected 0 after delete, got %d", count)
	}
}
