package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StepType is the kind of action to perform.
type StepType string

const (
	StepNavigate   StepType = "navigate"        // Navigate to URL
	StepClick      StepType = "click"            // Click an element (CSS selector)
	StepType_      StepType = "type"             // Type text into element
	StepUploadFile StepType = "upload_file"      // Upload a file via file input selector
	StepWait       StepType = "wait"             // Wait for element or timeout (ms)
	StepWaitNav    StepType = "wait_navigation"  // Wait for page navigation
	StepEvaluate   StepType = "evaluate"         // Evaluate JS expression
)

// Step is a single action in a workflow.
type Step struct {
	Type        StepType `json:"type"`
	Description string   `json:"description"`
	Selector    string   `json:"selector,omitempty"`
	Value       string   `json:"value,omitempty"`
	Timeout     int      `json:"timeout,omitempty"`
	Optional    bool     `json:"optional,omitempty"`
}

// Workflow is a named sequence of steps for a platform.
type Workflow struct {
	ID          string `json:"id"`
	Platform    string `json:"platform"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Steps       []Step `json:"steps"`
}

// UploadRequest is what the frontend sends to trigger an upload.
type UploadRequest struct {
	Platforms []string `json:"platforms"`
	FilePath  string   `json:"filePath"`
	Caption   string   `json:"caption"`
	Hashtags  []string `json:"hashtags"`
}

// GetWorkflowsDir returns the directory where workflow JSON files are stored.
// It checks for a "workflows" directory next to the executable first,
// then falls back to the current working directory.
func GetWorkflowsDir() string {
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Join(filepath.Dir(exe), "workflows")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "workflows"
	}
	return filepath.Join(cwd, "workflows")
}

// LoadWorkflow loads a workflow from <platform>_upload.json in the workflows directory.
func LoadWorkflow(platform string) (*Workflow, error) {
	dir := GetWorkflowsDir()
	filename := filepath.Join(dir, platform+"_upload.json")

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file %s: %w", filename, err)
	}

	var w Workflow
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("failed to parse workflow file %s: %w", filename, err)
	}

	return &w, nil
}

// SaveWorkflow saves a workflow to the workflows directory as <platform>_upload.json.
func SaveWorkflow(w *Workflow) error {
	dir := GetWorkflowsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}

	filename := filepath.Join(dir, w.Platform+"_upload.json")

	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflow: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write workflow file %s: %w", filename, err)
	}

	return nil
}

// PrepareSteps replaces template placeholders in step values with actual request data.
// Supported placeholders: {{caption}}, {{hashtags}}, {{file}}
func PrepareSteps(steps []Step, req UploadRequest) []Step {
	hashtags := strings.Join(req.Hashtags, " ")
	fullCaption := req.Caption
	if hashtags != "" {
		fullCaption = req.Caption + "\n\n" + hashtags
	}

	prepared := make([]Step, len(steps))
	for i, s := range steps {
		prepared[i] = s
		prepared[i].Value = strings.ReplaceAll(s.Value, "{{file}}", req.FilePath)
		prepared[i].Value = strings.ReplaceAll(prepared[i].Value, "{{caption}}", fullCaption)
		prepared[i].Value = strings.ReplaceAll(prepared[i].Value, "{{hashtags}}", hashtags)
	}
	return prepared
}

// DefaultWorkflows returns built-in default workflows for all supported platforms.
func DefaultWorkflows() []Workflow {
	return []Workflow{
		instagramWorkflow(),
		facebookWorkflow(),
		tiktokWorkflow(),
		youtubeWorkflow(),
		linkedinWorkflow(),
		xWorkflow(),
	}
}

func instagramWorkflow() Workflow {
	return Workflow{
		ID:          "instagram_upload",
		Platform:    "instagram",
		Name:        "Instagram Photo Upload",
		Description: "Upload a photo with caption and hashtags to Instagram via web",
		Steps: []Step{
			{Type: StepNavigate, Description: "Open Instagram", Value: "https://www.instagram.com/"},
			{Type: StepWait, Description: "Wait for page to load", Selector: "svg[aria-label='New post']", Timeout: 10000},
			{Type: StepClick, Description: "Click the new post button", Selector: "svg[aria-label='New post']"},
			{Type: StepWait, Description: "Wait for creation dialog", Selector: "div[role='dialog']", Timeout: 5000},
			{Type: StepUploadFile, Description: "Upload the media file", Selector: "div[role='dialog'] input[type='file']", Value: "{{file}}"},
			{Type: StepWait, Description: "Wait for file to process", Timeout: 3000},
			{Type: StepClick, Description: "Click Next button", Selector: "div[role='dialog'] button:has-text('Next')"},
			{Type: StepWait, Description: "Wait for filters screen", Timeout: 2000},
			{Type: StepClick, Description: "Click Next past filters", Selector: "div[role='dialog'] button:has-text('Next')"},
			{Type: StepWait, Description: "Wait for caption input", Selector: "div[role='dialog'] textarea[aria-label='Write a caption...']", Timeout: 5000},
			{Type: StepType_, Description: "Enter caption and hashtags", Selector: "div[role='dialog'] textarea[aria-label='Write a caption...']", Value: "{{caption}}"},
			{Type: StepClick, Description: "Click Share", Selector: "div[role='dialog'] button:has-text('Share')"},
			{Type: StepWait, Description: "Wait for upload to complete", Timeout: 15000},
		},
	}
}

func facebookWorkflow() Workflow {
	return Workflow{
		ID:          "facebook_upload",
		Platform:    "facebook",
		Name:        "Facebook Photo Upload",
		Description: "Upload a photo with caption to Facebook",
		Steps: []Step{
			{Type: StepNavigate, Description: "Open Facebook", Value: "https://www.facebook.com/"},
			{Type: StepWait, Description: "Wait for page to load", Selector: "div[role='main']", Timeout: 10000},
			{Type: StepClick, Description: "Click 'What's on your mind' composer", Selector: "div[role='main'] span:has-text('What\\'s on your mind')"},
			{Type: StepWait, Description: "Wait for post composer dialog", Selector: "div[role='dialog']", Timeout: 5000},
			{Type: StepClick, Description: "Click Photo/Video button", Selector: "div[role='dialog'] div[aria-label='Photo/video']"},
			{Type: StepWait, Description: "Wait for file input", Selector: "div[role='dialog'] input[type='file']", Timeout: 5000},
			{Type: StepUploadFile, Description: "Upload the media file", Selector: "div[role='dialog'] input[type='file']", Value: "{{file}}"},
			{Type: StepWait, Description: "Wait for upload to process", Timeout: 5000},
			{Type: StepType_, Description: "Enter caption", Selector: "div[role='dialog'] div[contenteditable='true']", Value: "{{caption}}"},
			{Type: StepClick, Description: "Click Post", Selector: "div[role='dialog'] div[aria-label='Post']"},
			{Type: StepWait, Description: "Wait for post to publish", Timeout: 10000},
		},
	}
}

func tiktokWorkflow() Workflow {
	return Workflow{
		ID:          "tiktok_upload",
		Platform:    "tiktok",
		Name:        "TikTok Video Upload",
		Description: "Upload a video with caption and hashtags to TikTok",
		Steps: []Step{
			{Type: StepNavigate, Description: "Open TikTok upload page", Value: "https://www.tiktok.com/upload"},
			{Type: StepWait, Description: "Wait for upload page", Selector: "input[type='file']", Timeout: 10000},
			{Type: StepUploadFile, Description: "Upload the media file", Selector: "input[type='file']", Value: "{{file}}"},
			{Type: StepWait, Description: "Wait for video to process", Timeout: 15000},
			{Type: StepClick, Description: "Click caption editor", Selector: "div[contenteditable='true'].public-DraftEditor-content"},
			{Type: StepType_, Description: "Enter caption and hashtags", Selector: "div[contenteditable='true'].public-DraftEditor-content", Value: "{{caption}}"},
			{Type: StepClick, Description: "Click Post button", Selector: "button:has-text('Post')"},
			{Type: StepWait, Description: "Wait for upload to complete", Timeout: 30000},
		},
	}
}

func youtubeWorkflow() Workflow {
	return Workflow{
		ID:          "youtube_upload",
		Platform:    "youtube",
		Name:        "YouTube Video Upload",
		Description: "Upload a video with title and description to YouTube Studio",
		Steps: []Step{
			{Type: StepNavigate, Description: "Open YouTube Studio upload", Value: "https://studio.youtube.com/"},
			{Type: StepWait, Description: "Wait for Studio to load", Selector: "#upload-icon", Timeout: 10000},
			{Type: StepClick, Description: "Click the Create/Upload button", Selector: "#upload-icon"},
			{Type: StepWait, Description: "Wait for upload dialog", Selector: "input[type='file']", Timeout: 5000},
			{Type: StepUploadFile, Description: "Upload the media file", Selector: "input[type='file']", Value: "{{file}}"},
			{Type: StepWait, Description: "Wait for video processing", Selector: "#textbox[aria-label='Add a title that describes your video']", Timeout: 15000},
			{Type: StepClick, Description: "Clear default title", Selector: "#textbox[aria-label='Add a title that describes your video']"},
			{Type: StepType_, Description: "Enter video title (caption)", Selector: "#textbox[aria-label='Add a title that describes your video']", Value: "{{caption}}"},
			{Type: StepClick, Description: "Click description field", Selector: "#textbox[aria-label='Tell viewers about your video']"},
			{Type: StepType_, Description: "Enter description with hashtags", Selector: "#textbox[aria-label='Tell viewers about your video']", Value: "{{hashtags}}"},
			{Type: StepClick, Description: "Click Next", Selector: "#next-button"},
			{Type: StepWait, Description: "Wait for next screen", Timeout: 2000},
			{Type: StepClick, Description: "Click Next again (video elements)", Selector: "#next-button"},
			{Type: StepWait, Description: "Wait for next screen", Timeout: 2000},
			{Type: StepClick, Description: "Click Next again (checks)", Selector: "#next-button"},
			{Type: StepWait, Description: "Wait for visibility screen", Timeout: 2000},
			{Type: StepClick, Description: "Select Public visibility", Selector: "tp-yt-paper-radio-button[name='PUBLIC']"},
			{Type: StepClick, Description: "Click Publish", Selector: "#done-button"},
			{Type: StepWait, Description: "Wait for upload to finish", Timeout: 30000},
		},
	}
}

func linkedinWorkflow() Workflow {
	return Workflow{
		ID:          "linkedin_upload",
		Platform:    "linkedin",
		Name:        "LinkedIn Photo Upload",
		Description: "Upload a photo with caption to LinkedIn",
		Steps: []Step{
			{Type: StepNavigate, Description: "Open LinkedIn feed", Value: "https://www.linkedin.com/feed/"},
			{Type: StepWait, Description: "Wait for page to load", Selector: "div.share-box-feed-entry__trigger", Timeout: 10000},
			{Type: StepClick, Description: "Click Start a post", Selector: "div.share-box-feed-entry__trigger"},
			{Type: StepWait, Description: "Wait for post composer", Selector: "div.share-creation-state__text-editor", Timeout: 5000},
			{Type: StepType_, Description: "Enter caption", Selector: "div.ql-editor[contenteditable='true']", Value: "{{caption}}"},
			{Type: StepClick, Description: "Click Add media button", Selector: "button[aria-label='Add media']"},
			{Type: StepWait, Description: "Wait for file input", Selector: "input[type='file']", Timeout: 5000},
			{Type: StepUploadFile, Description: "Upload the media file", Selector: "input[type='file']", Value: "{{file}}"},
			{Type: StepWait, Description: "Wait for upload to process", Timeout: 10000},
			{Type: StepClick, Description: "Click Post", Selector: "button.share-actions__primary-action"},
			{Type: StepWait, Description: "Wait for post to publish", Timeout: 10000},
		},
	}
}

func xWorkflow() Workflow {
	return Workflow{
		ID:          "x_upload",
		Platform:    "x",
		Name:        "X (Twitter) Photo Upload",
		Description: "Upload a photo with caption and hashtags to X",
		Steps: []Step{
			{Type: StepNavigate, Description: "Open X compose", Value: "https://x.com/compose/post"},
			{Type: StepWait, Description: "Wait for compose box", Selector: "div[data-testid='tweetTextarea_0']", Timeout: 10000},
			{Type: StepType_, Description: "Enter caption and hashtags", Selector: "div[data-testid='tweetTextarea_0']", Value: "{{caption}}"},
			{Type: StepClick, Description: "Click media upload button", Selector: "input[data-testid='fileInput']"},
			{Type: StepUploadFile, Description: "Upload the media file", Selector: "input[data-testid='fileInput']", Value: "{{file}}"},
			{Type: StepWait, Description: "Wait for media to upload", Timeout: 10000},
			{Type: StepClick, Description: "Click Post button", Selector: "button[data-testid='tweetButton']"},
			{Type: StepWait, Description: "Wait for post to publish", Timeout: 10000},
		},
	}
}
