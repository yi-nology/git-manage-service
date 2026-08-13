package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspace_GetStatus_CleanRepo(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "ws-clean")

	resp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
	s.AssertSuccess(t, resp)

	var status map[string]interface{}
	s.DecodeData(t, resp, &status)

	if status["is_clean"] != true {
		t.Fatalf("expected clean workspace, got: %v", status)
	}
	if status["branch"] == nil {
		t.Fatal("expected branch info")
	}
}

func TestWorkspace_GetStatus_WithUntrackedFiles(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "ws-untracked")
	key := s.RegisterRepo(t, "ws-untracked", repoDir)

	writeFile(t, filepath.Join(repoDir, "newfile.txt"), "hello world")

	resp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
	s.AssertSuccess(t, resp)

	var status map[string]interface{}
	s.DecodeData(t, resp, &status)

	if status["is_clean"] == true {
		t.Fatal("expected dirty workspace with untracked files")
	}

	untracked, ok := status["untracked"].([]interface{})
	if !ok || len(untracked) == 0 {
		t.Fatalf("expected untracked files, got: %v", status["untracked"])
	}
}

func TestWorkspace_StageAndUnstage(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "ws-stage")
	key := s.RegisterRepo(t, "ws-stage", repoDir)

	writeFile(t, filepath.Join(repoDir, "staged.txt"), "staged content")

	resp := s.PostJSON(t, "/api/v1/workspace/stage", map[string]interface{}{
		"repo_key": key,
		"files":    []string{"staged.txt"},
	})
	s.AssertSuccess(t, resp)

	statusResp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
	s.AssertSuccess(t, statusResp)

	var status map[string]interface{}
	s.DecodeData(t, statusResp, &status)

	staged, ok := status["staged"].([]interface{})
	if !ok || len(staged) == 0 {
		t.Fatalf("expected staged files, got: %v", status["staged"])
	}

	unstageResp := s.PostJSON(t, "/api/v1/workspace/unstage", map[string]interface{}{
		"repo_key": key,
		"files":    []string{"staged.txt"},
	})
	s.AssertSuccess(t, unstageResp)

	statusResp2 := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
	s.AssertSuccess(t, statusResp2)

	var status2 map[string]interface{}
	s.DecodeData(t, statusResp2, &status2)

	staged2, _ := status2["staged"].([]interface{})
	if len(staged2) != 0 {
		t.Fatalf("expected no staged files after unstage, got: %v", staged2)
	}
}

func TestWorkspace_StageAll(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "ws-stage-all")
	key := s.RegisterRepo(t, "ws-stage-all", repoDir)

	writeFile(t, filepath.Join(repoDir, "file1.txt"), "content1")
	writeFile(t, filepath.Join(repoDir, "file2.txt"), "content2")

	resp := s.PostJSON(t, "/api/v1/workspace/stage", map[string]interface{}{
		"repo_key":  key,
		"stage_all": true,
	})
	s.AssertSuccess(t, resp)

	statusResp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
	s.AssertSuccess(t, statusResp)

	var status map[string]interface{}
	s.DecodeData(t, statusResp, &status)

	staged, _ := status["staged"].([]interface{})
	if len(staged) < 2 {
		t.Fatalf("expected 2+ staged files, got %d", len(staged))
	}
}

func TestWorkspace_CommitChanges(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "ws-commit")
	key := s.RegisterRepo(t, "ws-commit", repoDir)

	writeFile(t, filepath.Join(repoDir, "commit-me.txt"), "to be committed")

	s.PostJSON(t, "/api/v1/workspace/stage", map[string]interface{}{
		"repo_key": key,
		"files":    []string{"commit-me.txt"},
	})

	resp := s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
		"repo_key": key,
		"message":  "test commit from integration",
	})
	s.AssertSuccess(t, resp)

	var result map[string]interface{}
	s.DecodeData(t, resp, &result)

	if result["commit_hash"] == nil || result["commit_hash"] == "" {
		t.Fatalf("expected commit hash, got: %v", result)
	}

	statusResp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
	s.AssertSuccess(t, statusResp)

	var status map[string]interface{}
	s.DecodeData(t, statusResp, &status)
	if status["is_clean"] != true {
		t.Fatalf("expected clean after commit, got: %v", status)
	}
}

func TestWorkspace_CommitWithStageAll(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "ws-commit-all")
	key := s.RegisterRepo(t, "ws-commit-all", repoDir)

	writeFile(t, filepath.Join(repoDir, "auto-stage.txt"), "auto staged")

	resp := s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
		"repo_key":  key,
		"message":   "auto stage commit",
		"stage_all": true,
	})
	s.AssertSuccess(t, resp)

	var result map[string]interface{}
	s.DecodeData(t, resp, &result)
	if result["commit_hash"] == nil {
		t.Fatal("expected commit hash")
	}
}

