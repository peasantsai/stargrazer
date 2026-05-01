package browser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"stargrazer/internal/automation"
	"stargrazer/internal/logger"
)

// StepHandler executes a single step against the manager's chromedp context.
// Implementations live alongside this file and self-register via init().
type StepHandler func(ctx context.Context, m *Manager, step automation.Step) error

// ErrNoHandler is returned by RunStep when no handler is registered for the step's Action.
var ErrNoHandler = errors.New("no step handler registered")

var (
	handlersMu sync.RWMutex
	handlers   = map[automation.Action]StepHandler{}
)

// RegisterHandler registers a StepHandler for the given Action. Panics if a
// handler is already registered for that action — registration is expected to
// happen exactly once per action, in init().
func RegisterHandler(a automation.Action, h StepHandler) {
	handlersMu.Lock()
	defer handlersMu.Unlock()
	if _, exists := handlers[a]; exists {
		panic(fmt.Sprintf("browser: duplicate handler registration for action %q", a))
	}
	handlers[a] = h
}

// RunStep dispatches a step to its registered handler.
func RunStep(ctx context.Context, m *Manager, step automation.Step) error {
	handlersMu.RLock()
	h, ok := handlers[step.Action]
	handlersMu.RUnlock()
	if !ok {
		return fmt.Errorf("%w: %q", ErrNoHandler, step.Action)
	}
	return h(ctx, m, step)
}

func lookupHandlerForTest(a automation.Action) (StepHandler, bool) {
	handlersMu.RLock()
	defer handlersMu.RUnlock()
	h, ok := handlers[a]
	return h, ok
}

func unregisterHandlerForTest(a automation.Action) {
	handlersMu.Lock()
	defer handlersMu.Unlock()
	delete(handlers, a)
}

func init() {
	RegisterHandler(automation.ActionNavigate, handleNavigate)
	RegisterHandler(automation.ActionClick, handleClick)
	RegisterHandler(automation.ActionType, handleType)
	RegisterHandler(automation.ActionWait, handleWait)
	RegisterHandler(automation.ActionEvaluate, handleEvaluate)
	RegisterHandler(automation.ActionScroll, handleScroll)
	RegisterHandler(automation.ActionDoubleClick, handleDoubleClick)
	RegisterHandler(automation.ActionHover, handleHover)
	RegisterHandler(automation.ActionKeyDown, handleKeyDown)
	RegisterHandler(automation.ActionKeyUp, handleKeyUp)
	RegisterHandler(automation.ActionSetViewport, handleSetViewport)
	RegisterHandler(automation.ActionWaitForElement, handleWaitForElement)
}

func handleNavigate(ctx context.Context, m *Manager, step automation.Step) error {
	return m.ExecNavigate(ctx, step.Target)
}

func handleClick(ctx context.Context, m *Manager, step automation.Step) error {
	return m.ExecClick(ctx, step.Target, step.Selectors)
}

func handleType(ctx context.Context, m *Manager, step automation.Step) error {
	return m.ExecType(ctx, step.Target, step.Value, step.Selectors)
}

func handleWait(_ context.Context, _ *Manager, step automation.Step) error {
	ms := 1000
	if n, err := strconv.Atoi(step.Value); err == nil && n > 0 {
		ms = n
	}
	logger.Info("cdpexec", fmt.Sprintf("wait %dms", ms))
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return nil
}

func handleEvaluate(ctx context.Context, m *Manager, step automation.Step) error {
	return m.ExecEvaluate(ctx, step.Value)
}

func handleScroll(ctx context.Context, m *Manager, step automation.Step) error {
	return m.ExecScroll(ctx, step.Target, step.Selectors)
}

func handleDoubleClick(ctx context.Context, _ *Manager, step automation.Step) error {
	sel := step.Target
	tctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()
	return chromedp.Run(tctx,
		chromedp.WaitVisible(sel, selectorOpt(sel)),
		chromedp.DoubleClick(sel, selectorOpt(sel)),
	)
}

func handleHover(ctx context.Context, _ *Manager, step automation.Step) error {
	// chromedp lacks a first-class Hover; dispatch a mouseover via JS on the resolved element.
	sel := step.Target
	js := fmt.Sprintf(`(()=> {
	  const el = document.querySelector(%q);
	  if (!el) return false;
	  const rect = el.getBoundingClientRect();
	  ['mouseover','mousemove','mouseenter'].forEach(t =>
	    el.dispatchEvent(new MouseEvent(t, {bubbles:true, clientX: rect.left+rect.width/2, clientY: rect.top+rect.height/2}))
	  );
	  return true;
	})()`, sel)
	tctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()
	var ok bool
	if err := chromedp.Run(tctx,
		chromedp.WaitVisible(sel, selectorOpt(sel)),
		chromedp.Evaluate(js, &ok),
	); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("hover: element %q not found at dispatch time", sel)
	}
	return nil
}

func handleKeyDown(ctx context.Context, _ *Manager, step automation.Step) error {
	if step.Value == "" {
		return fmt.Errorf("keyDown: step.Value (the key) is empty")
	}
	tctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()
	return chromedp.Run(tctx, dispatchKey(input.KeyDown, step.Value))
}

func handleKeyUp(ctx context.Context, _ *Manager, step automation.Step) error {
	if step.Value == "" {
		return fmt.Errorf("keyUp: step.Value (the key) is empty")
	}
	tctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()
	return chromedp.Run(tctx, dispatchKey(input.KeyUp, step.Value))
}

// dispatchKey returns a chromedp.Action that dispatches a single keyDown/keyUp
// to the currently-focused element via CDP Input.dispatchKeyEvent.
func dispatchKey(eventType input.KeyType, key string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		return input.DispatchKeyEvent(eventType).WithKey(key).WithCode(key).Do(ctx)
	})
}

func handleSetViewport(ctx context.Context, _ *Manager, step automation.Step) error {
	var dims struct {
		Width  int64 `json:"width"`
		Height int64 `json:"height"`
	}
	if err := json.Unmarshal([]byte(step.Value), &dims); err != nil {
		return fmt.Errorf("setViewport: parse step.Value as {width,height}: %w", err)
	}
	if dims.Width <= 0 || dims.Height <= 0 {
		return fmt.Errorf("setViewport: width=%d height=%d must both be positive", dims.Width, dims.Height)
	}
	tctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()
	return chromedp.Run(tctx, chromedp.EmulateViewport(dims.Width, dims.Height))
}

func handleWaitForElement(ctx context.Context, _ *Manager, step automation.Step) error {
	timeout := stepTimeout
	if n, err := strconv.Atoi(step.Value); err == nil && n > 0 {
		timeout = time.Duration(n) * time.Millisecond
	}
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return chromedp.Run(tctx, chromedp.WaitVisible(step.Target, selectorOpt(step.Target)))
}
