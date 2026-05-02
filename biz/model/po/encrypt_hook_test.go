package po

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/utils"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupModelTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	utils.InitEncryption()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}

	db.AutoMigrate(&Repo{}, &Credential{}, &SSHKey{}, &LLMProvider{}, &ProviderConfig{})
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})
	return db
}

func TestRepo_EncryptDecrypt(t *testing.T) {
	db := setupModelTestDB(t)

	repo := &Repo{
		Key:        "enc-test",
		Name:       "Enc Test",
		Path:       "/tmp/enc",
		AuthSecret: "super-secret-token",
	}
	db.Create(repo)

	var found Repo
	db.First(&found, repo.ID)

	if found.AuthSecret != "super-secret-token" {
		t.Errorf("AuthSecret round-trip failed: got %q", found.AuthSecret)
	}
}

func TestCredential_EncryptDecrypt(t *testing.T) {
	db := setupModelTestDB(t)

	cred := &Credential{
		Name:   "enc-cred",
		Type:   "http_token",
		Secret: "my-api-key-12345",
	}
	db.Create(cred)

	var found Credential
	db.First(&found, cred.ID)

	if found.Secret != "my-api-key-12345" {
		t.Errorf("Secret round-trip failed: got %q", found.Secret)
	}
}

func TestCredential_EmptySecret(t *testing.T) {
	db := setupModelTestDB(t)

	cred := &Credential{
		Name:   "empty-cred",
		Type:   "ssh_key",
		Secret: "",
	}
	db.Create(cred)

	var found Credential
	db.First(&found, cred.ID)

	if found.Secret != "" {
		t.Errorf("empty Secret should remain empty, got %q", found.Secret)
	}
}

func TestSSHKey_EncryptDecrypt(t *testing.T) {
	db := setupModelTestDB(t)

	key := &SSHKey{
		Name:       "enc-key",
		PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----",
		PublicKey:  "ssh-rsa AAAA...",
		Passphrase: "my-passphrase",
	}
	db.Create(key)

	var found SSHKey
	db.First(&found, key.ID)

	if found.PrivateKey != key.PrivateKey {
		t.Errorf("PrivateKey round-trip failed: got %q", found.PrivateKey)
	}
	if found.Passphrase != "my-passphrase" {
		t.Errorf("Passphrase round-trip failed: got %q", found.Passphrase)
	}
	if found.PublicKey != "ssh-rsa AAAA..." {
		t.Errorf("PublicKey should not be encrypted: got %q", found.PublicKey)
	}
}

func TestLLMProvider_EncryptDecrypt(t *testing.T) {
	db := setupModelTestDB(t)

	p := &LLMProvider{
		Name:    "enc-llm",
		Type:    "openai",
		APIKey:  "sk-proj-abc123def456",
		AIModel: "gpt-4",
	}
	db.Create(p)

	var found LLMProvider
	db.First(&found, p.ID)

	if found.APIKey != "sk-proj-abc123def456" {
		t.Errorf("APIKey round-trip failed: got %q", found.APIKey)
	}
}

func TestProviderConfig_EncryptDecrypt(t *testing.T) {
	db := setupModelTestDB(t)

	cfg := &ProviderConfig{
		Name:           "enc-provider",
		Platform:       "gitlab",
		WebhookSecret:  "wh-secret-12345",
	}
	db.Create(cfg)

	var found ProviderConfig
	db.First(&found, cfg.ID)

	if found.WebhookSecret != "wh-secret-12345" {
		t.Errorf("WebhookSecret round-trip failed: got %q", found.WebhookSecret)
	}
}

func TestRepo_EncryptDecrypt_MultipleFields(t *testing.T) {
	db := setupModelTestDB(t)

	repo := &Repo{
		Key:        "multi-enc",
		Name:       "Multi Enc",
		Path:       "/tmp/multi",
		AuthSecret: "secret1",
	}
	db.Create(repo)

	var raw Repo
	db.Raw("SELECT * FROM repos WHERE id = ?", repo.ID).Scan(&raw)

	if raw.AuthSecret == "secret1" {
		t.Error("AuthSecret should be encrypted in raw DB storage")
	}

	var found Repo
	db.First(&found, repo.ID)
	if found.AuthSecret != "secret1" {
		t.Error("AuthSecret should be decrypted after AfterFind")
	}
}
