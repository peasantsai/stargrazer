package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultWorkflowsReturns6(t *testing.T) {
	wfs := DefaultWorkflows()
	if len(wfs) != 6 {
		t.Errorf("expected 6 workflows, got %d", len(wfs))
	}
}

func TestDefaultWorkflowsHaveExpectedPlatforms(t *testing.T) {
	wfs := DefaultWorkflows()
	expected := map[string]bool{
		"instagram": false,
		"facebook":  false,
		"tiktok":    false,
		"youtube":   false,
		"linkedin":  false,
		"x":         false,
	}

	for _, w := range wfs {
		if _, ok := expected[w.Platform]; ok {
			expected[w.Platform] = true
		}
	}

	for platform, found := range expected {
		if !found {
			t.Errorf("workflow for platform %q not found in DefaultWorkflows()", platform)
		}
	}
}

func TestDefaultWorkflowsHaveRequiredFields(t *testing.T) {
	for _, w := range DefaultWorkflows() {
		t.Run(w.Platform, func(t *testing.T) {
			if w.ID == "" {
				t.Error("ID is empty")
			}
			if w.Name == "" {
				t.Error("Name is empty")
			}
			if w.Description == "" {
				t.Error("Description is empty")
			}
			if len(w.Steps) == 0 {
				t.Error("Steps is empty")
			}
		})
	}
}

func TestPrepareStepsSubstitutesCaption(t *testing.T) {
	steps := []Step{
		{Type: StepType_, Value: "{{caption}}"},
	}
	req := UploadRequest{
		Caption:  "Hello world",
		Hashtags: []string{},
	}

	prepared := PrepareSteps(steps, req)
	if len(prepared) != 1 {
		t.Fatalf("expected 1 step, got %d", len(prepared))
	}
	if prepared[0].Value != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", prepared[0].Value)
	}
}

func TestPrepareStepsSubstitutesHashtags(t *testing.T) {
	steps := []Step{
		{Type: StepType_, Value: "{{hashtags}}"},
	}
	req := UploadRequest{
		Caption:  "My post",
		Hashtags: []string{"#go", "#testing"},
	}

	prepared := PrepareSteps(steps, req)
	if prepared[0].Value != "#go #testing" {
		t.Errorf("expected '#go #testing', got %q", prepared[0].Value)
	}
}

func TestPrepareStepsSubstitutesFile(t *testing.T) {
	steps := []Step{
		{Type: StepUploadFile, Value: "{{file}}"},
	}
	req := UploadRequest{
		FilePath: "/path/to/image.png",
	}

	prepared := PrepareSteps(steps, req)
	if prepared[0].Value != "/path/to/image.png" {
		t.Errorf("expected '/path/to/image.png', got %q", prepared[0].Value)
	}
}

func TestPrepareStepsCaptionIncludesHashtags(t *testing.T) {
	steps := []Step{
		{Type: StepType_, Value: "{{caption}}"},
	}
	req := UploadRequest{
		Caption:  "Check this out",
		Hashtags: []string{"#photo", "#nature"},
	}

	prepared := PrepareSteps(steps, req)
	expected := "Check this out\n\n#photo #nature"
	if prepared[0].Value != expected {
		t.Errorf("expected %q, got %q", expected, prepared[0].Value)
	}
}

func TestPrepareStepsMultiplePlaceholders(t *testing.T) {
	steps := []Step{
		{Type: StepNavigate, Value: "https://example.com"},
		{Type: StepUploadFile, Value: "{{file}}"},
		{Type: StepType_, Value: "{{caption}}"},
		{Type: StepWait, Timeout: 5000},
	}
	req := UploadRequest{
		FilePath: "/tmp/video.mp4",
		Caption:  "My video",
		Hashtags: []string{"#viral"},
	}

	prepared := PrepareSteps(steps, req)

	if prepared[0].Value != "https://example.com" {
		t.Errorf("step 0: expected URL unchanged, got %q", prepared[0].Value)
	}
	if prepared[1].Value != "/tmp/video.mp4" {
		t.Errorf("step 1: expected file path, got %q", prepared[1].Value)
	}
	expectedCaption := "My video\n\n#viral"
	if prepared[2].Value != expectedCaption {
		t.Errorf("step 2: expected %q, got %q", expectedCaption, prepared[2].Value)
	}
	if prepared[3].Timeout != 5000 {
		t.Errorf("step 3: expected timeout 5000, got %d", prepared[3].Timeout)
	}
}

