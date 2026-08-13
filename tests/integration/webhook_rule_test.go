package integration

import (
	"testing"
)

// TestWebhookRule_CRUD exercises the hand-written webhook rule CRUD endpoints
// (GET/POST/PUT/DELETE /api/v1/webhook/rules) added to replace the mock UI.
func TestWebhookRule_CRUD(t *testing.T) {
	s := SetupSuite(t)

	// Create
	resp := s.PostJSON(t, "/api/v1/webhook/rules", map[string]interface{}{
		"name":               "test-rule",
		"provider_config_id": 1,
		"event_type_pattern": "push,tag",
		"repo_pattern":       "foo/*",
		"action":             "sync",
		"action_config":      map[string]interface{}{"sync_task_key": "tk-1"},
		"enabled":            true,
	})
	s.AssertSuccess(t, resp)

	var created map[string]interface{}
	s.DecodeData(t, resp, &created)
	id := created["id"]
	if id == nil {
		t.Fatal("expected id in created rule")
	}
	if created["event_type_pattern"] != "push,tag" {
		t.Fatalf("event_type_pattern mismatch: %v", created["event_type_pattern"])
	}
	if created["action_config"] == nil {
		t.Fatal("expected action_config to be persisted")
	}

	// List
	listResp := s.GetJSON(t, "/api/v1/webhook/rules")
	s.AssertSuccess(t, listResp)
	var listResult map[string]interface{}
	s.DecodeData(t, listResp, &listResult)
	items, _ := listResult["items"].([]interface{})
	if len(items) == 0 {
		t.Fatal("expected at least one rule in list")
	}

	// Update
	updateResp := s.PutJSON(t, "/api/v1/webhook/rules/"+jsonID(id), map[string]interface{}{
		"name":               "test-rule-updated",
		"provider_config_id": 1,
		"event_type_pattern": "push",
		"repo_pattern":       "",
		"action":             "code_review",
		"action_config":      map[string]interface{}{},
		"enabled":            false,
	})
	s.AssertSuccess(t, updateResp)
	var updated map[string]interface{}
	s.DecodeData(t, updateResp, &updated)
	if updated["name"] != "test-rule-updated" {
		t.Fatalf("name not updated: %v", updated["name"])
	}
	if updated["enabled"] != false {
		t.Fatalf("enabled not updated: %v", updated["enabled"])
	}

	// Delete
	delResp := s.Delete(t, "/api/v1/webhook/rules/"+jsonID(id))
	s.AssertSuccess(t, delResp)

	// List should reflect deletion
	listResp2 := s.GetJSON(t, "/api/v1/webhook/rules")
	s.AssertSuccess(t, listResp2)
	var listResult2 map[string]interface{}
	s.DecodeData(t, listResp2, &listResult2)
	items2, _ := listResult2["items"].([]interface{})
	for _, it := range items2 {
		if m, ok := it.(map[string]interface{}); ok && jsonID(m["id"]) == jsonID(id) {
			t.Fatal("deleted rule still present in list")
		}
	}
}
