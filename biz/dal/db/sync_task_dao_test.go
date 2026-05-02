package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestSyncTaskDAO_CRUD(t *testing.T) {
	SetupTestDB(t)

	repo := &po.Repo{Key: "src-repo", Name: "Source", Path: "/tmp/src"}
	NewRepoDAO().Create(repo)
	targetRepo := &po.Repo{Key: "tgt-repo", Name: "Target", Path: "/tmp/tgt"}
	NewRepoDAO().Create(targetRepo)

	dao := NewSyncTaskDAO()

	task := &po.SyncTask{
		Key:           "sync-1",
		SourceRepoKey: "src-repo",
		TargetRepoKey: "tgt-repo",
		Enabled:       true,
	}
	if err := dao.Create(task); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindByKey("sync-1")
	if err != nil {
		t.Fatalf("FindByKey failed: %v", err)
	}
	if found.SourceRepoKey != "src-repo" {
		t.Errorf("source mismatch: got %s", found.SourceRepoKey)
	}

	found.Enabled = false
	dao.Save(found)

	tasks, err := dao.FindByRepoKey("src-repo")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	if err := dao.Delete(task); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestSyncTaskDAO_FindAllWithRepos(t *testing.T) {
	SetupTestDB(t)
	NewRepoDAO().Create(&po.Repo{Key: "r1", Name: "R1", Path: "/tmp/r1"})
	NewRepoDAO().Create(&po.Repo{Key: "r2", Name: "R2", Path: "/tmp/r2"})
	dao := NewSyncTaskDAO()
	dao.Create(&po.SyncTask{Key: "s1", SourceRepoKey: "r1", TargetRepoKey: "r2"})
	dao.Create(&po.SyncTask{Key: "s2", SourceRepoKey: "r2", TargetRepoKey: "r1"})

	tasks, err := dao.FindAllWithRepos()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestSyncTaskDAO_CountByRepoKey(t *testing.T) {
	SetupTestDB(t)
	NewRepoDAO().Create(&po.Repo{Key: "r1", Name: "R1", Path: "/tmp/r1"})
	NewRepoDAO().Create(&po.Repo{Key: "r2", Name: "R2", Path: "/tmp/r2"})
	dao := NewSyncTaskDAO()
	dao.Create(&po.SyncTask{Key: "s1", SourceRepoKey: "r1", TargetRepoKey: "r2"})
	dao.Create(&po.SyncTask{Key: "s2", SourceRepoKey: "r1", TargetRepoKey: "r2"})

	count, err := dao.CountByRepoKey("r1")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestSyncTaskDAO_FindEnabledWithCron(t *testing.T) {
	SetupTestDB(t)
	NewRepoDAO().Create(&po.Repo{Key: "r1", Name: "R1", Path: "/tmp/r1"})
	NewRepoDAO().Create(&po.Repo{Key: "r2", Name: "R2", Path: "/tmp/r2"})
	dao := NewSyncTaskDAO()
	dao.Create(&po.SyncTask{Key: "s1", SourceRepoKey: "r1", TargetRepoKey: "r2", Enabled: true, Cron: "0 * * * *"})
	dao.Create(&po.SyncTask{Key: "s2", SourceRepoKey: "r1", TargetRepoKey: "r2", Enabled: false, Cron: "0 0 * * *"})
	dao.Create(&po.SyncTask{Key: "s3", SourceRepoKey: "r1", TargetRepoKey: "r2", Enabled: true, Cron: ""})

	tasks, err := dao.FindEnabledWithCron()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 enabled+cron task, got %d", len(tasks))
	}
}
