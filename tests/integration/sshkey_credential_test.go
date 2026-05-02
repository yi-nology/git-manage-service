package integration

import (
	"testing"
)

func TestSSHKey_CRUD(t *testing.T) {
	s := SetupSuite(t)

	t.Run("CreateSSHKey", func(t *testing.T) {
		private, _ := generateTestRSAKey(t)

		resp := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
			"name":        "test-key-1",
			"description": "Test SSH key for integration",
			"private_key": private,
		})
		s.AssertSuccess(t, resp)

		var result map[string]interface{}
		s.DecodeData(t, resp, &result)

		if result["id"] == nil {
			t.Fatal("expected id in response")
		}
		if result["name"] != "test-key-1" {
			t.Fatalf("name mismatch: %v", result["name"])
		}
		if result["key_type"] == nil {
			t.Fatal("expected key_type")
		}
		if result["public_key"] == nil {
			t.Fatal("expected public_key to be auto-extracted")
		}
	})

	t.Run("ListSSHKeys", func(t *testing.T) {
		resp := s.GetJSON(t, "/api/v1/system/db-ssh-keys")
		s.AssertSuccess(t, resp)

		var list []interface{}
		s.DecodeData(t, resp, &list)
		if len(list) == 0 {
			t.Fatal("expected at least one SSH key")
		}
	})

	t.Run("GetSSHKey", func(t *testing.T) {
		private, _ := generateTestRSAKey(t)

		resp := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
			"name":        "test-key-2",
			"private_key": private,
		})
		s.AssertSuccess(t, resp)

		var created map[string]interface{}
		s.DecodeData(t, resp, &created)
		id := created["id"]

		getResp := s.GetJSON(t, "/api/v1/system/db-ssh-keys/"+jsonID(id))
		s.AssertSuccess(t, getResp)

		var fetched map[string]interface{}
		s.DecodeData(t, getResp, &fetched)
		if fetched["name"] != "test-key-2" {
			t.Fatalf("name mismatch: %v", fetched["name"])
		}
	})

	t.Run("UpdateSSHKey", func(t *testing.T) {
		private, _ := generateTestRSAKey(t)

		resp := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
			"name":        "test-key-3",
			"private_key": private,
		})
		s.AssertSuccess(t, resp)

		var created map[string]interface{}
		s.DecodeData(t, resp, &created)
		id := created["id"]

		updateResp := s.PutJSON(t, "/api/v1/system/db-ssh-keys/"+jsonID(id), map[string]interface{}{
			"description": "Updated description",
		})
		s.AssertSuccess(t, updateResp)

		var updated map[string]interface{}
		s.DecodeData(t, updateResp, &updated)
		if updated["description"] != "Updated description" {
			t.Fatalf("description mismatch: %v", updated["description"])
		}
	})

	t.Run("DeleteSSHKey", func(t *testing.T) {
		private, _ := generateTestRSAKey(t)

		resp := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
			"name":        "test-key-del",
			"private_key": private,
		})
		s.AssertSuccess(t, resp)

		var created map[string]interface{}
		s.DecodeData(t, resp, &created)
		id := created["id"]

		delResp := s.Delete(t, "/api/v1/system/db-ssh-keys/"+jsonID(id))
		s.AssertSuccess(t, delResp)

		getResp := s.GetJSON(t, "/api/v1/system/db-ssh-keys/"+jsonID(id))
		s.AssertError(t, getResp)
	})

	t.Run("CreateDuplicateName", func(t *testing.T) {
		private, _ := generateTestRSAKey(t)

		resp1 := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
			"name":        "dup-key",
			"private_key": private,
		})
		s.AssertSuccess(t, resp1)

		private2, _ := generateTestRSAKey(t)
		resp2 := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
			"name":        "dup-key",
			"private_key": private2,
		})
		s.AssertError(t, resp2)
	})

	t.Run("CreateWithoutPrivateKey", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
			"name": "no-key",
		})
		s.AssertError(t, resp)
	})
}

