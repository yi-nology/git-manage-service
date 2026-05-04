package integration

import (
	"testing"
)

func TestBranch_ListBranches(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "branch-list")

	resp := s.GetJSON(t, "/api/v1/branch/list?repo_key="+key+"&type=local")
	s.AssertSuccess(t, resp)

	var result map[string]interface{}
	s.DecodeData(t, resp, &result)

	list, ok := result["list"].([]interface{})
	if !ok {
		t.Fatalf("expected list, got: %v", result)
	}
	if len(list) == 0 {
		t.Fatal("expected at least one branch")
	}

	first := list[0].(map[string]interface{})
	if first["is_current"] != true {
		t.Fatal("expected first branch to be current")
	}
}

func TestBranch_CreateBranch(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "branch-create")

	resp := s.PostJSON(t, "/api/v1/branch/create", map[string]interface{}{
		"repo_key": key,
		"name":     "feature/test-branch",
	})
	s.AssertSuccess(t, resp)

	listResp := s.GetJSON(t, "/api/v1/branch/list?repo_key="+key+"&type=local")
	s.AssertSuccess(t, listResp)

	var result map[string]interface{}
	s.DecodeData(t, listResp, &result)

	list, _ := result["list"].([]interface{})
	found := false
	for _, b := range list {
		branch := b.(map[string]interface{})
		if branch["name"] == "feature/test-branch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected feature/test-branch in branch list")
	}
}

func TestBranch_CheckoutBranch(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "branch-checkout")

	createResp := s.PostJSON(t, "/api/v1/branch/create", map[string]interface{}{
		"repo_key": key,
		"name":     "dev",
	})
	s.AssertSuccess(t, createResp)

	checkoutResp := s.PostJSON(t, "/api/v1/branch/checkout", map[string]interface{}{
		"repo_key": key,
		"name":     "dev",
	})
	s.AssertSuccess(t, checkoutResp)

	statusResp := s.GetJSON(t, "/api/v1/workspace/status?repo_key="+key)
	s.AssertSuccess(t, statusResp)

	var status map[string]interface{}
	s.DecodeData(t, statusResp, &status)
	if status["branch"] != "dev" {
		t.Fatalf("expected branch 'dev', got: %v", status["branch"])
	}
}

func TestBranch_CreateFromBaseRef(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "branch-baseref")

	resp := s.PostJSON(t, "/api/v1/branch/create", map[string]interface{}{
		"repo_key": key,
		"name":     "release/v1",
		"base_ref": "HEAD",
	})
	s.AssertSuccess(t, resp)
}

func TestBranch_DeleteBranch(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "branch-delete")

	s.PostJSON(t, "/api/v1/branch/create", map[string]interface{}{
		"repo_key": key,
		"name":     "to-delete",
	})

	s.PostJSON(t, "/api/v1/branch/checkout", map[string]interface{}{
		"repo_key": key,
		"name":     "master",
	})

	resp := s.PostJSON(t, "/api/v1/branch/delete", map[string]interface{}{
		"repo_key": key,
		"name":     "to-delete",
	})
	s.AssertSuccess(t, resp)

	listResp := s.GetJSON(t, "/api/v1/branch/list?repo_key="+key+"&type=local")
	s.AssertSuccess(t, listResp)

	var result map[string]interface{}
	s.DecodeData(t, listResp, &result)

	list, _ := result["list"].([]interface{})
	for _, b := range list {
		branch := b.(map[string]interface{})
		if branch["name"] == "to-delete" {
			t.Fatal("branch should have been deleted")
		}
	}
}

func TestBranch_ListRemote(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "branch-remote")

	resp := s.GetJSON(t, "/api/v1/branch/list?repo_key="+key+"&type=remote")
	s.AssertSuccess(t, resp)
}

func TestBranch_KeywordSearch(t *testing.T) {
	s := SetupSuite(t)

	key := s.CreateTestRepo(t, "branch-search")

	s.PostJSON(t, "/api/v1/branch/create", map[string]interface{}{
		"repo_key": key,
		"name":     "feature/search-a",
	})
	s.PostJSON(t, "/api/v1/branch/create", map[string]interface{}{
		"repo_key": key,
		"name":     "feature/search-b",
	})
	s.PostJSON(t, "/api/v1/branch/create", map[string]interface{}{
		"repo_key": key,
		"name":     "bugfix/other",
	})

	resp := s.GetJSON(t, "/api/v1/branch/list?repo_key="+key+"&keyword=search&type=local")
	s.AssertSuccess(t, resp)

	var result map[string]interface{}
	s.DecodeData(t, resp, &result)

	list, _ := result["list"].([]interface{})
	if len(list) < 2 {
		t.Fatalf("expected at least 2 branches matching 'search', got %d", len(list))
	}
}
