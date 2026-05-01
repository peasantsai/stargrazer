package template_test

import (
	"errors"
	"reflect"
	"testing"

	"stargrazer/internal/automation"
	"stargrazer/internal/db/dbtest"
	"stargrazer/internal/template"
)

func newRepo(t *testing.T) *template.SQLiteRepo {
	t.Helper()
	return template.NewSQLiteRepo(dbtest.NewMemDB(t))
}

func ptr(s string) *string { return &s }

func TestSQLiteRepo_Save_RoundTrip(t *testing.T) {
	repo := newRepo(t)
	in := &template.Template{
		Name:        "login-fb",
		Description: "Log in to Facebook",
		PlatformID:  ptr("facebook"),
		Steps: []automation.Step{
			{Action: automation.ActionNavigate, Target: "https://facebook.com/login"},
			{Action: automation.ActionType, Target: "input[name=email]", Value: "{{username}}"},
		},
		RequiredVars: []string{"username", "password"},
	}
	if err := repo.Save(in); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(in.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "login-fb" || got.PlatformID == nil || *got.PlatformID != "facebook" {
		t.Errorf("name/platform mismatch: %+v", got)
	}
	if !reflect.DeepEqual(got.Steps, in.Steps) {
		t.Errorf("steps mismatch: got %v, want %v", got.Steps, in.Steps)
	}
	if !reflect.DeepEqual(got.RequiredVars, in.RequiredVars) {
		t.Errorf("required_vars mismatch: got %v, want %v", got.RequiredVars, in.RequiredVars)
	}
}

func TestSQLiteRepo_Save_RejectsNestedTemplate(t *testing.T) {
	repo := newRepo(t)
	in := &template.Template{
		Name:       "bad",
		PlatformID: nil,
		Steps: []automation.Step{
			{Action: automation.ActionTemplate, Target: "other-template-id"},
		},
		RequiredVars: []string{},
	}
	err := repo.Save(in)
	if err == nil || !errors.Is(err, template.ErrNestedTemplate) {
		t.Fatalf("got %v, want ErrNestedTemplate", err)
	}
}

func TestSQLiteRepo_List_GlobalUnionPlatform(t *testing.T) {
	repo := newRepo(t)
	must := func(tt *template.Template) {
		if err := repo.Save(tt); err != nil {
			t.Fatal(err)
		}
	}
	must(&template.Template{Name: "global-helper", PlatformID: nil, RequiredVars: []string{}})
	must(&template.Template{Name: "fb-login", PlatformID: ptr("facebook"), RequiredVars: []string{}})
	must(&template.Template{Name: "ig-upload", PlatformID: ptr("instagram"), RequiredVars: []string{}})

	got, err := repo.List("facebook")
	if err != nil {
		t.Fatal(err)
	}
	names := make(map[string]bool, len(got))
	for _, x := range got {
		names[x.Name] = true
	}
	if !names["global-helper"] || !names["fb-login"] {
		t.Errorf("missing expected: got %v", names)
	}
	if names["ig-upload"] {
		t.Errorf("instagram template leaked into facebook list: %v", names)
	}
}

func TestSQLiteRepo_GetByName_PlatformAndGlobalAreSeparate(t *testing.T) {
	repo := newRepo(t)
	must := func(tt *template.Template) {
		if err := repo.Save(tt); err != nil {
			t.Fatal(err)
		}
	}
	must(&template.Template{Name: "shared", PlatformID: nil, RequiredVars: []string{}})
	must(&template.Template{Name: "shared", PlatformID: ptr("facebook"), RequiredVars: []string{}})

	g, err := repo.GetByName(nil, "shared")
	if err != nil {
		t.Fatal(err)
	}
	if g.PlatformID != nil {
		t.Errorf("global lookup returned platform-scoped row")
	}
	p, err := repo.GetByName(ptr("facebook"), "shared")
	if err != nil {
		t.Fatal(err)
	}
	if p.PlatformID == nil || *p.PlatformID != "facebook" {
		t.Errorf("platform lookup returned global row")
	}
}

func TestSQLiteRepo_Delete_NotFound(t *testing.T) {
	repo := newRepo(t)
	if err := repo.Delete("missing"); !errors.Is(err, template.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}
