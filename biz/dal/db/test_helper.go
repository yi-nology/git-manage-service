package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/utils"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) {
	t.Helper()
	utils.InitEncryption()

	var err error
	DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}

	err = DB.AutoMigrate(
		&po.Repo{},
		&po.SyncTask{},
		&po.SyncRun{},
		&po.AuditLog{},
		&po.SystemConfig{},
		&po.CommitStat{},
		&po.NotificationChannel{},
		&po.NotificationEventTemplate{},
		&po.SSHKey{},
		&po.BackupRecord{},
		&po.Credential{},
		&po.LintRule{},
		&po.CommitAnalysis{},
		&po.CommitPattern{},
		&po.SyncRecommendation{},
		&po.ProviderConfig{},
		&po.ChangeRequest{},
		&po.WebhookEvent{},
		&po.WebhookRule{},
		&po.RepoProviderBinding{},
		&po.ReviewTask{},
		&po.ReviewFinding{},
		&po.ReviewComment{},
		&po.MergeCheckResult{},
		&po.ReviewRepoConfig{},
		&po.LLMProvider{},
		&po.BranchRuleSet{},
		&po.BranchRuleOverride{},
		&po.ReviewRule{},
		&po.MaintenanceRecord{},
		&po.AuthorIdentity{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, _ := DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})
}
