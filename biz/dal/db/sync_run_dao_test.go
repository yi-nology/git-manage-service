package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestSyncRunDAO_CRUD(t *testing.T) {
	SetupTestDB(t)

	NewRepoDAO().Create(&po.Repo{Key: "sr1", Name: "SR1", Path: "/tmp/sr1"})
	NewRepoDAO().Create(&po.Repo{Key: "sr2", Name: "SR2", Path: "/tmp/sr2"})
	taskDAO := NewSyncTaskDAO()
	taskDAO.Create(&po.SyncTask{Key: "st1", SourceRepoKey: "sr1", TargetRepoKey: "sr2"})

	dao := NewSyncRunDAO()

	run := &po.SyncRun{
		TaskKey:       "st1",
		TriggerSource: "manual",
		Status:        "success",
		Details:       "Sync completed",
	}
	if err := dao.Create(run); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindLatest(1)
	if err != nil {
		t.Fatalf("FindLatest failed: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected 1 run, got %d", len(found))
	}
	if found[0].Status != "success" {
		t.Errorf("status mismatch: got %s", found[0].Status)
	}

	found[0].Status = "failed"
	dao.Save(found[0])

	if err := dao.Delete(run.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestSyncRunDAO_FindByTaskKeys(t *testing.T) {
	SetupTestDB(t)
	NewRepoDAO().Create(&po.Repo{Key: "sr1", Name: "SR1", Path: "/tmp/sr1"})
	NewRepoDAO().Create(&po.Repo{Key: "sr2", Name: "SR2", Path: "/tmp/sr2"})
	taskDAO := NewSyncTaskDAO()
	taskDAO.Create(&po.SyncTask{Key: "tk1", SourceRepoKey: "sr1", TargetRepoKey: "sr2"})
	taskDAO.Create(&po.SyncTask{Key: "tk2", SourceRepoKey: "sr1", TargetRepoKey: "sr2"})

	dao := NewSyncRunDAO()
	dao.Create(&po.SyncRun{TaskKey: "tk1", Status: "success"})
	dao.Create(&po.SyncRun{TaskKey: "tk2", Status: "failed"})
	dao.Create(&po.SyncRun{TaskKey: "tk1", Status: "success"})

	runs, err := dao.FindByTaskKeys([]string{"tk1"}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs for tk1, got %d", len(runs))
	}
}