func TestPrepareStepsDoesNotMutateOriginal(t *testing.T) {
	steps := []Step{
		{Type: StepType_, Value: "{{caption}}"},
	}
	req := UploadRequest{Caption: "mutated?"}

	_ = PrepareSteps(steps, req)

	if steps[0].Value != "{{caption}}" {
		t.Errorf("original step was mutated: %q", steps[0].Value)
	}
}

func TestGetWorkflowsDirReturnsNonEmpty(t *testing.T) {
	dir := GetWorkflowsDir()
	if dir == "" {
		t.Error("GetWorkflowsDir() returned empty string")
	}
}

func TestLoadWorkflowErrorForMissing(t *testing.T) {
	// Ensure the workflow file doesn't exist.
	_, err := LoadWorkflow("nonexistent_platform_xyz")
	if err == nil {
		t.Error("expected error for missing workflow, got nil")
	}
}

func TestLoadWorkflowFromTempDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal workflow JSON in the temp dir.
	wfJSON := `{
		"id": "test_upload",
		"platform": "test",
		"name": "Test Upload",
		"description": "A test workflow",
		"steps": [
			{"type": "navigate", "description": "Open page", "value": "https://example.com"}
		]
	}`

	wfDir := filepath.Join(tmpDir, "workflows")
	os.MkdirAll(wfDir, 0755)
	os.WriteFile(filepath.Join(wfDir, "test_upload.json"), []byte(wfJSON), 0644)

	// LoadWorkflow uses GetWorkflowsDir() which we can't override easily,
	// so instead test SaveWorkflow + reload pattern.
	w := &Workflow{
		ID:          "save_test",
		Platform:    "savetest",
		Name:        "Save Test",
		Description: "Testing save",
		Steps: []Step{
			{Type: StepNavigate, Value: "https://example.com"},
		},
	}

	// Save to a custom temp workflows dir by temporarily changing cwd.
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	err := SaveWorkflow(w)
	if err != nil {
		t.Fatalf("SaveWorkflow error: %v", err)
	}

	loaded, err := LoadWorkflow("savetest")
	if err != nil {
		t.Fatalf("LoadWorkflow error: %v", err)
	}
	if loaded.ID != "save_test" {
		t.Errorf("expected ID 'save_test', got %q", loaded.ID)
	}
	if loaded.Platform != "savetest" {
		t.Errorf("expected Platform 'savetest', got %q", loaded.Platform)
	}
	if len(loaded.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(loaded.Steps))
	}
}

func TestStepTypesAreCorrectConstants(t *testing.T) {
	tests := []struct {
		name     string
		stepType StepType
		expected string
	}{
		{"navigate", StepNavigate, "navigate"},
		{"click", StepClick, "click"},
		{"type", StepType_, "type"},
		{"upload_file", StepUploadFile, "upload_file"},
		{"wait", StepWait, "wait"},
		{"wait_navigation", StepWaitNav, "wait_navigation"},
		{"evaluate", StepEvaluate, "evaluate"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if string(tc.stepType) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.stepType))
			}
		})
	}
}

func TestDefaultWorkflowsUseValidStepTypes(t *testing.T) {
	validTypes := map[StepType]bool{
		StepNavigate:   true,
		StepClick:      true,
		StepType_:      true,
		StepUploadFile: true,
		StepWait:       true,
		StepWaitNav:    true,
		StepEvaluate:   true,
	}

	for _, w := range DefaultWorkflows() {
		t.Run(w.Platform, func(t *testing.T) {
			for i, s := range w.Steps {
				if !validTypes[s.Type] {
					t.Errorf("step %d has invalid type %q", i, s.Type)
				}
			}
		})
	}
}
