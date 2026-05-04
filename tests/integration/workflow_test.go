package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestE2E_FullGitWorkflow(t *testing.T) {
	s := SetupSuite(t)

	t.Run("Step1_CreateSSHKey", func(t *testing.T) {
		private, _ := generateTestRSAKey(t)
		resp := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
			"name":        "e2e-ssh-key",
			"description": "E2E test SSH key",
			"private_key": private,
		})
		s.AssertSuccess(t, resp)
	})

	private, _ := generateTestRSAKey(t)
	sshResp := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
		"name":        "e2e-ssh-key-real",
		"private_key": private,
	})
	var sshKey map[string]interface{}
	s.DecodeData(t, sshResp, &sshKey)
	sshKeyID := uint(sshKey["id"].(float64))

	t.Run("Step2_CreateCredential", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name":       "e2e-ssh-cred",
			"type":       "ssh_key",
			"ssh_key_id": sshKeyID,
		})
		s.AssertSuccess(t, resp)
	})

	credResp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
		"name":   "e2e-http-cred",
		"type":   "http_basic",
		"secret": "e2epassword",
	})
	var cred map[string]interface{}
	s.DecodeData(t, credResp, &cred)
	credID := cred["id"]

	t.Run("Step3_RegisterRepo", func(t *testing.T) {
		repoDir := s.CreateTestGitRepo(t, "e2e-workflow")

		resp := s.PostJSON(t, "/api/v1/repo/create", map[string]interface{}{
			"name":                  "e2e-workflow",
			"path":                  repoDir,
			"default_credential_id": credID,
		})
		s.AssertSuccess(t, resp)
	})

	repoDir := s.CreateTestGitRepo(t, "e2e-workflow-real")
	key := s.RegisterRepo(t, "e2e-workflow-real", repoDir)

	s.PostJSON(t, "/api/v1/repo/update", map[string]interface{}{
		"key":                   key,
		"default_credential_id": credID,
	})

	t.Run("Step4_VerifyCleanStatus", func(t *testing.T) {
		resp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
		s.AssertSuccess(t, resp)

		var status map[string]interface{}
		s.DecodeData(t, resp, &status)
		if status["isClean"] != true {
			t.Fatalf("initial repo should be clean")
		}
	})

	t.Run("Step5_CreateBranch", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/branch/create", map[string]interface{}{
			"repo_key": key,
			"name":     "feature/e2e-feature",
		})
		s.AssertSuccess(t, resp)
	})

	t.Run("Step6_CheckoutBranch", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/branch/checkout", map[string]interface{}{
			"repo_key": key,
			"name":     "feature/e2e-feature",
		})
		s.AssertSuccess(t, resp)
	})

	t.Run("Step7_ModifyFiles", func(t *testing.T) {
		writeFile(t, filepath.Join(repoDir, "feature.txt"), "E2E feature file\n")
		appendFile(t, filepath.Join(repoDir, "README.md"), "\n## E2E Test\nAdded by integration test\n")
		os.MkdirAll(filepath.Join(repoDir, "src"), 0o755)
		writeFile(t, filepath.Join(repoDir, "src", "main.go"), "package main\n\nfunc main() {}\n")
	})

	t.Run("Step8_CheckDirtyStatus", func(t *testing.T) {
		resp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
		s.AssertSuccess(t, resp)

		var status map[string]interface{}
		s.DecodeData(t, resp, &status)
		if status["isClean"] == true {
			t.Fatal("expected dirty after file modifications")
		}

		untracked, _ := status["untracked"].([]interface{})
		unstaged, _ := status["unstaged"].([]interface{})
		totalChanges := len(untracked) + len(unstaged)
		if totalChanges < 2 {
			t.Fatalf("expected multiple changes, got untracked=%d unstaged=%d", len(untracked), len(unstaged))
		}
	})

	t.Run("Step9_StageFiles", func(t *testing.T) {
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
			t.Fatalf("expected staged files, got %d", len(staged))
		}
	})

	t.Run("Step10_ViewDiff", func(t *testing.T) {
		resp := s.GetJSON(t, "/api/v1/workspace/diff?repo_key="+key)
		s.AssertSuccess(t, resp)

		var diff map[string]interface{}
		s.DecodeData(t, resp, &diff)

		files, _ := diff["files"].([]interface{})
		if len(files) == 0 {
			t.Fatal("expected diff output")
		}

		additions, _ := diff["totalAdditions"].(float64)
		if additions < 1 {
			t.Fatalf("expected additions, got %v", additions)
		}
	})

	t.Run("Step11_CommitChanges", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
			"repo_key":     key,
			"message":      "feat: add e2e feature files",
			"author_name":  "E2E Test",
			"author_email": "e2e@test.com",
		})
		s.AssertSuccess(t, resp)

		var result map[string]interface{}
		s.DecodeData(t, resp, &result)
		if result["commitHash"] == nil {
			t.Fatal("expected commit hash")
		}
		if result["pushed"] == true {
			t.Log("commit was also pushed (unexpected for local repo)")
		}
	})

	t.Run("Step12_VerifyCleanAfterCommit", func(t *testing.T) {
		resp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
		s.AssertSuccess(t, resp)

		var status map[string]interface{}
		s.DecodeData(t, resp, &status)
		if status["isClean"] != true {
			t.Fatal("expected clean after commit")
		}
	})

	t.Run("Step13_MoreChangesAndAmend", func(t *testing.T) {
		appendFile(t, filepath.Join(repoDir, "feature.txt"), "Additional content\n")

		s.PostJSON(t, "/api/v1/workspace/stage", map[string]interface{}{
			"repo_key":  key,
			"stage_all": true,
		})

		resp := s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
			"repo_key": key,
			"message":  "feat: additional changes",
		})
		s.AssertSuccess(t, resp)

		var result map[string]interface{}
		s.DecodeData(t, resp, &result)
		commitHash2 := result["commitHash"]
		if commitHash2 == nil {
			t.Fatal("expected second commit hash")
		}
	})

	t.Run("Step14_SwitchBackToMain", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/branch/checkout", map[string]interface{}{
			"repo_key": key,
			"name":     "master",
		})
		s.AssertSuccess(t, resp)

		statusResp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
		s.AssertSuccess(t, statusResp)

		var status map[string]interface{}
		s.DecodeData(t, statusResp, &status)
		if status["branch"] != "master" {
			t.Fatalf("expected master, got: %v", status["branch"])
		}
	})

	t.Run("Step15_DeleteBranch", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/branch/delete", map[string]interface{}{
			"repo_key": key,
			"name":     "feature/e2e-feature",
		})
		s.AssertSuccess(t, resp)
	})

	t.Run("Step16_DeleteRepo", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/repo/delete", map[string]interface{}{
			"key": key,
		})
		s.AssertSuccess(t, resp)

		listResp := s.GetJSON(t, "/api/v1/repo/list")
		s.AssertSuccess(t, listResp)

		var list []interface{}
		s.DecodeData(t, listResp, &list)
		for _, r := range list {
			repo := r.(map[string]interface{})
			if repo["key"] == key {
				t.Fatal("repo should have been deleted")
			}
		}
	})
}

