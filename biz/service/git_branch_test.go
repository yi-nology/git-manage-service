package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBranchCRUD(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "git-test-branch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	s := NewGitService()
	
	// Init Repo
	if _, err := s.RunCommand(tmpDir, "init"); err != nil {
		t.Fatal(err)
	}
	
	// Create a commit so we have a HEAD
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	s.RunCommand(tmpDir, "add", ".")
	s.RunCommand(tmpDir, "commit", "-m", "initial")

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
