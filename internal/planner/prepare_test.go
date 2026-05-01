package planner_test

import (
	"errors"
	"strings"
	"testing"

	"stargrazer/internal/automation"
	"stargrazer/internal/planner"
	"stargrazer/internal/profile"
	"stargrazer/internal/template"
)

// fakeTemplateRepo satisfies template.Repository against an in-memory map.
type fakeTemplateRepo struct {
	byID map[string]*template.Template
}

func (f *fakeTemplateRepo) Save(*template.Template) error { return nil }
func (f *fakeTemplateRepo) Get(id string) (*template.Template, error) {
	if t, ok := f.byID[id]; ok {
		return t, nil
	}
	return nil, template.ErrNotFound
}
func (f *fakeTemplateRepo) GetByName(*string, string) (*template.Template, error) {
	return nil, template.ErrNotFound
}
func (f *fakeTemplateRepo) List(string) ([]template.Template, error) { return nil, nil }
func (f *fakeTemplateRepo) Delete(string) error                      { return nil }

type fakeProfileRepo struct {
	byID map[string]*profile.Profile
}

func (f *fakeProfileRepo) Save(*profile.Profile) error { return nil }
func (f *fakeProfileRepo) Get(id string) (*profile.Profile, error) {
	if p, ok := f.byID[id]; ok {
		return p, nil
	}
	return nil, profile.ErrNotFound
}
func (f *fakeProfileRepo) GetByName(string) (*profile.Profile, error) {
	return nil, profile.ErrNotFound
}
func (f *fakeProfileRepo) List() ([]profile.Profile, error) { return nil, nil }
func (f *fakeProfileRepo) Delete(string) error              { return nil }

func newResolver(templates map[string]*template.Template, profiles map[string]*profile.Profile) planner.Resolver {
	if templates == nil {
		templates = map[string]*template.Template{}
	}
	if profiles == nil {
		profiles = map[string]*profile.Profile{}
	}
	return planner.NewResolver(&fakeTemplateRepo{byID: templates}, &fakeProfileRepo{byID: profiles})
}

func TestPreparePlan_DirectStepsNoTemplates(t *testing.T) {
	r := newResolver(nil, nil)
	a := &automation.Config{
		Steps: []automation.Step{
			{Action: automation.ActionNavigate, Target: "https://example.com"},
		},
	}
	p, err := r.PreparePlan(a, planner.RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Steps) != 1 || p.Steps[0].Target != "https://example.com" {
		t.Errorf("steps = %v", p.Steps)
	}
	if len(p.Warnings) != 0 {
		t.Errorf("warnings = %v", p.Warnings)
	}
	if len(p.Provenance) != 1 || p.Provenance[0].TemplateID != "" {
		t.Errorf("provenance = %v", p.Provenance)
	}
}

