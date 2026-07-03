package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo creates a git repo at dir and commits the given files in a
// single commit using the CLI (no go-git dependency).
func initTestRepo(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, string(out))
		}
	}
	run("init")
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "initial")
}

// commitFile adds and commits a single file in dir via CLI.
func commitFile(t *testing.T, dir, filename, content, message string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", filename)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v: %s", err, string(out))
	}
	cmd = exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", message)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v: %s", err, string(out))
	}
}

func TestBranchCRUD(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "git-test-branch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	s := NewGitService()

	// Init Repo + initial commit
	initTestRepo(t, tmpDir, map[string]string{"test.txt": "hello"})

	// Determine default branch name (master or main)
	branches, err := s.ListBranchesWithInfo(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) == 0 {
		t.Fatal("No branches found after init")
	}
	defaultBranch := branches[0].Name

	// 1. Test Create
	newBranch := "feature-test"
	if err := s.CreateBranch(tmpDir, newBranch, defaultBranch); err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Verify Created
	branches, _ = s.ListBranchesWithInfo(tmpDir)
	found := false
	for _, b := range branches {
		if b.Name == newBranch {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Branch %s not found after creation", newBranch)
	}

	// 2. Test Rename
	renamedBranch := "feature-renamed"
	if err := s.RenameBranch(tmpDir, newBranch, renamedBranch); err != nil {
		t.Fatalf("RenameBranch failed: %v", err)
	}

	branches, _ = s.ListBranchesWithInfo(tmpDir)
	foundOld := false
	foundNew := false
	for _, b := range branches {
		if b.Name == newBranch {
			foundOld = true
		}
		if b.Name == renamedBranch {
			foundNew = true
		}
	}
	if foundOld {
		t.Error("Old branch name still exists")
	}
	if !foundNew {
		t.Error("New branch name not found")
	}

	// 3. Test Description
	desc := "This is a test branch"
	if err := s.SetBranchDescription(tmpDir, renamedBranch, desc); err != nil {
		t.Fatalf("SetBranchDescription failed: %v", err)
	}
	gotDesc, err := s.GetBranchDescription(tmpDir, renamedBranch)
	if err != nil {
		t.Fatalf("GetBranchDescription failed: %v", err)
	}
	if strings.TrimSpace(gotDesc) != desc {
		t.Errorf("Description mismatch. Got '%s', want '%s'", gotDesc, desc)
	}

	// 4. Test Delete
	if err := s.DeleteBranch(tmpDir, renamedBranch, true); err != nil {
		t.Fatalf("DeleteBranch failed: %v", err)
	}

	branches, _ = s.ListBranchesWithInfo(tmpDir)
	for _, b := range branches {
		if b.Name == renamedBranch {
			t.Error("Branch still exists after delete")
		}
	}
}

func TestGetBranchMetrics(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "git-test-metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	s := NewGitService()

	// Init Repo + initial commit
	initTestRepo(t, tmpDir, map[string]string{"file0.txt": "content"})

	// Create 2 more commits
	commitFile(t, tmpDir, "file1.txt", "content", "commit 1")
	commitFile(t, tmpDir, "file2.txt", "content", "commit 2")

	// Determine default branch name
	branches, err := s.ListBranchesWithInfo(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) == 0 {
		t.Fatal("No branches found after init")
	}
	defaultBranch := branches[0].Name

	// Test Metrics
	metrics, err := s.GetBranchMetrics(tmpDir, defaultBranch)
	if err != nil {
		t.Fatalf("GetBranchMetrics failed: %v", err)
	}

	if count, ok := metrics["commit_count"]; !ok || count != 3 {
		t.Errorf("Expected commit_count 3, got %v", metrics["commit_count"])
	}
}
