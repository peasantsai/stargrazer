package browser

import (
	"context"
	"errors"
	"testing"

	"stargrazer/internal/automation"
)

func TestRunStep_UnknownActionReturnsError(t *testing.T) {
	err := RunStep(context.Background(), nil, automation.Step{Action: "no-such-action"})
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
	if !errors.Is(err, ErrNoHandler) {
		t.Errorf("expected ErrNoHandler, got %v", err)
	}
}

func TestRunStep_DispatchesToRegisteredHandler(t *testing.T) {
	const probe automation.Action = "test-probe"
	called := false
	RegisterHandler(probe, func(ctx context.Context, m *Manager, step automation.Step) error {
		called = true
		return nil
	})
	t.Cleanup(func() { unregisterHandlerForTest(probe) })

	if err := RunStep(context.Background(), nil, automation.Step{Action: probe}); err != nil {
		t.Fatalf("RunStep: %v", err)
	}
	if !called {
		t.Error("registered handler was not called")
	}
}

func TestRunStep_PropagatesHandlerError(t *testing.T) {
	const probe automation.Action = "test-probe-err"
	want := errors.New("boom")
	RegisterHandler(probe, func(ctx context.Context, m *Manager, step automation.Step) error {
		return want
	})
	t.Cleanup(func() { unregisterHandlerForTest(probe) })

	got := RunStep(context.Background(), nil, automation.Step{Action: probe})
	if !errors.Is(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestEveryActionConstantHasRegisteredHandler(t *testing.T) {
	for _, a := range automation.AllActions() {
		if _, ok := lookupHandlerForTest(a); !ok {
			t.Errorf("Action %q has no registered StepHandler", a)
		}
	}
}
