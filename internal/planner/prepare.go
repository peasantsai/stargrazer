// Package planner resolves an automation.Config + RunOptions into a flat,
// fully-substituted []automation.Step ready for the existing browser executor.
// One pass: template inlining (depth limit 1), then {{var}} substitution on
// Target and Value (Selectors are left literal).
package planner

import (
	"errors"
	"fmt"
	"sort"

	"stargrazer/internal/automation"
	"stargrazer/internal/profile"
	"stargrazer/internal/template"
)

// ErrNilAutomation is returned when PreparePlan is called with nil.
var ErrNilAutomation = errors.New("planner: nil automation")

// RunOptions controls plan resolution. Vars wins over schedule profile,
// which wins over automation default profile (DSG-001-P3 Decision 3).
type RunOptions struct {
	Vars      map[string]any
	ProfileID string
}

// Plan is the resolver's output: flat steps + parallel provenance + warnings.
type Plan struct {
	Steps      []automation.Step
	Provenance []StepProvenance
	Warnings   []string
}

// StepProvenance carries each plan step's origin so the run journal can trace
// failures back to the authored automation step (and template, if any).
type StepProvenance struct {
	AutomationStepIndex int
	TemplateID          string // "" if not from a template
	TemplateStepIndex   int    // -1 if not from a template
}

// Resolver is the seam consumed by App.RunAutomation/TestAutomation.
type Resolver interface {
	PreparePlan(a *automation.Config, opts RunOptions) (*Plan, error)
}

type resolver struct {
	templates template.Repository
	profiles  profile.Repository
}

// NewResolver constructs the production resolver.
func NewResolver(templates template.Repository, profiles profile.Repository) Resolver {
	return &resolver{templates: templates, profiles: profiles}
}

func (r *resolver) PreparePlan(a *automation.Config, opts RunOptions) (*Plan, error) {
	if a == nil {
		return nil, ErrNilAutomation
	}

	plan := &Plan{}
	warningsSet := make(map[string]bool)
	addWarning := func(w string) {
		if !warningsSet[w] {
			warningsSet[w] = true
			plan.Warnings = append(plan.Warnings, w)
		}
	}

	// 1. Variable bag, lowest → highest precedence.
	vars := map[string]any{}
	if a.DefaultProfileID != "" {
		p, err := r.profiles.Get(a.DefaultProfileID)
		if err == nil {
			mergeVars(vars, p.Vars)
		} else if errors.Is(err, profile.ErrNotFound) {
			addWarning(fmt.Sprintf("missing profile (automation default): %s", a.DefaultProfileID))
		} else {
			return nil, fmt.Errorf("load default profile %s: %w", a.DefaultProfileID, err)
		}
	}
	if opts.ProfileID != "" {
		p, err := r.profiles.Get(opts.ProfileID)
		if err == nil {
			mergeVars(vars, p.Vars)
		} else if errors.Is(err, profile.ErrNotFound) {
			addWarning(fmt.Sprintf("missing profile (run): %s", opts.ProfileID))
		} else {
			return nil, fmt.Errorf("load run profile %s: %w", opts.ProfileID, err)
		}
	}
	mergeVars(vars, opts.Vars)

	// 2. Inline templates (depth-limit 1).
	for i, s := range a.Steps {
		if s.Action != automation.ActionTemplate {
			plan.Steps = append(plan.Steps, s)
			plan.Provenance = append(plan.Provenance, StepProvenance{AutomationStepIndex: i, TemplateID: "", TemplateStepIndex: -1})
			continue
		}
		t, err := r.templates.Get(s.Target)
		if err != nil {
			if errors.Is(err, template.ErrNotFound) {
				addWarning(fmt.Sprintf("missing template: %s", s.Target))
				continue
			}
			return nil, fmt.Errorf("load template %s: %w", s.Target, err)
		}
		for _, k := range t.RequiredVars {
			if _, ok := vars[k]; !ok {
				addWarning(fmt.Sprintf("template %s requires {{%s}} but it is unset", t.Name, k))
			}
		}
		for j, ts := range t.Steps {
			if ts.Action == automation.ActionTemplate {
				addWarning(fmt.Sprintf("nested template ignored inside %s: %s", t.Name, ts.Target))
				continue
			}
			plan.Steps = append(plan.Steps, ts)
			plan.Provenance = append(plan.Provenance, StepProvenance{AutomationStepIndex: i, TemplateID: t.ID, TemplateStepIndex: j})
		}
	}

	// 3. Substitute placeholders in Target and Value.
	for i := range plan.Steps {
		newTarget, missingT := substitute(plan.Steps[i].Target, vars)
		newValue, missingV := substitute(plan.Steps[i].Value, vars)
		plan.Steps[i].Target = newTarget
		plan.Steps[i].Value = newValue
		for _, k := range missingT {
			addWarning(fmt.Sprintf("unset var: %s", k))
		}
		for _, k := range missingV {
			addWarning(fmt.Sprintf("unset var: %s", k))
		}
	}

	sort.Strings(plan.Warnings)
	return plan, nil
}

func mergeVars(dst, src map[string]any) {
	for k, v := range src {
		dst[k] = v
	}
}
