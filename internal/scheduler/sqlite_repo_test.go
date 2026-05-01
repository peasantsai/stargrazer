package scheduler

import (
	"testing"
	"time"

	"stargrazer/internal/db/dbtest"
)

func newJob() *Job {
	return &Job{
		ID:         "job-1",
		Name:       "Keep-alive Facebook",
		Type:       JobTypeKeepAlive,
		Platforms:  []string{"facebook"},
		CronExpr:   "0 */12 * * *",
		Status:     JobStatusActive,
		CreatedAt:  time.Now(),
		LastResult: "",
	}
}

func TestSQLiteRepo_ListEmpty(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	got, err := repo.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(got))
	}
}

func TestSQLiteRepo_SaveInsertThenGetThenList(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	j := newJob()
	if err := repo.Save(j); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.Get(j.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != j.Name || got.Type != j.Type {
		t.Errorf("round-trip mismatch: got=%+v", got)
	}
	if len(got.Platforms) != 1 || got.Platforms[0] != "facebook" {
		t.Errorf("Platforms: %v", got.Platforms)
	}

	all, _ := repo.List()
	if len(all) != 1 {
		t.Errorf("expected 1 job, got %d", len(all))
	}
}

func TestSQLiteRepo_SaveUpdatePreservesID(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	j := newJob()
	_ = repo.Save(j)
	j.Status = JobStatusPaused
	j.LastResult = "manual pause"
	if err := repo.Save(j); err != nil {
		t.Fatalf("Save update: %v", err)
	}
	got, _ := repo.Get(j.ID)
	if got.Status != JobStatusPaused {
		t.Errorf("Status not updated: %v", got.Status)
	}
	if got.LastResult != "manual pause" {
		t.Errorf("LastResult not updated: %q", got.LastResult)
	}
}

func TestSQLiteRepo_DeleteRemovesRow(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	j := newJob()
	_ = repo.Save(j)
	if err := repo.Delete(j.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.Get(j.ID); err == nil {
		t.Error("expected error on Get after Delete")
	}
}

func TestSQLiteRepo_UploadConfigRoundTrip(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	j := newJob()
	j.ID = "upload-job"
	j.Type = JobTypeUpload
	j.UploadConfig = &UploadConfig{
		FilePath: "/tmp/photo.jpg",
		Caption:  "hello",
		Hashtags: []string{"#a", "#b"},
	}
	_ = repo.Save(j)
	got, _ := repo.Get(j.ID)
	if got.UploadConfig == nil {
		t.Fatal("UploadConfig nil after round-trip")
	}
	if got.UploadConfig.FilePath != "/tmp/photo.jpg" {
		t.Errorf("FilePath: %q", got.UploadConfig.FilePath)
	}
	if len(got.UploadConfig.Hashtags) != 2 {
		t.Errorf("Hashtags: %v", got.UploadConfig.Hashtags)
	}
}