func TestCredential_CRUD(t *testing.T) {
	s := SetupSuite(t)

	t.Run("CreateHTTPBasicCredential", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name":        "test-http-cred",
			"type":        "http_basic",
			"description": "Test HTTP basic credential",
			"username":    "testuser",
			"secret":      "testpass",
		})
		s.AssertSuccess(t, resp)

		var result map[string]interface{}
		s.DecodeData(t, resp, &result)

		if result["id"] == nil {
			t.Fatal("expected id")
		}
		if result["type"] != "http_basic" {
			t.Fatalf("type mismatch: %v", result["type"])
		}
		if result["has_secret"] != true {
			t.Fatal("expected has_secret=true")
		}
	})

	t.Run("CreateHTTPTokenCredential", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name":   "test-token-cred",
			"type":   "http_token",
			"secret": "ghp_testtoken123",
		})
		s.AssertSuccess(t, resp)
	})

	t.Run("CreateCredentialWithSSHKey", func(t *testing.T) {
		private, _ := generateTestRSAKey(t)

		sshResp := s.PostJSON(t, "/api/v1/system/db-ssh-keys", map[string]interface{}{
			"name":        "cred-ssh-key",
			"private_key": private,
		})
		s.AssertSuccess(t, sshResp)

		var sshKey map[string]interface{}
		s.DecodeData(t, sshResp, &sshKey)
		sshKeyID := uint(sshKey["id"].(float64))

		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name":       "test-ssh-cred",
			"type":       "ssh_key",
			"ssh_key_id": sshKeyID,
		})
		s.AssertSuccess(t, resp)

		var cred map[string]interface{}
		s.DecodeData(t, resp, &cred)
		if cred["ssh_key_name"] == nil {
			t.Fatal("expected ssh_key_name to be populated")
		}
	})

	t.Run("ListCredentials", func(t *testing.T) {
		resp := s.GetJSON(t, "/api/v1/credentials/")
		s.AssertSuccess(t, resp)

		var list []interface{}
		s.DecodeData(t, resp, &list)
		if len(list) == 0 {
			t.Fatal("expected at least one credential")
		}
	})

	t.Run("GetCredential", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name":     "get-test-cred",
			"type":     "http_basic",
			"username": "user1",
			"secret":   "pass1",
		})
		s.AssertSuccess(t, resp)

		var created map[string]interface{}
		s.DecodeData(t, resp, &created)
		id := created["id"]

		getResp := s.GetJSON(t, "/api/v1/credentials/"+jsonID(id))
		s.AssertSuccess(t, getResp)

		var fetched map[string]interface{}
		s.DecodeData(t, getResp, &fetched)
		if fetched["name"] != "get-test-cred" {
			t.Fatalf("name mismatch: %v", fetched["name"])
		}
	})

	t.Run("UpdateCredential", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name":     "update-test-cred",
			"type":     "http_basic",
			"username": "user1",
			"secret":   "pass1",
		})
		s.AssertSuccess(t, resp)

		var created map[string]interface{}
		s.DecodeData(t, resp, &created)
		id := created["id"]

		updateResp := s.PutJSON(t, "/api/v1/credentials/"+jsonID(id), map[string]interface{}{
			"description": "Updated desc",
			"username":    "newuser",
		})
		s.AssertSuccess(t, updateResp)

		var updated map[string]interface{}
		s.DecodeData(t, updateResp, &updated)
		if updated["username"] != "newuser" {
			t.Fatalf("username mismatch: %v", updated["username"])
		}
	})

	t.Run("DeleteCredential", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name":     "del-test-cred",
			"type":     "http_basic",
			"username": "user1",
			"secret":   "pass1",
		})
		s.AssertSuccess(t, resp)

		var created map[string]interface{}
		s.DecodeData(t, resp, &created)
		id := created["id"]

		delResp := s.Delete(t, "/api/v1/credentials/"+jsonID(id))
		s.AssertSuccess(t, delResp)

		getResp := s.GetJSON(t, "/api/v1/credentials/"+jsonID(id))
		s.AssertError(t, getResp)
	})

	t.Run("CreateCredentialInvalidType", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name": "bad-type-cred",
			"type": "invalid",
		})
		s.AssertError(t, resp)
	})

	t.Run("CreateCredentialMissingName", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"type":   "http_basic",
			"secret": "pass",
		})
		s.AssertError(t, resp)
	})

	t.Run("CreateCredentialDuplicateName", func(t *testing.T) {
		s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name":   "dup-cred",
			"type":   "http_basic",
			"secret": "pass",
		})

		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name":   "dup-cred",
			"type":   "http_basic",
			"secret": "pass2",
		})
		s.AssertError(t, resp)
	})

	t.Run("CreateSSHCredWithoutKey", func(t *testing.T) {
		resp := s.PostJSON(t, "/api/v1/credentials/", map[string]interface{}{
			"name": "ssh-no-key",
			"type": "ssh_key",
		})
		s.AssertError(t, resp)
	})
}
