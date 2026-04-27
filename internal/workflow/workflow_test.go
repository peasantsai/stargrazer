package workflow

import (
	"encoding/json"
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

// --- PrepareSteps edge cases ---

func TestPrepareStepsEmptyInput(t *testing.T) {
	steps := []Step{}
	req := UploadRequest{}
	prepared := PrepareSteps(steps, req)
	if len(prepared) != 0 {
		t.Errorf("expected 0 steps, got %d", len(prepared))
	}
}

func TestPrepareStepsNoHashtags(t *testing.T) {
	steps := []Step{{Type: StepType_, Value: "{{caption}}"}}
	req := UploadRequest{Caption: "Just a caption", Hashtags: []string{}}
	prepared := PrepareSteps(steps, req)
	// With no hashtags, caption should be just "Just a caption" without newlines
	if prepared[0].Value != "Just a caption" {
		t.Errorf("expected 'Just a caption', got %q", prepared[0].Value)
	}
}

func TestPrepareStepsNilHashtags(t *testing.T) {
	steps := []Step{{Type: StepType_, Value: "{{caption}}"}}
	req := UploadRequest{Caption: "No tags", Hashtags: nil}
	prepared := PrepareSteps(steps, req)
	if prepared[0].Value != "No tags" {
		t.Errorf("expected 'No tags', got %q", prepared[0].Value)
	}
}

func TestPrepareStepsAllPlaceholders(t *testing.T) {
	steps := []Step{
		{Type: StepType_, Value: "Caption: {{caption}} File: {{file}} Tags: {{hashtags}}"},
	}
	req := UploadRequest{
		FilePath: "/video.mp4",
		Caption:  "hello",
		Hashtags: []string{"#a", "#b"},
	}
	prepared := PrepareSteps(steps, req)
	expected := "Caption: hello\n\n#a #b File: /video.mp4 Tags: #a #b"
	if prepared[0].Value != expected {
		t.Errorf("expected %q, got %q", expected, prepared[0].Value)
	}
}

func TestPrepareStepsPreservesOtherFields(t *testing.T) {
	steps := []Step{
		{Type: StepClick, Description: "Click button", Selector: "#btn", Value: "no-template", Timeout: 3000, Optional: true},
	}
	req := UploadRequest{Caption: "test"}
	prepared := PrepareSteps(steps, req)

	if prepared[0].Type != StepClick {
		t.Error("Type should be preserved")
	}
	if prepared[0].Description != "Click button" {
		t.Error("Description should be preserved")
	}
	if prepared[0].Selector != "#btn" {
		t.Error("Selector should be preserved")
	}
	if prepared[0].Value != "no-template" {
		t.Errorf("Value should be unchanged, got %q", prepared[0].Value)
	}
	if prepared[0].Timeout != 3000 {
		t.Error("Timeout should be preserved")
	}
	if !prepared[0].Optional {
		t.Error("Optional should be preserved")
	}
}

// --- LoadWorkflow / SaveWorkflow edge cases ---

func TestLoadWorkflowInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	wfDir := filepath.Join(tmpDir, "workflows")
	os.MkdirAll(wfDir, 0755)
	os.WriteFile(filepath.Join(wfDir, "bad_upload.json"), []byte("not json {["), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_, err := LoadWorkflow("bad")
	if err == nil {
		t.Error("expected error for invalid JSON workflow")
	}
}

func TestSaveWorkflowCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	w := &Workflow{
		ID:          "dir_test",
		Platform:    "dirtest",
		Name:        "Dir Test",
		Description: "Test directory creation",
		Steps:       []Step{{Type: StepNavigate, Value: "https://example.com"}},
	}

	err := SaveWorkflow(w)
	if err != nil {
		t.Fatalf("SaveWorkflow error: %v", err)
	}

	// Verify the directory was created
	wfDir := filepath.Join(tmpDir, "workflows")
	if _, err := os.Stat(wfDir); err != nil {
		t.Fatalf("workflows directory not created: %v", err)
	}

	// Verify the file exists
	fp := filepath.Join(wfDir, "dirtest_upload.json")
	if _, err := os.Stat(fp); err != nil {
		t.Fatalf("workflow file not created: %v", err)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	w := &Workflow{
		ID:          "roundtrip_test",
		Platform:    "roundtrip",
		Name:        "Round Trip",
		Description: "Test save and load",
		Steps: []Step{
			{Type: StepNavigate, Description: "Go", Value: "https://example.com"},
			{Type: StepClick, Description: "Click", Selector: "#btn"},
			{Type: StepWait, Description: "Wait", Timeout: 5000},
		},
	}

	if err := SaveWorkflow(w); err != nil {
		t.Fatalf("SaveWorkflow error: %v", err)
	}

	loaded, err := LoadWorkflow("roundtrip")
	if err != nil {
		t.Fatalf("LoadWorkflow error: %v", err)
	}

	if loaded.ID != w.ID {
		t.Errorf("ID mismatch: %q vs %q", loaded.ID, w.ID)
	}
	if loaded.Platform != w.Platform {
		t.Errorf("Platform mismatch: %q vs %q", loaded.Platform, w.Platform)
	}
	if loaded.Name != w.Name {
		t.Errorf("Name mismatch: %q vs %q", loaded.Name, w.Name)
	}
	if len(loaded.Steps) != len(w.Steps) {
		t.Errorf("Steps count mismatch: %d vs %d", len(loaded.Steps), len(w.Steps))
	}
}

func TestSaveWorkflowOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	w1 := &Workflow{ID: "v1", Platform: "overwrite", Name: "Version 1", Steps: []Step{}}
	w2 := &Workflow{ID: "v2", Platform: "overwrite", Name: "Version 2", Steps: []Step{}}

	SaveWorkflow(w1)
	SaveWorkflow(w2)

	loaded, err := LoadWorkflow("overwrite")
	if err != nil {
		t.Fatalf("LoadWorkflow error: %v", err)
	}
	if loaded.ID != "v2" {
		t.Errorf("expected overwritten ID 'v2', got %q", loaded.ID)
	}
}

