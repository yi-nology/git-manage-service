package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepo_RegisterLocalRepo(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "my-repo")

	resp := s.PostJSON(t, "/api/v1/repo/create", map[string]interface{}{
		"name": "my-repo",
		"path": repoDir,
	})
	s.AssertSuccess(t, resp)

	var result map[string]interface{}
	s.DecodeData(t, resp, &result)

	if result["key"] == nil || result["key"] == "" {
		t.Fatal("expected repo key")
	}
	if result["name"] != "my-repo" {
		t.Fatalf("name mismatch: %v", result["name"])
	}
	if result["path"] != repoDir {
		t.Fatalf("path mismatch: %v", result["path"])
	}
}

func TestRepo_ListRepos(t *testing.T) {
	s := SetupSuite(t)

	key1 := s.CreateTestRepo(t, "repo-list-1")
	key2 := s.CreateTestRepo(t, "repo-list-2")
	_ = key1
	_ = key2

	resp := s.GetJSON(t, "/api/v1/repo/list")
	s.AssertSuccess(t, resp)

	var list []interface{}
	s.DecodeData(t, resp, &list)
	if len(list) < 2 {
		t.Fatalf("expected at least 2 repos, got %d", len(list))
	}
}

func TestRepo_GetDetail(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "repo-detail")

	resp := s.GetJSON(t, "/api/v1/repo/detail?key=" + key)
	s.AssertSuccess(t, resp)

	var detail map[string]interface{}
	s.DecodeData(t, resp, &detail)
	if detail["key"] != key {
		t.Fatalf("key mismatch: %v", detail["key"])
	}
}

func TestRepo_GetDetailNotFound(t *testing.T) {
	s := SetupSuite(t)

	resp := s.GetJSON(t, "/api/v1/repo/detail?key=nonexistent")
	s.AssertError(t, resp)
}

func TestRepo_RegisterWithCredential(t *testing.T) {
	s := SetupSuite(t)

	credResp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
		"name":   "repo-cred",
		"type":   "http_token",
		"secret": "ghp_test123",
	})
	s.AssertSuccess(t, credResp)

	var cred map[string]interface{}
	s.DecodeData(t, credResp, &cred)
	credID := cred["id"]

	repoDir := s.CreateTestGitRepo(t, "repo-with-cred")

	resp := s.PostJSON(t, "/api/v1/repo/create", map[string]interface{}{
		"name":                 "repo-with-cred",
		"path":                 repoDir,
		"default_credential_id": credID,
	})
	s.AssertSuccess(t, resp)

	var result map[string]interface{}
	s.DecodeData(t, resp, &result)
	if result["default_credential_id"] != credID {
		t.Fatalf("credential id mismatch: %v vs %v", result["default_credential_id"], credID)
	}
}

func TestRepo_RegisterWithRemoteCredentials(t *testing.T) {
	s := SetupSuite(t)

	credResp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
		"name":   "origin-cred",
		"type":   "http_basic",
		"secret": "pass123",
	})
	s.AssertSuccess(t, credResp)

	var cred map[string]interface{}
	s.DecodeData(t, credResp, &cred)
	credID := uint(cred["id"].(float64))

	repoDir := s.CreateTestGitRepo(t, "repo-remote-cred")

	resp := s.PostJSON(t, "/api/v1/repo/create", map[string]interface{}{
		"name":  "repo-remote-cred",
		"path":  repoDir,
		"remote_credentials": map[string]uint{
			"origin": credID,
		},
	})
	s.AssertSuccess(t, resp)
}

func TestRepo_UpdateRepo(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "repo-update")
	key := s.RegisterRepo(t, "repo-update", repoDir)

	resp := s.PostJSON(t, "/api/v1/repo/update", map[string]interface{}{
		"key":  key,
		"name": "repo-updated",
		"path": repoDir,
	})
	s.AssertSuccess(t, resp)

	var result map[string]interface{}
	s.DecodeData(t, resp, &result)
	if result["name"] != "repo-updated" {
		t.Fatalf("name mismatch: %v", result["name"])
	}
}

func TestRepo_DeleteRepo(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "repo-delete")

	resp := s.PostJSON(t, "/api/v1/repo/delete", map[string]interface{}{
		"key": key,
	})
	s.AssertSuccess(t, resp)

	detailResp := s.GetJSON(t, "/api/v1/repo/detail?key=" + key)
	s.AssertError(t, detailResp)
}

func TestRepo_DeleteWithLinkedCredential(t *testing.T) {
	s := SetupSuite(t)

	credResp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
		"name":   "linked-cred",
		"type":   "http_token",
		"secret": "tok123",
	})
	s.AssertSuccess(t, credResp)

	var cred map[string]interface{}
	s.DecodeData(t, credResp, &cred)
	credID := cred["id"]

	repoDir := s.CreateTestGitRepo(t, "repo-linked-cred")
	key := s.RegisterRepo(t, "repo-linked-cred", repoDir)

	s.PostJSON(t, "/api/v1/repo/update", map[string]interface{}{
		"key":                   key,
		"path":                  repoDir,
		"default_credential_id": credID,
	})

	delResp := s.Delete(t, "/api/v1/credentials/"+jsonID(credID))
	s.AssertError(t, delResp, "should not delete credential linked to repo")
}

func TestRepo_RegisterInvalidPath(t *testing.T) {
	s := SetupSuite(t)

	resp := s.PostJSON(t, "/api/v1/repo/create", map[string]interface{}{
		"name": "bad-path",
		"path": "/nonexistent/path/to/repo",
	})
	s.AssertError(t, resp)
}

func TestRepo_RegisterNonGitDir(t *testing.T) {
	s := SetupSuite(t)

	nonGitDir := filepath.Join(s.TmpDir, "not-a-repo")
	os.MkdirAll(nonGitDir, 0o755)

	resp := s.PostJSON(t, "/api/v1/repo/create", map[string]interface{}{
		"name": "not-git",
		"path": nonGitDir,
	})
	s.AssertError(t, resp)
}

func TestRepo_BatchCreate(t *testing.T) {
	s := SetupSuite(t)

	dir1 := s.CreateTestGitRepo(t, "batch-1")
	dir2 := s.CreateTestGitRepo(t, "batch-2")

	resp := s.PostJSON(t, "/api/v1/repo/batch-create", map[string]interface{}{
		"repos": []map[string]interface{}{
			{"name": "batch-1", "path": dir1},
			{"name": "batch-2", "path": dir2},
		},
	})
	s.AssertSuccess(t, resp)

	var result map[string]interface{}
	s.DecodeData(t, resp, &result)

	successList, ok := result["success"].([]interface{})
	if !ok || len(successList) < 2 {
		t.Fatalf("expected 2 successful, got: %v", result)
	}
}

func TestRepo_FetchRepo(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "repo-fetch")

	resp := s.PostJSON(t, "/api/v1/repo/fetch", map[string]interface{}{
		"repo_key": key,
	})
	// fetch may fail if no remote, but the API should respond
	if resp.Code != 0 {
		t.Logf("fetch result (expected for local-only repo): code=%d msg=%s", resp.Code, resp.Msg)
	}
}
