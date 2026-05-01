package browser

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

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
