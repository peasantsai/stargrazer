package profile_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"stargrazer/internal/db/dbtest"
	"stargrazer/internal/profile"
)

func newRepo(t *testing.T) *profile.SQLiteRepo {
	t.Helper()
	return profile.NewSQLiteRepo(dbtest.NewMemDB(t))
}

func TestSQLiteRepo_Save_RoundTrip(t *testing.T) {
	repo := newRepo(t)
	in := &profile.Profile{
		Name: "Morning EN",
		Vars: map[string]any{
			"caption":  "good morning",
			"hashtags": []any{"#morning", "#en"},
			"file":     "C:/posts/m.jpg",
		},
	}
	if err := repo.Save(in); err != nil {
		t.Fatal(err)
	}
	if in.ID == "" {
		t.Fatal("Save did not assign ID")
	}
	if in.CreatedAt.IsZero() || in.UpdatedAt.IsZero() {
		t.Fatal("Save did not set timestamps")
	}
	got, err := repo.Get(in.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != in.Name {
		t.Errorf("name: got %q, want %q", got.Name, in.Name)
	}
	if !reflect.DeepEqual(got.Vars, in.Vars) {
		t.Errorf("vars: got %v, want %v", got.Vars, in.Vars)
	}
}

func TestSQLiteRepo_Save_UpdatePreservesCreatedAt(t *testing.T) {
	repo := newRepo(t)
	p := &profile.Profile{Name: "P", Vars: map[string]any{"k": "v"}}
	if err := repo.Save(p); err != nil {
		t.Fatal(err)
	}
	created := p.CreatedAt
	time.Sleep(2 * time.Millisecond)
	p.Vars = map[string]any{"k": "v2"}
	if err := repo.Save(p); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.CreatedAt.Equal(created) {
		t.Errorf("CreatedAt changed: was %v, now %v", created, got.CreatedAt)
	}
	if !got.UpdatedAt.After(created) {
		t.Errorf("UpdatedAt %v not after CreatedAt %v", got.UpdatedAt, created)
	}
}

func TestSQLiteRepo_Get_NotFound(t *testing.T) {
	repo := newRepo(t)
	_, err := repo.Get("missing")
	if !errors.Is(err, profile.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestSQLiteRepo_GetByName(t *testing.T) {
	repo := newRepo(t)
	p := &profile.Profile{Name: "uniqueName", Vars: map[string]any{"k": "v"}}
	if err := repo.Save(p); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByName("uniqueName")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != p.ID {
		t.Errorf("got ID %q, want %q", got.ID, p.ID)
	}
}

func TestSQLiteRepo_List_OrderedByName(t *testing.T) {
	repo := newRepo(t)
	for _, n := range []string{"Z", "A", "M"} {
		if err := repo.Save(&profile.Profile{Name: n, Vars: map[string]any{}}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Name != "A" || got[1].Name != "M" || got[2].Name != "Z" {
		t.Errorf("List ordering: got %v", []string{got[0].Name, got[1].Name, got[2].Name})
	}
}

func TestSQLiteRepo_Delete(t *testing.T) {
	repo := newRepo(t)
	p := &profile.Profile{Name: "tmp", Vars: map[string]any{}}
	if err := repo.Save(p); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(p.ID); !errors.Is(err, profile.ErrNotFound) {
		t.Fatalf("after delete, Get returned %v, want ErrNotFound", err)
	}
}

func TestSQLiteRepo_Delete_NotFound(t *testing.T) {
	repo := newRepo(t)
	if err := repo.Delete("missing"); !errors.Is(err, profile.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestSQLiteRepo_Save_DuplicateNameRejected(t *testing.T) {
	repo := newRepo(t)
	if err := repo.Save(&profile.Profile{Name: "dup", Vars: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	err := repo.Save(&profile.Profile{Name: "dup", Vars: map[string]any{}})
	if err == nil {
		t.Fatal("expected duplicate-name error")
	}
}
