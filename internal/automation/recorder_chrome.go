package automation

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ChromeRecorderImporter parses Chrome DevTools Recorder JSON exports.
// Format reference: https://developer.chrome.com/docs/devtools/recorder/reference
type ChromeRecorderImporter struct{}

func init() { RegisterImporter(&ChromeRecorderImporter{}) }

func (c *ChromeRecorderImporter) Format() string { return "chrome-devtools-recorder" }

func (c *ChromeRecorderImporter) Detect(raw []byte) bool {
	var probe struct {
		Steps []json.RawMessage `json:"steps"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	return probe.Steps != nil
}

// rawRecorderStep is the union shape of every recorder step. Fields we don't
// use are omitted; json.Unmarshal happily ignores unknown keys.
type rawRecorderStep struct {
	Type      string     `json:"type"`
	URL       string     `json:"url,omitempty"`
	Selectors [][]string `json:"selectors,omitempty"`
	Target    string     `json:"target,omitempty"`
	Value     string     `json:"value,omitempty"`
	Key       string     `json:"key,omitempty"`
	Width     int        `json:"width,omitempty"`
	Height    int        `json:"height,omitempty"`
	Timeout   int        `json:"timeout,omitempty"`
}

type rawRecorderFile struct {
	Title string            `json:"title"`
	Steps []rawRecorderStep `json:"steps"`
}

func (c *ChromeRecorderImporter) Import(raw []byte, platformID string) (AutomationDraft, error) {
	var f rawRecorderFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return AutomationDraft{}, fmt.Errorf("parse recorder JSON: %w", err)
	}
	draft := AutomationDraft{
		Title:       firstNonEmpty(f.Title, "Imported recording"),
		Description: fmt.Sprintf("Imported from Chrome DevTools Recorder (%d raw steps)", len(f.Steps)),
	}
	for i, s := range f.Steps {
		step, ok, warning := convertRecorderStep(s)
		if warning != "" {
			draft.Warnings = append(draft.Warnings, fmt.Sprintf("step %d: %s", i+1, warning))
		}
		if ok {
			step.Label = bestLabelFromSelectors(s.Selectors)
			draft.Steps = append(draft.Steps, step)
		}
	}
	return draft, nil
}

// convertRecorderStep returns (step, true, "") on success, (zero, false, msg)
// on a non-fatal skip.
func convertRecorderStep(s rawRecorderStep) (Step, bool, string) {
	primary := primarySelector(s.Selectors)
	switch s.Type {
	case "navigate":
		if s.URL == "" {
			return Step{}, false, "navigate without url"
		}
		return Step{Action: ActionNavigate, Target: s.URL}, true, ""

	case "click":
		return Step{Action: ActionClick, Target: primary, Selectors: s.Selectors}, true, ""

	case "doubleClick":
		return Step{Action: ActionDoubleClick, Target: primary, Selectors: s.Selectors}, true, ""

	case "change":
		return Step{Action: ActionType, Target: primary, Value: s.Value, Selectors: s.Selectors}, true, ""

	case "keyDown":
		if s.Key == "" {
			return Step{}, false, "keyDown without key"
		}
		return Step{Action: ActionKeyDown, Value: s.Key}, true, ""

	case "keyUp":
		if s.Key == "" {
			return Step{}, false, "keyUp without key"
		}
		return Step{Action: ActionKeyUp, Value: s.Key}, true, ""

	case "hover":
		return Step{Action: ActionHover, Target: primary, Selectors: s.Selectors}, true, ""

	case "scroll":
		return Step{Action: ActionScroll, Target: primary, Selectors: s.Selectors}, true, ""

	case "setViewport":
		if s.Width <= 0 || s.Height <= 0 {
			return Step{}, false, "setViewport with non-positive dims"
		}
		dims, _ := json.Marshal(struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		}{s.Width, s.Height})
		return Step{Action: ActionSetViewport, Value: string(dims)}, true, ""

	case "waitForElement":
		val := ""
		if s.Timeout > 0 {
			val = strconv.Itoa(s.Timeout)
		}
		return Step{Action: ActionWaitForElement, Target: primary, Value: val, Selectors: s.Selectors}, true, ""

	case "emulateNetworkConditions", "customStep", "":
		return Step{}, false, fmt.Sprintf("unsupported recorder type %q — skipped", s.Type)
	default:
		return Step{}, false, fmt.Sprintf("unknown recorder type %q — skipped", s.Type)
	}
}

func primarySelector(sels [][]string) string {
	if len(sels) == 0 || len(sels[0]) == 0 {
		return ""
	}
	return sels[0][0]
}

func bestLabelFromSelectors(sels [][]string) string {
	for _, group := range sels {
		for _, s := range group {
			if strings.HasPrefix(s, "text/") {
				return strings.TrimPrefix(s, "text/")
			}
			if strings.HasPrefix(s, "aria/") {
				rest := strings.TrimPrefix(s, "aria/")
				if i := strings.IndexByte(rest, '['); i >= 0 {
					rest = rest[:i]
				}
				return strings.TrimSpace(rest)
			}
		}
	}
	return primarySelector(sels)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
