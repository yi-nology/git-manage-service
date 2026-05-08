package db

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type migrationStep struct {
	Version string
	Name    string
	Run     func(tx *gorm.DB) error
}

func RunMigrations() error {
	return runMigrations(DB, []migrationStep{
		{
			Version: "2026050401_repo_provider_bindings",
			Name:    "migrate repo provider bindings",
			Run: func(tx *gorm.DB) error {
				if !tx.Migrator().HasTable(&po.RepoProviderBinding{}) {
					return nil
				}
				return migrateRepoProviderBindings(tx)
			},
		},
		{
			Version: "2026050402_ai_invocations",
			Name:    "create AI invocation audit table",
			Run: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&po.AIInvocation{})
			},
		},
		{
			Version: "2026050801_mirror_tables",
			Name:    "create mirror and mirror_sync_logs tables",
			Run: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&po.Mirror{}, &po.MirrorSyncLog{})
			},
		},
		{
			Version: "2026050802_review_task_platform_fields",
			Name:    "add platform_owner and platform_repo to review_tasks",
			Run: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&po.ReviewTask{})
			},
		},
		{
			Version: "2026050803_review_rule_prompt_fields",
			Name:    "add rule_type and prompt_text to review_rules",
			Run: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&po.ReviewRule{})
			},
		},
		{
			Version: "2026050804_llm_provider_embedding_fields",
			Name:    "add is_embedding and embedding_model to llm_providers",
			Run: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&po.LLMProvider{})
			},
		},
		{
			Version: "2026050805_review_task_process_log",
			Name:    "add process_log column to review_tasks",
			Run: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&po.ReviewTask{})
			},
		},
		{
			Version: "2026050806_review_composite_indexes",
			Name:    "add composite indexes for review tables",
			Run: func(tx *gorm.DB) error {
				indexes := []struct {
					name  string
					table string
					cols  string
				}{
					{"idx_review_tasks_provider_mriid", "review_tasks", "provider_config_id, mri_id"},
					{"idx_review_findings_task_severity", "review_findings", "task_id, severity"},
					{"idx_review_comments_task_type", "review_comments", "task_id, comment_type"},
					{"idx_review_findings_created_at", "review_findings", "created_at"},
				}
				for _, idx := range indexes {
					if !tx.Migrator().HasIndex(idx.table, idx.name) {
						if err := tx.Exec(fmt.Sprintf("CREATE INDEX %s ON %s (%s)", idx.name, idx.table, idx.cols)).Error; err != nil {
							return err
						}
					}
				}
				return nil
			},
		},
		{
			Version: "2026050807_review_task_raw_diff",
			Name:    "add raw_diff column to review_tasks",
			Run: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&po.ReviewTask{})
			},
		},
	})
}