func TestE2E_ModifyAndCommitWorkflow(t *testing.T) {
	s := SetupSuite(t)
	key := s.CreateTestRepo(t, "e2e-modify")
	repoDir := filepath.Join(s.TmpDir, "e2e-modify")

	writeFile(t, filepath.Join(repoDir, "hello.go"), "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n")

	s.PostJSON(t, "/api/v1/workspace/stage", map[string]interface{}{
		"repo_key": key,
		"files":    []string{"hello.go"},
	})

	commitResp := s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
		"repo_key": key,
		"message":  "add hello.go",
	})
	s.AssertSuccess(t, commitResp)

	var commit1 map[string]interface{}
	s.DecodeData(t, commitResp, &commit1)
	hash1 := commit1["commitHash"].(string)

	writeFile(t, filepath.Join(repoDir, "hello.go"), "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello world\")\n}\n")

	s.PostJSON(t, "/api/v1/workspace/stage", map[string]interface{}{
		"repo_key":  key,
		"stage_all": true,
	})

	commitResp2 := s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
		"repo_key": key,
		"message":  "update greeting",
	})
	s.AssertSuccess(t, commitResp2)

	var commit2 map[string]interface{}
	s.DecodeData(t, commitResp2, &commit2)
	hash2 := commit2["commitHash"].(string)

	if hash1 == hash2 {
		t.Fatal("two commits should have different hashes")
	}

	diffResp := s.GetJSON(t, "/api/v1/workspace/diff?repo_key="+key)
	s.AssertSuccess(t, diffResp)

	var diff map[string]interface{}
	s.DecodeData(t, diffResp, &diff)
	files, _ := diff["files"].([]interface{})
	if len(files) != 0 {
		t.Fatal("expected no diff after commit")
	}
}

