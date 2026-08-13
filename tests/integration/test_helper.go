package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/router"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	"github.com/yi-nology/git-manage-service/biz/utils"
	"github.com/yi-nology/git-manage-service/pkg/configs"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type APITestSuite struct {
	BaseURL string
	DB      *gorm.DB
	TmpDir  string
	client  *http.Client
	hertz   *server.Hertz
}

type APIResponse struct {
	Code  int32           `json:"code"`
	Msg   string          `json:"msg"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

func SetupSuite(t *testing.T) *APITestSuite {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	hlog.SetLevel(hlog.LevelWarn)

	configs.GlobalConfig = configs.Config{
		Database: configs.DatabaseConfig{
			Type: "sqlite",
			Path: ":memory:",
		},
		Storage: configs.StorageConfig{
			Type:      "local",
			LocalPath: t.TempDir(),
		},
	}

	utils.InitEncryption()

	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	setupModels := []interface{}{
		&po.Repo{},
		&po.Credential{},
		&po.SSHKey{},
		&po.AuditLog{},
		&po.SystemConfig{},
		&po.CommitStat{},
		&po.NotificationChannel{},
		&po.NotificationEventTemplate{},
		&po.LintRule{},
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
	}
	if err := gormDB.AutoMigrate(setupModels...); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	db.DB = gormDB

	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(1)

	audit.InitAuditService()
	stats.InitStatsService()

	tmpDir := t.TempDir()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()

	h := server.Default(server.WithListener(ln))
	router.GeneratedRegister(h)

	go func() {
		if err := h.Run(); err != nil {
			t.Logf("server stopped: %v", err)
		}
	}()

	if !waitReady(addr, 5*time.Second) {
		t.Fatal("server did not start in time")
	}

	s := &APITestSuite{
		BaseURL: "http://" + addr,
		DB:      gormDB,
		TmpDir:  tmpDir,
		hertz:   h,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
	}

	t.Cleanup(func() {
		s.client.CloseIdleConnections()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.Shutdown(ctx)
		stats.StatsSvc.Wait()
		sqlDB, _ := gormDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	return s
}

func waitReady(addr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func (s *APITestSuite) DoRequest(t *testing.T, method, path string, body interface{}) *APIResponse {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, s.BaseURL+path, bodyReader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	apiResp := &APIResponse{}
	if err := json.Unmarshal(respBytes, apiResp); err != nil {
		t.Fatalf("unmarshal response: %v (body: %s)", err, string(respBytes))
	}

	return apiResp
}

func (s *APITestSuite) GetJSON(t *testing.T, path string) *APIResponse {
	return s.DoRequest(t, http.MethodGet, path, nil)
}

func (s *APITestSuite) PostJSON(t *testing.T, path string, body interface{}) *APIResponse {
	return s.DoRequest(t, http.MethodPost, path, body)
}

func (s *APITestSuite) PutJSON(t *testing.T, path string, body interface{}) *APIResponse {
	return s.DoRequest(t, http.MethodPut, path, body)
}

func (s *APITestSuite) Delete(t *testing.T, path string) *APIResponse {
	return s.DoRequest(t, http.MethodDelete, path, nil)
}

func (s *APITestSuite) AssertSuccess(t *testing.T, resp *APIResponse, msgAndArgs ...interface{}) {
	t.Helper()
	if resp.Code != 0 {
		detail := resp.Msg
		if resp.Error != "" {
			detail = fmt.Sprintf("%s: %s", resp.Msg, resp.Error)
		}
		t.Fatalf("expected code 0, got %d: %s %v", resp.Code, detail, msgAndArgs)
	}
}

func (s *APITestSuite) AssertError(t *testing.T, resp *APIResponse, msgAndArgs ...interface{}) {
	t.Helper()
	if resp.Code == 0 {
		t.Fatalf("expected non-zero code, got 0 %v", msgAndArgs)
	}
}

func (s *APITestSuite) DecodeData(t *testing.T, resp *APIResponse, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(resp.Data, target); err != nil {
		t.Fatalf("decode data: %v (raw: %s)", err, string(resp.Data))
	}
}

func (s *APITestSuite) CreateTestGitRepo(t *testing.T, name string) (repoDir string) {
	t.Helper()

	repoDir = filepath.Join(s.TmpDir, name)
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")

	testFile := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "initial commit")

	return repoDir
}

func (s *APITestSuite) RegisterRepo(t *testing.T, name, repoDir string) string {
	t.Helper()

	resp := s.PostJSON(t, "/api/v1/repo/create", map[string]interface{}{
		"name": name,
		"path": repoDir,
	})
	s.AssertSuccess(t, resp, "register repo")

	var result map[string]interface{}
	s.DecodeData(t, resp, &result)

	key, _ := result["key"].(string)
	if key == "" {
		t.Fatal("repo key is empty")
	}
	return key
}

func (s *APITestSuite) CreateTestRepo(t *testing.T, name string) string {
	t.Helper()
	repoDir := s.CreateTestGitRepo(t, name)
	return s.RegisterRepo(t, name, repoDir)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE=2024-01-01T00:00:00Z",
		"GIT_COMMITTER_DATE=2024-01-01T00:00:00Z",
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func generateTestRSAKey(t *testing.T) (private, public string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return string(privPEM), string(pubPEM)
}

func jsonID(v interface{}) string {
	switch id := v.(type) {
	case float64:
		if id == float64(uint(id)) {
			return fmt.Sprintf("%d", uint(id))
		}
		return fmt.Sprintf("%v", id)
	case string:
		return id
	default:
		return fmt.Sprintf("%v", id)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func appendFile(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatalf("open file %s: %v", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("append file %s: %v", path, err)
	}
}
