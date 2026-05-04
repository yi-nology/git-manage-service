package db

import (
	"testing"
)

func TestSystemConfigDAO_GetSet(t *testing.T) {
	SetupTestDB(t)
	dao := NewSystemConfigDAO()

	val, err := dao.GetConfig("nonexistent")
	if err == nil {
		t.Log("GetConfig returns empty for nonexistent key (no error)")
	}
	_ = val

	if err := dao.SetConfig("test-key", "test-value"); err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	val, err = dao.GetConfig("test-key")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if val != "test-value" {
		t.Errorf("value mismatch: got %s", val)
	}
}

func TestSystemConfigDAO_Update(t *testing.T) {
	SetupTestDB(t)
	dao := NewSystemConfigDAO()
	dao.SetConfig("key1", "value1")
	dao.SetConfig("key1", "value2")

	val, _ := dao.GetConfig("key1")
	if val != "value2" {
		t.Errorf("expected value2 after update, got %s", val)
	}
}

func TestSystemConfigDAO_GetAll(t *testing.T) {
	SetupTestDB(t)
	dao := NewSystemConfigDAO()
	dao.SetConfig("k1", "v1")
	dao.SetConfig("k2", "v2")

	all, err := dao.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 configs, got %d", len(all))
	}
	if all["k1"] != "v1" {
		t.Errorf("k1 mismatch: got %s", all["k1"])
	}
}

func TestSystemConfigDAO_EmptyGetAll(t *testing.T) {
	SetupTestDB(t)
	dao := NewSystemConfigDAO()
	all, err := dao.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 configs, got %d", len(all))
	}
}
