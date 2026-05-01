package automation

import (
	"testing"
	"time"

	"stargrazer/internal/db/dbtest"
)

func TestSQLiteRepo_ListEmpty(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	got, err := repo.List("instagram")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestSQLiteRepo_SaveInsertThenList(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	cfg := Config{
		PlatformID:  "facebook",
		Name:        "Post photo",
		Description: "Upload a single photo",
		Steps: []Step{
			{Action: ActionNavigate, Target: "https://www.facebook.com"},
			{Action: ActionClick, Target: "#composer"},
		},
	}
	saved, err := repo.Save(cfg)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID == "" {
		t.Fatal("Save did not assign an ID")
	}
	if saved.CreatedAt.IsZero() {
		t.Error("Save did not set CreatedAt")
	}

	list, err := repo.List("facebook")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 row, got %d", len(list))
	}
	if list[0].Name != "Post photo" {
		t.Errorf("Name: got %q", list[0].Name)
	}
	if len(list[0].Steps) != 2 {
		t.Errorf("Steps: got %d", len(list[0].Steps))
	}
	if list[0].Steps[0].Action != ActionNavigate {
		t.Errorf("first step action: got %q", list[0].Steps[0].Action)
	}
}

func TestSQLiteRepo_SaveUpdateExistingID(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	first, _ := repo.Save(Config{PlatformID: "x", Name: "v1", Steps: []Step{{Action: ActionNavigate, Target: "https://x.com"}}})

	updated := first
	updated.Name = "v2"
	if _, err := repo.Save(updated); err != nil {
		t.Fatalf("Save update: %v", err)
	}

	list, _ := repo.List("x")
	if len(list) != 1 {
		t.Fatalf("expected 1 row after update, got %d", len(list))
	}
	if list[0].Name != "v2" {
		t.Errorf("Name not updated: %q", list[0].Name)
	}
	// CreatedAt must not change across updates.
	if !list[0].CreatedAt.Equal(first.CreatedAt) {
		t.Errorf("CreatedAt changed across update: first=%v after=%v", first.CreatedAt, list[0].CreatedAt)
	}
}

func TestSQLiteRepo_DeleteRemovesRow(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	saved, _ := repo.Save(Config{PlatformID: "tiktok", Name: "drop", Steps: []Step{{Action: ActionNavigate, Target: "https://tiktok.com"}}})

	ok, err := repo.Delete("tiktok", saved.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !ok {
		t.Error("Delete returned false on present row")
	}

	list, _ := repo.List("tiktok")
	if len(list) != 0 {
		t.Errorf("expected 0 rows after delete, got %d", len(list))
	}
}

func TestSQLiteRepo_DeleteUnknownReturnsFalse(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	ok, err := repo.Delete("instagram", "no-such-id")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if ok {
		t.Error("Delete returned true on missing row")
	}
}

func TestSQLiteRepo_RecordRunBumpsCountAndLastRun(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	saved, _ := repo.Save(Config{PlatformID: "youtube", Name: "rec", Steps: []Step{{Action: ActionNavigate, Target: "https://youtube.com"}}})

	before := time.Now()
	if err := repo.RecordRun("youtube", saved.ID); err != nil {
		t.Fatalf("RecordRun: %v", err)
	}
	list, _ := repo.List("youtube")
	if list[0].RunCount != 1 {
		t.Errorf("RunCount: got %d", list[0].RunCount)
	}
	if list[0].LastRun.Before(before) {
		t.Errorf("LastRun not updated: %v", list[0].LastRun)
	}
}