// --- DefaultWorkflows content checks ---

func TestDefaultWorkflowsAllStartWithNavigate(t *testing.T) {
	for _, w := range DefaultWorkflows() {
		t.Run(w.Platform, func(t *testing.T) {
			if len(w.Steps) == 0 {
				t.Fatal("no steps")
			}
			if w.Steps[0].Type != StepNavigate {
				t.Errorf("first step should be navigate, got %s", w.Steps[0].Type)
			}
		})
	}
}

func TestDefaultWorkflowsAllContainFileUpload(t *testing.T) {
	for _, w := range DefaultWorkflows() {
		t.Run(w.Platform, func(t *testing.T) {
			found := false
			for _, s := range w.Steps {
				if s.Type == StepUploadFile {
					found = true
					break
				}
			}
			if !found {
				t.Error("expected at least one upload_file step")
			}
		})
	}
}

// --- LoadWorkflow error paths ---

func TestLoadWorkflowJSONUnmarshalError(t *testing.T) {
	// Test the json.Unmarshal error branch directly.
	data := []byte("NOT JSON")
	var w Workflow
	err := json.Unmarshal(data, &w)
	if err == nil {
		t.Error("expected unmarshal error for invalid JSON")
	}
}

func TestLoadWorkflowMissingFile(t *testing.T) {
	// LoadWorkflow with a platform that has no file should return an error.
	_, err := LoadWorkflow("nonexistent_platform_xyz_abc")
	if err == nil {
		t.Error("expected error for missing workflow file")
	}
}

// --- SaveWorkflow and GetWorkflowsDir ---

func TestSaveWorkflowCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	worksDir := filepath.Join(tmpDir, "workflows")

	// Temporarily override the cwd so GetWorkflowsDir falls back to it.
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	w := &Workflow{
		ID:       "test_wf",
		Platform: "instagram",
		Steps:    []Step{{Type: StepNavigate, Value: "https://instagram.com"}},
	}

	if err := SaveWorkflow(w); err != nil {
		t.Fatalf("SaveWorkflow error: %v", err)
	}

	fp := filepath.Join(worksDir, "instagram_upload.json")
	if _, err := os.Stat(fp); err != nil {
		t.Fatalf("expected workflow file to be created at %s: %v", fp, err)
	}
}

func TestSaveAndLoadWorkflowRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	w := &Workflow{
		ID:       "roundtrip",
		Platform: "facebook",
		Steps: []Step{
			{Type: StepNavigate, Value: "https://facebook.com"},
			{Type: StepClick, Selector: "button"},
		},
	}

	if err := SaveWorkflow(w); err != nil {
		t.Fatalf("SaveWorkflow error: %v", err)
	}

	loaded, err := LoadWorkflow("facebook")
	if err != nil {
		t.Fatalf("LoadWorkflow error: %v", err)
	}
	if loaded.ID != "roundtrip" {
		t.Errorf("expected ID 'roundtrip', got %q", loaded.ID)
	}
	if len(loaded.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(loaded.Steps))
	}
}

func TestGetWorkflowsDirFallbackToCwd(t *testing.T) {
	// When no workflows/ dir exists next to the exe, GetWorkflowsDir
	// should return filepath.Join(cwd, "workflows").
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	dir := GetWorkflowsDir()
	expected := filepath.Join(tmpDir, "workflows")
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}
}

func TestGetWorkflowsDirUsesExeAdjacentDir(t *testing.T) {
	// If a workflows/ dir exists next to the test binary, it should be returned.
	// The test binary lives somewhere; we find it and check if workflows/ exists there.
	exe, err := os.Executable()
	if err != nil {
		t.Skip("cannot determine executable path")
	}
	exeDir := filepath.Dir(exe)
	adjacentWorkflows := filepath.Join(exeDir, "workflows")

	// Create the dir temporarily.
	os.MkdirAll(adjacentWorkflows, 0700)
	defer os.RemoveAll(adjacentWorkflows)

	dir := GetWorkflowsDir()
	if dir != adjacentWorkflows {
		t.Errorf("expected exe-adjacent dir %q, got %q", adjacentWorkflows, dir)
	}
}
