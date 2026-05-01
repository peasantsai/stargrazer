package browser

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"stargrazer/internal/automation"
)

// chromedpFixture spins up a fresh headless chromedp context loaded with the
// given inline HTML. Returned cancel must be deferred by the caller. Each test
// gets its own browser, isolated from the user's `wails dev` Chromium.
func chromedpFixture(t *testing.T, html string) (context.Context, context.CancelFunc) {
	t.Helper()
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", true))...)
	ctx, ctxCancel := chromedp.NewContext(allocCtx)
	encoded := base64.StdEncoding.EncodeToString([]byte(html))
	if err := chromedp.Run(ctx, chromedp.Navigate("data:text/html;base64,"+encoded)); err != nil {
		ctxCancel()
		allocCancel()
		t.Fatalf("navigate fixture: %v", err)
	}
	return ctx, func() { ctxCancel(); allocCancel() }
}

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
	// ActionTemplate is a planner-level sentinel; the resolver inlines it
	// before the executor sees the step. No browser handler exists for it.
	plannerOnly := map[automation.Action]bool{
		automation.ActionTemplate: true,
	}
	for _, a := range automation.AllActions() {
		if plannerOnly[a] {
			continue
		}
		if _, ok := lookupHandlerForTest(a); !ok {
			t.Errorf("Action %q has no registered StepHandler", a)
		}
	}
}

func TestHandleDoubleClick(t *testing.T) {
	ctx, cancel := chromedpFixture(t, `<html><body><button id="b" ondblclick="this.dataset.dbl=1">go</button></body></html>`)
	defer cancel()

	step := automation.Step{Action: automation.ActionDoubleClick, Target: "#b"}
	if err := handleDoubleClick(ctx, nil, step); err != nil {
		t.Fatalf("handleDoubleClick: %v", err)
	}
	var attr string
	if err := chromedp.Run(ctx, chromedp.AttributeValue("#b", "data-dbl", &attr, nil)); err != nil {
		t.Fatalf("read attr: %v", err)
	}
	if attr != "1" {
		t.Errorf("expected data-dbl=1, got %q", attr)
	}
}

func TestHandleHover(t *testing.T) {
	ctx, cancel := chromedpFixture(t,
		`<html><body><div id="t" onmouseover="this.dataset.hov=1" style="width:100px;height:100px"></div></body></html>`)
	defer cancel()

	step := automation.Step{Action: automation.ActionHover, Target: "#t"}
	if err := handleHover(ctx, nil, step); err != nil {
		t.Fatalf("handleHover: %v", err)
	}
	var attr string
	if err := chromedp.Run(ctx, chromedp.AttributeValue("#t", "data-hov", &attr, nil)); err != nil {
		t.Fatalf("read attr: %v", err)
	}
	if attr != "1" {
		t.Errorf("expected data-hov=1, got %q", attr)
	}
}

func TestHandleKeyDown_KeyUp_OnFocusedInput(t *testing.T) {
	ctx, cancel := chromedpFixture(t,
		`<html><body><input id="i" onkeydown="this.dataset.kd=event.key" onkeyup="this.dataset.ku=event.key"></body></html>`)
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Focus("#i")); err != nil {
		t.Fatalf("focus: %v", err)
	}
	if err := handleKeyDown(ctx, nil, automation.Step{Action: automation.ActionKeyDown, Value: "A"}); err != nil {
		t.Fatalf("handleKeyDown: %v", err)
	}
	if err := handleKeyUp(ctx, nil, automation.Step{Action: automation.ActionKeyUp, Value: "A"}); err != nil {
		t.Fatalf("handleKeyUp: %v", err)
	}

	var kd, ku string
	if err := chromedp.Run(ctx,
		chromedp.AttributeValue("#i", "data-kd", &kd, nil),
		chromedp.AttributeValue("#i", "data-ku", &ku, nil),
	); err != nil {
		t.Fatalf("read attrs: %v", err)
	}
	if kd != "A" || ku != "A" {
		t.Errorf("expected kd=A ku=A, got kd=%q ku=%q", kd, ku)
	}
}

func TestHandleSetViewport(t *testing.T) {
	ctx, cancel := chromedpFixture(t, `<html><body><div id="x"></div></body></html>`)
	defer cancel()

	step := automation.Step{Action: automation.ActionSetViewport, Value: `{"width":640,"height":480}`}
	if err := handleSetViewport(ctx, nil, step); err != nil {
		t.Fatalf("handleSetViewport: %v", err)
	}
	var w, h int64
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`window.innerWidth`, &w),
		chromedp.Evaluate(`window.innerHeight`, &h),
	); err != nil {
		t.Fatalf("read dims: %v", err)
	}
	if w != 640 || h != 480 {
		t.Errorf("expected 640x480, got %dx%d", w, h)
	}
}

func TestHandleWaitForElement_AppearsLater(t *testing.T) {
	ctx, cancel := chromedpFixture(t, `<html><body>
<script>setTimeout(()=>{
  const d=document.createElement('div'); d.id='late'; document.body.appendChild(d);
}, 200)</script>
</body></html>`)
	defer cancel()

	step := automation.Step{Action: automation.ActionWaitForElement, Target: "#late", Value: "2000"}
	if err := handleWaitForElement(ctx, nil, step); err != nil {
		t.Fatalf("handleWaitForElement: %v", err)
	}
	var nodes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes("#late", &nodes, chromedp.AtLeast(1))); err != nil {
		t.Fatalf("post-wait query: %v", err)
	}
	if len(nodes) == 0 {
		t.Error("#late never appeared")
	}
}
