package planner

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var placeholderRE = regexp.MustCompile(`\{\{(\w+)\}\}`)

// substitute walks s once, replacing every {{key}} with the value from vars.
// Returns the substituted string plus the deduped list of missing keys
// (placeholders whose key was not in vars).
//
// Value handling:
//   - string                     → as-is
//   - []string / []any of strings → joined with " " (matches workflow.PrepareSteps for hashtags)
//   - everything else            → JSON-marshalled (e.g. numbers, bools)
func substitute(s string, vars map[string]any) (string, []string) {
	if !strings.Contains(s, "{{") {
		return s, nil
	}
	missingSet := make(map[string]bool)
	out := placeholderRE.ReplaceAllStringFunc(s, func(match string) string {
		key := match[2 : len(match)-2]
		v, ok := vars[key]
		if !ok {
			missingSet[key] = true
			return match
		}
		return formatVar(v)
	})
	if len(missingSet) == 0 {
		return out, nil
	}
	missing := make([]string, 0, len(missingSet))
	for k := range missingSet {
		missing = append(missing, k)
	}
	return out, missing
}

func formatVar(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []string:
		return strings.Join(x, " ")
	case []any:
		parts := make([]string, len(x))
		for i, item := range x {
			if s, ok := item.(string); ok {
				parts[i] = s
			} else {
				parts[i] = fmt.Sprintf("%v", item)
			}
		}
		return strings.Join(parts, " ")
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}