func TestPreparePlan_TemplateInlining(t *testing.T) {
	templates := map[string]*template.Template{
		"login-fb": {
			ID:   "login-fb",
			Name: "login-fb",
			Steps: []automation.Step{
				{Action: automation.ActionNavigate, Target: "https://fb.com"},
				{Action: automation.ActionType, Target: "input[name=email]", Value: "{{username}}"},
			},
		},
	}
	r := newResolver(templates, nil)
	a := &automation.Config{
		Steps: []automation.Step{
			{Action: automation.ActionTemplate, Target: "login-fb"},
			{Action: automation.ActionClick, Target: "#post"},
		},
	}
	p, err := r.PreparePlan(a, planner.RunOptions{Vars: map[string]any{"username": "u@x.com"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Steps) != 3 {
		t.Fatalf("got %d steps, want 3", len(p.Steps))
	}
	if p.Steps[0].Target != "https://fb.com" || p.Steps[1].Value != "u@x.com" || p.Steps[2].Target != "#post" {
		t.Errorf("steps mismatched: %v", p.Steps)
	}
	if p.Provenance[0].TemplateID != "login-fb" || p.Provenance[1].TemplateID != "login-fb" {
		t.Errorf("template provenance missing on inlined steps: %v", p.Provenance)
	}
	if p.Provenance[2].TemplateID != "" {
		t.Errorf("direct step provenance leaked TemplateID: %v", p.Provenance[2])
	}
}

func TestPreparePlan_MissingTemplateWarns(t *testing.T) {
	r := newResolver(nil, nil)
	a := &automation.Config{
		Steps: []automation.Step{
			{Action: automation.ActionTemplate, Target: "ghost"},
			{Action: automation.ActionClick, Target: "#post"},
		},
	}
	p, err := r.PreparePlan(a, planner.RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Steps) != 1 {
		t.Errorf("expected 1 surviving step, got %d", len(p.Steps))
	}
	found := false
	for _, w := range p.Warnings {
		if strings.Contains(w, "missing template") && strings.Contains(w, "ghost") {
			found = true
		}
	}
	if !found {
		t.Errorf("missing template warning not surfaced: %v", p.Warnings)
	}
}

func TestPreparePlan_NestedTemplateIgnoredWithWarning(t *testing.T) {
	templates := map[string]*template.Template{
		"outer": {
			ID:   "outer",
			Name: "outer",
			Steps: []automation.Step{
				{Action: automation.ActionTemplate, Target: "inner"},
				{Action: automation.ActionClick, Target: "#x"},
			},
		},
	}
	r := newResolver(templates, nil)
	a := &automation.Config{
		Steps: []automation.Step{
			{Action: automation.ActionTemplate, Target: "outer"},
		},
	}
	p, err := r.PreparePlan(a, planner.RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Steps) != 1 || p.Steps[0].Target != "#x" {
		t.Errorf("expected only #x to survive, got %v", p.Steps)
	}
	found := false
	for _, w := range p.Warnings {
		if strings.Contains(w, "nested template") {
			found = true
		}
	}
	if !found {
		t.Errorf("nested template warning not surfaced: %v", p.Warnings)
	}
}

func TestPreparePlan_VarPrecedence(t *testing.T) {
	profiles := map[string]*profile.Profile{
		"default-prof": {ID: "default-prof", Name: "default", Vars: map[string]any{"caption": "from-default", "file": "from-default.jpg"}},
		"sched-prof":   {ID: "sched-prof", Name: "sched", Vars: map[string]any{"caption": "from-sched"}},
	}
	r := newResolver(nil, profiles)
	a := &automation.Config{
		DefaultProfileID: "default-prof",
		Steps: []automation.Step{
			{Action: automation.ActionType, Value: "{{caption}} / {{file}}"},
		},
	}
	p, err := r.PreparePlan(a, planner.RunOptions{
		ProfileID: "sched-prof",
		Vars:      map[string]any{"caption": "from-runopts"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Steps[0].Value != "from-runopts / from-default.jpg" {
		t.Errorf("precedence wrong: %q", p.Steps[0].Value)
	}
}

func TestPreparePlan_RequiredVarsLint(t *testing.T) {
	templates := map[string]*template.Template{
		"login": {
			ID:           "login",
			Name:         "login",
			RequiredVars: []string{"username", "password"},
			Steps: []automation.Step{
				{Action: automation.ActionClick, Target: "#login"},
			},
		},
	}
	r := newResolver(templates, nil)
	a := &automation.Config{
		Steps: []automation.Step{
			{Action: automation.ActionTemplate, Target: "login"},
		},
	}
	p, err := r.PreparePlan(a, planner.RunOptions{Vars: map[string]any{"username": "u"}})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(p.Warnings, "\n")
	if !strings.Contains(got, "{{password}}") {
		t.Errorf("expected password lint warning; warnings=%v", p.Warnings)
	}
	for _, w := range p.Warnings {
		if strings.Contains(w, "{{username}}") && strings.Contains(w, "requires") {
			t.Errorf("username should not warn; got %q", w)
		}
	}
}

func TestPreparePlan_UnsetVarWarns(t *testing.T) {
	r := newResolver(nil, nil)
	a := &automation.Config{
		Steps: []automation.Step{
			{Action: automation.ActionType, Value: "{{caption}}"},
			{Action: automation.ActionType, Value: "{{caption}}"},
		},
	}
	p, err := r.PreparePlan(a, planner.RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, w := range p.Warnings {
		if strings.Contains(w, "caption") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("unset-var warnings not deduped: %v", p.Warnings)
	}
}

func TestPreparePlan_MissingProfileWarns(t *testing.T) {
	r := newResolver(nil, nil)
	a := &automation.Config{Steps: []automation.Step{{Action: automation.ActionClick, Target: "#x"}}}
	p, err := r.PreparePlan(a, planner.RunOptions{ProfileID: "missing"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, w := range p.Warnings {
		if strings.Contains(w, "missing profile") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing-profile warning: %v", p.Warnings)
	}
}

func TestPreparePlan_SelectorsNotSubstituted(t *testing.T) {
	r := newResolver(nil, nil)
	a := &automation.Config{
		Steps: []automation.Step{
			{Action: automation.ActionClick, Target: "#x", Selectors: [][]string{{"#{{leak}}"}}},
		},
	}
	p, err := r.PreparePlan(a, planner.RunOptions{Vars: map[string]any{"leak": "evil"}})
	if err != nil {
		t.Fatal(err)
	}
	if p.Steps[0].Selectors[0][0] != "#{{leak}}" {
		t.Errorf("Selectors got substituted: %v", p.Steps[0].Selectors)
	}
}

func TestPreparePlan_EmptyAutomation(t *testing.T) {
	r := newResolver(nil, nil)
	p, err := r.PreparePlan(&automation.Config{Steps: nil}, planner.RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Steps) != 0 || len(p.Warnings) != 0 {
		t.Errorf("expected empty plan, got %+v", p)
	}
}

func TestPreparePlan_HardErrorOnNilAutomation(t *testing.T) {
	r := newResolver(nil, nil)
	_, err := r.PreparePlan(nil, planner.RunOptions{})
	if err == nil || !errors.Is(err, planner.ErrNilAutomation) {
		t.Fatalf("got %v, want ErrNilAutomation", err)
	}
}
