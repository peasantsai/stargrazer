// Package automation persists user-defined automation workflows.
// Production read/write goes through SQLiteRepo (sqlite_repo.go), which
// satisfies Repository (repo.go).
package automation

import "time"

// Step is one command in an automation sequence.
type Step struct {
	Action Action `json:"action"`
	// Target is a CSS selector for click/type/scroll, or a URL for navigate.
	Target string `json:"target"`
	// Value is text to type, JS to evaluate, or milliseconds to wait.
	Value string `json:"value"`
	// Label is a human-readable description shown in the UI.
	Label string `json:"label"`
	// Selectors holds Chrome Recorder-style fallback selector groups.
	// Each outer element is an alternative strategy (CSS, XPath, ARIA, text, pierce).
	// The execution engine tries each until one succeeds.
	// When empty, Target is used as the sole selector.
	Selectors [][]string `json:"selectors,omitempty"`
}

// Config is a named, platform-scoped automation workflow.
type Config struct {
	ID               string    `json:"id"`
	PlatformID       string    `json:"platformId"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Steps            []Step    `json:"steps"`
	CreatedAt        time.Time `json:"createdAt"`
	LastRun          time.Time `json:"lastRun,omitempty"`
	RunCount         int       `json:"runCount"`
	DefaultProfileID string    `json:"defaultProfileId,omitempty"`
}
