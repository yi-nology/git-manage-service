package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestNotificationChannelDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewNotificationChannelDAO()

	ch := &po.NotificationChannel{
		Name:    "test-dingtalk",
		Type:    "dingtalk",
		Enabled: true,
	}
	if err := dao.Create(ch); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindByID(ch.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "test-dingtalk" {
		t.Errorf("name mismatch: got %s", found.Name)
	}

	found.Enabled = false
	dao.Save(found)

	if err := dao.Delete(ch); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestNotificationChannelDAO_FindByType(t *testing.T) {
	SetupTestDB(t)
	dao := NewNotificationChannelDAO()
	dao.Create(&po.NotificationChannel{Name: "d1", Type: "dingtalk", Enabled: true})
	dao.Create(&po.NotificationChannel{Name: "f1", Type: "feishu", Enabled: true})
	dao.Create(&po.NotificationChannel{Name: "d2", Type: "dingtalk", Enabled: true})

	dingtalks, err := dao.FindByType("dingtalk")
	if err != nil {
		t.Fatal(err)
	}
	if len(dingtalks) != 2 {
		t.Errorf("expected 2 dingtalk channels, got %d", len(dingtalks))
	}
}

func TestNotificationChannelDAO_FindEnabled(t *testing.T) {
	SetupTestDB(t)
	dao := NewNotificationChannelDAO()
	dao.Create(&po.NotificationChannel{Name: "e1", Type: "dingtalk", Enabled: true})
	dao.Create(&po.NotificationChannel{Name: "e2", Type: "feishu", Enabled: false})

	enabled, err := dao.FindEnabled()
	if err != nil {
		t.Fatal(err)
	}
	if len(enabled) != 1 {
		t.Errorf("expected 1 enabled, got %d", len(enabled))
	}
}

func TestNotificationChannelDAO_FindAll(t *testing.T) {
	SetupTestDB(t)
	dao := NewNotificationChannelDAO()
	dao.Create(&po.NotificationChannel{Name: "c1", Type: "dingtalk"})
	dao.Create(&po.NotificationChannel{Name: "c2", Type: "email"})

	all, err := dao.FindAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 channels, got %d", len(all))
	}
}
