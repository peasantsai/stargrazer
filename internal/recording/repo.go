// Package recording persists Chrome DevTools Recorder JSON exports along with
// the parsed []automation.Step they produce. Converting to an automation copies
// parsed_steps into a new automations row but leaves the recording untouched.
package recording

import (
	"errors"
	"time"

	"stargrazer/internal/automation"
)

// ErrNotFound is returned when Get/Delete misses.
var ErrNotFound = errors.New("recording: not found")

// Recording is one persisted recorder export.
type Recording struct {
	ID          string            `json:"id"`
	PlatformID  string            `json:"platformId"`
	Title       string            `json:"title"`
	Source      string            `json:"source"`
	RawJSON     string            `json:"rawJson"`
	ParsedSteps []automation.Step `json:"parsedSteps"`
	Warnings    []string          `json:"warnings"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// Repository is the persistence seam for recordings.
type Repository interface {
	Save(r *Recording) error
	Get(id string) (*Recording, error)
	List(platformID string) ([]Recording, error)
	Delete(id string) error
}