func TestE2E_MultiBranchWorkflow(t *testing.T) {
	s := SetupSuite(t)
	key := s.CreateTestRepo(t, "e2e-multi-branch")
	repoDir := filepath.Join(s.TmpDir, "e2e-multi-branch")

	s.PostJSON(t, "/api/v1/branch/create", map[string]interface{}{
		"repo_key": key,
		"name":     "develop",
	})

	s.PostJSON(t, "/api/v1/branch/checkout", map[string]interface{}{
		"repo_key": key,
		"name":     "develop",
	})

	writeFile(t, filepath.Join(repoDir, "dev-file.txt"), "develop branch content")

	s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
		"repo_key":  key,
		"message":   "add dev file",
		"stage_all": true,
	})

	s.PostJSON(t, "/api/v1/branch/create", map[string]interface{}{
		"repo_key": key,
		"name":     "feature/from-develop",
		"base_ref": "develop",
	})

	s.PostJSON(t, "/api/v1/branch/checkout", map[string]interface{}{
		"repo_key": key,
		"name":     "feature/from-develop",
	})

	writeFile(t, filepath.Join(repoDir, "feature-file.txt"), "feature work")

	s.PostJSON(t, "/api/v1/workspace/commit", map[string]interface{}{
		"repo_key":  key,
		"message":   "add feature work",
		"stage_all": true,
	})

	listResp := s.GetJSON(t, "/api/v1/branch/list?repo_key="+key+"&type=local")
	s.AssertSuccess(t, listResp)

	var result map[string]interface{}
	s.DecodeData(t, listResp, &result)
	list, _ := result["list"].([]interface{})

	branchNames := map[string]bool{}
	for _, b := range list {
		branch := b.(map[string]interface{})
		branchNames[branch["name"].(string)] = true
	}

	for _, expected := range []string{"develop", "feature/from-develop"} {
		if !branchNames[expected] {
			t.Fatalf("expected branch %s in list", expected)
		}
	}
}

func TestE2E_CredentialAndRepoLifecycle(t *testing.T) {
	s := SetupSuite(t)

	private, _ := generateTestRSAKey(t)
	sshResp := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
		"name":        "lifecycle-ssh",
		"private_key": private,
	})
	s.AssertSuccess(t, sshResp)

	var sshKey map[string]interface{}
	s.DecodeData(t, sshResp, &sshKey)
	sshID := uint(sshKey["id"].(float64))

	credResp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
		"name":       "lifecycle-cred",
		"type":       "ssh_key",
		"ssh_key_id": sshID,
	})
	s.AssertSuccess(t, credResp)

	var cred map[string]interface{}
	s.DecodeData(t, credResp, &cred)
	credID := cred["id"]

	repoDir := s.CreateTestGitRepo(t, "lifecycle-repo")
	key := s.RegisterRepo(t, "lifecycle-repo", repoDir)

	s.PostJSON(t, "/api/v1/repo/update", map[string]interface{}{
		"key":                   key,
		"path":                  repoDir,
		"default_credential_id": credID,
	})

	detailResp := s.GetJSON(t, "/api/v1/repo/detail?key="+key)
	s.AssertSuccess(t, detailResp)

	var detail map[string]interface{}
	s.DecodeData(t, detailResp, &detail)
	if detail["default_credential_id"] != credID {
		t.Fatalf("credential not linked")
	}

	s.PostJSON(t, "/api/v1/repo/delete", map[string]interface{}{"key": key})

	delCredResp := s.Delete(t, "/api/v1/credentials/"+jsonID(credID))
	s.AssertSuccess(t, delCredResp, "credential should be deletable after repo removed")

	delSSHResp := s.Delete(t, fmt.Sprintf("/api/v1/system/db-ssh-keys/%d", sshID))
	s.AssertSuccess(t, delSSHResp, "ssh key should be deletable after credential removed")
}
