package planner

import (
	"reflect"
	"sort"
	"testing"
)

func TestSubstitute_StringValue(t *testing.T) {
	out, missing := substitute("hello {{caption}}", map[string]any{"caption": "world"})
	if out != "hello world" {
		t.Errorf("got %q, want %q", out, "hello world")
	}
	if len(missing) != 0 {
		t.Errorf("unexpected missing: %v", missing)
	}
}

func TestSubstitute_StringSliceJoinedWithSpace(t *testing.T) {
	out, _ := substitute("tags: {{hashtags}}", map[string]any{"hashtags": []any{"#a", "#b", "#c"}})
	if out != "tags: #a #b #c" {
		t.Errorf("got %q, want %q", out, "tags: #a #b #c")
	}
}

func TestSubstitute_MissingKeyReportsAndPreserves(t *testing.T) {
	out, missing := substitute("{{caption}} and {{ghost}}", map[string]any{"caption": "x"})
	if out != "x and {{ghost}}" {
		t.Errorf("got %q, want %q", out, "x and {{ghost}}")
	}
	if !reflect.DeepEqual(missing, []string{"ghost"}) {
		t.Errorf("missing = %v, want [ghost]", missing)
	}
}

func TestSubstitute_MultipleMissingDeduped(t *testing.T) {
	_, missing := substitute("{{a}} {{b}} {{a}}", map[string]any{})
	sort.Strings(missing)
	want := []string{"a", "b"}
	if !reflect.DeepEqual(missing, want) {
		t.Errorf("missing = %v, want %v", missing, want)
	}
}

func TestSubstitute_NoPlaceholders(t *testing.T) {
	out, missing := substitute("plain text", map[string]any{"x": "y"})
	if out != "plain text" || len(missing) != 0 {
		t.Errorf("got (%q,%v)", out, missing)
	}
}

func TestSubstitute_OtherTypeJSONMarshalled(t *testing.T) {
	out, _ := substitute("count: {{n}}", map[string]any{"n": 42})
	if out != "count: 42" {
		t.Errorf("got %q, want %q", out, "count: 42")
	}
}