func runMigrations(gdb *gorm.DB, steps []migrationStep) error {
	if gdb == nil {
		return fmt.Errorf("database is not initialized")
	}

	if err := gdb.AutoMigrate(&po.SchemaMigration{}); err != nil {
		return fmt.Errorf("migrate schema_migrations table: %w", err)
	}

	for _, step := range steps {
		if err := gdb.Transaction(func(tx *gorm.DB) error {
			applied, err := isMigrationApplied(tx, step.Version)
			if err != nil {
				return err
			}
			if applied {
				return nil
			}

			if step.Run != nil {
				if err := step.Run(tx); err != nil {
					return fmt.Errorf("run migration %s: %w", step.Version, err)
				}
			}

			if err := recordMigration(tx, step.Version, step.Name); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func isMigrationApplied(gdb *gorm.DB, version string) (bool, error) {
	var count int64
	if err := gdb.Model(&po.SchemaMigration{}).Where("version = ?", version).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return count > 0, nil
}

func recordMigration(gdb *gorm.DB, version, name string) error {
	migration := &po.SchemaMigration{
		Version:   version,
		Name:      name,
		AppliedAt: time.Now(),
	}
	if err := gdb.Create(migration).Error; err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}
	return nil
}

// MigrateCredentials 将现有认证数据迁移到凭证池
// 幂等操作：重复执行不会重复创建
func MigrateCredentials() {
	credDAO := NewCredentialDAO()

	// 1. 为每个 SSH 密钥创建对应的凭证
	migrateSSHKeysToCredentials(credDAO)

	// 2. 从 Repo 的认证配置中提取凭证
	migrateRepoAuthToCredentials(credDAO)

	log.Println("Credential migration completed.")
}

func migrateSSHKeysToCredentials(credDAO *CredentialDAO) {
	sshKeyDAO := NewSSHKeyDAO()
	keys, err := sshKeyDAO.FindAll()
	if err != nil {
		log.Printf("Warning: failed to load SSH keys for migration: %v", err)
		return
	}

	for _, key := range keys {
		credName := fmt.Sprintf("SSH: %s", key.Name)
		exists, _ := credDAO.ExistsByName(credName)
		if exists {
			continue
		}

		// 检查是否已有引用此 ssh_key_id 的凭证
		existing, _ := credDAO.FindBySSHKeyID(key.ID)
		if len(existing) > 0 {
			continue
		}

		cred := &po.Credential{
			Name:        credName,
			Type:        "ssh_key",
			Description: key.Description,
			SSHKeyID:    key.ID,
		}
		if err := credDAO.Create(cred); err != nil {
			log.Printf("Warning: failed to create credential for SSH key %s: %v", key.Name, err)
		}
	}
}

func migrateRepoAuthToCredentials(credDAO *CredentialDAO) {
	repoDAO := NewRepoDAO()
	repos, err := repoDAO.FindAll()
	if err != nil {
		log.Printf("Warning: failed to load repos for migration: %v", err)
		return
	}

	for i := range repos {
		repo := &repos[i]
		changed := false

		// 处理主认证
		if repo.AuthType != "" && repo.AuthType != "none" && repo.DefaultCredentialID == 0 {
			credID := findOrCreateCredentialFromAuth(credDAO, repo.AuthType, repo.AuthKey, repo.AuthSecret, "", 0)
			if credID > 0 {
				repo.DefaultCredentialID = credID
				changed = true
			}
		}

		// 处理远程认证
		if repo.RemoteAuthsJSON != "" && len(repo.RemoteCredentials) == 0 {
			var remoteAuths map[string]domain.AuthInfo
			if err := json.Unmarshal([]byte(repo.RemoteAuthsJSON), &remoteAuths); err == nil {
				remoteCreds := make(map[string]uint)
				for remoteName, authInfo := range remoteAuths {
					if authInfo.Type == "" || authInfo.Type == "none" {
						continue
					}
					credID := findOrCreateCredentialFromAuth(credDAO, authInfo.Type, authInfo.Key, authInfo.Secret, authInfo.Source, authInfo.SSHKeyID)
					if credID > 0 {
						remoteCreds[remoteName] = credID
					}
				}
				if len(remoteCreds) > 0 {
					repo.RemoteCredentials = remoteCreds
					changed = true
				}
			}
		}

		if changed {
			if err := repoDAO.Save(repo); err != nil {
				log.Printf("Warning: failed to update repo %s with credential refs: %v", repo.Key, err)
			}
		}
	}
}

func findOrCreateCredentialFromAuth(credDAO *CredentialDAO, authType, authKey, authSecret, source string, sshKeyID uint) uint {
	switch authType {
	case "ssh":
		if source == "database" && sshKeyID > 0 {
			// 查找已有的凭证
			existing, _ := credDAO.FindBySSHKeyID(sshKeyID)
			if len(existing) > 0 {
				return existing[0].ID
			}
			// 创建新凭证
			cred := &po.Credential{
				Name:     fmt.Sprintf("SSH Key #%d (migrated)", sshKeyID),
				Type:     "ssh_key",
				SSHKeyID: sshKeyID,
			}
			if err := credDAO.Create(cred); err == nil {
				return cred.ID
			}
		} else if authKey != "" {
			// 本地密钥 - 按路径去重
			name := fmt.Sprintf("SSH Local: %s", authKey)
			if existing, err := credDAO.FindByName(name); err == nil {
				return existing.ID
			}
			cred := &po.Credential{
				Name:       name,
				Type:       "ssh_key",
				SSHKeyPath: authKey,
				Secret:     authSecret,
			}
			if err := credDAO.Create(cred); err == nil {
				return cred.ID
			}
		}
	case "http":
		if authKey != "" {
			name := fmt.Sprintf("HTTP: %s (migrated)", authKey)
			if existing, err := credDAO.FindByName(name); err == nil {
				return existing.ID
			}
			cred := &po.Credential{
				Name:     name,
				Type:     "http_basic",
				Username: authKey,
				Secret:   authSecret,
			}
			if err := credDAO.Create(cred); err == nil {
				return cred.ID
			}
		}
	}
	return 0
}

func MigrateRepoProviderBindings() {
	if err := migrateRepoProviderBindings(DB); err != nil {
		log.Printf("Warning: repo-provider binding migration failed: %v", err)
	}
}

func migrateRepoProviderBindings(gdb *gorm.DB) error {
	if gdb == nil {
		return fmt.Errorf("database is not initialized")
	}

	var repos []po.Repo
	err := gdb.Find(&repos).Error
	if err != nil {
		return fmt.Errorf("load repos for binding migration: %w", err)
	}

	for i := range repos {
		repo := &repos[i]
		if repo.ProviderConfigID == 0 || repo.PlatformOwner == "" || repo.PlatformRepo == "" {
			continue
		}

		var count int64
		err := gdb.Model(&po.RepoProviderBinding{}).
			Where("repo_id = ? AND provider_config_id = ? AND status = ?", repo.ID, repo.ProviderConfigID, "active").
			Count(&count).Error
		if err != nil {
			return fmt.Errorf("check binding for repo %s: %w", repo.Key, err)
		}
		exists := count > 0
		if exists {
			continue
		}

		binding := &po.RepoProviderBinding{
			RepoID:           repo.ID,
			ProviderConfigID: repo.ProviderConfigID,
			PlatformOwner:    repo.PlatformOwner,
			PlatformRepo:     repo.PlatformRepo,
			PlatformRepoID:   repo.PlatformRepoID,
			RemoteName:       "origin",
			IsPrimary:        true,
			Status:           "active",
		}

		if err := gdb.Create(binding).Error; err != nil {
			return fmt.Errorf("create binding for repo %s: %w", repo.Key, err)
		} else {
			log.Printf("Migrated binding: repo=%s provider=%d owner=%s repo=%s",
				repo.Key, repo.ProviderConfigID, repo.PlatformOwner, repo.PlatformRepo)
		}
	}

	log.Println("Repo-Provider binding migration completed.")
	return nil
}
