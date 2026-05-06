package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestRunMigrationsRecordsAndSkipsApplied(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}

	runs := 0
	steps := []migrationStep{
		{
			Version: "test_migration",
			Name:    "test migration",
			Run: func(tx *gorm.DB) error {
				runs++
				return nil
			},
		},
	}

	if err := runMigrations(gdb, steps); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if err := runMigrations(gdb, steps); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	if runs != 1 {
		t.Fatalf("expected migration to run once, got %d", runs)
	}

	var count int64
	if err := gdb.Model(&po.SchemaMigration{}).Where("version = ?", "test_migration").Count(&count).Error; err != nil {
		t.Fatalf("failed to count migration records: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one migration record, got %d", count)
	}
}

func TestRunMigrationsRejectsNilDB(t *testing.T) {
	err := runMigrations(nil, nil)
	if err == nil {
		t.Fatal("expected nil DB error")
	}
}

func TestRunMigrationsRollsBackFailedStep(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}

	steps := []migrationStep{
		{
			Version: "failed_migration",
			Name:    "failed migration",
			Run: func(tx *gorm.DB) error {
				if err := tx.Create(&po.SchemaMigration{
					Version:   "side_effect",
					Name:      "side effect",
					AppliedAt: time.Now(),
				}).Error; err != nil {
					return err
				}
				return fmt.Errorf("boom")
			},
		},
	}

	if err := runMigrations(gdb, steps); err == nil {
		t.Fatal("expected migration failure")
	}

	var count int64
	if err := gdb.Model(&po.SchemaMigration{}).
		Where("version IN ?", []string{"failed_migration", "side_effect"}).
		Count(&count).Error; err != nil {
		t.Fatalf("failed to count migration records: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected failed migration transaction to roll back, got %d records", count)
	}
}