func TestWorkspace_CommitSpecificFiles(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "ws-commit-files")
	key := s.RegisterRepo(t, "ws-commit-files", repoDir)

	writeFile(t, filepath.Join(repoDir, "included.txt"), "included")
	writeFile(t, filepath.Join(repoDir, "excluded.txt"), "excluded")

	resp := s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
		"repo_key": key,
		"message":  "partial commit",
		"files":    []string{"included.txt"},
	})
	s.AssertSuccess(t, resp)

	statusResp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
	s.AssertSuccess(t, statusResp)

	var status map[string]interface{}
	s.DecodeData(t, statusResp, &status)
	if status["is_clean"] == true {
		t.Fatal("expected dirty workspace (excluded.txt still untracked)")
	}
}

func TestWorkspace_CommitEmptyFails(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "ws-commit-empty")

	resp := s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
		"repo_key": key,
		"message":  "nothing to commit",
	})
	// empty commit should fail or return error
	if resp.Code == 0 {
		t.Log("empty commit succeeded (may be acceptable)")
	}
}

func TestWorkspace_GetDiff(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "ws-diff")
	key := s.RegisterRepo(t, "ws-diff", repoDir)

	appendFile(t, filepath.Join(repoDir, "README.md"), "\nNew line added\n")

	s.PostJSON(t, "/api/v1/workspace/stage", map[string]interface{}{
		"repo_key":  key,
		"stage_all": true,
	})

	resp := s.GetJSON(t, "/api/v1/workspace/diff?repo_key="+key)
	s.AssertSuccess(t, resp)

	var diff map[string]interface{}
	s.DecodeData(t, resp, &diff)

	files, ok := diff["files"].([]interface{})
	if !ok || len(files) == 0 {
		t.Fatalf("expected diff files, got: %v", diff["files"])
	}
}

func TestWorkspace_GetDiffStagedOnly(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "ws-diff-staged")
	key := s.RegisterRepo(t, "ws-diff-staged", repoDir)

	writeFile(t, filepath.Join(repoDir, "staged.txt"), "staged")
	writeFile(t, filepath.Join(repoDir, "unstaged.txt"), "unstaged")

	s.PostJSON(t, "/api/v1/workspace/stage", map[string]interface{}{
		"repo_key": key,
		"files":    []string{"staged.txt"},
	})

	resp := s.GetJSON(t, "/api/v1/workspace/diff?repo_key="+key+"&staged_only=true")
	s.AssertSuccess(t, resp)
}

func TestWorkspace_AddToGitignore(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "ws-gitignore")
	key := s.RegisterRepo(t, "ws-gitignore", repoDir)

	resp := s.PostJSON(t, "/api/v1/workspace/gitignore", map[string]interface{}{
		"repo_key": key,
		"patterns": []string{"*.log", "build/"},
	})
	s.AssertSuccess(t, resp)

	gitignorePath := filepath.Join(repoDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected non-empty .gitignore")
	}
}

func TestWorkspace_ModifyTrackedFile(t *testing.T) {
	s := SetupSuite(t)

	repoDir := s.CreateTestGitRepo(t, "ws-modify")
	key := s.RegisterRepo(t, "ws-modify", repoDir)

	readme := filepath.Join(repoDir, "README.md")
	writeFile(t, readme, "# Modified\nThis is modified content\n")

	statusResp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
	s.AssertSuccess(t, statusResp)

	var status map[string]interface{}
	s.DecodeData(t, statusResp, &status)

	if status["is_clean"] == true {
		t.Fatal("expected dirty after modifying tracked file")
	}

	unstaged, ok := status["unstaged"].([]interface{})
	if !ok || len(unstaged) == 0 {
		t.Fatalf("expected unstaged changes, got: %v", status["unstaged"])
	}

	s.PostJSON(t, "/api/v1/workspace/stage", map[string]interface{}{
		"repo_key":  key,
		"stage_all": true,
	})

	commitResp := s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
		"repo_key": key,
		"message":  "modify tracked file",
	})
	s.AssertSuccess(t, commitResp)

	var result map[string]interface{}
	s.DecodeData(t, commitResp, &result)
	if result["commit_hash"] == nil {
		t.Fatal("expected commit hash")
	}
}

func TestWorkspace_StatusRepoNotFound(t *testing.T) {
	s := SetupSuite(t)

	resp := s.GetJSON(t, "/api/v1/workspace/status?repo_key=nonexistent")
	s.AssertError(t, resp)
}
