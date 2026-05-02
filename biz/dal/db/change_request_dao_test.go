package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestChangeRequestDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewChangeRequestDAO()

	cr := &po.ChangeRequest{
		RepoID:           1,
		CRNumber:         42,
		Title:            "Fix bug",
		State:            "opened",
		SourceBranch:     "feature",
		TargetBranch:     "main",
		AuthorName:       "dev",
		ProviderConfigID: 1,
	}
	if err := dao.Create(cr); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindByID(cr.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Title != "Fix bug" {
		t.Errorf("title mismatch: got %s", found.Title)
	}

	byNumber, err := dao.FindByRepoAndNumber(1, 42)
	if err != nil {
		t.Fatalf("FindByRepoAndNumber failed: %v", err)
	}
	if byNumber.Title != "Fix bug" {
		t.Errorf("title mismatch: got %s", byNumber.Title)
	}

	found.State = "merged"
	dao.Save(found)

	if err := dao.DeleteByRepo(1); err != nil {
		t.Fatalf("DeleteByRepo failed: %v", err)
	}
	_, err = dao.FindByID(cr.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestChangeRequestDAO_FindByRepo(t *testing.T) {
	SetupTestDB(t)
	dao := NewChangeRequestDAO()
	dao.Create(&po.ChangeRequest{RepoID: 1, CRNumber: 1, Title: "CR1", State: "opened"})
	dao.Create(&po.ChangeRequest{RepoID: 1, CRNumber: 2, Title: "CR2", State: "merged"})
	dao.Create(&po.ChangeRequest{RepoID: 2, CRNumber: 1, Title: "CR3", State: "opened"})

	crs, count, err := dao.FindByRepo(1, "", "", "", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2 CRs, got %d", count)
	}
	if len(crs) != 2 {
		t.Errorf("expected 2 CRs in page, got %d", len(crs))
	}
}

func TestChangeRequestDAO_FindByRepo_FilterByState(t *testing.T) {
	SetupTestDB(t)
	dao := NewChangeRequestDAO()
	dao.Create(&po.ChangeRequest{RepoID: 1, CRNumber: 1, Title: "CR1", State: "opened"})
	dao.Create(&po.ChangeRequest{RepoID: 1, CRNumber: 2, Title: "CR2", State: "merged"})

	crs, count, _ := dao.FindByRepo(1, "merged", "", "", 1, 10)
	if count != 1 {
		t.Errorf("expected 1 merged CR, got %d", count)
	}
	if len(crs) != 1 || crs[0].State != "merged" {
		t.Error("filter by state failed")
	}
}
