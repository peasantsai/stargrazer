// Package template persists named, parameterised step sequences referenced
// from automations via {action:"template", target:"<id>"} sentinels and
// inlined by the planner at run time.
package template

import (
	"errors"
	"time"

	"stargrazer/internal/automation"
)

// ErrNotFound is returned when Get/GetByName/Delete misses.
var ErrNotFound = errors.New("template: not found")

// ErrNestedTemplate rejects templates whose Steps contain a template-action.
// Depth limit is 1 in P3 (DSG-001-P3 spec).
var ErrNestedTemplate = errors.New("template: nested templates not allowed (depth limit 1)")

// Template is a named, optionally platform-scoped step sequence.
type Template struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	PlatformID   *string           `json:"platformId"`
	Steps        []automation.Step `json:"steps"`
	RequiredVars []string          `json:"requiredVars"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

// Repository is the persistence seam for step templates.
type Repository interface {
	Save(t *Template) error
	Get(id string) (*Template, error)
	GetByName(platformID *string, name string) (*Template, error)
	List(platformID string) ([]Template, error)
	Delete(id string) error
}
