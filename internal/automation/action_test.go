package automation

import "testing"

func TestActionConstants_StringValues(t *testing.T) {
	cases := map[Action]string{
		ActionNavigate:       "navigate",
		ActionClick:          "click",
		ActionDoubleClick:    "doubleClick",
		ActionType:           "type",
		ActionKeyDown:        "keyDown",
		ActionKeyUp:          "keyUp",
		ActionWait:           "wait",
		ActionWaitForElement: "waitForElement",
		ActionEvaluate:       "evaluate",
		ActionScroll:         "scroll",
		ActionHover:          "hover",
		ActionSetViewport:    "setViewport",
	}
	for a, want := range cases {
		if string(a) != want {
			t.Errorf("Action %v: got %q, want %q", a, string(a), want)
		}
	}
}

func TestAllActions_ContainsEveryConstant(t *testing.T) {
	got := AllActions()
	want := []Action{
		ActionNavigate, ActionClick, ActionDoubleClick, ActionType,
		ActionKeyDown, ActionKeyUp, ActionWait, ActionWaitForElement,
		ActionEvaluate, ActionScroll, ActionHover, ActionSetViewport,
	}
	if len(got) != len(want) {
		t.Fatalf("AllActions length: got %d, want %d", len(got), len(want))
	}
	seen := make(map[Action]bool)
	for _, a := range got {
		seen[a] = true
	}
	for _, w := range want {
		if !seen[w] {
			t.Errorf("AllActions missing %q", w)
		}
	}
}
