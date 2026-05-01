package automation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChromeRecorderImporter_Detect(t *testing.T) {
	imp := &ChromeRecorderImporter{}
	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{"with title and steps", `{"title":"x","steps":[{"type":"navigate","url":"https://x"}]}`, true},
		{"with steps only", `{"steps":[{"type":"click"}]}`, true},
		{"no steps key", `{"title":"x"}`, false},
		{"not json", `not json at all`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := imp.Detect([]byte(c.raw)); got != c.want {
				t.Errorf("Detect(%q) = %v, want %v", c.raw, got, c.want)
			}
		})
	}
}

func TestChromeRecorderImporter_Import_GoldenFiles(t *testing.T) {
	files, err := filepath.Glob("testdata/facebook_upload_post_*.json")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no testdata files found")
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read %s: %v", f, err)
			}
			imp := &ChromeRecorderImporter{}
			if !imp.Detect(raw) {
				t.Fatalf("Detect returned false for %s", f)
			}
			draft, err := imp.Import(raw, "facebook")
			if err != nil {
				t.Fatalf("Import: %v", err)
			}
			if draft.Title == "" {
				t.Error("Title empty")
			}
			if len(draft.Steps) == 0 {
				t.Error("Steps empty")
			}
			known := make(map[Action]bool)
			for _, a := range AllActions() {
				known[a] = true
			}
			for i, s := range draft.Steps {
				if !known[s.Action] {
					t.Errorf("step %d: unknown Action %q", i, s.Action)
				}
			}
		})
	}
}

func TestChromeRecorderImporter_Import_HandlesEachActionType(t *testing.T) {
	raw := []byte(`{
	  "title": "all-types",
	  "steps": [
	    {"type":"setViewport","width":1280,"height":720},
	    {"type":"navigate","url":"https://example.com"},
	    {"type":"click","selectors":[["#b"]]},
	    {"type":"doubleClick","selectors":[["#b"]]},
	    {"type":"change","value":"hello","selectors":[["#i"]]},
	    {"type":"keyDown","key":"Enter"},
	    {"type":"keyUp","key":"Enter"},
	    {"type":"hover","selectors":[["#h"]]},
	    {"type":"scroll","selectors":[["#s"]]},
	    {"type":"waitForElement","selectors":[["#w"]],"timeout":2500}
	  ]
	}`)
	imp := &ChromeRecorderImporter{}
	draft, err := imp.Import(raw, "facebook")
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	want := []Action{
		ActionSetViewport, ActionNavigate, ActionClick, ActionDoubleClick,
		ActionType, ActionKeyDown, ActionKeyUp, ActionHover, ActionScroll, ActionWaitForElement,
	}
	if len(draft.Steps) != len(want) {
		t.Fatalf("got %d steps, want %d", len(draft.Steps), len(want))
	}
	for i, w := range want {
		if draft.Steps[i].Action != w {
			t.Errorf("step %d: got %q, want %q", i, draft.Steps[i].Action, w)
		}
	}
}

func TestChromeRecorderImporter_Import_UnknownTypeBecomesWarning(t *testing.T) {
	raw := []byte(`{"title":"x","steps":[{"type":"emulateNetworkConditions"}]}`)
	imp := &ChromeRecorderImporter{}
	draft, err := imp.Import(raw, "facebook")
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if len(draft.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(draft.Warnings))
	}
	if len(draft.Steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(draft.Steps))
	}
}

func TestPickImporter_ReturnsChromeForRecorderJSON(t *testing.T) {
	got := PickImporter([]byte(`{"title":"x","steps":[]}`))
	if got == nil {
		t.Fatal("PickImporter returned nil for valid recorder JSON")
	}
	if got.Format() != "chrome-devtools-recorder" {
		t.Errorf("Format: got %q, want chrome-devtools-recorder", got.Format())
	}
}
