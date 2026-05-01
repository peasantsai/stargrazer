package recording_test

import (
	"errors"
	"reflect"
	"testing"

	"stargrazer/internal/automation"
	"stargrazer/internal/db/dbtest"
	"stargrazer/internal/recording"
)

func newRepo(t *testing.T) *recording.SQLiteRepo {
	t.Helper()
	return recording.NewSQLiteRepo(dbtest.NewMemDB(t))
}

func TestSQLiteRepo_Save_RoundTrip(t *testing.T) {
	repo := newRepo(t)
	in := &recording.Recording{
		PlatformID: "facebook",
		Title:      "login-flow.json",
		Source:     "chrome-devtools-recorder",
		RawJSON:    `{"title":"login","steps":[]}`,
		ParsedSteps: []automation.Step{
			{Action: automation.ActionNavigate, Target: "https://facebook.com/login"},
		},
		Warnings: []string{"unknown step type: foo"},
	}
	if err := repo.Save(in); err != nil {
		t.Fatal(err)
	}
	if in.ID == "" {
		t.Fatal("Save did not assign ID")
	}
	got, err := repo.Get(in.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != in.Title || got.RawJSON != in.RawJSON {
		t.Errorf("title/raw mismatch")
	}
	if !reflect.DeepEqual(got.ParsedSteps, in.ParsedSteps) {
		t.Errorf("parsed_steps mismatch: got %v, want %v", got.ParsedSteps, in.ParsedSteps)
	}
	if !reflect.DeepEqual(got.Warnings, in.Warnings) {
		t.Errorf("warnings mismatch: got %v, want %v", got.Warnings, in.Warnings)
	}
}

func TestSQLiteRepo_List_PlatformScoped(t *testing.T) {
	repo := newRepo(t)
	for _, plat := range []string{"facebook", "facebook", "instagram"} {
		if err := repo.Save(&recording.Recording{
			PlatformID:  plat,
			Title:       "rec",
			Source:      "chrome-devtools-recorder",
			RawJSON:     "{}",
			ParsedSteps: []automation.Step{},
			Warnings:    []string{},
		}); err != nil {
			t.Fatal(err)
		}
	}
	fb, err := repo.List("facebook")
	if err != nil {
		t.Fatal(err)
	}
	if len(fb) != 2 {
		t.Errorf("facebook count: got %d, want 2", len(fb))
	}
	ig, err := repo.List("instagram")
	if err != nil {
		t.Fatal(err)
	}
	if len(ig) != 1 {
		t.Errorf("instagram count: got %d, want 1", len(ig))
	}
}

func TestSQLiteRepo_Get_NotFound(t *testing.T) {
	repo := newRepo(t)
	_, err := repo.Get("missing")
	if !errors.Is(err, recording.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestSQLiteRepo_Delete(t *testing.T) {
	repo := newRepo(t)
	r := &recording.Recording{
		PlatformID:  "x",
		Title:       "tmp",
		Source:      "chrome-devtools-recorder",
		RawJSON:     "{}",
		ParsedSteps: []automation.Step{},
		Warnings:    []string{},
	}
	if err := repo.Save(r); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(r.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(r.ID); !errors.Is(err, recording.ErrNotFound) {
		t.Fatalf("after delete: got %v, want ErrNotFound", err)
	}
}

func TestSQLiteRepo_Save_RequiresPlatform(t *testing.T) {
	repo := newRepo(t)
	err := repo.Save(&recording.Recording{
		Title:       "x",
		Source:      "chrome-devtools-recorder",
		RawJSON:     "{}",
		ParsedSteps: []automation.Step{},
		Warnings:    []string{},
	})
	if err == nil {
		t.Fatal("expected platform-required error")
	}
}
